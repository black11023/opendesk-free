package ocr

import (
 "bytes"
 "context"
 "errors"
 "image"
 "image/color"
 "image/png"
 "strings"
 "testing"
 "github.com/pdfcpu/pdfcpu/pkg/api"
 "opendesk-free/internal/engine"
)
type fakeRunner struct {calls int; lastLanguage string; lastFormat string; out []byte; err error}
func(f *fakeRunner)Recognize(ctx context.Context,b []byte,lang,format string)([]byte,error){
 f.calls++;f.lastLanguage=lang;f.lastFormat=format
 if f.err!=nil{return nil,f.err}
 return f.out,nil
}
func imageFixture(t *testing.T)engine.Input {
 t.Helper()
 img:=image.NewRGBA(image.Rect(0,0,80,60))
 for y:=0;y<60;y++{for x:=0;x<80;x++{img.Set(x,y,color.RGBA{255,255,255,255})}}
 var b bytes.Buffer;if e:=png.Encode(&b,img);e!=nil{t.Fatal(e)}
 return engine.Input{Name:"scan.png",Data:b.Bytes()}
}
func TestTextOCRUsesMockEngineAndNames(t *testing.T){
 f:=&fakeRunner{out:[]byte("Hello from OCR")}
 a:=imageFixture(t)
 result,e:=Process(context.Background(),f,[]engine.Input{a,a},"eng+vie","txt")
 if e!=nil{t.Fatal(e)}
 if f.calls!=2||f.lastLanguage!="eng+vie"||f.lastFormat!="txt"{t.Fatal(f)}
 if result.ContentType!="text/plain; charset=utf-8"||!bytes.Contains(result.Data,[]byte("Hello from OCR")){t.Fatal("wrong OCR result")}
 if strings.Count(string(result.Data),"===== scan.png =====")!=2{t.Fatalf("%q",result.Data)}
}
func TestOCRStrictValidation(t *testing.T){
 a:=imageFixture(t)
 for _,tc:=range []struct{files []engine.Input;lang,format string}{
  {[]engine.Input{a},"bad","txt"},
  {[]engine.Input{a},"eng","jpeg"},
  {[]engine.Input{{Name:"fake.jpg",Data:a.Data}},"eng","txt"},
  {[]engine.Input{{Name:"bad.png",Data:[]byte("not an image")}},"eng","txt"},
  {[]engine.Input{{Name:"bad.txt",Data:a.Data}},"eng","txt"},
 }{
  f:=&fakeRunner{out:[]byte("hello")}
  _,e:=Process(context.Background(),f,tc.files,tc.lang,tc.format)
  if e==nil||f.calls!=0{t.Errorf("invalid upload accepted: %+v",tc)}
 }
}
func TestOCRLanguagePackGates(t *testing.T){
 if selectedLanguages("eng",Status{English:true})!=nil{t.Fatal("English should work")}
 if selectedLanguages("vie",Status{Vietnamese:true})!=nil{t.Fatal("Vietnamese should work")}
 if selectedLanguages("eng+vie",Status{English:true})==nil{t.Fatal("missing Vietnamese should fail")}
 if selectedLanguages("vie",Status{English:true})==nil{t.Fatal("missing Vietnamese should fail")}
}
func TestOCRFailureIsFailClosed(t *testing.T){
 f:=&fakeRunner{err:errors.New("no recognition")}
 _,e:=Process(context.Background(),f,[]engine.Input{imageFixture(t)},"eng","txt")
 if e==nil{t.Fatal("OCR failure must never yield a file")}
}
func TestSearchablePDFVerification(t *testing.T){
 a:=imageFixture(t)
 p,e:=engine.Run("images-pdf",[]engine.Input{a});if e!=nil{t.Fatal(e)}
 f:=&fakeRunner{out:p.Data}
 res,e:=Process(context.Background(),f,[]engine.Input{a},"eng","pdf")
 if e!=nil{t.Fatal(e)}
 n,e:=api.PageCount(context.Background(),bytes.NewReader(res.Data),nil)
 if e!=nil||n!=1{t.Fatalf("OCR PDF page count %d %v",n,e)}
 res,e=Process(context.Background(),f,[]engine.Input{a,a},"eng","pdf")
 if e!=nil{t.Fatal(e)}
 n,e=api.PageCount(context.Background(),bytes.NewReader(res.Data),nil)
 if e!=nil||n!=2{t.Fatalf("merged OCR PDF page count %d %v",n,e)}
}
