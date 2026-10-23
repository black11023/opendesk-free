package office
import(
 "archive/zip"
 "bytes"
 "context"
 "errors"
 "os"
 "path/filepath"
 "strings"
 "testing"
 "opendesk-free/internal/engine"
)
type mock struct{calls int;args []string;temp []string;err error;missing bool;bad bool}
func(m *mock)Run(ctx context.Context,binary string,args []string)error{
 m.calls++;m.args=append([]string(nil),args...)
 in,outdir:=args[len(args)-1],args[len(args)-2];m.temp=append(m.temp,filepath.Dir(in))
 if m.err!=nil{return m.err};if m.missing{return nil}
 b,e:=os.ReadFile("../../tests/fixtures/pdf_two_pages.pdf");if e!=nil{return e}
 if m.bad{b=[]byte("invalid")}
 stem:=strings.TrimSuffix(filepath.Base(in),filepath.Ext(in))
 return os.WriteFile(filepath.Join(outdir,stem+".pdf"),b,0600)
}
func officeFile(n string)engine.Input{return engine.Input{Name:n,Data:[]byte{'P','K',3,4,0,0,0,0,1}}}
func TestPDFAndZIP(t *testing.T){
 for _,cases:=range [][]engine.Input{{officeFile("document.docx")},{officeFile("work.xlsx"),officeFile("slides.pptx")}}{
  runner:=&mock{};out,e:=ProcessWith(context.Background(),runner,"soffice",cases)
  if e!=nil{t.Fatal(e)};if runner.calls!=len(cases){t.Fatal("wrong number of conversions")}
  if len(cases)==1{if out.ContentType!="application/pdf"||!bytes.HasPrefix(out.Data,[]byte("%PDF-")){t.Fatal("single PDF invalid")}}else{
   z,e:=zip.NewReader(bytes.NewReader(out.Data),int64(len(out.Data)));if e!=nil||len(z.File)!=2{t.Fatalf("invalid ZIP: %v",e)}
  }
  if !strings.Contains(strings.Join(runner.args," "),"-env:UserInstallation=file:")||!strings.Contains(strings.Join(runner.args," "),"--headless"){t.Fatal("unsafe command flags")}
  for _,dir:=range runner.temp{if _,e:=os.Stat(dir);!os.IsNotExist(e){t.Fatal("temp leak")}}
 }
}
func TestFailClosed(t *testing.T){
 for _,f:=range []engine.Input{{Name:"bad.docx",Data:[]byte("not office")},{Name:"bad.txt",Data:[]byte{'P','K',3,4,0,0,0,0}},{Name:"old.doc",Data:[]byte{'P','K',3,4,0,0,0,0}}}{
  ex:=&mock{};if _,e:=ProcessWith(context.Background(),ex,"soffice",[]engine.Input{f});e==nil||ex.calls!=0{t.Fatalf("unsafe input %s",f.Name)}
 }
 for _,ex:=range []*mock{{err:errors.New("exit failure")},{missing:true},{bad:true}}{
  if _,e:=ProcessWith(context.Background(),ex,"soffice",[]engine.Input{officeFile("doc.docx")});e==nil{t.Fatal("missing/invalid output accepted")}
  for _,dir:=range ex.temp{if _,e:=os.Stat(dir);!os.IsNotExist(e){t.Fatal("temp not removed")}}
 }
 ctx,cancel:=context.WithCancel(context.Background());cancel()
 if _,e:=ProcessWith(ctx,&mock{},"soffice",[]engine.Input{officeFile("a.docx")});e==nil{t.Fatal("cancelled request accepted")}
}
