package engine
import (
 "bytes"
 "image"
 "image/color"
 "image/jpeg"
 "image/png"
 "testing"
)
func fixtureImage(t *testing.T,format string) Input{
 t.Helper()
 img:=image.NewRGBA(image.Rect(0,0,20,20))
 for y:=0;y<20;y++{for x:=0;x<20;x++{img.Set(x,y,color.RGBA{R:100,G:150,B:200,A:255})}}
 var b bytes.Buffer
 if format=="png"{if e:=png.Encode(&b,img);e!=nil{t.Fatal(e)};return Input{Name:"one.png",Data:b.Bytes()}}
 if e:=jpeg.Encode(&b,img,&jpeg.Options{Quality:85});e!=nil{t.Fatal(e)}
 return Input{Name:"two.jpg",Data:b.Bytes()}
}
func TestImagesToPDFPNGAndJPEG(t *testing.T){
 files:=[]Input{fixtureImage(t,"png"),fixtureImage(t,"jpeg")}
 out,e:=Run("images-pdf",files);if e!=nil{t.Fatal(e)}
 if out.ContentType!="application/pdf"{t.Fatalf("unexpected content type %s",out.ContentType)}
 checkPages(t,out.Data,2)
}
func TestImagesToPDFRejectsInvalid(t *testing.T){
 for _,items:=range [][]Input{
  {{Name:"invalid.png",Data:[]byte("not an image")}},
  {{Name:"invalid.txt",Data:[]byte("not an image")}},
  {{Name:"fake.jpg",Data:fixtureImage(t,"png").Data}},
 }{
  if _,e:=Run("images-pdf",items);e==nil{t.Errorf("unexpectedly accepted %+v",items)}
 }
}
