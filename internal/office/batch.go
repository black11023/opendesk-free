package office

// The separate Office batch workflow is intentionally best-effort: every
// source has a CSV row; good PDFs survive an unrelated invalid document.
// The ordinary Office → PDF operation remains all-or-nothing.
import (
 "archive/zip"
 "bytes"
 "context"
 "crypto/sha256"
 "encoding/csv"
 "errors"
 "fmt"
 "strings"

 "opendesk-free/internal/engine"
)
func BatchWith(ctx context.Context,runner Executor,binary string,files []engine.Input)(engine.Result,error){
 if len(files)<2||len(files)>MaxFiles{return engine.Result{},fmt.Errorf("choose 2–%d documents for Office batch",MaxFiles)}
 if runner==nil||strings.TrimSpace(binary)==""{return engine.Result{},errors.New("LibreOffice is not available")}
 var sourceBytes int64
 for _,f:=range files{sourceBytes+=int64(len(f.Data));if sourceBytes>engine.MaxBytes{return engine.Result{},errors.New("Office batch input exceeds 40 MB")}}
 if e:=ctx.Err();e!=nil{return engine.Result{},e}
 var out bytes.Buffer
 archive:=zip.NewWriter(&out)
 var table bytes.Buffer
 table.Write([]byte{0xEF,0xBB,0xBF})
 csvWriter:=csv.NewWriter(&table)
 if e:=csvWriter.Write([]string{"index","source","status","output","source_bytes","sha256","error"});e!=nil{return engine.Result{},e}
 successes:=0
 for i,source:=range files{
  if e:=ctx.Err();e!=nil{return engine.Result{},e}
  output,e:=ProcessWith(ctx,runner,binary,[]engine.Input{source})
  if err:=ctx.Err();err!=nil{return engine.Result{},err}
  name:=fmt.Sprintf("%02d_%s.pdf",i+1,engine.SafeName(strings.TrimSuffix(source.Name,fileExt(source.Name))))
  status,message,download:="PASS","",""
  if e!=nil {
   status,message="FAIL",e.Error()
  }else{
   // Preserve the exact selection order in numbered ZIP entry names.
   f,err:=archive.Create(name);if err!=nil{return engine.Result{},err}
   if _,err=f.Write(output.Data);err!=nil{return engine.Result{},err}
   successes++
   download=name
  }
  digest:=sha256.Sum256(source.Data)
  row:=[]string{fmt.Sprint(i+1),csvSafe(engine.SafeName(source.Name)),status,download,fmt.Sprint(len(source.Data)),fmt.Sprintf("%x",digest),csvSafe(message)}
  if e:=csvWriter.Write(row);e!=nil{return engine.Result{},e}
  if out.Len()>MaxOutputBytes{return engine.Result{},errors.New("Office batch ZIP exceeds 80 MB")}
 }
 csvWriter.Flush();if e:=csvWriter.Error();e!=nil{return engine.Result{},e}
 entry,e:=archive.Create("OpenDesk_Office_Batch_Report.csv");if e!=nil{return engine.Result{},e}
 if _,e=entry.Write(table.Bytes());e!=nil{return engine.Result{},e}
 if e=archive.Close();e!=nil{return engine.Result{},e}
 if out.Len()>MaxOutputBytes{return engine.Result{},errors.New("Office batch ZIP exceeds 80 MB")}
 if successes==0 { // A report-only ZIP still helps the user diagnose failure.
  return engine.Result{Filename:"OpenDesk_Office_Batch_Results.zip",ContentType:"application/zip",Data:out.Bytes()},nil
 }
 return engine.Result{Filename:"OpenDesk_Office_Batch_Results.zip",ContentType:"application/zip",Data:out.Bytes()},nil
}
func fileExt(n string)string{
 i:=strings.LastIndexByte(n,'.');if i<0{return ""}
 return n[i:]
}
func csvSafe(s string)string{
 first:=strings.TrimLeft(s," \t\r\n")
 if first!=""&&strings.ContainsRune("=+-@",rune(first[0])){return "'"+s}
 return s
}
func RunBatch(ctx context.Context,files []engine.Input)(engine.Result,error){
 executable,e:=FindBinary();if e!=nil{return engine.Result{},e}
 return BatchWith(ctx,LocalExecutor{},executable,files)
}
