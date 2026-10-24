// Package engine implements local-only document operations without network calls.
package engine

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
 "context"
	"encoding/csv"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

const MaxFiles = 50
const MaxBytes = 40 << 20

type Input struct {
	Name string
	Data []byte
}
type Result struct {
	Filename, ContentType string
	Data                  []byte
}

func SafeName(s string) string {
	s = strings.ReplaceAll(s, "\\", "/")
	s = path.Base(s)
	s = strings.TrimSpace(s)
	s = strings.Trim(s, ". ")
	var b strings.Builder
	for _, r := range s {
		if r < 32 || r == 127 || strings.ContainsRune(`<>:"/\\|?*`, r) {
			b.WriteRune('_')
		} else {
			b.WriteRune(r)
		}
	}
	name := b.String()
	if len(name) > 160 {
		name = string([]rune(name)[:min(len([]rune(name)), 75)])
	}
	if name == "" || name == "." || name == ".." {
		return "file"
	}
	return name
}

func Run(op string, files []Input, pageSelections ...string) (Result, error) {
 return RunContext(context.Background(),op,files,pageSelections...)
}

// RunContext allows the local HTTP request to cancel CPU-heavy PDF processing.
func RunContext(ctx context.Context,op string, files []Input, pageSelections ...string) (Result, error) {
 if e:=ctx.Err();e!=nil{return Result{},e}
	if len(files) == 0 {
		return Result{}, errors.New("Please select at least one file / Chọn ít nhất một tệp")
	}
	if len(files) > MaxFiles {
		return Result{}, fmt.Errorf("too many files: max %d", MaxFiles)
	}
	var total int64
	for _, f := range files {
		total += int64(len(f.Data))
		if len(f.Data) == 0 {
			return Result{}, fmt.Errorf("empty file: %s", SafeName(f.Name))
		}
	}
	if total > MaxBytes {
		return Result{}, fmt.Errorf("total exceeds %d MB", MaxBytes>>20)
	}
	switch op {
	case "organize":
		return organize(files)
	case "csv-xlsx":
		return csvXLSX(files)
	case "docx-txt":
		return docxText(files)
	case "duplicates":
		return duplicates(files)
	case "join-text":
		return joinText(files)
	case "images-pdf":
		return imagesPDF(ctx,files)
	case "pdf-merge", "pdf-split", "pdf-extract", "pdf-rotate":
		return pdfProcess(ctx,op, files, pageSelections)
	default:
		return Result{}, errors.New("unknown operation")
	}
}
func zipFiles(name string, files map[string][]byte) (Result, error) {
	var b bytes.Buffer
	zw := zip.NewWriter(&b)
	paths := make([]string, 0, len(files))
	for k := range files {
		paths = append(paths, k)
	}
	sort.Strings(paths)
	for _, p := range paths {
		w, err := zw.Create(p)
		if err != nil {
			return Result{}, err
		}
		if _, err = w.Write(files[p]); err != nil {
			return Result{}, err
		}
	}
	if err := zw.Close(); err != nil {
		return Result{}, err
	}
	return Result{Filename: name, ContentType: "application/zip", Data: b.Bytes()}, nil
}
func uniqueName(p string, used map[string]int) string {
	key := strings.ToLower(p)
	if used[key] == 0 {
		used[key] = 1
		return p
	}
	ext := path.Ext(p)
	base := strings.TrimSuffix(p, ext)
	for i := 2; ; i++ {
		n := fmt.Sprintf("%s_%d%s", base, i, ext)
		k := strings.ToLower(n)
		if used[k] == 0 {
			used[k] = 1
			return n
		}
	}
}
func category(name string) string {
	switch strings.ToLower(path.Ext(name)) {
	case ".pdf":
		return "Documents/PDF"
	case ".docx", ".doc", ".xlsx", ".xls", ".pptx", ".ppt", ".odt", ".ods", ".csv", ".txt", ".md":
		return "Documents/Office"
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg", ".bmp":
		return "Images"
	case ".mp3", ".wav", ".ogg", ".flac", ".m4a":
		return "Audio"
	case ".mp4", ".mkv", ".mov", ".avi", ".webm":
		return "Video"
	case ".zip", ".7z", ".rar", ".tar", ".gz":
		return "Archives"
	default:
		return "Other"
	}
}
func organize(files []Input) (Result, error) {
	out := map[string][]byte{}
	used := map[string]int{}
	var b bytes.Buffer
	w := csv.NewWriter(&b)
	_ = w.Write([]string{"Original name", "Saved in ZIP", "Bytes", "SHA256"})
	for _, f := range files {
		nm := SafeName(f.Name)
		target := uniqueName(category(nm)+"/"+nm, used)
		out[target] = f.Data
		h := sha256.Sum256(f.Data)
		_ = w.Write([]string{csvSafe(nm), csvSafe(target), strconv.Itoa(len(f.Data)), hex.EncodeToString(h[:])})
	}
	w.Flush()
	out["MANIFEST.csv"] = b.Bytes()
	return zipFiles("OpenDesk_Organized.zip", out)
}
func csvSafe(s string) string {
	if len(s) > 0 && strings.ContainsAny(string(s[0]), "=+-@\t\r") {
		return "'" + s
	}
	return s
}
func duplicates(files []Input) (Result, error) {
	hashes := make([]string, len(files))
	counts := map[string]int{}
	for i, f := range files {
		h := sha256.Sum256(f.Data)
		hs := hex.EncodeToString(h[:])
		hashes[i] = hs
		counts[hs]++
	}
	var b bytes.Buffer
	b.Write([]byte{0xef, 0xbb, 0xbf})
	w := csv.NewWriter(&b)
	_ = w.Write([]string{"File", "Bytes", "SHA256", "Duplicate count", "Status"})
	for i, f := range files {
		cnt := counts[hashes[i]]
		status := "UNIQUE"
		if cnt > 1 {
			status = "DUPLICATE"
		}
		_ = w.Write([]string{csvSafe(SafeName(f.Name)), strconv.Itoa(len(f.Data)), hashes[i], strconv.Itoa(cnt), status})
	}
	w.Flush()
	return Result{"OpenDesk_Duplicates.csv", "text/csv; charset=utf-8", b.Bytes()}, nil
}
func joinText(files []Input) (Result, error) {
	var b bytes.Buffer
	for _, f := range files {
		ext := strings.ToLower(path.Ext(f.Name))
		if ext != ".txt" && ext != ".md" {
			return Result{}, fmt.Errorf("only .txt and .md are supported: %s", SafeName(f.Name))
		}
		if !utf8.Valid(f.Data) {
			return Result{}, fmt.Errorf("only UTF-8 text supported: %s", SafeName(f.Name))
		}
		fmt.Fprintf(&b, "\n===== %s =====\n", SafeName(f.Name))
		b.Write(bytes.TrimPrefix(f.Data, []byte{0xef, 0xbb, 0xbf}))
		b.WriteString("\n")
	}
	return Result{"OpenDesk_Combined.txt", "text/plain; charset=utf-8", b.Bytes()}, nil
}
func docxText(files []Input) (Result, error) {
	out := map[string][]byte{}
	used := map[string]int{}
	for _, f := range files {
		if strings.ToLower(path.Ext(f.Name)) != ".docx" {
			return Result{}, fmt.Errorf("only .docx supported: %s", SafeName(f.Name))
		}
		z, err := zip.NewReader(bytes.NewReader(f.Data), int64(len(f.Data)))
		if err != nil {
			return Result{}, fmt.Errorf("bad DOCX %s: %w", SafeName(f.Name), err)
		}
		var text []byte
		found := false
		for _, item := range z.File {
			if item.Name != "word/document.xml" {
				continue
			}
			if item.UncompressedSize64 > 20<<20 {
				return Result{}, errors.New("DOCX content too large")
			}
			rc, e := item.Open()
			if e != nil {
				return Result{}, e
			}
			data, e := io.ReadAll(io.LimitReader(rc, 20<<20+1))
			_ = rc.Close()
			if e != nil {
				return Result{}, e
			}
			if len(data) > 20<<20 {
				return Result{}, errors.New("DOCX XML too large")
			}
			text, e = docXMLToText(data)
			if e != nil {
				return Result{}, e
			}
			found = true
			break
		}
		if !found {
			return Result{}, fmt.Errorf("not a DOCX document: %s", SafeName(f.Name))
		}
		stem := strings.TrimSuffix(SafeName(f.Name), path.Ext(SafeName(f.Name)))
		out[uniqueName(stem+".txt", used)] = text
	}
	if len(out) == 1 {
		for k, v := range out {
			return Result{k, "text/plain; charset=utf-8", v}, nil
		}
	}
	return zipFiles("OpenDesk_DOCX_Text.zip", out)
}
func docXMLToText(data []byte) ([]byte, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var b strings.Builder
	inText := false
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "t":
				inText = true
			case "tab":
				b.WriteByte('\t')
			case "br", "cr":
				b.WriteByte('\n')
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "t":
				inText = false
			case "p":
				b.WriteByte('\n')
			}
		case xml.CharData:
			if inText {
				b.Write([]byte(t))
			}
		}
	}
	return []byte(strings.TrimSpace(b.String()) + "\n"), nil
}
func csvXLSX(files []Input) (Result, error) {
	var b bytes.Buffer
	zw := zip.NewWriter(&b)
	write := func(n, content string) error {
		w, e := zw.Create(n)
		if e != nil {
			return e
		}
		_, e = io.WriteString(w, content)
		return e
	}
	names := make([]string, len(files))
	used := map[string]int{}
	var workbook strings.Builder
	workbook.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets>`)
	var rels strings.Builder
	rels.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	for i, f := range files {
		if strings.ToLower(path.Ext(f.Name)) != ".csv" {
			return Result{}, fmt.Errorf("only .csv supported: %s", SafeName(f.Name))
		}
		if !utf8.Valid(f.Data) {
			return Result{}, fmt.Errorf("CSV must be UTF-8: %s", SafeName(f.Name))
		}
		nm := strings.TrimSuffix(SafeName(f.Name), path.Ext(SafeName(f.Name)))
		nm = strings.Map(func(r rune) rune {
			if strings.ContainsRune(`[]:*?/\\`, r) {
				return '_'
			}
			return r
		}, nm)
		if len([]rune(nm)) > 31 {
			nm = string([]rune(nm)[:31])
		}
		if nm == "" {
			nm = "Sheet"
		}
		nm = uniqueSheet(nm, used)
		names[i] = nm
		var sheet strings.Builder
		sheet.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>`)
		s := bytes.TrimPrefix(f.Data, []byte{0xef, 0xbb, 0xbf})
		delim := detectDelimiter(s)
		r := csv.NewReader(bytes.NewReader(s))
		r.Comma = delim
		r.FieldsPerRecord = -1
		row := 0
		for {
			record, e := r.Read()
			if e == io.EOF {
				break
			}
			if e != nil {
				return Result{}, fmt.Errorf("invalid CSV %s: %w", SafeName(f.Name), e)
			}
			row++
			if row > 200000 {
				return Result{}, errors.New("CSV has more than 200000 rows")
			}
			if len(record) > 16384 {
				return Result{}, errors.New("CSV has more than 16384 columns")
			}
			fmt.Fprintf(&sheet, `<row r="%d">`, row)
			for col, val := range record {
				if len(val) > 32767 {
					return Result{}, errors.New("cell exceeds Excel 32767 character limit")
				}
				coord := colName(col) + strconv.Itoa(row)
				// All imported values are inline strings: identifiers and formulas are never executed.
				fmt.Fprintf(&sheet, `<c r="%s" t="inlineStr"><is><t xml:space="preserve">%s</t></is></c>`, coord, xmlText(val))
			}
			sheet.WriteString(`</row>`)
		}
		sheet.WriteString(`</sheetData></worksheet>`)
		if e := write(fmt.Sprintf("xl/worksheets/sheet%d.xml", i+1), sheet.String()); e != nil {
			return Result{}, e
		}
		fmt.Fprintf(&workbook, `<sheet name="%s" sheetId="%d" r:id="rId%d"/>`, xmlText(nm), i+1, i+1)
		fmt.Fprintf(&rels, `<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet%d.xml"/>`, i+1, i+1)
	}
	workbook.WriteString(`</sheets></workbook>`)
	rels.WriteString(`</Relationships>`)
	var types strings.Builder
	types.WriteString(`<?xml version="1.0" encoding="UTF-8"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>`)
	for i := range files {
		fmt.Fprintf(&types, `<Override PartName="/xl/worksheets/sheet%d.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>`, i+1)
	}
	types.WriteString(`</Types>`)
	for n, v := range map[string]string{"[Content_Types].xml": types.String(), "_rels/.rels": `<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`, "xl/workbook.xml": workbook.String(), "xl/_rels/workbook.xml.rels": rels.String()} {
		if e := write(n, v); e != nil {
			return Result{}, e
		}
	}
	if e := zw.Close(); e != nil {
		return Result{}, e
	}
	return Result{"OpenDesk_Converted.xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", b.Bytes()}, nil
}
func detectDelimiter(src []byte) rune {
	s := string(src)
	if p := strings.IndexByte(s, '\n'); p >= 0 {
		s = s[:p]
	}
	var best rune = ','
	max := strings.Count(s, ",")
	for _, v := range []rune{';', '\t'} {
		n := strings.Count(s, string(v))
		if n > max {
			max = n
			best = v
		}
	}
	return best
}
func colName(n int) string {
	var s string
	for n >= 0 {
		s = string(rune('A'+n%26)) + s
		n = n/26 - 1
	}
	return s
}
func xmlText(s string) string {
	s = strings.Map(func(r rune) rune {
		if r == 9 || r == 10 || r == 13 || r >= 32 && r <= 0xD7FF || r >= 0xE000 && r <= 0xFFFD || r >= 0x10000 && r <= 0x10FFFF {
			return r
		}
		return -1
	}, s)
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}
func uniqueSheet(n string, used map[string]int) string {
	key := strings.ToLower(n)
	if used[key] == 0 {
		used[key] = 1
		return n
	}
	for i := 2; ; i++ {
		suffix := fmt.Sprintf("_%d", i)
		r := []rune(n)
		if len(r)+len([]rune(suffix)) > 31 {
			r = r[:31-len([]rune(suffix))]
		}
		s := string(r) + suffix
		k := strings.ToLower(s)
		if used[k] == 0 {
			used[k] = 1
			return s
		}
	}
}
