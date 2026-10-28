package preservetranslate

// XLSX translation intentionally changes only plain shared strings and
// inline string cells. Values, formulas, styles, sheet structure, merges,
// named ranges and unknown OOXML parts remain byte-for-byte unchanged.
import(
 "archive/zip"
 "bytes"
 "context"
 "crypto/sha256"
 "encoding/csv"
 "encoding/hex"
 "encoding/xml"
 "errors"
 "fmt"
 "io"
 "path"
 "regexp"
 "sort"
 "strconv"
 "strings"
 "unicode/utf8"
 "opendesk-free/internal/engine"
)
const maxExcelBytes=12<<20
const maxExcelSegments=600
const maxExcelChars=125000
var excelSI=regexp.MustCompile(`(?s)<si(?:\s[^>]*)?>(.*?)</si>`)
var excelT=regexp.MustCompile(`(?s)<t(?:\s[^>]*)?>(.*?)</t>`)
var excelCell=regexp.MustCompile(`(?s)<c(?:\s[^>]*)?>.*?</c>`)
var excelV=regexp.MustCompile(`(?s)<v>([0-9]+)</v>`)
var excelAttr=regexp.MustCompile(`([A-Za-z]+)="([^"]*)"`)
var safeRef=regexp.MustCompile(`^[A-Z]{1,3}[1-9][0-9]{0,6}$`)
type excelSegment struct {
 Part string
 Ref string
 Old string
 Start,End int
 SharedIndex int
}
func officeTextValue(value []byte)(string,error){
 var d struct{Value string `xml:",chardata"`}
 if e:=xml.Unmarshal(append(append([]byte("<root>"),value...),[]byte("</root>")...),&d);e!=nil{return "",e}
 return d.Value,nil
}
func xmlTextBounds(doc []byte,offset int)(start,end int,text string,err error){
 matches:=excelT.FindAllSubmatchIndex(doc,-1)
 if len(matches)!=1{return 0,0,"",errors.New("BLOCKED: only a single plain text element is supported")}
 loc:=matches[0]
 raw:=doc[loc[2]:loc[3]]
 if bytes.Contains(raw,[]byte("<")){return 0,0,"",errors.New("BLOCKED: nested XML in spreadsheet string")}
 value,e:=officeTextValue(raw);if e!=nil{return 0,0,"",e}
 if len(value)>4000{return 0,0,"",errors.New("BLOCKED: spreadsheet text exceeds 4000 bytes")}
 return offset+loc[2],offset+loc[3],value,nil
}
func plainXLSX(part string,data []byte)error{
 if e:=checkXML(data);e!=nil{return fmt.Errorf("BLOCKED: invalid XML %s: %w",part,e)}
 for _,unsafe:=range [][]byte{[]byte("<!--"),[]byte("<![CDATA["),[]byte("<!DOCTYPE"),[]byte("<!ENTITY"),[]byte("<mc:AlternateContent")}{
  if bytes.Contains(data,unsafe){return fmt.Errorf("BLOCKED: unsupported XML comments/content in %s",part)}
 }
 return nil
}
func excelCells(part string,data []byte)([]excelSegment,map[int][]string,error){
 if err:=plainXLSX(part,data);err!=nil{return nil,nil,err}
 if !bytes.Contains(data,[]byte(`xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"`)){
  return nil,nil,errors.New("BLOCKED: unsupported SpreadsheetML namespace")
 }
 // These containers can carry user-visible strings outside supported cells.
 for _,tag:=range []string{"<headerFooter","<dataValidations","<tableParts","<controls","<extLst","<legacyDrawing","<drawing"}{
  if bytes.Contains(data,[]byte(tag)){return nil,nil,fmt.Errorf("BLOCKED: %s has unsupported text-bearing structure %s",part,tag)}
 }
 found:=map[string]bool{}
 refs:=map[int][]string{}
 segments:=[]excelSegment{}
 cells:=excelCell.FindAllIndex(data,-1)
 seenInline:=0
 for _,loc:=range cells{
  cell:=data[loc[0]:loc[1]]
  tagEnd:=bytes.IndexByte(cell,'>');if tagEnd<0{return nil,nil,errors.New("BLOCKED: invalid Excel cell")}
  attrs:=map[string]string{}
  for _,a:=range excelAttr.FindAllSubmatch(cell[:tagEnd],-1){attrs[string(a[1])]=string(a[2])}
  ref:=attrs["r"]
  if !safeRef.MatchString(ref){return nil,nil,fmt.Errorf("BLOCKED: invalid or absent Excel cell reference %s",ref)}
  if found[ref]{return nil,nil,fmt.Errorf("BLOCKED: duplicate cell reference %s",ref)}
  found[ref]=true
  switch attrs["t"]{
  case "s":
   values:=excelV.FindAllSubmatch(cell,-1)
   if len(values)!=1{return nil,nil,fmt.Errorf("BLOCKED: shared string index missing at %s",ref)}
   index,e:=strconv.Atoi(string(values[0][1]));if e!=nil||index<0{return nil,nil,errors.New("BLOCKED: invalid shared string index")}
   refs[index]=append(refs[index],part+"!"+ref)
  case "inlineStr":
   if bytes.Contains(cell,[]byte("<f"))||bytes.Contains(cell,[]byte("<r>"))||bytes.Contains(cell,[]byte("<r "))||
      bytes.Count(cell,[]byte("<is>"))!=1{return nil,nil,fmt.Errorf("BLOCKED: complex inline Excel string at %s",ref)}
   inner:=excelT.FindAllSubmatchIndex(cell,-1)
   if len(inner)!=1{return nil,nil,fmt.Errorf("BLOCKED: unsupported inline string at %s",ref)}
   seenInline++
   start,end,old,e:=xmlTextBounds(cell,loc[0]);if e!=nil{return nil,nil,fmt.Errorf("%s!%s: %w",part,ref,e)}
   if strings.TrimSpace(old)!=""{
    segments=append(segments,excelSegment{Part:part,Ref:ref,Old:old,Start:start,End:end,SharedIndex:-1})
   }
  }
 }
 if seenInline!=len(excelT.FindAllIndex(data,-1)){
  return nil,nil,fmt.Errorf("BLOCKED: worksheet has unsupported free text: %s",part)
 }
 return segments,refs,nil
}
func sharedStrings(data []byte,refs map[int][]string)([]excelSegment,error){
 if err:=plainXLSX("xl/sharedStrings.xml",data);err!=nil{return nil,err}
 if !bytes.Contains(data,[]byte(`xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"`)){
  return nil,errors.New("BLOCKED: unsupported shared string namespace")
 }
 sis:=excelSI.FindAllIndex(data,-1)
 for idx:=range refs{if idx>=len(sis){return nil,fmt.Errorf("BLOCKED: workbook references invalid shared string index %d",idx)}}
 segments:=[]excelSegment{}
 foundT:=0
 for index,loc:=range sis{
  si:=data[loc[0]:loc[1]]
  matches:=excelT.FindAllSubmatchIndex(si,-1)
  foundT+=len(matches)
  if len(matches)!=1||bytes.Contains(si,[]byte("<r>"))||bytes.Contains(si,[]byte("<r "))||
      bytes.Contains(si,[]byte("<rPh"))||bytes.Contains(si,[]byte("<phoneticPr")){
   return nil,fmt.Errorf("BLOCKED: shared string %d has complex formatting",index)
  }
  start,end,original,err:=xmlTextBounds(si,loc[0]);if err!=nil{return nil,err}
  if _,used:=refs[index];used&&strings.TrimSpace(original)!=""{
   segments=append(segments,excelSegment{Part:"xl/sharedStrings.xml",Ref:strings.Join(refs[index],"; "),Old:original,Start:start,End:end,SharedIndex:index})
  }
 }
 if foundT!=len(excelT.FindAllIndex(data,-1)){return nil,errors.New("BLOCKED: shared strings outside ordinary si records")}
 return segments,nil
}
func TranslateXLSX(ctx context.Context,files []engine.Input,from,to string,provider TextTranslator)(engine.Result,error){
 if len(files)!=1{return engine.Result{},errors.New("choose exactly one XLSX workbook")}
 from,err:=canonicalTag(from);if err!=nil{return engine.Result{},err}
 to,err=canonicalTag(to);if err!=nil{return engine.Result{},err}
 if from==to{return engine.Result{},errors.New("source/target language must differ")}
 f:=files[0]
 if strings.ToLower(path.Ext(f.Name))!=".xlsx"||len(f.Data)<100||len(f.Data)>maxExcelBytes{
  return engine.Result{},errors.New("select a valid XLSX up to 12 MB")
 }
 gate:=inspectOOXML(ctx,f)
 if gate.Status!="READY"{return engine.Result{},fmt.Errorf("BLOCKED by FormGuard (%s): %s",gate.Status,gate.Findings)}
 z,err:=zip.NewReader(bytes.NewReader(f.Data),int64(len(f.Data)));if err!=nil{return engine.Result{},err}
 docs:=map[string][]byte{}
 sheets:=[]string{}
 for _,part:=range z.File{
  if err:=ctx.Err();err!=nil{return engine.Result{},err}
  low:=strings.ToLower(part.Name)
  if strings.HasPrefix(low,"xl/pivottables/")||strings.HasPrefix(low,"xl/tables/")||
    strings.HasPrefix(low,"xl/externallinks/")||strings.HasPrefix(low,"xl/connections")||
    strings.HasPrefix(low,"xl/querytables/")||strings.HasPrefix(low,"xl/comments")||
    strings.HasPrefix(low,"xl/drawings/")||strings.HasPrefix(low,"xl/charts/")||
    strings.HasPrefix(low,"xl/slicer")||strings.HasPrefix(low,"xl/metadata"){
    return engine.Result{},fmt.Errorf("BLOCKED: unsupported workbook part %s",part.Name)
  }
  r,e:=part.Open();if e!=nil{return engine.Result{},e}
  size,e:=io.Copy(io.Discard,io.LimitReader(r,maxPartBytes+1))
  closed:=r.Close()
  if e!=nil||closed!=nil||size>maxPartBytes{return engine.Result{},fmt.Errorf("BLOCKED: damaged XLSX part %s",part.Name)}
  if low=="xl/sharedstrings.xml"||strings.HasPrefix(low,"xl/worksheets/sheet")&&strings.HasSuffix(low,".xml"){
   r,e=part.Open();if e!=nil{return engine.Result{},e}
   body,e:=io.ReadAll(io.LimitReader(r,maxPartBytes+1));r.Close()
   if e!=nil||len(body)>maxPartBytes{return engine.Result{},errors.New("BLOCKED: worksheet part exceeds limit")}
   docs[part.Name]=body
   if low!="xl/sharedstrings.xml"{sheets=append(sheets,part.Name)}
  }
 }
 if len(sheets)==0{return engine.Result{},errors.New("BLOCKED: no supported worksheets found")}
 sort.Strings(sheets)
 byPart:=map[string][]excelSegment{}
 links:=map[int][]string{}
 for _,sheet:=range sheets{
  segs,refs,e:=excelCells(sheet,docs[sheet]);if e!=nil{return engine.Result{},e}
  byPart[sheet]=segs
  for idx,values:=range refs{links[idx]=append(links[idx],values...)}
 }
 var sharedName string
 for n:=range docs{if strings.EqualFold(n,"xl/sharedStrings.xml"){sharedName=n;break}}
 if sharedName==""&&len(links)>0{return engine.Result{},errors.New("BLOCKED: missing sharedStrings.xml")}
 if sharedName!=""{
  segs,e:=sharedStrings(docs[sharedName],links);if e!=nil{return engine.Result{},e}
  byPart[sharedName]=segs
 }
 all:=[]excelSegment{}
 for _,sheet:=range sheets{all=append(all,byPart[sheet]...)}
 if sharedName!=""{all=append(all,byPart[sharedName]...)}
 if len(all)==0{return engine.Result{},errors.New("BLOCKED: no supported text cells in workbook")}
 if len(all)>maxExcelSegments{return engine.Result{},errors.New("BLOCKED: too many text cells for this beta (600 maximum)")}
 masked:=make([]string,len(all));locks:=make([][]LockedSpan,len(all))
 numChars:=0
 for i,segment:=range all{
  numChars+=len([]rune(segment.Old));if numChars>maxExcelChars{return engine.Result{},errors.New("BLOCKED: workbook text exceeds safe character budget")}
  masked[i],locks[i],err=Protect(segment.Old);if err!=nil{return engine.Result{},err}
 }
 translations,err:=provider.Translate(ctx,from,to,masked);if err!=nil{return engine.Result{},err}
 if len(translations)!=len(all){return engine.Result{},errors.New("BLOCKED: incomplete model translation")}
 finished:=make([]string,len(all))
 for i,v:=range translations{
  finished[i],err=Restore(v,locks[i]);if err!=nil{return engine.Result{},fmt.Errorf("BLOCKED: %s!%s: protected value changed: %w",all[i].Part,all[i].Ref,err)}
  if len(finished[i])>12000||!utf8.ValidString(finished[i]){return engine.Result{},errors.New("BLOCKED: invalid translated text")}
 }
 replacements:=map[string][]struct{start,end int;escaped []byte}{}
 for i,seg:=range all{
  var escaped bytes.Buffer
  if err:=xml.EscapeText(&escaped,[]byte(finished[i]));err!=nil{return engine.Result{},err}
  // XLSX cells without xml:space=preserve can lose leading/trailing whitespace
  // after a model edit. Refuse rather than silently altering significant spaces.
  if strings.TrimSpace(finished[i])!=finished[i]&&strings.TrimSpace(seg.Old)==seg.Old{
   return engine.Result{},fmt.Errorf("BLOCKED: translated text introduced boundary whitespace at %s",seg.Ref)
  }
  replacements[seg.Part]=append(replacements[seg.Part],struct{start,end int;escaped []byte}{seg.Start,seg.End,escaped.Bytes()})
 }
 modified:=map[string][]byte{}
 for part,edits:=range replacements{
  original:=docs[part]
  sort.Slice(edits,func(i,j int)bool{return edits[i].start>edits[j].start})
  output:=append([]byte(nil),original...)
  for _,ed:=range edits{
   if ed.start<0||ed.end>len(output)||ed.start>ed.end{return engine.Result{},errors.New("BLOCKED: invalid XLSX edit offset")}
   patch:=make([]byte,0,len(output)-(ed.end-ed.start)+len(ed.escaped))
   patch=append(patch,output[:ed.start]...)
   patch=append(patch,ed.escaped...)
   patch=append(patch,output[ed.end:]...)
   output=patch
  }
  if err:=checkXML(output);err!=nil{return engine.Result{},fmt.Errorf("BLOCKED: invalid translated XML: %w",err)}
  modified[part]=output
 }
 var workbook bytes.Buffer
 writer:=zip.NewWriter(&workbook)
 for _,file:=range z.File{
  if err:=ctx.Err();err!=nil{return engine.Result{},err}
  if bytes,changed:=modified[file.Name];changed{
   header:=file.FileHeader
   w,e:=writer.CreateHeader(&header);if e!=nil{return engine.Result{},e}
   if _,e=w.Write(bytes);e!=nil{return engine.Result{},e}
  }else if e:=writer.Copy(file);e!=nil{return engine.Result{},e}
 }
 if err:=writer.Close();err!=nil{return engine.Result{},err}
 var audit bytes.Buffer
 audit.Write([]byte{0xef,0xbb,0xbf})
 csvw:=csv.NewWriter(&audit)
 _=csvw.Write([]string{"Excel XML part","Cell(s) / shared-string references","Shared-string index","Language from","Language to","Original","Translated","Protected spans","Status"})
 for i,seg:=range all{
  idx:="inline";if seg.SharedIndex>=0{idx=strconv.Itoa(seg.SharedIndex)}
  if err:=csvw.Write([]string{seg.Part,quotedCSV(seg.Ref),idx,from,to,quotedCSV(seg.Old),quotedCSV(finished[i]),fmt.Sprint(len(locks[i])),"REVIEW_MEANING_AND_CELL_OVERFLOW"});err!=nil{return engine.Result{},err}
 }
 csvw.Flush();if err:=csvw.Error();err!=nil{return engine.Result{},err}
 origSum:=sha256.Sum256(f.Data);outSum:=sha256.Sum256(workbook.Bytes())
 manifest:=fmt.Sprintf("Original SHA256: %s\nTranslated SHA256: %s\nRoute: %s -> %s\nEdited text records: %d\nFormulas, styles and numeric cell XML outside translated text are unmodified. Layout overflow and accuracy require manual review.\n",hex.EncodeToString(origSum[:]),hex.EncodeToString(outSum[:]),from,to,len(all))
 name:=strings.TrimSuffix(engine.SafeName(f.Name),path.Ext(f.Name))+"_"+to+".xlsx"
 var bundle bytes.Buffer;wz:=zip.NewWriter(&bundle)
 for _,item:=range []struct{name string;data []byte}{{name,workbook.Bytes()},{"Translation_Audit.csv",audit.Bytes()},{"MANIFEST.txt",[]byte(manifest)}}{
  w,e:=wz.Create(item.name);if e!=nil{return engine.Result{},e}
  if _,e=w.Write(item.data);e!=nil{return engine.Result{},e}
 }
 if err:=wz.Close();err!=nil{return engine.Result{},err}
 return engine.Result{Filename:"OpenDesk_XLSX_Translation_REVIEW.zip",ContentType:"application/zip",Data:bundle.Bytes()},nil
}
