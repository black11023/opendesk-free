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
