package engine

import (
 "bytes"
 "context"
 "crypto/sha256"
 "encoding/csv"
 "fmt"
 "strings"
 "time"

 "github.com/pdfcpu/pdfcpu/pkg/api"
)
// PDF batch audit is strictly read-only. Each source produces one CSV row,
// including malformed PDFs, instead of discarding the entire batch.
func pdfAudit(parent context.Context, files []Input)(Result,error){
 ctx,cancel:=context.WithTimeout(parent,90*time.Second);defer cancel()
 var report bytes.Buffer
 report.Write([]byte{0xEF,0xBB,0xBF}) // Excel-friendly UTF-8 BOM
 w:=csv.NewWriter(&report)
 if e:=w.Write([]string{"index","filename","status","page_count","input_bytes","sha256","error"});e!=nil{return Result{},e}
 for i,f:=range files{
  if e:=ctx.Err();e!=nil{return Result{},e}
  name:=SafeName(f.Name)
  status,reason,pageCount:="PASS","",0
  switch{
  case !strings.EqualFold(fileExt(f.Name),".pdf"):
   status,reason="FAIL","File extension is not PDF"
  case len(f.Data)<8||!bytes.HasPrefix(f.Data,[]byte("%PDF-")):
   status,reason="FAIL","Invalid PDF header or empty file"
  default:
   n,e:=api.PageCount(ctx,bytes.NewReader(f.Data),nil)
   if e!=nil||n<1{status,reason="FAIL","Invalid PDF structure"}else{pageCount=n}
  }
  digest:=sha256.Sum256(f.Data)
  columns:=[]string{fmt.Sprint(i+1),safeCSVValue(name),status,fmt.Sprint(pageCount),fmt.Sprint(len(f.Data)),fmt.Sprintf("%x",digest),reason}
  if e:=w.Write(columns);e!=nil{return Result{},e}
 }
 w.Flush()
 if e:=w.Error();e!=nil{return Result{},e}
 return Result{Filename:"OpenDesk_PDF_Batch_Audit.csv",ContentType:"text/csv; charset=utf-8",Data:report.Bytes()},nil
}
// CSV quoting alone does NOT prevent formula execution in Excel.
func safeCSVValue(s string)string{
 head:=strings.TrimLeft(s," \t\r\n")
 if head!=""&&strings.ContainsRune("=+-@",rune(head[0])){return "'"+s}
 return s
}
