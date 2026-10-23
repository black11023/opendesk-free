package engine

import (
  "archive/zip"
  "bytes"
  "context"
  "io"
  "os"
  "strings"
  "testing"
  "github.com/pdfcpu/pdfcpu/pkg/api"
)
func twoPageFixture(t *testing.T) Input {
 t.Helper()
 data,err:=os.ReadFile("../../tests/fixtures/pdf_two_pages.pdf")
 if err!=nil {t.Fatal(err)}
 return Input{Name:"fixture.pdf",Data:data}
}
func checkPages(t *testing.T,data []byte,want int) {
 t.Helper()
 got,err:=api.PageCount(context.Background(),bytes.NewReader(data),nil)
 if err!=nil||got!=want {t.Fatalf("expected %d pages, got %d (%v)",want,got,err)}
}
func TestMergePDFPages(t *testing.T) {
 original:=twoPageFixture(t)
 out,err:=Run("pdf-merge",[]Input{original,original})
 if err!=nil{t.Fatal(err)}
 if out.ContentType!="application/pdf"{t.Fatalf("unexpected type %s",out.ContentType)}
 checkPages(t,out.Data,4)
}
func TestRotatePDFPages(t *testing.T){
 out,err:=Run("pdf-rotate",[]Input{twoPageFixture(t)})
 if err!=nil{t.Fatal(err)}
 checkPages(t,out.Data,2)
}
func TestSplitPDFIntoPages(t *testing.T){
 out,err:=Run("pdf-split",[]Input{twoPageFixture(t)})
 if err!=nil{t.Fatal(err)}
 z,err:=zip.NewReader(bytes.NewReader(out.Data),int64(len(out.Data)))
 if err!=nil{t.Fatal(err)}
 if len(z.File)!=2 {t.Fatalf("expected 2 files, got %d",len(z.File))}
 for _,f:=range z.File {
  if !strings.EqualFold(".pdf", f.Name[len(f.Name)-4:]){t.Fatalf("unexpected name %q",f.Name)}
  rc,err:=f.Open();if err!=nil{t.Fatal(err)}
  b,err:=io.ReadAll(rc);rc.Close();if err!=nil{t.Fatal(err)}
  checkPages(t,b,1)
 }
}
func TestPDFRejectsWrongFilesAndCounts(t *testing.T){
 valid:=twoPageFixture(t)
 cases:=[]struct{op string;files []Input}{
  {"pdf-merge",[]Input{valid}},
  {"pdf-rotate",[]Input{valid,valid}},
  {"pdf-split",[]Input{{Name:"not-pdf.txt",Data:valid.Data}}},
  {"pdf-rotate",[]Input{{Name:"fake.pdf",Data:[]byte("%PDF-corrupt")}}},
 }
 for _,c:=range cases{if _,err:=Run(c.op,c.files);err==nil{t.Errorf("accepted invalid %s",c.op)}}
}
