package engine

import (
 "bytes"
 "context"
 "encoding/csv"
 "strings"
 "testing"
)

func TestPDFAuditReportsEachSourceAndRejectsSpreadsheetInjection(t *testing.T){
 good:=twoPageFixture(t)
 good.Name="valid.pdf"
 invalid:=Input{Name:"corrupt.pdf",Data:[]byte("%PDF-corrupt")}
 unsafe:=Input{Name:"=HYPERLINK(123).pdf",Data:good.Data}
 wrong:=Input{Name:"renamed.txt",Data:good.Data}
 empty:=Input{Name:"empty.pdf"}
 output,e:=Run("pdf-audit",[]Input{good,invalid,unsafe,wrong,empty})
 if e!=nil{t.Fatal(e)}
 if output.Filename!="OpenDesk_PDF_Batch_Audit.csv"||!strings.HasPrefix(output.ContentType,"text/csv"){t.Fatal("bad report metadata")}
 if !bytes.HasPrefix(output.Data,[]byte{0xef,0xbb,0xbf}){t.Fatal("UTF-8 BOM missing")}
 r:=csv.NewReader(bytes.NewReader(output.Data[3:]))
 rows,e:=r.ReadAll();if e!=nil{t.Fatal(e)}
 if len(rows)!=6{t.Fatalf("expected header and five result rows, got %d",len(rows))}
 if rows[1][2]!="PASS"||rows[1][3]!="2"{t.Fatalf("valid PDF not inspected correctly: %v",rows[1])}
 if rows[2][2]!="FAIL"||rows[2][6]==""{t.Fatalf("corrupt PDF was not flagged: %v",rows[2])}
 if !strings.HasPrefix(rows[3][1],"'=")||rows[3][2]!="PASS"{t.Fatalf("unsafe CSV filename or valid source incorrectly handled: %v",rows[3])}
 if rows[4][2]!="FAIL"||rows[5][2]!="FAIL"{t.Fatalf("invalid name or empty file not flagged: %v / %v",rows[4],rows[5])}
 for i:=1;i<len(rows);i++{if len(rows[i][5])!=64{t.Fatalf("missing sha256 in row %d",i)}}
}
func TestPDFAuditHonorsCancellation(t *testing.T){
 ctx,cancel:=context.WithCancel(context.Background());cancel()
 _,e:=RunContext(ctx,"pdf-audit",[]Input{{Name:"a.pdf",Data:[]byte("%PDF-bad")}})
 if e==nil{t.Fatal("cancelled batch accepted")}
}
