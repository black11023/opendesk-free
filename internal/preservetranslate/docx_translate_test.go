package preservetranslate
import(
 "archive/zip"
 "bytes"
 "context"
 "crypto/sha256"
 "encoding/hex"
 "errors"
 "io"
 "strings"
 "testing"
 "opendesk-free/internal/engine"
)
type fakeTranslator struct {Broken bool}
func (f fakeTranslator)Translate(_ context.Context,_,_ string,segments []string)([]string,error){
 out:=make([]string,len(segments))
 for i,s:=range segments{out[i]=strings.ReplaceAll(s,"Hello","Xin chào")}
 if f.Broken{out[0]=strings.ReplaceAll(out[0],"__ODS_LOCK_0001__","9999")}
 return out,nil
}
func docxFixture(t *testing.T,paragraphs string,extra map[string]string)[]byte{
 t.Helper();var b bytes.Buffer;w:=zip.NewWriter(&b)
 files:=map[string]string{
 "[Content_Types].xml":"<Types/>",
 "word/document.xml":`<?xml version="1.0"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>`+paragraphs+`</w:body></w:document>`,
 "word/styles.xml":"<styles>Styles must stay byte-identical</styles>",
 }
 for k,v:=range extra{files[k]=v}
 for k,v:=range files{p,e:=w.Create(k);if e!=nil{t.Fatal(e)};if _,e=p.Write([]byte(v));e!=nil{t.Fatal(e)}}
 if e:=w.Close();e!=nil{t.Fatal(e)}
 return b.Bytes()
}
func unzipPart(t *testing.T,archive []byte,name string)[]byte{
 t.Helper();z,e:=zip.NewReader(bytes.NewReader(archive),int64(len(archive)));if e!=nil{t.Fatal(e)}
 for _,f:=range z.File{if f.Name==name{r,e:=f.Open();if e!=nil{t.Fatal(e)};defer r.Close();b,e:=io.ReadAll(r);if e!=nil{t.Fatal(e)};return b}}
 t.Fatalf("missing %s",name);return nil
}
func TestDOCXTranslatePreservesSourceStylesNumbersAndAudit(t *testing.T){
 source:=docxFixture(t,`<w:p><w:r><w:t>Hello DN20 costs 1,250.00 USD.</w:t></w:r></w:p><w:p><w:r><w:t>Hello world!</w:t></w:r></w:p>`,nil)
 before:=sha256.Sum256(source)
 result,e:=TranslateDOCX(context.Background(),[]engine.Input{{Name:"Tender.docx",Data:source}},"en","vi",fakeTranslator{})
 if e!=nil{t.Fatal(e)}
 if result.ContentType!="application/zip"{t.Fatal(result.ContentType)}
 target:=unzipPart(t,result.Data,"Tender_vi.docx")
 got:=string(unzipPart(t,target,"word/document.xml"))
 if !strings.Contains(got,"Xin chào DN20 costs 1,250.00 USD.")||!strings.Contains(got,"Xin chào world!"){t.Fatalf("translation or protected numeric lost: %s",got)}
 if !bytes.Equal(unzipPart(t,target,"word/styles.xml"),[]byte("<styles>Styles must stay byte-identical</styles>")){t.Fatal("styles changed")}
 if before!=sha256.Sum256(source){t.Fatal("input modified")}
 audit:=string(unzipPart(t,result.Data,"Translation_Audit.csv"))
 if !strings.Contains(audit,"REVIEW_LAYOUT_AND_MEANING")||!strings.Contains(audit,"DN20"){t.Fatal("audit incomplete")}
 manifest:=string(unzipPart(t,result.Data,"MANIFEST.txt"))
 if !strings.Contains(manifest,hex.EncodeToString(before[:])){t.Fatal("manifest lost provenance")}
}
func TestDOCXRejectsMultirunAndHighRiskBeforeModel(t *testing.T){
 cases:=[]string{
 `<w:p><w:r><w:t>Hello</w:t></w:r><w:r><w:t> World</w:t></w:r></w:p>`,
 `<w:p><w:hyperlink r:id="id1"><w:r><w:t>Hello</w:t></w:r></w:hyperlink></w:p>`,
 }
 for _,body:=range cases{
   f:=docxFixture(t,body,nil)
   _,e:=TranslateDOCX(context.Background(),[]engine.Input{{Name:"test.docx",Data:f}},"en","vi",fakeTranslator{})
   if e==nil||!strings.Contains(e.Error(),"BLOCKED"){t.Fatalf("expected fail-closed, got %v",e)}
 }
}
func TestDOCXRejectsUnchangedHeaderAndTamperedLock(t *testing.T){
 p:=`<w:p><w:r><w:t>Hello DN20</w:t></w:r></w:p>`
 f:=docxFixture(t,p,map[string]string{"word/header1.xml":"<header>Confidential</header>"})
 if _,e:=TranslateDOCX(context.Background(),[]engine.Input{{Name:"a.docx",Data:f}},"en","vi",fakeTranslator{});e==nil{t.Fatal("header not blocked")}
 f=docxFixture(t,p,nil)
 if _,e:=TranslateDOCX(context.Background(),[]engine.Input{{Name:"a.docx",Data:f}},"en","vi",fakeTranslator{Broken:true});e==nil||!strings.Contains(e.Error(),"BLOCKED"){t.Fatalf("token mutation not caught: %v",e)}
}
type failingTranslator struct {}
func (failingTranslator)Translate(context.Context,string,string,[]string)([]string,error){return nil,errors.New("no installed model")}
func TestDOCXMissingModelFailsAndNoPartialOutput(t *testing.T){
 f:=docxFixture(t,`<w:p><w:r><w:t>Hello</w:t></w:r></w:p>`,nil)
 if result,e:=TranslateDOCX(context.Background(),[]engine.Input{{Name:"a.docx",Data:f}},"en","vi",failingTranslator{});e==nil||len(result.Data)!=0{t.Fatal("missing model should not generate an output")}
 if _,e:=TranslateDOCX(context.Background(),[]engine.Input{{Name:"a.docx",Data:f}},"auto","vi",fakeTranslator{});e==nil{t.Fatal("unsupported auto source accepted")}
 if _,e:=TranslateDOCX(context.Background(),[]engine.Input{{Name:"a.docx",Data:f}},"en","en",fakeTranslator{});e==nil{t.Fatal("same source/dest accepted")}
}
