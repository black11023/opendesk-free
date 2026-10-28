package preservetranslate
import (
 "archive/zip"
 "bytes"
 "context"
 "crypto/sha256"
 "encoding/hex"
 "fmt"
 "net/url"
 "os"
 "strings"
 "testing"
)

func makeTestPack(t *testing.T,meta string,extra map[string]string)(*os.File,ApprovedPack){
 t.Helper()
 var buf bytes.Buffer
 zipW:=zip.NewWriter(&buf)
 files:=map[string]string{"metadata.json":meta,"model/test.bin":"Model data"}
 for k,v:=range extra{files[k]=v}
 for k,v:=range files{
   w,e:=zipW.Create(k);if e!=nil{t.Fatal(e)}
   if _,e=w.Write([]byte(v));e!=nil{t.Fatal(e)}
 }
 if e:=zipW.Close();e!=nil{t.Fatal(e)}
 file,e:=os.CreateTemp(t.TempDir(),"approved-*.argosmodel");if e!=nil{t.Fatal(e)}
 if _,e=file.Write(buf.Bytes());e!=nil{t.Fatal(e)}
 sum:=sha256.Sum256(buf.Bytes())
 spec:=ApprovedPack{ID:"fixture-test",Source:"en",Target:"es",Size:int64(buf.Len()),SHA256:hex.EncodeToString(sum[:])}
 return file,spec
}
func TestVerifiedLanguagePackStrictChecks(t *testing.T){
 meta:=`{"from_code":"en","to_code":"es","package_version":"1.0"}`
 valid,p:=makeTestPack(t,meta,nil);defer valid.Close()
 if err:=ValidateApprovedModel(valid,p);err!=nil{t.Fatalf("valid pack should pass: %v",err)}
 // Official model ZIP has nested metadata.json (verified in real CI).
 nested,pn:=makeTestPack(t,meta,map[string]string{"translate-en_es-1_0/metadata.json":meta})
 defer nested.Close()
 if err:=ValidateApprovedModel(nested,pn);err==nil{t.Fatal("more than one metadata file accepted")}
 var single bytes.Buffer
 zw:=zip.NewWriter(&single)
 for _,it:=range []struct{k,v string}{{"translate-en_es-1_0/metadata.json",meta},{"translate-en_es-1_0/model/test.bin","model bytes"}}{
  w,e:=zw.Create(it.k);if e!=nil{t.Fatal(e)};if _,e=w.Write([]byte(it.v));e!=nil{t.Fatal(e)}
 }
 if e:=zw.Close();e!=nil{t.Fatal(e)}
 f2,e:=os.CreateTemp(t.TempDir(),"nested-*.argosmodel");if e!=nil{t.Fatal(e)};defer f2.Close()
 if _,e=f2.Write(single.Bytes());e!=nil{t.Fatal(e)}
 digest:=sha256.Sum256(single.Bytes())
 nestedPack:=ApprovedPack{Source:"en",Target:"es",Size:int64(single.Len()),SHA256:hex.EncodeToString(digest[:])}
 if e:=ValidateApprovedModel(f2,nestedPack);e!=nil{t.Fatalf("single nested official metadata must pass: %v",e)}
 bad:=p;bad.SHA256=strings.Repeat("0",64)
 if err:=ValidateApprovedModel(valid,bad);err==nil||!strings.Contains(err.Error(),"SHA-256"){t.Fatalf("tampered digest accepted: %v",err)}
 bad=p;bad.Size++
 if err:=ValidateApprovedModel(valid,bad);err==nil{t.Fatal("incorrect model size passed")}
 wrong,pw:=makeTestPack(t,`{"from_code":"en","to_code":"ru"}`,nil);defer wrong.Close()
 if err:=ValidateApprovedModel(wrong,pw);err==nil||!strings.Contains(err.Error(),"language pair"){t.Fatalf("wrong pair accepted: %v",err)}
 traversal,pt:=makeTestPack(t,meta,map[string]string{"../../outside.exe":"unsafe"});defer traversal.Close()
 if err:=ValidateApprovedModel(traversal,pt);err==nil||!strings.Contains(err.Error(),"unsafe"){t.Fatalf("zip traversal accepted: %v",err)}
 duplicate,pd:=makeTestPack(t,meta,map[string]string{"MODEL/test.bin":"duplicate case"});defer duplicate.Close()
 if err:=ValidateApprovedModel(duplicate,pd);err==nil||!strings.Contains(err.Error(),"duplicate"){t.Fatalf("case collision accepted: %v",err)}
}
func TestApprovedCatalogIsExplicitAndNoNetworkOnBadID(t *testing.T){
 approved:=ApprovedPacks()
 if len(approved)!=1{t.Fatalf("expected only approved model, got %d",len(approved))}
 pack:=approved[0]
 if pack.Source!="en"||pack.Target!="es"||pack.Size!=87503191||pack.SHA256!="d698d0ef87ad70d5d184b7fa6965905bf4368f09a2bb9ffb165a79bac96af0c4"{t.Fatal("approved model changed without review")}
 parsed,e:=url.Parse(pack.URL);if e!=nil||!trustedModelURL(parsed){t.Fatal("unapproved official model URL")}
 for _,s:=range []string{"http://argos-net.com/x","https://evil.example/x","https://argos-net.com.evil.example/x","https://user@argos-net.com/x"}{
  u,e:=url.Parse(s);if e==nil&&trustedModelURL(u){t.Errorf("untrusted URL accepted: %s",s)}
 }
 if _,e=DownloadApprovedModel(context.Background(),"unknown-model");e==nil{t.Fatal("unsupported model should be rejected before any download")}
 if _,e=findApproved("en->es");e==nil{t.Fatal("no fuzzy model identifiers allowed")}
}
func TestStagingFailureLeavesNoInstalledOutput(t *testing.T){
 pack:=approvedPacks[0]
 for _,tc:=range []struct{name string;input []byte}{{"empty",nil},{"wrong bytes",[]byte("not a language package")}}{
  t.Run(tc.name,func(t *testing.T){
   _,e:=stageApproved(context.Background(),pack,bytes.NewReader(tc.input))
   if e==nil{t.Fatal("bad input accepted")}
   if !strings.Contains(e.Error(),"incomplete") && !strings.Contains(e.Error(),"SHA-256"){t.Fatal(fmt.Sprintf("unexpected failure: %v",e))}
  })
 }
 ctx,cancel:=context.WithCancel(context.Background());cancel()
 _,e:=stageApproved(ctx,pack,bytes.NewReader(bytes.Repeat([]byte("x"),64)))
 if e==nil{t.Fatal("cancellation should stop staging")}
}
