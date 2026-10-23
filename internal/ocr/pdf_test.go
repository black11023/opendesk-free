package ocr

import (
 "bytes"
 "context"
 "errors"
 "os"
 "path/filepath"
 "strings"
 "testing"

 "github.com/pdfcpu/pdfcpu/pkg/api"
 "opendesk-free/internal/engine"
)

type fakePDFExec struct {calls int; args []string; err error; omit bool; bad bool; wrongPages bool; tempDir string}
func (m *fakePDFExec) Execute(ctx context.Context,bin string,args []string) error {
 m.calls++;m.args=append([]string(nil),args...)
 if m.err!=nil{return m.err}
 if len(args)<3{return errors.New("bad arguments")}
 in,out:=args[len(args)-2],args[len(args)-1]
 m.tempDir=filepath.Dir(in)
 b,e:=os.ReadFile(in);if e!=nil{return e}
 if m.omit{return nil}
 if m.bad{b=[]byte("not PDF")}
 if m.wrongPages{
  var picked engine.Result
  picked,e=engine.Run("pdf-extract",[]engine.Input{{Name:"input.pdf",Data:b}},"1")
  if e!=nil{return e}
  b=picked.Data
 }
 return os.WriteFile(out,b,0600)
}
func pdfFixture(t *testing.T)engine.Input{
 t.Helper()
 b,e:=os.ReadFile("../../tests/fixtures/pdf_two_pages.pdf");if e!=nil{t.Fatal(e)}
 return engine.Input{Name:"scan.pdf",Data:b}
}
func TestPDFOCRArgvAndPageCount(t *testing.T){
 fake:=&fakePDFExec{}
 res,e:=ProcessPDFWith(context.Background(),fake,"ocrmypdf",pdfFixture(t),"eng+vie")
 if e!=nil{t.Fatal(e)}
 if fake.calls!=1{t.Fatalf("calls %d",fake.calls)}
 want:=[]string{"--skip-text","--output-type","pdf","--jobs","1","--optimize","0","--tesseract-timeout","25","-l","eng+vie"}
 if strings.Join(fake.args[:len(want)],"|")!=strings.Join(want,"|"){t.Fatalf("unsafe/unexpected command argv: %v",fake.args)}
 if !bytes.Equal(res.Data,pdfFixture(t).Data){t.Fatal("mock runner did not receive exact source PDF")}
 if res.Filename!="OpenDesk_OCR_PDF_Searchable.pdf"{t.Fatal(res.Filename)}
 got,e:=api.PageCount(context.Background(),bytes.NewReader(res.Data),nil);if e!=nil||got!=2{t.Fatalf("page check %d %v",got,e)}
 if _,e:=os.Stat(fake.tempDir);!os.IsNotExist(e){t.Fatal("temporary OCR data was not cleaned")}
}
func TestPDFOCRFailClosed(t *testing.T){
 source:=pdfFixture(t)
 for _,tc:=range []struct{file engine.Input;lang string;runner *fakePDFExec}{
  {source,"bad",&fakePDFExec{}},
  {engine.Input{Name:"fake.pdf",Data:[]byte("%PDF-broken")},"eng",&fakePDFExec{}},
  {engine.Input{Name:"wrong.txt",Data:source.Data},"eng",&fakePDFExec{}},
  {source,"eng",&fakePDFExec{err:errors.New("OCR failure")}},
  {source,"eng",&fakePDFExec{omit:true}},
  {source,"eng",&fakePDFExec{bad:true}},
  {source,"eng",&fakePDFExec{wrongPages:true}},
 } {
  _,e:=ProcessPDFWith(context.Background(),tc.runner,"ocrmypdf",tc.file,tc.lang)
  if e==nil{t.Fatalf("unexpected success for %+v",tc)}
  if tc.runner.tempDir!=""{if _,e:=os.Stat(tc.runner.tempDir);!os.IsNotExist(e){t.Fatal("temp files persisted on error")}}
 }
}
func TestPDFOCRRejectsOversizedSelection(t *testing.T){
 f:=pdfFixture(t)
 for _,lang:=range []string{"","ger","eng;rm -rf /","vie+eng"} {
  if _,e:=ProcessPDFWith(context.Background(),&fakePDFExec{},"ocrmypdf",f,lang);e==nil{t.Fatalf("accepted language %q",lang)}
 }
 if _,e:=ProcessPDFWith(context.Background(),nil,"ocrmypdf",f,"eng");e==nil{t.Fatal("nil runner accepted")}
}
