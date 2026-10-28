package preservetranslate

import (
 "archive/zip"
 "bytes"
 "context"
 "crypto/sha256"
 "encoding/csv"
 "encoding/hex"
 "strings"
 "testing"

 "opendesk-free/internal/engine"
)
func testArchive(t *testing.T,parts map[string]string)[]byte{
 t.Helper()
 var b bytes.Buffer; w:=zip.NewWriter(&b)
 for n,v:=range parts {out,e:=w.Create(n);if e!=nil{t.Fatal(e)};if _,e=out.Write([]byte(v));e!=nil{t.Fatal(e)}}
 if e:=w.Close();e!=nil{t.Fatal(e)}
 return b.Bytes()
}
func TestReadOnlyNativePreflightAndSourceHash(t *testing.T){
 doc:=testArchive(t,map[string]string{
 "[Content_Types].xml":`<Types/>`,
 "word/document.xml":`<w:document xmlns:w="urn:word"><w:body><w:p><w:r><w:t>Xin chào</w:t></w:r></w:p></w:body></w:document>`,
 })
 excel:=testArchive(t,map[string]string{
 "[Content_Types].xml":`<Types/>`,
 "xl/workbook.xml":`<workbook/>`,
 "xl/worksheets/sheet1.xml":`<worksheet><sheetData><row><c r="A1" t="n"><v>100</v></c><c r="B1"><f>A1*2</f><v>200</v></c></row></sheetData><mergeCells><mergeCell ref="A2:B2"/></mergeCells></worksheet>`,
 "xl/sharedStrings.xml":`<sst><si><t>Đơn giá</t></si></sst>`,
 })
 pdf:=[]byte("%PDF-should-not-be-accepted")
 original:=append([]byte(nil),doc...)
 r,e:=Preflight(context.Background(),[]engine.Input{{Name:"A.docx",Data:doc},{Name:"B.xlsx",Data:excel},{Name:"C.pdf",Data:pdf}})
 if e!=nil{t.Fatal(e)}
 if r.ContentType!="text/csv; charset=utf-8"||!bytes.HasPrefix(r.Data,[]byte{0xef,0xbb,0xbf}){t.Fatal("invalid CSV output")}
 rows,e:=csv.NewReader(bytes.NewReader(r.Data[3:])).ReadAll()
 if e!=nil{t.Fatal(e)}
 if len(rows)!=4{t.Fatalf("expected header+3 rows, got %d",len(rows))}
 sum:=sha256.Sum256(doc)
 if rows[1][2]!=hex.EncodeToString(sum[:]) || rows[1][3]!="READY"||rows[1][5]!="1"{t.Fatalf("doc unexpected: %v",rows[1])}
 if rows[2][3]!="READY"||rows[2][6]!="1"||rows[2][8]!="1"{t.Fatalf("xlsx formulas/merged incorrect: %v",rows[2])}
 if rows[3][3]!="BLOCKED" {t.Fatalf("pdf should be blocked: %v",rows[3])}
 if !bytes.Equal(doc,original){t.Fatal("source bytes modified")}
}
func TestMaliciousOOXMLRejected(t *testing.T){
 cases:=[]struct{name string;parts map[string]string;want string}{
 {"../escape.docx",map[string]string{"[Content_Types].xml":"<Types/>","word/document.xml":"<doc><t>x</t></doc>"},"Unsafe ZIP entry path"},
 {"macro.docx",map[string]string{"[Content_Types].xml":"<Types/>","word/document.xml":"<doc><t>x</t></doc>","word/vbaProject.bin":"x"},"Macros"},
 {"missing.docx",map[string]string{"[Content_Types].xml":"<Types/>"},"Required OOXML"},
 {"bad.docx",map[string]string{"[Content_Types].xml":"<Types/>","word/document.xml":"<broken"},"Invalid"},
 }
 for _,tc:=range cases {
   blob:=testArchive(t,tc.parts)
   if tc.name=="../escape.docx"{blob=testArchive(t,map[string]string{"../evil":"x","word/document.xml":"<doc><t>x</t></doc>","[Content_Types].xml":"<Types/>"})}
   row:=inspectOOXML(context.Background(),engine.Input{Name:tc.name,Data:blob})
   if row.Status!="BLOCKED"||!strings.Contains(row.Findings,tc.want){t.Fatalf("%s: %+v",tc.name,row)}
 }
}
func TestExternalRelationshipsAndSpreadsheetInjection(t *testing.T){
 b:=testArchive(t,map[string]string{
   "[Content_Types].xml":"<Types/>",
   "word/document.xml":"<doc><t>A</t></doc>",
   "word/_rels/document.xml.rels":`<Relationships><Relationship TargetMode="External" Target="https://example.com"/></Relationships>`,
 })
 rec:=inspectOOXML(context.Background(),engine.Input{Name:"=report.docx",Data:b})
 if rec.Status!="REVIEW"{t.Fatal(rec)}
 out,e:=Preflight(context.Background(),[]engine.Input{{Name:"=report.docx",Data:b}})
 if e!=nil{t.Fatal(e)}
 rows,e:=csv.NewReader(bytes.NewReader(out.Data[3:])).ReadAll();if e!=nil{t.Fatal(e)}
 if rows[1][0]!="'=report.docx"{t.Fatal("CSV formula injection not escaped",rows[1][0])}
}
func TestPreflightLimits(t *testing.T){
 if _,e:=Preflight(context.Background(),nil);e==nil{t.Fatal("empty selection accepted")}
 many:=make([]engine.Input,9)
 if _,e:=Preflight(context.Background(),many);e==nil{t.Fatal("more than eight accepted")}
 ctx,cancel:=context.WithCancel(context.Background());cancel()
 if _,e:=Preflight(ctx,[]engine.Input{{Name:"a.docx",Data:[]byte("x")}});e==nil{t.Fatal("cancelled not rejected")}
}
