package preservetranslate
import ("strings";"testing")
func good(id,from,to string) ModelPackage {return ModelPackage{ID:id,Source:from,Target:to,Installed:true,SHA256Verified:true,LicenseReviewed:true}}
func TestCatalogAndCopies(t *testing.T) {
 list:=Languages()
 if len(list)!=46 {t.Fatalf("expected 46 candidate languages, got %d",len(list))}
 seen:=map[string]bool{};for _,x:=range list{if x.Tag==""||x.EnglishName==""||x.VietnameseName==""||seen[x.Tag]{t.Fatalf("bad tag %+v",x)};seen[x.Tag]=true}
 for _,s:=range []string{"vi","en","ja","ko","fr","zh-Hans","zh-Hant","ar","pt-BR","ru","th"}{if !seen[s]{t.Fatalf("missing %s",s)}}
 list[0].Tag="CORRUPTED";if Languages()[0].Tag=="CORRUPTED"{t.Fatal("catalog mutated by caller")}
}
func TestDirectPreferredAndManyTargets(t *testing.T) {
 p:=[]ModelPackage{good("vi-en","vi","en"),good("en-ja","en","ja"),good("vi-fr","vi","fr"),good("en-fr","en","fr")}
 r,e:=PlanMultiTarget("VI",[]string{"fr","ja","en"},p,true)
 if e!=nil||len(r)!=3{t.Fatalf("%+v %v",r,e)}
 if r[0].ModelIDs[0]!="vi-fr"||len(r[0].Via)!=0{t.Fatal("not direct preferred")}
 if len(r[1].Via)!=1||r[1].Via[0]!="en"||!r[1].RequiresReview||len(r[1].ModelIDs)!=2{t.Fatal("unsafe silent pivot")}
 if r[2].Target!="en"||len(r[2].ModelIDs)!=1{t.Fatal("target order lost")}
}
func TestNoSilentPivotAndNoUnverifiedPack(t *testing.T){
 p:=[]ModelPackage{good("vi-en","vi","en"),good("en-ja","en","ja")}
 if _,e:=PlanMultiTarget("vi",[]string{"ja"},p,false);e==nil{t.Fatal("pivot without consent")}
 if _,e:=PlanMultiTarget("vi",[]string{"ja"},p,true);e!=nil{t.Fatal(e)}
 for k:=0;k<3;k++{m:=good("vi-en","vi","en");switch k{case 0:m.Installed=false;case 1:m.SHA256Verified=false;case 2:m.LicenseReviewed=false};if _,e:=PlanMultiTarget("vi",[]string{"en"},[]ModelPackage{m},false);e==nil{t.Fatalf("unverified model accepted: %+v",m)}}
 if r,e:=PlanMultiTarget("vi",[]string{"en","ko"},p,false);e==nil||r!=nil{t.Fatal("partial result escaped")}
}
func TestMixedDocumentKeepsAlreadyTargetLanguageUnchanged(t *testing.T){
 p:=[]ModelPackage{good("en-vi","en","vi")}
 r,e:=PlanMixedDocument([]Segment{{"a","en"},{"b","vi"}},[]string{"vi"},p,false)
 if e!=nil||len(r)!=2{t.Fatalf("%+v %v",r,e)}
 if r[0].Routes[0].Unchanged||r[0].Routes[0].ModelIDs[0]!="en-vi"{t.Fatal("English not routed")}
 if !r[1].Routes[0].Unchanged||len(r[1].Routes[0].ModelIDs)!=0{t.Fatal("Vietnamese must not be translated again")}
}
func TestScriptsAndMalformedConfigsFailClosed(t *testing.T){
 if _,e:=PlanMultiTarget("zh-Hant",[]string{"en"},[]ModelPackage{good("hans-en","zh-Hans","en")},false);e==nil{t.Fatal("simplified and traditional wrongly conflated")}
 p:=[]ModelPackage{good("vi-en","vi","en")}
 for _,source:=range []string{"auto","xx",""}{if _,e:=PlanMultiTarget(source,[]string{"en"},p,false);e==nil{t.Fatalf("source %s accepted",source)}}
 for _,target:=range [][]string{{"vi"},{"en","EN"},{},{"xyz"}}{if _,e:=PlanMultiTarget("vi",target,p,false);e==nil{t.Fatalf("bad targets accepted: %v",target)}}
 if _,e:=PlanMixedDocument([]Segment{{"x","vi"},{"x","en"}},[]string{"en"},p,false);e==nil{t.Fatal("duplicate segment ID")}
 if _,e:=PlanMixedDocument([]Segment{{"x","auto"}},[]string{"en"},p,false);e==nil||!strings.Contains(e.Error(),"segment x"){t.Fatal("auto source incorrectly accepted")}
}
