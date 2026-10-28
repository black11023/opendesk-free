package preservetranslate
import(
 "archive/zip"
 "bytes"
 "context"
 "crypto/sha256"
 "errors"
 "io"
 "strings"
 "testing"
 "opendesk-free/internal/engine"
)
const excelNS="http://schemas.openxmlformats.org/spreadsheetml/2006/main"
func xlsxFixture(t *testing.T,sheet,shared string,extra map[string]string)[]byte{
 t.Helper()
 all:=map[string]string{
 "[Content_Types].xml":`<?xml version="1.0"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/><Override PartName="/xl/sharedStrings.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sharedStrings+xml"/></Types>`,
 "_rels/.rels":`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`,
 "xl/_rels/workbook.xml.rels":`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/><Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/sharedStrings" Target="sharedStrings.xml"/></Relationships>`,
 "xl/workbook.xml":`<workbook xmlns="`+excelNS+`" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="Budget" sheetId="1" r:id="rId1"/></sheets></workbook>`,
 "xl/styles.xml":`<styleSheet xmlns="`+excelNS+`"><fonts count="1"><font><sz val="11"/><name val="Calibri"/></font></fonts><fills count="1"><fill><patternFill patternType="none"/></fill></fills><borders count="1"><border/></borders><cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs><cellXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/></cellXfs></styleSheet>`,
 "xl/worksheets/sheet1.xml":sheet,
 "xl/sharedStrings.xml":shared,
 }
 for k,v:=range extra{all[k]=v}
 var b bytes.Buffer;writer:=zip.NewWriter(&b)
 for k,v:=range all{
  p,e:=writer.Create(k);if e!=nil{t.Fatal(e)}
  if _,e=p.Write([]byte(v));e!=nil{t.Fatal(e)}
 }
 if e:=writer.Close();e!=nil{t.Fatal(e)}
 return b.Bytes()
}
func spreadsheetSheet(t *testing.T)string{
 t.Helper()
 return `<worksheet xmlns="`+excelNS+`"><dimension ref="A1:D3"/><sheetData><row r="1"><c r="A1" t="s"><v>0</v></c><c r="B1" s="0" t="n"><v>1250.25</v></c><c r="C1"><f>B1*2</f><v>2500.5</v></c></row><row r="2"><c r="A2" t="s"><v>1</v></c><c r="D2" t="inlineStr"><is><t>Hello 2026</t></is></c></row><row r="3"><c r="A3" t="s"><v>0</v></c></row></sheetData><mergeCells count="1"><mergeCell ref="A2:B2"/></mergeCells></worksheet>`
}
func spreadsheetShared()string{
 return `<sst xmlns="`+excelNS+`" count="3" uniqueCount="2"><si><t>Hello DN20</t></si><si><t>Hello world</t></si></sst>`
}
func extractZipPart(t *testing.T,raw []byte,name string)[]byte{
 t.Helper();z,e:=zip.NewReader(bytes.NewReader(raw),int64(len(raw)));if e!=nil{t.Fatal(e)}
 for _,p:=range z.File{if p.Name==name{r,e:=p.Open();if e!=nil{t.Fatal(e)};defer r.Close();b,e:=io.ReadAll(r);if e!=nil{t.Fatal(e)};return b}}
 t.Fatalf("ZIP entry missing: %s",name);return nil
}
func TestTranslateXLSXUnchangedFormulasNumbersStylesAndReferences(t *testing.T){
 original:=xlsxFixture(t,spreadsheetSheet(t),spreadsheetShared(),nil)
 inputDigest:=sha256.Sum256(original)
 out,e:=TranslateXLSX(context.Background(),[]engine.Input{{Name:"Rates.xlsx",Data:original}},"en","vi",fakeTranslator{})
 if e!=nil{t.Fatal(e)}
 if out.ContentType!="application/zip"{t.Fatal(out.ContentType)}
 workbook:=extractZipPart(t,out.Data,"Rates_vi.xlsx")
 sst:=string(extractZipPart(t,workbook,"xl/sharedStrings.xml"))
 sheet:=string(extractZipPart(t,workbook,"xl/worksheets/sheet1.xml"))
 if !strings.Contains(sst,"Xin chào DN20")||!strings.Contains(sst,"Xin chào world"){t.Fatalf("shared strings did not translate: %s",sst)}
 if !strings.Contains(sheet,"Xin chào 2026"){t.Fatalf("inline label not translated: %s",sheet)}
 if !strings.Contains(sheet,"<f>B1*2</f><v>2500.5</v>")||!strings.Contains(sheet,">1250.25</v>"){t.Fatal("formula or number changed")}
 for _,p:=range []string{"xl/styles.xml","xl/workbook.xml","xl/_rels/workbook.xml.rels"}{
  if !bytes.Equal(extractZipPart(t,workbook,p),extractZipPart(t,original,p)){t.Fatalf("untouched OOXML part modified: %s",p)}
 }
 if !strings.Contains(sheet,`<mergeCell ref="A2:B2"/>`){t.Fatal("merge removed")}
 if sha256.Sum256(original)!=inputDigest{t.Fatal("input mutated")}
 audit:=string(extractZipPart(t,out.Data,"Translation_Audit.csv"))
 if !strings.Contains(audit,"xl/worksheets/sheet1.xml!A1; xl/worksheets/sheet1.xml!A3")||!strings.Contains(audit,"REVIEW_MEANING_AND_CELL_OVERFLOW"){t.Fatalf("audit missing shared string references: %s",audit)}
 if !strings.Contains(string(extractZipPart(t,out.Data,"MANIFEST.txt")),"Original SHA256:"){t.Fatal("manifest missing")}
}
func TestTranslateXLSXFailClosedCases(t *testing.T){
 for _,tt:=range []struct{name,sheet,shared string;extra map[string]string}{
  {"missing-shared-index",strings.Replace(spreadsheetSheet(t),`<v>1</v>`,`<v>9</v>`,1),spreadsheetShared(),nil},
  {"rich-shared",spreadsheetSheet(t),`<sst xmlns="`+excelNS+`"><si><r><t>Hello</t></r></si><si><t>Hello world</t></si></sst>`,nil},
  {"header-footer",strings.Replace(spreadsheetSheet(t),"</worksheet>","<headerFooter><oddHeader>Hello</oddHeader></headerFooter></worksheet>",1),spreadsheetShared(),nil},
  {"external-link",spreadsheetSheet(t),spreadsheetShared(),map[string]string{"xl/externalLinks/externalLink1.xml":"<external/>"}},
  {"bad-namespace",strings.Replace(spreadsheetSheet(t),excelNS,"https://example.invalid/spreadsheet",1),spreadsheetShared(),nil},
  {"rich-inline",strings.Replace(spreadsheetSheet(t),"<is><t>Hello 2026</t></is>","<is><r><t>Hello</t></r><r><t>2026</t></r></is>",1),spreadsheetShared(),nil},
  {"duplicate-cell",strings.Replace(spreadsheetSheet(t),`<c r="A2" t="s"><v>1</v></c>`,`<c r="A1" t="s"><v>1</v></c>`,1),spreadsheetShared(),nil},
 }{
  t.Run(tt.name,func(t *testing.T){
   source:=xlsxFixture(t,tt.sheet,tt.shared,tt.extra)
   result,e:=TranslateXLSX(context.Background(),[]engine.Input{{Name:"Rates.xlsx",Data:source}},"en","vi",fakeTranslator{})
   if e==nil||len(result.Data)!=0||!strings.Contains(e.Error(),"BLOCKED"){t.Fatalf("must block safely, got %v",e)}
  })
 }
}
type noXLSXModel struct{}
func(noXLSXModel)Translate(context.Context,string,string,[]string)([]string,error){return nil,errors.New("model missing")}
func TestTranslateXLSXMissingModelAndParameterRules(t *testing.T){
 fixture:=xlsxFixture(t,spreadsheetSheet(t),spreadsheetShared(),nil)
 file:=[]engine.Input{{Name:"Rates.xlsx",Data:fixture}}
 if r,e:=TranslateXLSX(context.Background(),file,"en","vi",noXLSXModel{});e==nil||len(r.Data)>0{t.Fatal("missing model should not produce workbook")}
 if _,e:=TranslateXLSX(context.Background(),file,"en","en",fakeTranslator{});e==nil{t.Fatal("same language must fail")}
 if _,e:=TranslateXLSX(context.Background(),file,"auto","vi",fakeTranslator{});e==nil{t.Fatal("unknown source must fail")}
}
