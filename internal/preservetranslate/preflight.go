// Package preservetranslate contains the shared data-safety foundation for
// future multilingual translation. Preflight is read-only: it DOES NOT translate.
package preservetranslate

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
 "strings"

 "opendesk-free/internal/engine"
)

const (
 maxPreflightFiles = 8
 maxArchiveParts = 1500
 maxPartBytes = 24 << 20
 maxExpandedBytes = 120 << 20
)

type preflightRecord struct {
 Name, Format, SHA256, Status string
 Parts, TextNodes, Formulas, NumberCells, MergedRanges int
 Findings string
}
func (r *preflightRecord)flag(status, note string){
 if r.Status!="BLOCKED" && (status=="BLOCKED" || r.Status=="READY" && status=="REVIEW") {r.Status=status}
 if note!="" && !strings.Contains("|"+r.Findings+"|","|"+note+"|"){
   if r.Findings!="" {r.Findings+=" | "}
   r.Findings+=note
 }
}

func cellSafe(v string)string {
 v=strings.ReplaceAll(strings.ReplaceAll(v,"\r"," "),"\n"," ")
 v=strings.TrimSpace(v)
 if len(v)>1800 {v=v[:1800]}
 if len(v)>0 && strings.ContainsAny(v[:1],"=+-@\t"){return "'"+v}
 return v
}

// Preflight inspects native OOXML files without altering or executing them.
// It generates a human-readable audit CSV, never a translated document.
func Preflight(ctx context.Context,files []engine.Input)(engine.Result,error){
 if len(files)<1 || len(files)>maxPreflightFiles {return engine.Result{},fmt.Errorf("select 1 to %d Word/Excel/PowerPoint files",maxPreflightFiles)}
 var total int
 for _,f:=range files{total+=len(f.Data)}
 if total>engine.MaxBytes{return engine.Result{},errors.New("combined inputs exceed 40 MB")}
 var buf bytes.Buffer
 buf.Write([]byte{0xef,0xbb,0xbf}) // Excel-compatible UTF-8 BOM
 w:=csv.NewWriter(&buf)
 if err:=w.Write([]string{"Source file","Format","SHA-256 original","Readiness (not translation)","ZIP parts","Text nodes","Formula nodes","Numeric cells (estimate)","Merged ranges","Findings"});err!=nil{return engine.Result{},err}
 for _,f:=range files{
   if err:=ctx.Err();err!=nil{return engine.Result{},err}
   r:=inspectOOXML(ctx,f)
   row:=[]string{cellSafe(r.Name),r.Format,r.SHA256,r.Status,
     fmt.Sprint(r.Parts),fmt.Sprint(r.TextNodes),fmt.Sprint(r.Formulas),
     fmt.Sprint(r.NumberCells),fmt.Sprint(r.MergedRanges),cellSafe(r.Findings)}
   if err:=w.Write(row);err!=nil{return engine.Result{},err}
 }
 w.Flush();if err:=w.Error();err!=nil{return engine.Result{},err}
 return engine.Result{Filename:"OpenDesk_FormGuard_Preflight.csv",ContentType:"text/csv; charset=utf-8",Data:buf.Bytes()},nil
}

func inspectOOXML(ctx context.Context,f engine.Input)preflightRecord{
 digest:=sha256.Sum256(f.Data)
 r:=preflightRecord{Name:engine.SafeName(f.Name),SHA256:hex.EncodeToString(digest[:]),Status:"READY"}
 ext:=strings.ToLower(path.Ext(f.Name))
 root:=""
 switch ext{
 case ".docx":r.Format="DOCX";root="word/document.xml"
 case ".xlsx":r.Format="XLSX";root="xl/workbook.xml"
 case ".pptx":r.Format="PPTX";root="ppt/presentation.xml"
 default:
   r.Format="UNSUPPORTED";r.flag("BLOCKED","Only .docx, .xlsx and .pptx are accepted");return r
 }
 if len(f.Data)<4 || !bytes.HasPrefix(f.Data,[]byte{'P','K',3,4}){r.flag("BLOCKED","Invalid Office ZIP header");return r}
 z,err:=zip.NewReader(bytes.NewReader(f.Data),int64(len(f.Data)))
 if err!=nil{r.flag("BLOCKED","Invalid/corrupted Office ZIP");return r}
 if len(z.File)>maxArchiveParts{r.flag("BLOCKED","Archive part count limit exceeded");return r}
 names:=map[string]bool{}
 var expanded uint64
 for _,part:=range z.File{
   if err:=ctx.Err();err!=nil{r.flag("BLOCKED","Cancelled");return r}
   n:=strings.ReplaceAll(part.Name,"\\","/")
   canonical:=path.Clean(n)
   if n==""||strings.HasPrefix(n,"/")||canonical==".."||strings.HasPrefix(canonical,"../")||canonical!=strings.TrimSuffix(n,"/")||strings.Contains(n,":"){
     r.flag("BLOCKED","Unsafe ZIP entry path");return r
   }
   key:=strings.ToLower(canonical)
   if names[key]{r.flag("BLOCKED","Duplicate ZIP entry path");return r};names[key]=true
   r.Parts++
   if part.UncompressedSize64>maxPartBytes {
     r.flag("BLOCKED","Oversized Office part");return r
   }
   expanded+=part.UncompressedSize64
   if expanded>maxExpandedBytes{r.flag("BLOCKED","Expanded Office archive size limit exceeded");return r}
   if part.UncompressedSize64>(1<<20) && part.UncompressedSize64>300*(part.CompressedSize64+1) {
     r.flag("BLOCKED","Suspicious ZIP compression ratio");return r
   }
   lower:=strings.ToLower(n)
   if strings.Contains(lower,"vbaproject")||strings.Contains(lower,"activex/")||strings.Contains(lower,"_xmlsignatures/"){
     r.flag("BLOCKED","Macros, ActiveX or signed OOXML are not supported for safe translation")
   }
   if strings.Contains(lower,"/embeddings/")||strings.Contains(lower,"/charts/")||strings.Contains(lower,"/drawings/") {
     r.flag("REVIEW","Embedded objects/charts/drawings need layout review")
   }
   if strings.Contains(lower,"/comments")||strings.Contains(lower,"/people/") {r.flag("REVIEW","Comments or reviewers need manual review")}
   relevant:=strings.HasSuffix(lower,".xml") || strings.HasSuffix(lower,".rels")
   if !relevant {continue}
   if err:=inspectXML(ctx,part,ext,n,&r);err!=nil{
     r.flag("BLOCKED","Invalid or oversized OOXML XML part");return r
   }
 }
 if !names["[content_types].xml"] || !names[strings.ToLower(root)]{
   r.flag("BLOCKED","Required OOXML package part missing")
 }
 if r.TextNodes==0{r.flag("REVIEW","No ordinary text nodes found; content may be graphical or scanned")}
 if r.Status=="READY"{r.Findings="Structural preflight passed; translation/layout accuracy NOT verified"}
 return r
}

func inspectXML(ctx context.Context,part *zip.File,ext,name string,r *preflightRecord)error{
 rd,err:=part.Open();if err!=nil{return err};defer rd.Close()
 limited:=io.LimitReader(rd,maxPartBytes+1)
 decoder:=xml.NewDecoder(limited)
 inCell:=false;cellNumber:=false;cellFormula:=false
 for {
   if err:=ctx.Err();err!=nil{return err}
   token,err:=decoder.Token()
   if err==io.EOF{break}
   if err!=nil{return err}
   switch e:=token.(type){
   case xml.StartElement:
     switch e.Name.Local{
     case "t":
       if strings.Contains(name,"word/")||strings.Contains(name,"ppt/")||strings.Contains(name,"xl/"){r.TextNodes++}
     case "f":
       if ext==".xlsx"&&strings.Contains(name,"xl/worksheets/"){r.Formulas++;cellFormula=true}
     case "mergeCell":
       if ext==".xlsx" {r.MergedRanges++}
     case "c":
       if ext==".xlsx" && strings.Contains(name,"xl/worksheets/") {
         inCell=true;cellNumber=true;cellFormula=false
         for _,attr:=range e.Attr{if attr.Name.Local=="t" && attr.Value!="n"{cellNumber=false}}
       }
     case "Relationship":
       for _,attr:=range e.Attr{
         if attr.Name.Local=="TargetMode" && strings.EqualFold(attr.Value,"External"){
            r.flag("REVIEW","External relationships require manual review")
         }
       }
     case "ins","del","sdt","fldSimple":
       if ext==".docx"&&strings.Contains(name,"word/"){r.flag("REVIEW","Tracked changes, controls or fields need review")}
     }
   case xml.EndElement:
     if e.Name.Local=="c"&&inCell{
       if cellNumber&&!cellFormula{r.NumberCells++}
       inCell=false
     }
   }
 }
 return nil
}
