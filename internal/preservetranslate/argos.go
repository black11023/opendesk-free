package preservetranslate

// Optional strictly local Argos Translate worker. No internet, shell,
// model download or install. The user supplies Python+Argos+direct model.
import (
 "bytes"
 "context"
 "encoding/json"
 "errors"
 "fmt"
 "os/exec"
 "strings"
 "time"
 "opendesk-free/internal/procrun"
)
type ArgosProvider struct {Python string}
type ArgosStatus struct {
 Installed bool `json:"installed"`
 Detail string `json:"detail"`
 DirectPairs []string `json:"direct_pairs"`
}
type workerRequest struct {
 Command string `json:"command"`
 Source string `json:"source"`
 Target string `json:"target"`
 Segments []string `json:"segments"`
}
type workerResponse struct {
 Error string `json:"error"`
 DirectPairs []string `json:"direct_pairs"`
 Segments []string `json:"segments"`
}
const argosWorker = `
import json,sys
try:
    import argostranslate.package as pkg
    import argostranslate.translate as tr
    req=json.load(sys.stdin)
    installed=pkg.get_installed_packages()
    pairs=sorted({p.from_code+"->"+p.to_code for p in installed})
    result={"direct_pairs":pairs}
    if req.get("command")=="translate":
        source=req.get("source")
        target=req.get("target")
        if (source+"->"+target) not in pairs:
            raise RuntimeError("Direct offline Argos language model is not installed")
        languages=tr.get_installed_languages()
        a=next((l for l in languages if l.code==source),None)
        b=next((l for l in languages if l.code==target),None)
        if a is None or b is None:raise RuntimeError("Installed model languages not initialized")
        # Only choose an installed direct package, checked above. Argos
        # Language.get_translation() is its stable public API; some versions
        # do not expose a Language.translations attribute.
        direct=a.get_translation(b)
        if direct is None:raise RuntimeError("No direct translation model")
        segments=req.get("segments")
        if not isinstance(segments,list) or len(segments)>600:
            raise RuntimeError("Invalid segment count")
        result["segments"]=[direct.translate(x) for x in segments]
    print(json.dumps(result,ensure_ascii=False))
except Exception as ex:
    print(json.dumps({"error":str(ex)[:250]},ensure_ascii=False))
    sys.exit(1)
`
func FindArgosPython()(string,error){
 for _,name:=range []string{"python","python3"} {
   p,e:=exec.LookPath(name);if e==nil{return p,nil}
 }
 return "",errors.New("Python and offline Argos Translate are not installed")
}
func (p ArgosProvider)call(ctx context.Context,r workerRequest)(workerResponse,error){
 binary:=p.Python
 if binary==""{var e error;binary,e=FindArgosPython();if e!=nil{return workerResponse{},e}}
 input,e:=json.Marshal(r);if e!=nil{return workerResponse{},e}
 if len(input)>1<<20{return workerResponse{},errors.New("translation input too large")}
 step,cancel:=context.WithTimeout(ctx,2*time.Minute);defer cancel()
 cmd:=procrun.Command(step,binary,"-I","-c",argosWorker)
 cmd.Stdin=bytes.NewReader(input)
 var output,errorsOutput bytes.Buffer
 cmd.Stdout=&output;cmd.Stderr=&errorsOutput
 if e=cmd.Run();e!=nil{
   if step.Err()!=nil{return workerResponse{},fmt.Errorf("offline translation cancelled/timed out: %w",step.Err())}
   var failure workerResponse
   if json.Unmarshal(output.Bytes(),&failure)==nil&&failure.Error!=""{return workerResponse{},fmt.Errorf("Argos: %s",failure.Error)}
   return workerResponse{},fmt.Errorf("local Argos Python worker failed: %v",e)
 }
 if output.Len()>2<<20{return workerResponse{},errors.New("worker output too large")}
 var result workerResponse
 if e=json.Unmarshal(output.Bytes(),&result);e!=nil{return workerResponse{},fmt.Errorf("invalid offline model response: %w",e)}
 if result.Error!=""{return workerResponse{},errors.New(result.Error)}
 return result,nil
}
func (p ArgosProvider)Translate(ctx context.Context,from,to string,segments []string)([]string,error){
 result,e:=p.call(ctx,workerRequest{Command:"translate",Source:from,Target:to,Segments:segments})
 if e!=nil{return nil,e}
 if len(result.Segments)!=len(segments){return nil,errors.New("Argos returned incomplete paragraphs")}
 return result.Segments,nil
}
func ProbeArgos(ctx context.Context)ArgosStatus{
 python,e:=FindArgosPython()
 if e!=nil{return ArgosStatus{Detail:"Optional Python + offline Argos Translate not installed"}}
 step,cancel:=context.WithTimeout(ctx,10*time.Second);defer cancel()
 result,e:=(ArgosProvider{Python:python}).call(step,workerRequest{Command:"probe"})
 if e!=nil{return ArgosStatus{Detail:"Argos cannot be loaded. Install optional package and a direct language model."}}
 pairs:=make([]string,0,len(result.DirectPairs))
 for _,pair:=range result.DirectPairs{
   lang:=strings.Split(pair,"->");if len(lang)!=2{continue}
   if _,e:=canonicalTag(lang[0]);e!=nil{continue}
   if _,e:=canonicalTag(lang[1]);e!=nil{continue}
   pairs=append(pairs,pair)
 }
 detail:="Argos ready with direct models. Translation quality and layout require review."
 if len(pairs)==0{detail="Argos installed but no supported direct language model found"}
 return ArgosStatus{Installed:true,Detail:detail,DirectPairs:pairs}
}
