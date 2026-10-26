//go:build integration

package integration

import (
 "archive/zip"
 "bytes"
 "context"
 "io"
 "os"
 "os/exec"
 "path/filepath"
 "strings"
 "testing"
 "time"

 "github.com/pdfcpu/pdfcpu/pkg/api"

 "opendesk-free/internal/engine"
 "opendesk-free/internal/ocr"
 "opendesk-free/internal/office"
)
const fixtureDir="../../tests/generated"
func requireExecutable(t *testing.T,name string){
 t.Helper();if _,e:=exec.LookPath(name);e!=nil{t.Fatalf("real integration requires %s: %v",name,e)}
}
func source(t *testing.T,filename string)engine.Input{
 t.Helper();b,e:=os.ReadFile(filepath.Join(fixtureDir,filename));if e!=nil{t.Fatal(e)}
 return engine.Input{Name:filename,Data:b}
}
func pdfText(t *testing.T,pdf []byte)string{
 t.Helper();cmd:=exec.Command("pdftotext","-layout","-","-")
 cmd.Stdin=bytes.NewReader(pdf)
 output,e:=cmd.CombinedOutput()
 if e!=nil{t.Fatalf("pdftotext failed: %v %s",e,output)}
 return strings.ToUpper(string(output))
}
func TestRealLibreOfficeDOCXAndXLSX(t *testing.T){
 requireExecutable(t,"soffice");requireExecutable(t,"pdftotext")
 ctx,cancel:=context.WithTimeout(context.Background(),2*time.Minute);defer cancel()
 one,e:=office.Run(ctx,[]engine.Input{source(t,"real_office.docx")})
 if e!=nil{t.Fatalf("real DOCX to PDF conversion failed: %v",e)}
 if one.ContentType!="application/pdf"||!strings.Contains(pdfText(t,one.Data),"OPENDESK OFFICE TEST 2026"){
  t.Fatalf("DOCX output missing expected original text: %q",pdfText(t,one.Data))
 }
 batch,e:=office.Run(ctx,[]engine.Input{source(t,"real_office.docx"),source(t,"real_sheet.xlsx")})
 if e!=nil{t.Fatalf("real Office batch conversion failed: %v",e)}
 archive,e:=zip.NewReader(bytes.NewReader(batch.Data),int64(len(batch.Data)));if e!=nil{t.Fatal(e)}
 if len(archive.File)!=2{t.Fatalf("expected 2 converted PDFs, got %d",len(archive.File))}
 texts:=[]string{"OPENDESK OFFICE TEST 2026","OPENDESK SHEET 2026"}
 for i,entry:=range archive.File{
  rd,e:=entry.Open();if e!=nil{t.Fatal(e)}
  output,e:=io.ReadAll(rd);_ =rd.Close();if e!=nil{t.Fatal(e)}
  if !strings.Contains(pdfText(t,output),texts[i]){t.Fatalf("%s lost expected content %q: %q",entry.Name,texts[i],pdfText(t,output))}
 }
}
func TestRealTesseractImagesAndOCRmyPDFScan(t *testing.T){
 for _,tool:=range []string{"tesseract","ocrmypdf","pdftotext"}{requireExecutable(t,tool)}
 ctx,cancel:=context.WithTimeout(context.Background(),3*time.Minute);defer cancel()
 image:=source(t,"real_scanned_english.png")
 raw,e:=ocr.Process(ctx,ocr.LocalRunner{Executable:"tesseract"},[]engine.Input{image},"eng","txt")
 if e!=nil{t.Fatalf("real image OCR failed: %v",e)}
 if !strings.Contains(strings.ToUpper(string(raw.Data)),"OPENDESK"){t.Fatalf("Tesseract did not read expected word: %q",string(raw.Data))}
 scan,e:=engine.RunContext(ctx,"images-pdf",[]engine.Input{image})
 if e!=nil{t.Fatalf("raster PDF generation failed: %v",e)}
 result,e:=ocr.RunPDF(ctx,[]engine.Input{{Name:"scanned.pdf",Data:scan.Data}},"eng")
 if e!=nil{t.Fatalf("real OCRmyPDF failed: %v",e)}
 if !strings.Contains(pdfText(t,result.Data),"OPENDESK"){t.Fatalf("searchable PDF lacks OCR text: %q",pdfText(t,result.Data))}
}

func TestRealLibreOfficePowerPointAndBatch(t *testing.T){
 requireExecutable(t,"soffice")
 requireExecutable(t,"pdftotext")
 ctx,cancel:=context.WithTimeout(context.Background(),2*time.Minute)
 defer cancel()
 presentation:=source(t,"real_slides.pptx")
 result,e:=office.Run(ctx,[]engine.Input{presentation})
 if e!=nil{t.Fatalf("real PowerPoint conversion failed: %v",e)}
 if result.ContentType!="application/pdf"{t.Fatalf("unexpected PPTX output %s",result.ContentType)}
 pages,e:=api.PageCount(ctx,bytes.NewReader(result.Data),nil)
 if e!=nil||pages!=2{t.Fatalf("PowerPoint slide/page mismatch: pages=%d err=%v",pages,e)}
 words:=pdfText(t,result.Data)
 for _,heading:=range []string{"OPENDESK SLIDE ONE 2026","OPENDESK SLIDE TWO 2026"}{
  if !strings.Contains(words,heading){t.Fatalf("slide title %q lost in rendered PDF: %q",heading,words)}
 }
 files:=[]engine.Input{source(t,"real_office.docx"),source(t,"real_sheet.xlsx"),presentation}
 batch,e:=office.Run(ctx,files)
 if e!=nil{t.Fatalf("real three-format Office batch failed: %v",e)}
 archive,e:=zip.NewReader(bytes.NewReader(batch.Data),int64(len(batch.Data)))
 if e!=nil||len(archive.File)!=3{t.Fatalf("batch returned %d files, err=%v",len(archive.File),e)}
 for i,f:=range archive.File{
  handle,e:=f.Open();if e!=nil{t.Fatal(e)}
  pdf,e:=io.ReadAll(handle);_ =handle.Close();if e!=nil{t.Fatal(e)}
  pageCount,e:=api.PageCount(ctx,bytes.NewReader(pdf),nil)
  expectedPages:=1;if i==2{expectedPages=2}
  if e!=nil||pageCount!=expectedPages{t.Fatalf("batch file %q: pages=%d expected=%d err=%v",f.Name,pageCount,expectedPages,e)}
 }
}
