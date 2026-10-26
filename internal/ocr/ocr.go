// Package ocr provides an optional, local-only Tesseract command adapter.
// No OCR executable or language data is bundled or downloaded.
package ocr

import (
 "bytes"
 "context"
 "errors"
 "fmt"
 "image"
 _ "image/jpeg"
 _ "image/png"
 "io"
 "os"
 "os/exec"
 "path/filepath"
 "runtime"
 "strings"
 "time"
 "unicode/utf8"

 "github.com/pdfcpu/pdfcpu/pkg/api"
 "opendesk-free/internal/engine"
 "opendesk-free/internal/procrun"
)

const MaxOCRImages=10
const MaxPixels=int64(20_000_000)
const MaxOutputBytes=50<<20

type Status struct {
 Installed bool `json:"installed"`
 English bool `json:"eng"`
 Vietnamese bool `json:"vie"`
 Detail string `json:"detail"`
}
func FindBinary() (string,error){
 if p,e:=exec.LookPath("tesseract");e==nil{return p,nil}
 if runtime.GOOS=="windows" {
  for _,env:=range []string{"ProgramFiles","ProgramFiles(x86)"} {
   root:=os.Getenv(env);if root==""{continue}
   p:=filepath.Join(root,"Tesseract-OCR","tesseract.exe")
   if st,e:=os.Stat(p);e==nil&&!st.IsDir(){return p,nil}
  }
 }
 return "",errors.New("Tesseract OCR is not installed; install it locally and add it to PATH")
}
func Probe(ctx context.Context) Status {
 p,e:=FindBinary()
 if e!=nil{return Status{Detail:"Tesseract not found. See the local OCR setup guide."}}
 c,cancel:=context.WithTimeout(ctx,8*time.Second);defer cancel()
 cmd:=procrun.Command(c,p,"--list-langs")
 out,e:=cmd.CombinedOutput()
 if e!=nil{return Status{Installed:true,Detail:"Tesseract is installed but language data cannot be listed."}}
 langs:=map[string]bool{}
 for _,line:=range strings.Split(strings.ReplaceAll(string(out),"\r",""),"\n"){
  val:=strings.TrimSpace(line)
  if val!=""{langs[val]=true}
 }
 detail:="Tesseract ready"
 if !langs["eng"]&&!langs["vie"]{detail="Install eng.traineddata and/or vie.traineddata"}
 return Status{Installed:true,English:langs["eng"],Vietnamese:langs["vie"],Detail:detail}
}
type Runner interface {Recognize(ctx context.Context, image []byte, language, format string) ([]byte,error)}
type limitWriter struct {buf bytes.Buffer; max int}
func (w *limitWriter)Write(p []byte)(int,error){
 if w.buf.Len()+len(p)>w.max{return 0,errors.New("OCR output exceeds safety size limit")}
 return w.buf.Write(p)
}
type LocalRunner struct {Executable string}
func (r LocalRunner)Recognize(ctx context.Context, raw []byte, language, format string)([]byte,error){
 args:=[]string{"stdin","stdout","-l",language}
 if format=="pdf"{args=append(args,"pdf")}
 cmd:=procrun.Command(ctx,r.Executable,args...)
 cmd.Stdin=bytes.NewReader(raw)
 out:=&limitWriter{max:MaxOutputBytes}
 errOut:=&limitWriter{max:8192}
 cmd.Stdout=out;cmd.Stderr=errOut
 e:=cmd.Run()
 if e!=nil{
  if ctx.Err()!=nil{return nil,fmt.Errorf("OCR timed out or cancelled: %w",ctx.Err())}
  detail:=strings.TrimSpace(errOut.buf.String())
  if len(detail)>500{detail=detail[:500]}
  return nil,fmt.Errorf("Tesseract failed: %v %s",e,detail)
 }
 if out.buf.Len()==0{return nil,errors.New("Tesseract returned empty output")}
 return out.buf.Bytes(),nil
}
func selectedLanguages(language string, st Status)error{
 switch language {
 case "eng":
  if !st.English{return errors.New("English OCR language pack is missing (eng.traineddata)")}
 case "vie":
  if !st.Vietnamese{return errors.New("Vietnamese OCR language pack is missing (vie.traineddata)")}
 case "eng+vie":
  if !st.English||!st.Vietnamese{return errors.New("Both eng.traineddata and vie.traineddata are required")}
 default:return errors.New("Unsupported OCR language selection")
 }
 return nil
}
func validateInput(files []engine.Input)error{
 if len(files)==0||len(files)>MaxOCRImages{return fmt.Errorf("select 1–%d PNG/JPG images for OCR",MaxOCRImages)}
 var total int64
 for _,f:=range files{
  total+=int64(len(f.Data));if total>engine.MaxBytes{return errors.New("OCR input exceeds 40 MB total")}
  ext:=strings.ToLower(filepath.Ext(f.Name))
  if ext!=".png"&&ext!=".jpg"&&ext!=".jpeg"{return fmt.Errorf("only PNG/JPG OCR input supported: %s",engine.SafeName(f.Name))}
  cfg,format,e:=image.DecodeConfig(bytes.NewReader(f.Data))
  if e!=nil{return fmt.Errorf("invalid OCR image %s: %w",engine.SafeName(f.Name),e)}
  if (format=="png"&&ext!=".png")||(format=="jpeg"&&ext!=".jpeg"&&ext!=".jpg"){return fmt.Errorf("image extension/content mismatch: %s",engine.SafeName(f.Name))}
  if format!="png"&&format!="jpeg" {return fmt.Errorf("unsupported image type: %s",engine.SafeName(f.Name))}
  if cfg.Width<1||cfg.Height<1||int64(cfg.Width)*int64(cfg.Height)>MaxPixels{return fmt.Errorf("image exceeds 20 megapixels: %s",engine.SafeName(f.Name))}
 }
 return nil
}
func Process(ctx context.Context, runner Runner, files []engine.Input, language, format string)(engine.Result,error){
 if e:=validateInput(files);e!=nil{return engine.Result{},e}
 if language!="eng"&&language!="vie"&&language!="eng+vie"{return engine.Result{},errors.New("unsupported OCR language")}
 if format!="txt"&&format!="pdf"{return engine.Result{},errors.New("unsupported OCR output format")}
 var text bytes.Buffer
 var pdfInputs []io.ReadSeeker
 for _,f:=range files{
  stepCtx,cancel:=context.WithTimeout(ctx,45*time.Second)
  b,e:=runner.Recognize(stepCtx,f.Data,language,format)
  cancel()
  if e!=nil{return engine.Result{},fmt.Errorf("OCR %s: %w",engine.SafeName(f.Name),e)}
  if len(b)>MaxOutputBytes{return engine.Result{},errors.New("OCR output exceeds 50 MB")}
  if format=="txt"{
   if !utf8.Valid(b){return engine.Result{},errors.New("OCR output is not valid UTF-8")}
   fmt.Fprintf(&text,"===== %s =====\n",engine.SafeName(f.Name))
   text.Write(bytes.TrimSpace(b));text.WriteString("\n\n")
   if text.Len()>MaxOutputBytes{return engine.Result{},errors.New("OCR text output too large")}
  }else{
   if !bytes.HasPrefix(b,[]byte("%PDF-")){return engine.Result{},errors.New("OCR did not produce a PDF")}
   pages,e:=api.PageCount(ctx,bytes.NewReader(b),nil)
   if e!=nil||pages!=1{return engine.Result{},fmt.Errorf("OCR result is not a valid one-page PDF: %v",e)}
   pdfInputs=append(pdfInputs,bytes.NewReader(b))
  }
 }
 if format=="txt"{
  return engine.Result{Filename:"OpenDesk_OCR_Text.txt",ContentType:"text/plain; charset=utf-8",Data:text.Bytes()},nil
 }
 if len(pdfInputs)==1{
  data:=make([]byte,0)
  // A ReadSeeker can be read without exposing input paths to subprocess arguments.
  if _,e:=pdfInputs[0].Seek(0,io.SeekStart);e!=nil{return engine.Result{},e}
  b,e:=io.ReadAll(io.LimitReader(pdfInputs[0],MaxOutputBytes+1))
  if e!=nil||len(b)>MaxOutputBytes{return engine.Result{},errors.New("OCR PDF exceeds output size limit")}
  data=append(data,b...)
  return engine.Result{Filename:"OpenDesk_OCR_Searchable.pdf",ContentType:"application/pdf",Data:data},nil
 }
 var out bytes.Buffer
 if e:=api.MergeRaw(ctx,pdfInputs,&out,false,nil);e!=nil{return engine.Result{},fmt.Errorf("merge OCR PDF: %w",e)}
 pages,e:=api.PageCount(ctx,bytes.NewReader(out.Bytes()),nil)
 if e!=nil||pages!=len(files){return engine.Result{},errors.New("OCR merged PDF page-count mismatch")}
 return engine.Result{Filename:"OpenDesk_OCR_Searchable.pdf",ContentType:"application/pdf",Data:out.Bytes()},nil
}
func RunImages(ctx context.Context,files []engine.Input,language,format string)(engine.Result,error){
 p,e:=FindBinary();if e!=nil{return engine.Result{},e}
 status:=Probe(ctx);if !status.Installed{return engine.Result{},errors.New(status.Detail)}
 if e=selectedLanguages(language,status);e!=nil{return engine.Result{},e}
 return Process(ctx,LocalRunner{Executable:p},files,language,format)
}
