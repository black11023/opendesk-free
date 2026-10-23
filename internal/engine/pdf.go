package engine
import (
"bytes"
"context"
"errors"
"fmt"
"io"
"os"
"path/filepath"
"strings"
"time"
"github.com/pdfcpu/pdfcpu/pkg/api"
)
func pdfProcess(op string, files []Input) (Result,error) {
 if len(files)<1||len(files)>10{return Result{},errors.New("PDF operations accept 1–10 files")}
 if op=="pdf-merge"&&len(files)<2{return Result{},errors.New("Merge needs at least two PDFs")}
 if op!="pdf-merge"&&len(files)!=1{return Result{},errors.New("Select exactly one PDF")}
 ctx,cancel:=context.WithTimeout(context.Background(),60*time.Second);defer cancel()
 readers:=make([]io.ReadSeeker,0,len(files));totalPages:=0
 for _,f:=range files {
  if !strings.EqualFold(filepath.Ext(f.Name),".pdf")||!bytes.HasPrefix(f.Data,[]byte("%PDF-")){return Result{},fmt.Errorf("not a PDF: %s",SafeName(f.Name))}
  n,e:=api.PageCount(ctx,bytes.NewReader(f.Data),nil)
  if e!=nil{return Result{},fmt.Errorf("invalid PDF %s: %w",SafeName(f.Name),e)}
  if n<1||n>200{return Result{},errors.New("each PDF must have 1–200 pages")}
  totalPages+=n;if totalPages>300{return Result{},errors.New("maximum 300 pages per operation")}
  readers=append(readers,bytes.NewReader(f.Data))
 }
 switch op {
 case "pdf-merge":
  var out bytes.Buffer
  if e:=api.MergeRaw(ctx,readers,&out,false,nil);e!=nil{return Result{},fmt.Errorf("PDF merge: %w",e)}
  n,e:=api.PageCount(ctx,bytes.NewReader(out.Bytes()),nil);if e!=nil||n!=totalPages{return Result{},errors.New("merged PDF page count check failed")}
  return Result{Filename:"OpenDesk_Merged.pdf",ContentType:"application/pdf",Data:out.Bytes()},nil
 case "pdf-rotate":
  var out bytes.Buffer
  if e:=api.Rotate(ctx,readers[0],&out,90,nil,nil);e!=nil{return Result{},fmt.Errorf("PDF rotate: %w",e)}
  n,e:=api.PageCount(ctx,bytes.NewReader(out.Bytes()),nil);if e!=nil||n!=totalPages{return Result{},errors.New("rotated PDF page count check failed")}
  return Result{Filename:"OpenDesk_Rotated.pdf",ContentType:"application/pdf",Data:out.Bytes()},nil
 case "pdf-split":
  dir,e:=os.MkdirTemp("","opendesk-pdf-*");if e!=nil{return Result{},e};defer os.RemoveAll(dir)
  if e:=api.Split(ctx,readers[0],dir,"page",1,nil);e!=nil{return Result{},fmt.Errorf("PDF split: %w",e)}
  entries,e:=os.ReadDir(dir);if e!=nil{return Result{},e}
  out:=map[string][]byte{}
  for _,entry:=range entries {if entry.IsDir()||!strings.EqualFold(filepath.Ext(entry.Name()),".pdf"){continue}
   b,e:=os.ReadFile(filepath.Join(dir,entry.Name()));if e!=nil{return Result{},e}
   n,e:=api.PageCount(ctx,bytes.NewReader(b),nil);if e!=nil||n!=1{return Result{},errors.New("split PDF verification failed")}
   out[SafeName(entry.Name())]=b
  }
  if len(out)!=totalPages{return Result{},errors.New("PDF split page-count mismatch")}
  return zipFiles("OpenDesk_Split_Pages.zip",out)
 }
 return Result{},errors.New("unknown PDF operation")
}
