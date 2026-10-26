// Package ocr offers opt-in local OCRmyPDF support for scanned PDFs.
// The OCRmyPDF executable is never bundled or downloaded by OpenDesk.
package ocr

import (
 "bytes"
 "context"
 "errors"
 "fmt"
 "io"
 "os"
 "os/exec"
 "path/filepath"
 "strings"
 "time"

 "github.com/pdfcpu/pdfcpu/pkg/api"
 "opendesk-free/internal/engine"
 "opendesk-free/internal/procrun"
)

const PDFMaxPages = 100
const PDFMaxOutputBytes = 80 << 20

type PDFStatus struct {
 Installed bool `json:"installed"`
 Detail string `json:"detail"`
}
func PDFBinary() (string,error) {
 p,e:=exec.LookPath("ocrmypdf")
 if e!=nil{return "",errors.New("OCRmyPDF not found on PATH; install it separately and restart OpenDesk")}
 return p,nil
}
func PDFProbe(ctx context.Context) PDFStatus {
 bin,e:=PDFBinary()
 if e!=nil{return PDFStatus{Detail:"OCRmyPDF is not installed. See docs/OCR_LOCAL_SETUP.md."}}
 c,cancel:=context.WithTimeout(ctx,8*time.Second);defer cancel()
 cmd:=procrun.Command(c,bin,"--version")
 b,e:=cmd.Output()
 if e!=nil{return PDFStatus{Detail:"OCRmyPDF detected but the command cannot start; check its dependencies."}}
 version:=strings.TrimSpace(string(b))
 if len(version)>60{version=version[:60]}
 return PDFStatus{Installed:true,Detail:"OCRmyPDF "+version+" detected; processing still requires Tesseract and other upstream dependencies."}
}
type PDFExecutor interface { Execute(context.Context,string,[]string) error }
type CommandPDFExecutor struct{}
func (CommandPDFExecutor) Execute(ctx context.Context, executable string, args []string) error {
 cmd:=procrun.Command(ctx,executable,args...) // fixed argv, no shell or interpolation
 errbuf:=&limitWriter{max:8192}
 cmd.Stdout=io.Discard
 cmd.Stderr=errbuf
 err:=cmd.Run()
 if err!=nil{
  if ctx.Err()!=nil{return fmt.Errorf("PDF OCR timed out or was cancelled: %w",ctx.Err())}
  detail:=strings.TrimSpace(errbuf.buf.String())
  if len(detail)>500{detail=detail[:500]}
  return fmt.Errorf("OCRmyPDF failed (%v): %s",err,detail)
 }
 return nil
}
// ProcessPDFWith runs a provided adapter. This seam is used by CI tests
// without installing external binaries and does not claim OCR accuracy.
func ProcessPDFWith(ctx context.Context, executor PDFExecutor, executable string, file engine.Input, language string) (engine.Result,error){
 if language!="eng"&&language!="vie"&&language!="eng+vie"{return engine.Result{},errors.New("unsupported OCR language")}
 if strings.ToLower(filepath.Ext(file.Name))!=".pdf"||!bytes.HasPrefix(file.Data,[]byte("%PDF-")){
  return engine.Result{},errors.New("select a valid PDF file")
 }
 if len(file.Data)==0||len(file.Data)>engine.MaxBytes{return engine.Result{},errors.New("PDF input exceeds 40 MB limit")}
 if executor==nil||strings.TrimSpace(executable)==""{return engine.Result{},errors.New("OCRmyPDF runner unavailable")}
 count,e:=api.PageCount(ctx,bytes.NewReader(file.Data),nil)
 if e!=nil{return engine.Result{},fmt.Errorf("PDF validation failed: %w",e)}
 if count<1||count>PDFMaxPages{return engine.Result{},fmt.Errorf("OCR supports 1–%d PDF pages per operation (got %d)",PDFMaxPages,count)}
 if e=ctx.Err();e!=nil{return engine.Result{},e}
 dir,e:=os.MkdirTemp("","opendesk-ocr-pdf-*")
 if e!=nil{return engine.Result{},e}
 defer os.RemoveAll(dir)
 src:=filepath.Join(dir,"input.pdf");out:=filepath.Join(dir,"output.pdf")
 if e=os.WriteFile(src,file.Data,0600);e!=nil{return engine.Result{},e}
 args:=[]string{"--skip-text","--output-type","pdf","--jobs","1","--optimize","0","--tesseract-timeout","25","-l",language,src,out}
 if e=executor.Execute(ctx,executable,args);e!=nil{return engine.Result{},e}
 if e=ctx.Err();e!=nil{return engine.Result{},e}
 stat,e:=os.Stat(out)
 if e!=nil{return engine.Result{},errors.New("OCRmyPDF returned no output PDF")}
 if !stat.Mode().IsRegular()||stat.Size()<8||stat.Size()>PDFMaxOutputBytes{return engine.Result{},errors.New("OCR PDF output size is invalid or exceeds 80 MB")}
 fd,e:=os.Open(out);if e!=nil{return engine.Result{},e};defer fd.Close()
 data,e:=io.ReadAll(io.LimitReader(fd,PDFMaxOutputBytes+1))
 if e!=nil||len(data)>PDFMaxOutputBytes{return engine.Result{},errors.New("failed to read or oversized OCR PDF")}
 if !bytes.HasPrefix(data,[]byte("%PDF-")){return engine.Result{},errors.New("OCR output is not PDF")}
 outputPages,e:=api.PageCount(ctx,bytes.NewReader(data),nil)
 if e!=nil||outputPages!=count{return engine.Result{},fmt.Errorf("OCR PDF page count mismatch: input %d, output %d",count,outputPages)}
 return engine.Result{Filename:"OpenDesk_OCR_PDF_Searchable.pdf",ContentType:"application/pdf",Data:data},nil
}
func RunPDF(ctx context.Context,files []engine.Input,language string)(engine.Result,error){
 if len(files)!=1{return engine.Result{},errors.New("select exactly one scanned PDF for OCR")}
 binary,e:=PDFBinary();if e!=nil{return engine.Result{},e}
 status:=Probe(ctx)
 if e=selectedLanguages(language,status);e!=nil{return engine.Result{},e}
 runCtx,cancel:=context.WithTimeout(ctx,4*time.Minute);defer cancel()
 return ProcessPDFWith(runCtx,CommandPDFExecutor{},binary,files[0],language)
}
