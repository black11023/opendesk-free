package office

import (
 "archive/zip"
 "bytes"
 "context"
 "encoding/csv"
 "errors"
 "io"
 "strings"
 "testing"

 "opendesk-free/internal/engine"
)
type flakyOffice struct{mock;failure int}
func(f *flakyOffice)Run(ctx context.Context,binary string,args []string)error{
 if f.calls+1==f.failure {f.calls++;return errors.New("unsupported content")}
 return f.mock.Run(ctx,binary,args)
}
func TestOfficeBatchKeepsSuccessesAndPerFileFailures(t *testing.T){
 runner:=&flakyOffice{failure:2}
 docs:=[]engine.Input{officeFile("ok.docx"),officeFile("broken.xlsx"),officeFile("=evil.pptx")}
 result,e:=BatchWith(context.Background(),runner,"soffice",docs)
 if e!=nil{t.Fatal(e)}
 if result.ContentType!="application/zip"{t.Fatal("wrong MIME")}
 zipFile,e:=zip.NewReader(bytes.NewReader(result.Data),int64(len(result.Data)));if e!=nil{t.Fatal(e)}
 if len(zipFile.File)!=3{t.Fatalf("expected 2 PDFs + CSV, got %d entries",len(zipFile.File))}
 var rows [][]string
 for _,entry:=range zipFile.File{
  file,e:=entry.Open();if e!=nil{t.Fatal(e)}
  data,e:=io.ReadAll(file);_ =file.Close();if e!=nil{t.Fatal(e)}
  if strings.HasSuffix(entry.Name,".csv"){
   if !bytes.HasPrefix(data,[]byte{0xef,0xbb,0xbf}){t.Fatal("missing CSV BOM")}
   rows,e=csv.NewReader(bytes.NewReader(data[3:])).ReadAll();if e!=nil{t.Fatal(e)}
  }else if !bytes.HasPrefix(data,[]byte("%PDF-")){t.Fatalf("invalid PDF: %s",entry.Name)}
 }
 if len(rows)!=4||rows[1][2]!="PASS"||rows[2][2]!="FAIL"||rows[3][2]!="PASS"{t.Fatalf("bad per-file status: %v",rows)}
 if !strings.HasPrefix(rows[3][1],"'="){t.Fatalf("Excel injection: %s",rows[3][1])}
 if rows[2][6]==""||rows[2][3]!=""{t.Fatalf("failed row needs an error and no output filename: %v",rows[2])}
 for _,row:=range rows[1:]{if len(row[5])!=64{t.Fatal("missing SHA256")}}
}
func TestBatchRejectsInvalidSizeAndCancellation(t *testing.T){
 ctx,cancel:=context.WithCancel(context.Background());cancel()
 if _,e:=BatchWith(ctx,&mock{},"soffice",[]engine.Input{officeFile("a.docx"),officeFile("b.docx")});!errors.Is(e,context.Canceled){t.Fatalf("expected cancellation, got %v",e)}
 if _,e:=BatchWith(context.Background(),&mock{},"soffice",[]engine.Input{officeFile("a.docx")});e==nil{t.Fatal("single-file batch accepted")}
}
func TestBatchAllFailuresStillReports(t *testing.T){
 docs:=[]engine.Input{{Name:"invalid.txt",Data:[]byte("fake fake")},{Name:"broken.docx",Data:[]byte("fake docx")}}
 result,e:=BatchWith(context.Background(),&mock{},"soffice",docs)
 if e!=nil{t.Fatal(e)}
 arc,e:=zip.NewReader(bytes.NewReader(result.Data),int64(len(result.Data)))
 if e!=nil||len(arc.File)!=1||arc.File[0].Name!="OpenDesk_Office_Batch_Report.csv"{t.Fatalf("failure report missing: %v",e)}
}
