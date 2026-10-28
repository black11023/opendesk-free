package preservetranslate

// Multilingual planning only: no text translation and no document mutation.
// Listed languages are candidates; only verified locally installed packs
// allow a route to be planned. No network calls or model auto-downloads.

import (
 "errors"
 "fmt"
 "sort"
 "strings"
)

type Language struct { Tag, EnglishName, VietnameseName string }

const names = `ar|Arabic|Tiếng Ả Rập
az|Azerbaijani|Tiếng Azerbaijan
bn|Bengali|Tiếng Bengal
bg|Bulgarian|Tiếng Bulgaria
ca|Catalan|Tiếng Catalan
cs|Czech|Tiếng Séc
da|Danish|Tiếng Đan Mạch
nl|Dutch|Tiếng Hà Lan
en|English|Tiếng Anh
eo|Esperanto|Quốc tế ngữ
et|Estonian|Tiếng Estonia
fi|Finnish|Tiếng Phần Lan
fr|French|Tiếng Pháp
de|German|Tiếng Đức
el|Greek|Tiếng Hy Lạp
he|Hebrew|Tiếng Do Thái
hi|Hindi|Tiếng Hindi
hu|Hungarian|Tiếng Hungary
id|Indonesian|Tiếng Indonesia
ga|Irish|Tiếng Ireland
it|Italian|Tiếng Ý
ja|Japanese|Tiếng Nhật
ko|Korean|Tiếng Hàn
ky|Kyrgyz|Tiếng Kyrgyz
lv|Latvian|Tiếng Latvia
lt|Lithuanian|Tiếng Litva
ms|Malay|Tiếng Mã Lai
nb|Norwegian Bokmål|Tiếng Na Uy Bokmål
fa|Persian|Tiếng Ba Tư
pl|Polish|Tiếng Ba Lan
pt|Portuguese|Tiếng Bồ Đào Nha
pt-BR|Brazilian Portuguese|Tiếng Bồ Đào Nha (Brazil)
ro|Romanian|Tiếng Romania
ru|Russian|Tiếng Nga
sk|Slovak|Tiếng Slovakia
sl|Slovenian|Tiếng Slovenia
es|Spanish|Tiếng Tây Ban Nha
sv|Swedish|Tiếng Thụy Điển
tl|Tagalog|Tiếng Tagalog
th|Thai|Tiếng Thái
tr|Turkish|Tiếng Thổ Nhĩ Kỳ
uk|Ukrainian|Tiếng Ukraina
ur|Urdu|Tiếng Urdu
vi|Vietnamese|Tiếng Việt
zh-Hans|Simplified Chinese|Tiếng Trung (Giản thể)
zh-Hant|Traditional Chinese|Tiếng Trung (Phồn thể)`

var languageCatalog = func() []Language {
 var out []Language
 for _,line:=range strings.Split(names,"\n") {
   parts:=strings.Split(line,"|")
   if len(parts)==3 {out=append(out,Language{parts[0],parts[1],parts[2]})}
 }
 return out
}()

func Languages() []Language {return append([]Language(nil),languageCatalog...)}

func canonicalTag(tag string)(string,error) {
 tag=strings.TrimSpace(tag)
 if tag==""||strings.EqualFold(tag,"auto") {
   return "",errors.New("automatic language detection is not yet implemented; select a known source")
 }
 for _,l:=range languageCatalog {if strings.EqualFold(l.Tag,tag){return l.Tag,nil}}
 return "",fmt.Errorf("unsupported language tag %q; a listed language is not proof of an installed model",tag)
}

// ModelPackage is local verified-inventory metadata, not an advertising list.
// The future provider must prove Installed, SHA256Verified and LicenseReviewed.
type ModelPackage struct {
 ID, Source, Target string
 Installed, SHA256Verified, LicenseReviewed bool
}

type Route struct {
 Source, Target string
 ModelIDs []string
 Via []string
 RequiresReview bool
 Unchanged bool // For a segment already in a requested target language
 Warning string
}

// PlanMultiTarget is fail-closed: no partial batch, no silent pivot, no cloud.
// Source+target must differ. Up to 16 target documents are planned per batch.
func PlanMultiTarget(source string,targets []string,models []ModelPackage,allowPivot bool)([]Route,error){
 src,err:=canonicalTag(source);if err!=nil{return nil,err}
 if len(targets)<1||len(targets)>16 {return nil,errors.New("select 1 to 16 target languages")}
 type edge struct{to,id string}
 graph:=map[string][]edge{}
 for _,m:=range models {
   if m.ID==""||!m.Installed||!m.SHA256Verified||!m.LicenseReviewed {continue}
   from,e1:=canonicalTag(m.Source);to,e2:=canonicalTag(m.Target)
   if e1!=nil||e2!=nil||from==to {continue}
   graph[from]=append(graph[from],edge{to,m.ID})
 }
 for from:=range graph{sort.Slice(graph[from],func(i,j int)bool{return graph[from][i].id<graph[from][j].id})}
 direct:=func(a,b string)(string,bool) {
   for _,m:=range graph[a]{if m.to==b{return m.id,true}}
   return "",false
 }
 routes:=make([]Route,0,len(targets));seen:=map[string]bool{}
 for _,raw:=range targets {
   to,e:=canonicalTag(raw);if e!=nil{return nil,e}
   if to==src {return nil,fmt.Errorf("source and target are identical: %s",to)}
   if seen[to]{return nil,fmt.Errorf("duplicate target: %s",to)}
   seen[to]=true
   if id,ok:=direct(src,to);ok {
     routes=append(routes,Route{Source:src,Target:to,ModelIDs:[]string{id}})
     continue
   }
   if allowPivot&&src!="en"&&to!="en"{
     first,firstOK:=direct(src,"en")
     second,secondOK:=direct("en",to)
     if firstOK&&secondOK{
       routes=append(routes,Route{Source:src,Target:to,ModelIDs:[]string{first,second},Via:[]string{"en"},RequiresReview:true,Warning:"Translation via English may lose meaning; review required"})
       continue
     }
   }
   return nil,fmt.Errorf("BLOCKED: no verified offline translation route %s → %s",src,to)
 }
 return routes,nil
}

type Segment struct {ID, SourceLanguage string}
type SegmentLanguagePlan struct {SegmentID, Source string; Routes []Route}

// PlanMixedDocument accepts ALREADY IDENTIFIED source languages per segment.
// Auto-detection and reading actual DOCX/PDF/XLSX are NOT implemented here.
// Target-language segments are carried over unchanged, never retranslated.
func PlanMixedDocument(segments []Segment,targets []string,models []ModelPackage,allowPivot bool)([]SegmentLanguagePlan,error) {
 if len(segments)==0||len(segments)>50000{return nil,errors.New("expected 1 to 50000 identified segments")}
 if len(targets)==0||len(targets)>16{return nil,errors.New("select 1 to 16 target languages")}
 normalized:=make([]string,0,len(targets));seenTarget:=map[string]bool{}
 for _,raw:=range targets{
   tag,e:=canonicalTag(raw);if e!=nil{return nil,e}
   if seenTarget[tag]{return nil,fmt.Errorf("duplicate target: %s",tag)}
   seenTarget[tag]=true;normalized=append(normalized,tag)
 }
 out:=make([]SegmentLanguagePlan,0,len(segments));seen:=map[string]bool{}
 for _,s:=range segments {
   if s.ID==""||seen[s.ID]{return nil,fmt.Errorf("empty/duplicate segment identifier: %q",s.ID)}
   seen[s.ID]=true
   src,e:=canonicalTag(s.SourceLanguage);if e!=nil{return nil,fmt.Errorf("segment %s: %w",s.ID,e)}
   routes:=make([]Route,0,len(normalized))
   for _,to:=range normalized {
     if to==src{
       routes=append(routes,Route{Source:src,Target:to,Unchanged:true})
       continue
     }
     plan,e:=PlanMultiTarget(src,[]string{to},models,allowPivot)
     if e!=nil{return nil,fmt.Errorf("segment %s: %w",s.ID,e)}
     routes=append(routes,plan[0])
   }
   out=append(out,SegmentLanguagePlan{SegmentID:s.ID,Source:src,Routes:routes})
 }
 return out,nil
}
