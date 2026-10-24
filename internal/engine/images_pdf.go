package engine

import (
 "bytes"
 "context"
 "errors"
 "fmt"
 "image"
 _ "image/jpeg"
 _ "image/png"
 "io"
 "path/filepath"
 "strings"
 "time"
 "github.com/pdfcpu/pdfcpu/pkg/api"
)

// imagesPDF builds one PDF page per image using the pre-existing pdfcpu dependency.
// Each image is decoded for dimensions before passing to pdfcpu. No source file is
// changed, no executable is downloaded, and no external API is contacted.
func imagesPDF(parent context.Context,files []Input) (Result, error) {
 if len(files)==0||len(files)>25 {return Result{},errors.New("select 1–25 PNG/JPG images")}
 images:=make([]io.Reader,0,len(files))
 for _,f:=range files{
  ext:=strings.ToLower(filepath.Ext(f.Name))
  if ext!=".png"&&ext!=".jpg"&&ext!=".jpeg"{return Result{},fmt.Errorf("only PNG/JPG is supported: %s",SafeName(f.Name))}
  cfg,format,e:=image.DecodeConfig(bytes.NewReader(f.Data))
  if e!=nil{return Result{},fmt.Errorf("invalid image %s: %w",SafeName(f.Name),e)}
  if (format=="png"&&ext!=".png")||(format=="jpeg"&&ext!=".jpg"&&ext!=".jpeg"){return Result{},fmt.Errorf("image extension does not match content: %s",SafeName(f.Name))}
  if format!="png"&&format!="jpeg"{return Result{},fmt.Errorf("unsupported image data: %s",SafeName(f.Name))}
  if cfg.Width<1||cfg.Height<1||int64(cfg.Width)*int64(cfg.Height)>20_000_000{return Result{},fmt.Errorf("image dimensions outside 20 MP limit: %s",SafeName(f.Name))}
  images=append(images,bytes.NewReader(f.Data))
 }
 ctx,cancel:=context.WithTimeout(parent,90*time.Second)
 defer cancel()
 var out bytes.Buffer
 if e:=api.ImportImages(ctx,nil,&out,images,api.DefaultImportConfig(),nil);e!=nil{return Result{},fmt.Errorf("image to PDF: %w",e)}
 if e:=verifyPDF(ctx,out.Bytes(),len(files));e!=nil{return Result{},e}
 return Result{Filename:"OpenDesk_Images.pdf",ContentType:"application/pdf",Data:out.Bytes()},nil
}
