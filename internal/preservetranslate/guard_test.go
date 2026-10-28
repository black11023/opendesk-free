package preservetranslate

import (
  "strings"
  "testing"
)

func TestProtectRestoreVietnameseNumericAndMaterialCodes(t *testing.T) {
  src:="Đơn giá 1.250.000 VND; mã PPR-DN20, ô A12; tiến độ 31/12/2026; 5.5%."
  protected, spans, err:=Protect(src)
  if err!=nil{t.Fatal(err)}
  if len(spans)<4{t.Fatalf("expected multiple protected spans, got %d",len(spans))}
  for _, s:=range spans {
    if !strings.Contains(protected,s.Marker){t.Errorf("missing marker %s",s.Marker)}
  }
  translated:=strings.Replace(protected,"Đơn giá","Unit price",1)
  out,err:=Restore(translated,spans)
  if err!=nil{t.Fatal(err)}
  if !strings.Contains(out,"1.250.000")||!strings.Contains(out,"PPR-DN20")||!strings.Contains(out,"31/12/2026"){
    t.Errorf("numeric/code tokens were not preserved: %s",out)
  }
  if src!="Đơn giá 1.250.000 VND; mã PPR-DN20, ô A12; tiến độ 31/12/2026; 5.5%."{t.Fatal("input unexpectedly mutated")}
}

func TestMarkerIntegrityRejectsMissingDuplicateReorderAndInjection(t *testing.T) {
  s,spans,e:=Protect("DN20; 2500")
  if e!=nil||len(spans)!=2{t.Fatalf("invalid setup %v %v",e,spans)}
  cases:=[]string{
    strings.Replace(s,spans[0].Marker,"",1),
    s+" "+spans[0].Marker,
    strings.Replace(strings.Replace(s,spans[0].Marker,"TMP",1),spans[1].Marker,spans[0].Marker,1),
    s+" __ODS_LOCK_9999__",
  }
  for _,c:=range cases {
    if _,err:=Restore(c,spans);err==nil{t.Fatalf("must reject invalid tokens: %s",c)}
  }
}
func TestMarkersCannotBeInjectedInSource(t *testing.T) {
  _,_,err:=Protect("text __ODS_LOCK_0001__")
  if err==nil{t.Fatal("expected marker injection rejection")}
}
func TestNoNumericSpansStillWorks(t *testing.T) {
  s,spans,e:=Protect("Xin chào thế giới")
  if e!=nil||s!="Xin chào thế giới"||len(spans)!=0{t.Fatalf("unexpected: %q %v %v",s,spans,e)}
  _,e=Restore("__ODS_LOCK_0001__",nil)
  if e==nil{t.Fatal("unexpected marker must fail")}
}
