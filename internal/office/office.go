package office
import(
 "archive/zip"
 "bytes"
 "context"
 "errors"
 "fmt"
 "io"
 "net/url"
 "os"
 "os/exec"
 "path/filepath"
 "runtime"
 "strings"
 "time"
 "github.com/pdfcpu/pdfcpu/pkg/api"
 "opendesk-free/internal/engine"
 "opendesk-free/internal/procrun"
)
const MaxFiles=8
const MaxOutputBytes=80<<20
type Status struct{Installed bool `json:"installed"`; Detail string `json:"detail"`}
func FindBinary()(string,error){
 for _,name:=range []string{"soffice","libreoffice"}{if p,e:=exec.LookPath(name);e==nil{return p,nil}}
 if runtime.GOOS=="windows"{for _,key:=range []string{"ProgramFiles","ProgramFiles(x86)"}{
  if root:=os.Getenv(key);root!=""{p:=filepath.Join(root,"LibreOffice","program","soffice.exe");if st,e:=os.Stat(p);e==nil&&st.Mode().IsRegular(){return p,nil}}
 }}
 if runtime.GOOS=="darwin"{p:="/Applications/LibreOffice.app/Contents/MacOS/soffice";if st,e:=os.Stat(p);e==nil&&st.Mode().IsRegular(){return p,nil}}
 return "",errors.New("LibreOffice not installed; see docs/OFFICE_LOCAL_SETUP.md")
}
func Probe(ctx context.Context)Status{
 p,e:=FindBinary();if e!=nil{return Status{Detail:e.Error()}}
 c,cancel:=context.WithTimeout(ctx,8*time.Second);defer cancel()
 out,e:=procrun.Command(c,p,"--version").CombinedOutput()
 if e!=nil{return Status{Detail:"LibreOffice found but unable to start"}}
 ver:=strings.TrimSpace(string(out));if len(ver)>80{ver=ver[:80]}
 return Status{Installed:true,Detail:ver}
}
type Executor interface{Run(context.Context,string,[]string)error}
type LocalExecutor struct{}
type boundedWriter struct{b bytes.Buffer;max int}
func(w *boundedWriter)Write(p []byte)(int,error){if w.b.Len()+len(p)>w.max{return 0,errors.New("tool log too large")};return w.b.Write(p)}
func(LocalExecutor)Run(ctx context.Context,binary string,args []string)error{
 cmd:=procrun.Command(ctx,binary,args...)
 out,errout:=&boundedWriter{max:8192},&boundedWriter{max:8192}
 cmd.Stdout=out;cmd.Stderr=errout
 e:=cmd.Run()
 if e!=nil{
  if ctx.Err()!=nil{return fmt.Errorf("conversion timed out: %w",ctx.Err())}
  detail:=strings.TrimSpace(errout.b.String());if len(detail)>350{detail=detail[:350]}
  return fmt.Errorf("LibreOffice failed: %v %s",e,detail)
 }
 return nil
}
func validate(files []engine.Input)error{
 if len(files)<1||len(files)>MaxFiles{return fmt.Errorf("select 1–%d Office files",MaxFiles)}
 var total int64
 for _,f:=range files{
  ext:=strings.ToLower(filepath.Ext(f.Name))
  modern:=ext==".docx"||ext==".xlsx"||ext==".pptx"||ext==".odt"||ext==".ods"||ext==".odp"
  legacy:=ext==".doc"||ext==".xls"||ext==".ppt"
  if !modern&&!legacy{return fmt.Errorf("unsupported Office file: %s",engine.SafeName(f.Name))}
  if len(f.Data)<8{return fmt.Errorf("invalid Office file: %s",engine.SafeName(f.Name))}
  if modern&&!bytes.HasPrefix(f.Data,[]byte{'P','K',3,4}){return fmt.Errorf("not a ZIP-based Office file: %s",engine.SafeName(f.Name))}
  if legacy&&!bytes.HasPrefix(f.Data,[]byte{0xd0,0xcf,0x11,0xe0,0xa1,0xb1,0x1a,0xe1}){return fmt.Errorf("not a legacy Office file: %s",engine.SafeName(f.Name))}
  total+=int64(len(f.Data));if total>engine.MaxBytes{return errors.New("Office input exceeds 40 MB")}
 }
 return nil
}
func fileURI(p string)string{
 s:=filepath.ToSlash(p);if filepath.VolumeName(p)!=""{s="/"+s}
 return (&url.URL{Scheme:"file",Path:s}).String()
}
func ProcessWith(ctx context.Context,runner Executor,binary string,files []engine.Input)(engine.Result,error){
 if e:=validate(files);e!=nil{return engine.Result{},e}
 if runner==nil||strings.TrimSpace(binary)==""{return engine.Result{},errors.New("LibreOffice not available")}
 if e:=ctx.Err();e!=nil{return engine.Result{},e}
 dir,e:=os.MkdirTemp("","opendesk-office-*");if e!=nil{return engine.Result{},e};defer os.RemoveAll(dir)
 outDir:=filepath.Join(dir,"output");if e=os.MkdirAll(outDir,0700);e!=nil{return engine.Result{},e}
 profile:=fileURI(filepath.Join(dir,"profile"))
 outputs:=make([][]byte,0,len(files));names:=make([]string,0,len(files))
 totalBytes,totalPages:=0,0
 for i,f:=range files{
  if e=ctx.Err();e!=nil{return engine.Result{},e}
  base:=fmt.Sprintf("work_%02d",i+1)
  in:=filepath.Join(dir,base+strings.ToLower(filepath.Ext(f.Name)))
  if e=os.WriteFile(in,f.Data,0600);e!=nil{return engine.Result{},e}
  argv:=[]string{"-env:UserInstallation="+profile,"--headless","--nologo","--nodefault","--norestore","--nofirststartwizard","--convert-to","pdf","--outdir",outDir,in}
  step,cancel:=context.WithTimeout(ctx,50*time.Second);e=runner.Run(step,binary,argv);cancel()
  if e!=nil{return engine.Result{},fmt.Errorf("convert %s: %w",engine.SafeName(f.Name),e)}
  path:=filepath.Join(outDir,base+".pdf")
  stat,e:=os.Stat(path);if e!=nil{return engine.Result{},fmt.Errorf("LibreOffice did not produce PDF for %s",engine.SafeName(f.Name))}
  if !stat.Mode().IsRegular()||stat.Size()<8||stat.Size()>MaxOutputBytes{return engine.Result{},errors.New("invalid Office PDF size")}
  fd,e:=os.Open(path);if e!=nil{return engine.Result{},e}
  b,re:=io.ReadAll(io.LimitReader(fd,MaxOutputBytes+1));_ =fd.Close()
  if re!=nil||len(b)>MaxOutputBytes{return engine.Result{},errors.New("Office PDF too large")}
  if !bytes.HasPrefix(b,[]byte("%PDF-")){return engine.Result{},errors.New("invalid converted PDF")}
  count,e:=api.PageCount(ctx,bytes.NewReader(b),nil)
  if e!=nil||count<1||count>250{return engine.Result{},errors.New("invalid Office PDF page count")}
  totalPages+=count;totalBytes+=len(b)
  if totalPages>500||totalBytes>MaxOutputBytes{return engine.Result{},errors.New("Office output exceeds 500 pages or 80 MB")}
  stem:=engine.SafeName(strings.TrimSuffix(f.Name,filepath.Ext(f.Name)))
  names=append(names,fmt.Sprintf("%02d_%s.pdf",i+1,stem));outputs=append(outputs,b)
 }
 if len(outputs)==1{return engine.Result{Filename:names[0],ContentType:"application/pdf",Data:outputs[0]},nil}
 var buf bytes.Buffer;w:=zip.NewWriter(&buf)
 for i,b:=range outputs{f,e:=w.Create(names[i]);if e!=nil{return engine.Result{},e};if _,e=f.Write(b);e!=nil{return engine.Result{},e}}
 if e=w.Close();e!=nil{return engine.Result{},e}
 if buf.Len()>MaxOutputBytes{return engine.Result{},errors.New("ZIP output exceeds 80 MB")}
 return engine.Result{Filename:"OpenDesk_Office_PDFs.zip",ContentType:"application/zip",Data:buf.Bytes()},nil
}
func Run(ctx context.Context,files []engine.Input)(engine.Result,error){b,e:=FindBinary();if e!=nil{return engine.Result{},e};return ProcessWith(ctx,LocalExecutor{},b,files)}
