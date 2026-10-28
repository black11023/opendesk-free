//go:build translation_integration

package preservetranslate

import (
 "bytes"
 "archive/zip"
 "context"
 "io"
 "os"
 "os/exec"
 "path/filepath"
 "time"
 "strings"
 "testing"
 "opendesk-free/internal/engine"
)
func TestRealArgosDOCX(t *testing.T){
 ctx:=context.Background()
 status:=ProbeArgos(ctx)
 if !status.Installed{t.Fatalf("real Argos not installed: %+v",status)}
 if !strings.Contains(strings.Join(status.DirectPairs,","),"en->es"){t.Fatalf("real direct en->es model not available: %+v",status)}

 source:=docxFixture(t,`<w:p><w:r><w:t>Hello world.</w:t></w:r></w:p><w:p><w:r><w:t>Good morning.</w:t></w:r></w:p><w:sectPr/>`,
  map[string]string{
   "[Content_Types].xml":`<?xml version="1.0" encoding="UTF-8"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/></Types>`,
   "_rels/.rels":`<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`,
   "word/styles.xml":`<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"/>`,
  })
 original:=append([]byte(nil),source...)
 output,err:=TranslateDOCX(ctx,[]engine.Input{{Name:"Real.docx",Data:source}},"en","es",ArgosProvider{})
 if err!=nil{t.Fatalf("real Argos translation failed: %v",err)}
 resultDoc:=unzipPart(t,output.Data,"Real_es.docx")
 xml:=string(unzipPart(t,resultDoc,"word/document.xml"))
 if strings.Contains(xml,"Hello world.")||strings.Contains(xml,"Good morning.") {
   t.Fatalf("real model failed to translate text: %s",xml)
 }
 if !strings.Contains(xml,"<w:document")||!strings.Contains(xml,"<w:p>"){t.Fatal("DOCX format lost")}
 if !bytes.Equal(source,original){t.Fatal("original changed")}
 if !strings.Contains(string(unzipPart(t,output.Data,"Translation_Audit.csv")),"REVIEW_LAYOUT_AND_MEANING"){t.Fatal("audit missing")}

 // Validate an actual standards-conformant output using LibreOffice, not only
 // synthetic ZIP/XML parsing. This is a preview fidelity gate, not a claim
 // of identical page layout for arbitrary source documents.
 dir:=t.TempDir()
 translatedPath:=filepath.Join(dir,"Real_es.docx")
 if err:=os.WriteFile(translatedPath,resultDoc,0600);err!=nil{t.Fatal(err)}
 soffice,e:=exec.LookPath("soffice");if e!=nil{t.Fatalf("LibreOffice Writer required: %v",e)}
 officeCtx,cancel:=context.WithTimeout(ctx,45*time.Second);defer cancel()
 cmd:=exec.CommandContext(officeCtx,soffice,"-env:UserInstallation=file://"+filepath.ToSlash(filepath.Join(dir,"profile")),"--headless","--convert-to","pdf","--outdir",dir,translatedPath)
 logs,e:=cmd.CombinedOutput()
 if e!=nil{t.Fatalf("real LibreOffice failed to open translated DOCX: %v: %s",e,logs)}
 pdfPath:=filepath.Join(dir,"Real_es.pdf")
 pdf,e:=os.Open(pdfPath);if e!=nil{t.Fatalf("translated DOCX could not render to PDF: %v",e)}
 defer pdf.Close()
 hdr:=make([]byte,5)
 if _,e=io.ReadFull(pdf,hdr);e!=nil||string(hdr)!="%PDF-"{t.Fatalf("Office produced invalid PDF: %v",e)}
 rendered,e:=exec.CommandContext(officeCtx,"pdftotext",pdfPath,"-").Output()
 if e!=nil||len(strings.TrimSpace(string(rendered)))<7{t.Fatalf("rendered translation is missing: %v, text=%s",e,rendered)}
 if strings.Contains(string(rendered),"Hello world."){t.Fatalf("Office still renders original source text: %s",rendered)}
 t.Log("PASS: genuine Argos inference, byte-unchanged source, audit and LibreOffice open/render to searchable PDF")
}

func TestRealApprovedModelInstall(t *testing.T){
 path:=os.Getenv("ARGOS_TEST_MODEL")
 if path==""{t.Fatal("real pinned Argos model path required")}
 file,err:=os.Open(path);if err!=nil{t.Fatal(err)}
 defer file.Close()
 receipt,err:=InstallApprovedModel(context.Background(),"argos-en-es-1.0",file)
 if err!=nil{t.Fatalf("actual Go model-pack installer rejected official pinned model: %v",err)}
 if receipt.Status!="READY"||receipt.Pair!="en->es"||receipt.SHA256!=approvedPacks[0].SHA256{t.Fatalf("wrong installer receipt: %+v",receipt)}
 t.Log("PASS: offline Go intake, SHA-256/ZIP validation, Argos package install and live inference")
}


func TestRealArgosXLSX(t *testing.T){
 ctx:=context.Background()
 status:=ProbeArgos(ctx)
 if !status.Installed||!strings.Contains(strings.Join(status.DirectPairs,","),"en->es"){
   t.Fatalf("installed direct Argos en->es required: %+v",status)
 }
 // Reuse the same standards-conformant XLSX package that LibreOffice
 // integration tests already open successfully; add one formula, one number
 // and a plain inline translation label without altering the package metadata.
 fixture,err:=os.ReadFile("../../tests/generated/real_sheet.xlsx")
 if err!=nil{t.Fatalf("real XLSX fixture not generated: %v",err)}
 archive,err:=zip.NewReader(bytes.NewReader(fixture),int64(len(fixture)))
 if err!=nil{t.Fatal(err)}
 var built bytes.Buffer
 wz:=zip.NewWriter(&built)
 for _,part:=range archive.File{
  if part.Name!="xl/worksheets/sheet1.xml"{
    if err=wz.Copy(part);err!=nil{t.Fatal(err)}
    continue
  }
  r,e:=part.Open();if e!=nil{t.Fatal(e)}
  original,e:=io.ReadAll(r);r.Close();if e!=nil{t.Fatal(e)}
  target:=`<c r="A1" t="inlineStr"><is><t>OPENDESK SHEET 2026</t></is></c>`
  replacement:=`<c r="A1" t="inlineStr"><is><t>Hello world.</t></is></c><c r="B1" t="n"><v>1250.25</v></c><c r="C1"><f>B1*2</f><v>2500.5</v></c>`
  if !bytes.Contains(original,[]byte(target)){t.Fatal("generated fixture did not have expected inline cell")}
  edited:=bytes.Replace(original,[]byte(target),[]byte(replacement),1)
  header:=part.FileHeader
  out,e:=wz.CreateHeader(&header);if e!=nil{t.Fatal(e)}
  if _,e=out.Write(edited);e!=nil{t.Fatal(e)}
 }
 if err=wz.Close();err!=nil{t.Fatal(err)}
 source:=built.Bytes()
 preserved:=append([]byte(nil),source...)
 response,err:=TranslateXLSX(ctx,[]engine.Input{{Name:"Budget.xlsx",Data:source}},"en","es",ArgosProvider{})
 if err!=nil{t.Fatalf("actual EN->ES Argos XLSX translation failed: %v",err)}
 result:=extractZipPart(t,response.Data,"Budget_es.xlsx")
 sheet:=string(extractZipPart(t,result,"xl/worksheets/sheet1.xml"))
 if strings.Contains(sheet,"Hello world."){t.Fatalf("real model did not translate Excel text: %s",sheet)}
 if !strings.Contains(sheet,"<f>B1*2</f><v>2500.5</v>")||!strings.Contains(sheet,"<v>1250.25</v>"){
  t.Fatal("formula or number changed during translation")
 }
 if !bytes.Equal(preserved,source){t.Fatal("source workbook was changed")}
 for _,name:=range []string{"xl/workbook.xml","xl/_rels/workbook.xml.rels","_rels/.rels"}{
  if !bytes.Equal(extractZipPart(t,source,name),extractZipPart(t,result,name)){t.Fatalf("unchanged workbook part was modified: %s",name)}
 }
 if !strings.Contains(string(extractZipPart(t,response.Data,"Translation_Audit.csv")),"REVIEW_MEANING_AND_CELL_OVERFLOW"){
  t.Fatal("complete XLSX audit absent")
 }
 dir:=t.TempDir();output:=filepath.Join(dir,"Budget_es.xlsx")
 if err=os.WriteFile(output,result,0600);err!=nil{t.Fatal(err)}
 soffice,err:=exec.LookPath("soffice");if err!=nil{t.Fatalf("LibreOffice required: %v",err)}
 officeCtx,cancel:=context.WithTimeout(ctx,60*time.Second);defer cancel()
 cmd:=exec.CommandContext(officeCtx,soffice,"-env:UserInstallation=file://"+filepath.ToSlash(filepath.Join(dir,"profile-xlsx")),"--headless","--convert-to","pdf","--outdir",dir,output)
 logs,err:=cmd.CombinedOutput()
 if err!=nil{t.Fatalf("LibreOffice rejected translated XLSX: %v: %s",err,logs)}
 pdfName:=filepath.Join(dir,"Budget_es.pdf")
 pdf,err:=os.Open(pdfName);if err!=nil{t.Fatalf("LibreOffice did not create PDF: %v; stdout=%s",err,logs)}
 defer pdf.Close()
 hdr:=make([]byte,5)
 if _,err=io.ReadFull(pdf,hdr);err!=nil||string(hdr)!="%PDF-"{t.Fatalf("invalid PDF: %v",err)}
 extracted,err:=exec.CommandContext(officeCtx,"pdftotext",pdfName,"-").Output()
 if err!=nil||len(strings.TrimSpace(string(extracted)))<5{t.Fatalf("rendered Excel PDF lacks searchable text: %v %s",err,extracted)}
 if strings.Contains(string(extracted),"Hello world."){t.Fatal("LibreOffice rendered untranslated text")}
 t.Log("PASS: real Argos XLSX translation, unchanged numeric/formula fields and native LibreOffice open/render to PDF")
}
