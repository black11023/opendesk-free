//go:build translation_integration

package preservetranslate

import (
 "bytes"
 "context"
 "strings"
 "testing"
 "opendesk-free/internal/engine"
)
func TestRealArgosDOCX(t *testing.T){
 ctx:=context.Background()
 status:=ProbeArgos(ctx)
 if !status.Installed{t.Fatalf("real Argos not installed: %+v",status)}
 if !strings.Contains(strings.Join(status.DirectPairs,","),"en->es"){t.Fatalf("real direct en->es model not available: %+v",status)}
 source:=docxFixture(t,`<w:p><w:r><w:t>Hello world.</w:t></w:r></w:p><w:p><w:r><w:t>Good morning.</w:t></w:r></w:p>`,nil)
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
}
