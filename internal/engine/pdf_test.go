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

func TestSelectedPagesAndOrdering(t *testing.T){
 p:=twoPageFixture(t)
 selected,e:=Run("pdf-extract",[]Input{p},"2");if e!=nil{t.Fatal(e)}
 checkPages(t,selected.Data,1)
 reversed,e:=Run("pdf-extract",[]Input{p},"2,1");if e!=nil{t.Fatal(e)}
 checkPages(t,reversed.Data,2)
 merged,e:=Run("pdf-merge",[]Input{p,p},"2","1");if e!=nil{t.Fatal(e)}
 checkPages(t,merged.Data,2)
 parts,e:=Run("pdf-split",[]Input{p},"2");if e!=nil{t.Fatal(e)}
 z,e:=zip.NewReader(bytes.NewReader(parts.Data),int64(len(parts.Data)));if e!=nil{t.Fatal(e)}
 if len(z.File)!=1||z.File[0].Name!="OpenDesk_Page_0002.pdf"{t.Fatalf("wrong selected output %+v",z.File)}
}
func TestPageParserStrictBounds(t *testing.T){
 for _,tc:=range []struct{spec string;pages int;want []int}{
  {"2",550,[]int{2}},
  {"7,2,5",10,[]int{7,2,5}},
  {"3-5",10,[]int{3,4,5}},
  {"5-3",10,[]int{5,4,3}},
  {"2,2",3,[]int{2,2}},
 }{
  got,e:=ParsePageSelection(tc.spec,tc.pages);if e!=nil{t.Fatal(e)}
  if len(got)!=len(tc.want){t.Fatalf("%s: got %v",tc.spec,got)}
  for i:=range got{if got[i]!=tc.want[i]{t.Fatalf("%s: got %v",tc.spec,got)}}
 }
 for _,spec:=range []string{"0","3","1-4","x","1,,2","1-","-1","1--2","1.5","1;2"}{
  if _,e:=ParsePageSelection(spec,2);e==nil{t.Errorf("expected rejection for %q",spec)}
 }
 if _,e:=ParsePageSelection("",550);e==nil{t.Fatal("large PDF must require page selection")}
}
func TestPDFInvalidPageSelectionDoesNotReturnPartialResult(t *testing.T){
 src:=twoPageFixture(t)
 for _,tc:=range []struct{op string;files []Input;selection []string}{
  {"pdf-extract",[]Input{src},[]string{"3"}},
  {"pdf-split",[]Input{src},[]string{"1,1"}},
  {"pdf-merge",[]Input{src,src},[]string{"1"}},
  {"pdf-merge",[]Input{src,src},[]string{"1","5"}},
 } {
  if _,e:=Run(tc.op,tc.files,tc.selection...);e==nil{t.Errorf("%s: expected failure",tc.op)}
 }
}
