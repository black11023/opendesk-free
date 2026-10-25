package engine

import (
 "bytes"
 "context"
 "errors"
 "fmt"
 "io"
 "strconv"
 "strings"
 "time"

 "github.com/pdfcpu/pdfcpu/pkg/api"
)

const pdfMaxInputPages = 5000
const pdfMaxOutputPages = 300

// ParsePageSelection validates 1-based page numbers, preserving the requested
// sequence (and repeats) for extraction/merge. Empty means all pages.
// Allowed: "2", "1,5,9", "10-15", "7,2,5" or descending "6-4".
func ParsePageSelection(spec string, pageCount int) ([]int, error) {
 if pageCount<1 || pageCount>pdfMaxInputPages{return nil,fmt.Errorf("PDF must have 1–%d pages",pdfMaxInputPages)}
 spec=strings.TrimSpace(spec)
 if len(spec)>600{return nil,errors.New("page selection is too long")}
 if spec=="" {
  if pageCount>pdfMaxOutputPages{return nil,fmt.Errorf("PDF has %d pages: select up to %d output pages",pageCount,pdfMaxOutputPages)}
  out:=make([]int,pageCount);for i:=range out{out[i]=i+1};return out,nil
 }
 var pages []int
 parts:=strings.Split(spec,",")
 for _,part:=range parts {
  part=strings.TrimSpace(part)
  if part==""{return nil,errors.New("empty page number; use e.g. 2,5,10-12")}
  pair:=strings.Split(part,"-")
  if len(pair)>2{return nil,fmt.Errorf("invalid page range %q",part)}
  get:=func(s string)(int,error){
   s=strings.TrimSpace(s)
   if s==""{return 0,errors.New("missing page number")}
   for _,r:=range s{if r<'0'||r>'9'{return 0,fmt.Errorf("invalid page number %q",s)}}
   n,e:=strconv.Atoi(s);if e!=nil{return 0,e}
   if n<1||n>pageCount{return 0,fmt.Errorf("page %d outside document range 1–%d",n,pageCount)}
   return n,nil
  }
  a,e:=get(pair[0]);if e!=nil{return nil,e}
  b:=a
  if len(pair)==2{b,e=get(pair[1]);if e!=nil{return nil,e}}
  step:=1;if b<a{step=-1}
  for n:=a;;n+=step{
   pages=append(pages,n)
   if len(pages)>pdfMaxOutputPages{return nil,fmt.Errorf("select no more than %d output pages",pdfMaxOutputPages)}
   if n==b{break}
  }
 }
 return pages,nil
}

func pageSpec(nums []int) []string {
 parts:=make([]string,len(nums))
 for i,n:=range nums{parts[i]=strconv.Itoa(n)}
 return parts
}
func verifyPDF(ctx context.Context, data []byte, want int) error {
 got,e:=api.PageCount(ctx,bytes.NewReader(data),nil)
 if e!=nil{return fmt.Errorf("cannot verify result PDF: %w",e)}
 if got!=want{return fmt.Errorf("output pages %d != expected %d",got,want)}
 return nil
}
func pdfProcess(parent context.Context,op string, files []Input, selection []string) (Result,error){
 if len(files)<1||len(files)>10{return Result{},errors.New("PDF operations accept 1–10 files")}
 if op=="pdf-rotate"&&len(files)>1{return batchRotatePDF(parent,files)}
 if op=="pdf-merge"&&len(files)<2{return Result{},errors.New("merge requires two or more PDFs")}
 if op!="pdf-merge"&&len(files)!=1{return Result{},errors.New("select exactly one PDF")}
 if len(selection)!=0&&len(selection)!=len(files){return Result{},errors.New("page selections must match the number of files")}
 ctx,cancel:=context.WithTimeout(parent,90*time.Second);defer cancel()
 sets:=make([][]int,len(files))
 total:=0
 for i,f:=range files{
  if !strings.EqualFold(fileExt(f.Name),".pdf")||!bytes.HasPrefix(f.Data,[]byte("%PDF-")){
   return Result{},fmt.Errorf("not a PDF: %s",SafeName(f.Name))
  }
  count,e:=api.PageCount(ctx,bytes.NewReader(f.Data),nil)
  if e!=nil{return Result{},fmt.Errorf("invalid PDF %s: %w",SafeName(f.Name),e)}
  spec:=""
  if len(selection)!=0{spec=selection[i]}
  if op=="pdf-rotate"&&spec!=""{return Result{},errors.New("rotate currently operates on the entire PDF")}
  if op=="pdf-rotate"{
   if count>pdfMaxOutputPages{return Result{},fmt.Errorf("rotation currently supports PDFs up to %d pages",pdfMaxOutputPages)}
   sets[i]=make([]int,count)
   for j:=range sets[i]{sets[i][j]=j+1}
  }else{
   nums,e:=ParsePageSelection(spec,count);if e!=nil{return Result{},fmt.Errorf("%s: %w",SafeName(f.Name),e)}
   sets[i]=nums
  }
  total+=len(sets[i])
  if total>pdfMaxOutputPages{return Result{},fmt.Errorf("select up to %d total output pages",pdfMaxOutputPages)}
 }
 switch op{
 case "pdf-merge":
  readers:=make([]io.ReadSeeker,0,len(files))
  for i,f:=range files {
   // Collect honors exact selected sequence, including reverse order.
   var picked bytes.Buffer
   if e:=api.Collect(ctx,bytes.NewReader(f.Data),&picked,pageSpec(sets[i]),nil);e!=nil{
    return Result{},fmt.Errorf("select pages from %s: %w",SafeName(f.Name),e)
   }
   if e:=verifyPDF(ctx,picked.Bytes(),len(sets[i]));e!=nil{return Result{},e}
   readers=append(readers,bytes.NewReader(picked.Bytes()))
  }
  var out bytes.Buffer
  if e:=api.MergeRaw(ctx,readers,&out,false,nil);e!=nil{return Result{},fmt.Errorf("merge: %w",e)}
  if e:=verifyPDF(ctx,out.Bytes(),total);e!=nil{return Result{},e}
  return Result{Filename:"OpenDesk_Merged_Selected.pdf",ContentType:"application/pdf",Data:out.Bytes()},nil
 case "pdf-extract":
  var out bytes.Buffer
  if e:=api.Collect(ctx,bytes.NewReader(files[0].Data),&out,pageSpec(sets[0]),nil);e!=nil{
   return Result{},fmt.Errorf("extract: %w",e)
  }
  if e:=verifyPDF(ctx,out.Bytes(),len(sets[0]));e!=nil{return Result{},e}
  return Result{Filename:"OpenDesk_Selected_Pages.pdf",ContentType:"application/pdf",Data:out.Bytes()},nil
 case "pdf-split":
  if len(sets[0])>100{return Result{},errors.New("separate PDF ZIP supports up to 100 selected pages")}
  seen:=map[int]bool{}
  for _,n:=range sets[0]{
   if seen[n]{return Result{},fmt.Errorf("page %d selected twice: duplicates are not supported for separate ZIP",n)}
   seen[n]=true
  }
  out:=map[string][]byte{}
  digest:=func(r io.Reader,n int) error{
   b,e:=io.ReadAll(io.LimitReader(r,MaxBytes+1));if e!=nil{return e}
   if len(b)>MaxBytes{return errors.New("extracted page exceeds size limit")}
   if e=verifyPDF(ctx,b,1);e!=nil{return e}
   out[fmt.Sprintf("OpenDesk_Page_%04d.pdf",n)]=b
   return nil
  }
  if e:=api.ExtractPages(ctx,bytes.NewReader(files[0].Data),pageSpec(sets[0]),digest,nil);e!=nil{
   return Result{},fmt.Errorf("split: %w",e)
  }
  if len(out)!=len(sets[0]){return Result{},errors.New("split page count verification failed")}
  return zipFiles("OpenDesk_Selected_Pages.zip",out)
 case "pdf-rotate":
  var out bytes.Buffer
  if e:=api.Rotate(ctx,bytes.NewReader(files[0].Data),&out,90,nil,nil);e!=nil{return Result{},fmt.Errorf("rotate: %w",e)}
  if e:=verifyPDF(ctx,out.Bytes(),total);e!=nil{return Result{},e}
  return Result{Filename:"OpenDesk_Rotated.pdf",ContentType:"application/pdf",Data:out.Bytes()},nil
 }
 return Result{},errors.New("unknown PDF operation")
}
// Batch rotate produces a ZIP with numbered entries and verified page counts.
func batchRotatePDF(ctx context.Context,files []Input)(Result,error){
 if len(files)<2||len(files)>10{return Result{},errors.New("select 2–10 PDFs for batch rotation")}
 outputs:=make(map[string][]byte,len(files))
 totalPages,totalBytes:=0,0
 for i,f:=range files{
  if e:=ctx.Err();e!=nil{return Result{},e}
  if !strings.EqualFold(fileExt(f.Name),".pdf")||!bytes.HasPrefix(f.Data,[]byte("%PDF-")){return Result{},fmt.Errorf("invalid PDF: %s",SafeName(f.Name))}
  n,e:=api.PageCount(ctx,bytes.NewReader(f.Data),nil)
  if e!=nil||n<1{return Result{},fmt.Errorf("cannot read PDF: %s",SafeName(f.Name))}
  totalPages+=n
  if totalPages>300{return Result{},errors.New("batch rotate supports at most 300 pages")}
  rotated,e:=pdfProcess(ctx,"pdf-rotate",[]Input{f},nil)
  if e!=nil{return Result{},fmt.Errorf("rotate %s: %w",SafeName(f.Name),e)}
  totalBytes+=len(rotated.Data)
  if totalBytes>80<<20{return Result{},errors.New("batch PDF results exceed 80 MB")}
  label:=strings.TrimSuffix(SafeName(f.Name),fileExt(f.Name))
  outputs[fmt.Sprintf("%02d_%s_rotated.pdf",i+1,label)]=rotated.Data
 }
 output,e:=zipFiles("OpenDesk_Rotated_PDFs.zip",outputs)
 if e!=nil{return Result{},e}
 if len(output.Data)>80<<20{return Result{},errors.New("batch PDF ZIP exceeds 80 MB")}
 return output,nil
}
func fileExt(name string)string{p:=strings.LastIndex(name,".");if p<0{return ""};return name[p:]}
