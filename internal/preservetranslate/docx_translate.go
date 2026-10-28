package preservetranslate

// Safe DOCX translation core. Only plain one-run paragraphs are accepted;
// complex formatting is blocked rather than silently discarded.
import (
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
 "strings"
 "unicode/utf8"
 "opendesk-free/internal/engine"
)
type TextTranslator interface { Translate(context.Context,string,string,[]string)([]string,error) }
const maxDocxBytes=12<<20
const maxDocxParagraphs=600
var pRE=regexp.MustCompile(`(?s)<w:p(?:\s[^>]*)?>.*?</w:p>`)
var tRE=regexp.MustCompile(`(?s)<w:t(?:\s[^>]*)?>(.*?)</w:t>`)
var unsupportedMarkup=regexp.MustCompile(`<w:(?:instrText|fldChar|fldSimple|sdt|hyperlink|ins|del|drawing|object|altChunk|tab|br|cr|sym|pict|customXml|smartTag|delText|bookmarkStart|bookmarkEnd|commentRangeStart|commentRangeEnd|commentReference|footnoteReference|endnoteReference|moveTo|moveFrom|proofErr|softHyphen|noBreakHyphen|permStart|permEnd)(?:\s|/|>)`)
type wordSegment struct { Index, Start, End int; Old string }
func checkXML(data []byte)error{
 d:=xml.NewDecoder(bytes.NewReader(data))
 for {_,e:=d.Token();if e==io.EOF{return nil};if e!=nil{return e}}
}
func wordSegments(data []byte)([]wordSegment,error){
 if !bytes.Contains(data,[]byte("xmlns:w=")){return nil,errors.New("unsupported Word XML prefix")}
 // Regex locators are intentionally allowed only for plain OOXML; do not
 // match fake w:p/w:t tags hidden in comments, CDATA or alternate payloads.
 for _,unsafe:=range [][]byte{[]byte("<!--"),[]byte("<![CDATA["),[]byte("<!DOCTYPE"),[]byte("<!ENTITY"),[]byte("<mc:AlternateContent")}{
   if bytes.Contains(data,unsafe){return nil,errors.New("BLOCKED: XML comments/alternate content need a structural editor")}
 }
 if e:=checkXML(data);e!=nil{return nil,e}
 if unsupportedMarkup.Match(data){return nil,errors.New("BLOCKED: document has fields, revisions, links, complex drawings or breaks")}
 paragraphs:=pRE.FindAllIndex(data,-1)
 all:=tRE.FindAllIndex(data,-1)
 if len(paragraphs)==0||len(paragraphs)>maxDocxParagraphs{return nil,errors.New("BLOCKED: no paragraphs or paragraph limit exceeded")}
 out:=[]wordSegment{};seen:=0
 for i,loc:=range paragraphs{
   p:=data[loc[0]:loc[1]]
   text:=tRE.FindAllSubmatchIndex(p,-1)
   seen+=len(text)
   if len(text)>1{return nil,fmt.Errorf("BLOCKED: paragraph %d has multiple formatted text runs",i+1)}
   if len(text)==0{continue}
   m:=text[0]
   var decoded struct {Text string `xml:",chardata"`}
   if e:=xml.Unmarshal([]byte("<root>"+string(p[m[2]:m[3]])+"</root>"),&decoded);e!=nil{return nil,e}
   if strings.TrimSpace(decoded.Text)==""{continue}
   if len(decoded.Text)>4000{return nil,fmt.Errorf("BLOCKED: paragraph %d exceeds 4000 bytes",i+1)}
   out=append(out,wordSegment{Index:i+1,Start:loc[0]+m[2],End:loc[0]+m[3],Old:decoded.Text})
 }
 if seen!=len(all){return nil,errors.New("BLOCKED: Word text outside ordinary paragraphs")}
 if len(out)==0{return nil,errors.New("BLOCKED: nothing to translate")}
 return out,nil
}
func quotedCSV(s string)string{
 if len(s)>0&&strings.ContainsAny(s[:1],"=+-@\t\r"){return "'"+s}
 return s
}
func TranslateDOCX(ctx context.Context,files []engine.Input,from,to string,provider TextTranslator)(engine.Result,error){
 if len(files)!=1{return engine.Result{},errors.New("choose one DOCX")}
 from,e:=canonicalTag(from);if e!=nil{return engine.Result{},e}
 to,e=canonicalTag(to);if e!=nil{return engine.Result{},e}
 if from==to{return engine.Result{},errors.New("source and target languages must be different")}
 f:=files[0]
 if strings.ToLower(path.Ext(f.Name))!=".docx"||len(f.Data)>maxDocxBytes||len(f.Data)<100{return engine.Result{},errors.New("select a valid DOCX under 12 MB")}
 pre:=inspectOOXML(ctx,f)
 if pre.Status!="READY"{return engine.Result{},fmt.Errorf("BLOCKED by FormGuard (%s): %s",pre.Status,pre.Findings)}
 z,e:=zip.NewReader(bytes.NewReader(f.Data),int64(len(f.Data)));if e!=nil{return engine.Result{},e}
 var original []byte
 for _,part:=range z.File {
   low:=strings.ToLower(part.Name)
   if strings.HasPrefix(low,"word/header")||strings.HasPrefix(low,"word/footer")||strings.HasPrefix(low,"word/footnotes")||strings.HasPrefix(low,"word/endnotes")||strings.HasPrefix(low,"customxml/")||strings.HasPrefix(low,"word/glossary/"){
     return engine.Result{},fmt.Errorf("BLOCKED: unsupported text-bearing part %s",part.Name)
   }
   // Test CRC and deflate integrity of ALL Office parts, not only XML.
   // A corrupt picture/media part otherwise survives zip.Copy unnoticed.
   rr,openErr:=part.Open()
   if openErr!=nil{return engine.Result{},fmt.Errorf("BLOCKED: invalid package part %s: %w",part.Name,openErr)}
   n,copyErr:=io.Copy(io.Discard,io.LimitReader(rr,maxPartBytes+1))
   closeErr:=rr.Close()
   if copyErr!=nil||closeErr!=nil||n>maxPartBytes{return engine.Result{},fmt.Errorf("BLOCKED: corrupt or oversized Office part %s",part.Name)}
   if low=="word/document.xml"{
     r,err:=part.Open();if err!=nil{return engine.Result{},err}
     original,err=io.ReadAll(io.LimitReader(r,maxPartBytes+1));r.Close()
     if err!=nil||len(original)>maxPartBytes{return engine.Result{},errors.New("main Word XML too large")}
   }
 }
 segments,e:=wordSegments(original);if e!=nil{return engine.Result{},e}
 masked:=make([]string,len(segments));locks:=make([][]LockedSpan,len(segments))
 sumChars:=0
 for i,s:=range segments {
   sumChars+=len([]rune(s.Old));if sumChars>125000{return engine.Result{},errors.New("translation exceeds 125000 characters")}
   masked[i],locks[i],e=Protect(s.Old);if e!=nil{return engine.Result{},e}
 }
 if e=ctx.Err();e!=nil{return engine.Result{},e}
 translated,e:=provider.Translate(ctx,from,to,masked)
 if e!=nil{return engine.Result{},e}
 if len(translated)!=len(segments){return engine.Result{},errors.New("BLOCKED: translation missing paragraphs")}
 restored:=make([]string,len(segments))
 for i,s:=range translated{
   restored[i],e=Restore(s,locks[i]);if e!=nil{return engine.Result{},fmt.Errorf("BLOCKED: paragraph %d protected value changed: %w",segments[i].Index,e)}
   if len(restored[i])>12000||!utf8.ValidString(restored[i]){return engine.Result{},errors.New("BLOCKED: translated text invalid or too long")}
 }
 updated:=append([]byte(nil),original...)
 for i:=len(segments)-1;i>=0;i--{
   s:=segments[i];var encoded bytes.Buffer
   if e=xml.EscapeText(&encoded,[]byte(restored[i]));e!=nil{return engine.Result{},e}
   buf:=make([]byte,0,len(updated)-(s.End-s.Start)+encoded.Len())
   buf=append(buf,updated[:s.Start]...);buf=append(buf,encoded.Bytes()...);buf=append(buf,updated[s.End:]...)
   updated=buf
 }
 if e=checkXML(updated);e!=nil{return engine.Result{},fmt.Errorf("BLOCKED: translated Word XML invalid: %w",e)}
 var word bytes.Buffer;wz:=zip.NewWriter(&word)
 for _,p:=range z.File{
   if e=ctx.Err();e!=nil{return engine.Result{},e}
   if p.Name!="word/document.xml"{if e=wz.Copy(p);e!=nil{return engine.Result{},e};continue}
   h:=p.FileHeader;var w io.Writer
   w,e=wz.CreateHeader(&h);if e!=nil{return engine.Result{},e}
   if _,e=w.Write(updated);e!=nil{return engine.Result{},e}
 }
 if e=wz.Close();e!=nil{return engine.Result{},e}
 var audit bytes.Buffer;audit.Write([]byte{0xef,0xbb,0xbf})
 csvw:=csv.NewWriter(&audit)
 _=csvw.Write([]string{"Part","Paragraph","Source","Target","Original","Translation","Protected tokens","Review status"})
 for i,s:=range segments{
   if e=csvw.Write([]string{"word/document.xml",fmt.Sprint(s.Index),from,to,quotedCSV(s.Old),quotedCSV(restored[i]),fmt.Sprint(len(locks[i])),"REVIEW_LAYOUT_AND_MEANING"});e!=nil{return engine.Result{},e}
 }
 csvw.Flush();if e=csvw.Error();e!=nil{return engine.Result{},e}
 name:=strings.TrimSuffix(engine.SafeName(f.Name),path.Ext(f.Name))+"_"+to+".docx"
 sh:=sha256.Sum256(f.Data);oh:=sha256.Sum256(word.Bytes())
 manifest:=fmt.Sprintf("Original SHA256: %s\nTranslated SHA256: %s\nDirect model route: %s -> %s\nParagraphs: %d\nWARNING: layout/translation quality must be manually reviewed. Original unchanged.\n",hex.EncodeToString(sh[:]),hex.EncodeToString(oh[:]),from,to,len(segments))
 var out bytes.Buffer;archive:=zip.NewWriter(&out)
 for _,entry:=range []struct{name string;data []byte}{{name,word.Bytes()},{"Translation_Audit.csv",audit.Bytes()},{"MANIFEST.txt",[]byte(manifest)}}{
   var wr io.Writer;wr,e=archive.Create(entry.name);if e!=nil{return engine.Result{},e}
   if _,e=wr.Write(entry.data);e!=nil{return engine.Result{},e}
 }
 if e=archive.Close();e!=nil{return engine.Result{},e}
 return engine.Result{Filename:"OpenDesk_DOCX_Translation_REVIEW.zip",ContentType:"application/zip",Data:out.Bytes()},nil
}
