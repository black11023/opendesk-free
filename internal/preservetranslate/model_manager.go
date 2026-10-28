package preservetranslate

// Approved language models are never bundled, silently downloaded or run
// from untrusted ZIP content. A user must explicitly request an install;
// the exact archive is verified against an immutable SHA-256 allowlist first.
import (
 "archive/zip"
 "bytes"
 "context"
 "crypto/sha256"
 "crypto/subtle"
 "encoding/hex"
 "encoding/json"
 "errors"
 "fmt"
 "io"
 "net/http"
 "net/url"
 "os"
 "path"
 "strings"
 "time"
 "opendesk-free/internal/procrun"
)

type ApprovedPack struct {
 ID string `json:"id"`
 Name string `json:"name"`
 Source string `json:"source"`
 Target string `json:"target"`
 Version string `json:"version"`
 Size int64 `json:"size_bytes"`
 SHA256 string `json:"sha256"`
 URL string `json:"url"`
 LicenseNotice string `json:"license_notice"`
}
var approvedPacks=[]ApprovedPack{
 {ID:"argos-en-es-1.0",Name:"English → Spanish (Argos)",Source:"en",Target:"es",Version:"1.0",
  Size:87503191,SHA256:"d698d0ef87ad70d5d184b7fa6965905bf4368f09a2bb9ffb165a79bac96af0c4",
  URL:"https://argos-net.com/v1/translate-en_es-1_0.argosmodel",
  LicenseNotice:"Third-party model licensing must be reviewed independently. Model is not bundled or redistributed by OpenDesk."},
}
type PackReceipt struct {
 Status string `json:"status"`
 ID string `json:"id"`
 SHA256 string `json:"sha256"`
 Pair string `json:"pair"`
 Detail string `json:"detail"`
}
const maxModelParts=300
const maxModelExpanded=int64(800<<20)
const maxModelPart=int64(450<<20)
func ApprovedPacks()[]ApprovedPack {return append([]ApprovedPack(nil),approvedPacks...)}
func findApproved(id string)(ApprovedPack,error){
 for _,p:=range approvedPacks {if p.ID==id {return p,nil}}
 return ApprovedPack{},errors.New("unsupported or unverified language model; only listed SHA-256 pinned packages are accepted")
}
func ValidateApprovedModel(file *os.File,p ApprovedPack)error {
 stat,err:=file.Stat();if err!=nil{return err}
 if stat.Size()!=p.Size{return fmt.Errorf("BLOCKED: model size %d differs from approved %d bytes",stat.Size(),p.Size)}
 if _,err=file.Seek(0,io.SeekStart);err!=nil{return err}
 sum:=sha256.New()
 if _,err=io.Copy(sum,file);err!=nil{return err}
 want,err:=hex.DecodeString(p.SHA256);if err!=nil{return err}
 if subtle.ConstantTimeCompare(sum.Sum(nil),want)!=1{return errors.New("BLOCKED: model SHA-256 does not match the approved original")}
 z,err:=zip.NewReader(file,stat.Size());if err!=nil{return fmt.Errorf("BLOCKED: invalid ZIP: %w",err)}
 if len(z.File)<2||len(z.File)>maxModelParts{return errors.New("BLOCKED: invalid model package part count")}
 seen:=map[string]bool{}
 var expanded int64
 foundMeta:=false
 for _,part:=range z.File{
  n:=strings.ReplaceAll(part.Name,"\\","/")
  clean:=path.Clean(strings.TrimSuffix(n,"/"))
  if n==""||strings.HasPrefix(n,"/")||clean=="."||clean==".."||strings.HasPrefix(clean,"../")||
   clean!=strings.TrimSuffix(n,"/")||strings.Contains(n,":")||len(n)>250||
   part.Mode()&os.ModeSymlink!=0 {return errors.New("BLOCKED: unsafe model ZIP entry")}
  key:=strings.ToLower(clean)
  if seen[key]{return errors.New("BLOCKED: duplicate model ZIP entry")}
  seen[key]=true
  if part.UncompressedSize64>uint64(maxModelPart) {return errors.New("BLOCKED: oversized model component")}
  expanded+=int64(part.UncompressedSize64)
  if expanded>maxModelExpanded{return errors.New("BLOCKED: model expansion exceeds limit")}
  if part.UncompressedSize64>1<<20&&part.UncompressedSize64>800*(part.CompressedSize64+1){
    return errors.New("BLOCKED: suspicious compression ratio")
  }
  rd,err:=part.Open();if err!=nil{return fmt.Errorf("BLOCKED: model ZIP read error: %w",err)}
  if part.Name=="metadata.json"{
    var meta struct {From string `json:"from_code"`;To string `json:"to_code"`}
    body,readErr:=io.ReadAll(io.LimitReader(rd,64<<10))
    closeErr:=rd.Close()
    if readErr!=nil||closeErr!=nil{return errors.New("BLOCKED: invalid model metadata CRC")}
    if err=json.Unmarshal(body,&meta);err!=nil||meta.From!=p.Source||meta.To!=p.Target{return errors.New("BLOCKED: unapproved model language pair")}
    foundMeta=true
    continue
  }
  // Check CRC on every part, not just the manifest.
  count,readErr:=io.Copy(io.Discard,io.LimitReader(rd,maxModelPart+1))
  closeErr:=rd.Close()
  if readErr!=nil||closeErr!=nil||count>maxModelPart{return fmt.Errorf("BLOCKED: corrupt model component %s",part.Name)}
 }
 if !foundMeta{return errors.New("BLOCKED: missing root model metadata")}
 return nil
}
func stageApproved(ctx context.Context,p ApprovedPack,src io.Reader)(*os.File,error){
 f,err:=os.CreateTemp("","opendesk-verified-*.argosmodel");if err!=nil{return nil,err}
 ok:=false
 defer func(){if !ok{f.Close();os.Remove(f.Name())}}()
 written,err:=io.Copy(f,io.LimitReader(&cancelReader{ctx:ctx,source:src},p.Size+1))
 if err!=nil{return nil,err}
 if written!=p.Size{return nil,fmt.Errorf("BLOCKED: incomplete or oversized model (%d/%d bytes)",written,p.Size)}
 if err=ValidateApprovedModel(f,p);err!=nil{return nil,err}
 if _,err=f.Seek(0,io.SeekStart);err!=nil{return nil,err}
 ok=true
 return f,nil
}
type cancelReader struct {ctx context.Context;source io.Reader}
func (c *cancelReader)Read(p []byte)(int,error){
 if err:=c.ctx.Err();err!=nil{return 0,err}
 return c.source.Read(p)
}
func trustedModelURL(u *url.URL)bool{
 return u.Scheme=="https"&&u.Host=="argos-net.com"&&u.User==nil&&u.Fragment==""
}
func verifiedHTTPClient()*http.Client{
 return &http.Client{Timeout:14*time.Minute,CheckRedirect:func(req *http.Request,via []*http.Request)error{
   if len(via)>5||!trustedModelURL(req.URL){return errors.New("BLOCKED: model redirect to unapproved origin")}
   return nil
 }}
}
func DownloadApprovedModel(ctx context.Context,id string)(PackReceipt,error){
 p,err:=findApproved(id);if err!=nil{return PackReceipt{},err}
 if !ProbeArgos(ctx).Installed{return PackReceipt{},errors.New("Python + Argos runtime missing; model download cannot be installed yet")}
 u,err:=url.Parse(p.URL);if err!=nil||!trustedModelURL(u){return PackReceipt{},errors.New("unapproved model download URL")}
 req,err:=http.NewRequestWithContext(ctx,http.MethodGet,p.URL,nil);if err!=nil{return PackReceipt{},err}
 resp,err:=verifiedHTTPClient().Do(req);if err!=nil{return PackReceipt{},err}
 defer resp.Body.Close()
 if resp.StatusCode!=http.StatusOK{return PackReceipt{},fmt.Errorf("model vendor responded HTTP %d",resp.StatusCode)}
 if resp.ContentLength>p.Size{return PackReceipt{},errors.New("BLOCKED: model vendor reported oversized file")}
 return InstallApprovedModel(ctx,id,resp.Body)
}
func InstallApprovedModel(ctx context.Context,id string,data io.Reader)(PackReceipt,error){
 p,err:=findApproved(id);if err!=nil{return PackReceipt{},err}
 if !ProbeArgos(ctx).Installed{return PackReceipt{},errors.New("Argos Translate runtime not ready; install the optional Python engine before selecting model packages")}
 f,err:=stageApproved(ctx,p,data);if err!=nil{return PackReceipt{},err}
 name:=f.Name()
 defer os.Remove(name)
 if err=f.Close();err!=nil{return PackReceipt{},err}
 result,err:=installWithArgos(ctx,name,p)
 if err!=nil{return PackReceipt{},err}
 return PackReceipt{Status:"READY",ID:p.ID,SHA256:p.SHA256,Pair:p.Source+"->"+p.Target,Detail:result},nil
}
func installWithArgos(ctx context.Context,modelFile string,p ApprovedPack)(string,error){
 python,err:=FindArgosPython();if err!=nil{return "",err}
 step,cancel:=context.WithTimeout(ctx,4*time.Minute);defer cancel()
 // Use installed Argos API only. No subprocess shell, pip, or network call.
 script:=`import json,sys
try:
 import argostranslate.package as pkg
 import argostranslate.translate as tr
 file,source,target=sys.argv[1:4]
 installed=pkg.get_installed_packages()
 if not any(p.from_code==source and p.to_code==target for p in installed):
  pkg.install_from_path(file)
  installed=pkg.get_installed_packages()
 if not any(p.from_code==source and p.to_code==target for p in installed):
  raise RuntimeError("Argos model did not become installed")
 langs=tr.get_installed_languages()
 a=next((x for x in langs if x.code==source),None)
 b=next((x for x in langs if x.code==target),None)
 if a is None or b is None: raise RuntimeError("Argos installed language missing")
 trans=a.get_translation(b)
 if trans is None or not trans.translate("Hello world.").strip():
  raise RuntimeError("Argos inference smoke check failed")
 print("Installed and tested direct offline "+source+"->"+target)
except Exception as ex:
 print("BLOCKED: "+str(ex)[:220])
 sys.exit(1)
`
 cmd:=procrun.Command(step,python,"-I","-c",script,modelFile,p.Source,p.Target)
 output:=&boundedBuffer{Max:8<<10}
 stderr:=&boundedBuffer{Max:8<<10}
 cmd.Stdout=output
 cmd.Stderr=stderr
 if err=cmd.Run();err!=nil{
  if step.Err()!=nil{return "",step.Err()}
  return "",fmt.Errorf("Argos install or local inference failed: %s",strings.TrimSpace(output.String()))
 }
 if output.Len()>1000{return "",errors.New("unexpected model installer output")}
 if !strings.HasPrefix(output.String(),"Installed and tested direct offline "){return "",errors.New("model installer did not confirm inference")}
 return strings.TrimSpace(output.String()),nil
}
