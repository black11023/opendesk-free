//go:build translation_integration

package preservetranslate

import (
 "bytes"
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
   t.Fatalf("installed direct Argos en->es model required: %+v",status)
 }
 sheet:=strings.Replace(spreadsheetSheet(t),"Hello 2026","Hello there.",1)
 labels:=`<sst xmlns="`+excelNS+`" count="3" uniqueCount="2"><si><t>Hello world.</t></si><si><t>Good morning.</t></si></sst>`
 source:=xlsxFixture(t,sheet,labels,nil)
 original:=append([]byte(nil),source...)
 response,e:=TranslateXLSX(ctx,[]engine.Input{{Name:"Budget.xlsx",Data:source}},"en","es",ArgosProvider{})
 if e!=nil{t.Fatalf("actual EN→ES Argos XLSX translation failed: %v",e)}
 result:=extractZipPart(t,response.Data,"Budget_es.xlsx")
 sst:=string(extractZipPart(t,result,"xl/sharedStrings.xml"))
 if strings.Contains(sst,"Hello world.")||strings.Contains(sst,"Good morning."){
  t.Fatalf("real installed Argos model did not translate shared Excel labels: %s",sst)
 }
 if !bytes.Equal(original,source){t.Fatal("original XLSX changed")}
 if !bytes.Equal(extractZipPart(t,result,"xl/styles.xml"),extractZipPart(t,source,"xl/styles.xml")){
  t.Fatal("Excel styles changed")
 }
 worksheet:=string(extractZipPart(t,result,"xl/worksheets/sheet1.xml"))
 if !strings.Contains(worksheet,"<f>B1*2</f><v>2500.5</v>")||!strings.Contains(worksheet,"<v>1250.25</v>"){
  t.Fatal("formula or numeric values changed")
 }
 if !strings.Contains(string(extractZipPart(t,response.Data,"Translation_Audit.csv")),"REVIEW_MEANING_AND_CELL_OVERFLOW"){
  t.Fatal("translated workbook lacks review audit")
 }
 dir:=t.TempDir()
 xlsxFile:=filepath.Join(dir,"Budget_es.xlsx")
 if e:=os.WriteFile(xlsxFile,result,0600);e!=nil{t.Fatal(e)}
 officeCtx,cancel:=context.WithTimeout(ctx,60*time.Second);defer cancel()
 soffice,e:=exec.LookPath("soffice");if e!=nil{t.Fatalf("LibreOffice required: %v",e)}
 cmd:=exec.CommandContext(officeCtx,soffice,"-env:UserInstallation=file://"+filepath.ToSlash(filepath.Join(dir,"profile-xlsx")),"--headless","--convert-to","pdf","--outdir",dir,xlsxFile)
 stdout,e:=cmd.CombinedOutput()
 if e!=nil{t.Fatalf("actual LibreOffice could not open translated XLSX: %v: %s",e,stdout)}
 renderedPDF:=filepath.Join(dir,"Budget_es.pdf")
 pdf,e:=os.Open(renderedPDF);if e!=nil{t.Fatalf("LibreOffice did not render translated XLSX: %v",e)}
 defer pdf.Close()
 buf:=make([]byte,5)
 if _,e=io.ReadFull(pdf,buf);e!=nil||string(buf)!="%PDF-"{t.Fatal("LibreOffice rendered an invalid PDF")}
 extracted,e:=exec.CommandContext(officeCtx,"pdftotext",renderedPDF,"-").Output()
 if e!=nil||len(strings.TrimSpace(string(extracted)))<8{t.Fatalf("Excel PDF is not searchable: %v %s",e,extracted)}
 if strings.Contains(string(extracted),"Hello world."){t.Fatal("LibreOffice still renders source label")}
 t.Log("PASS: real model XLSX inference, original unchanged, native formulas/styles, complete review audit and LibreOffice rendering")
}
