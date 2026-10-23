package engine

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"io"
	"strings"
	"testing"
)

func TestOrganizeCollisionNoFileLoss(t *testing.T) {
	r, e := Run("organize", []Input{{"a.pdf", []byte("1")}, {"a.pdf", []byte("2")}, {"photo.png", []byte("x")}})
	if e != nil {
		t.Fatal(e)
	}
	z, e := zip.NewReader(bytes.NewReader(r.Data), int64(len(r.Data)))
	if e != nil {
		t.Fatal(e)
	}
	if len(z.File) != 4 {
		t.Fatalf("expected 4 entries got %d", len(z.File))
	}
}
func TestCSVXLSXPreserveFormulaAsText(t *testing.T) {
	r, e := Run("csv-xlsx", []Input{{"A.csv", []byte("part,qty\n=2+2,0015\n")}})
	if e != nil {
		t.Fatal(e)
	}
	z, e := zip.NewReader(bytes.NewReader(r.Data), int64(len(r.Data)))
	if e != nil {
		t.Fatal(e)
	}
	ok := false
	for _, f := range z.File {
		if f.Name == "xl/worksheets/sheet1.xml" {
			rc, _ := f.Open()
			b, _ := io.ReadAll(rc)
			_ = rc.Close()
			if !strings.Contains(string(b), `t="inlineStr"`) || !strings.Contains(string(b), "=2+2") || !strings.Contains(string(b), "0015") {
				t.Fatal("did not preserve text fields")
			}
			ok = true
		}
	}
	if !ok {
		t.Fatal("missing sheet")
	}
}
func TestDuplicates(t *testing.T) {
	r, e := Run("duplicates", []Input{{"1", []byte("abc")}, {"2", []byte("abc")}, {"3", []byte("def")}})
	if e != nil {
		t.Fatal(e)
	}
	rows, e := csv.NewReader(bytes.NewReader(bytes.TrimPrefix(r.Data, []byte{0xef, 0xbb, 0xbf}))).ReadAll()
	if e != nil {
		t.Fatal(e)
	}
	if rows[1][4] != "DUPLICATE" || rows[3][4] != "UNIQUE" {
		t.Fatal(rows)
	}
}
func TestRejectedAndSanitized(t *testing.T) {
	if _, e := Run("csv-xlsx", []Input{{"a.txt", []byte("bad")}}); e == nil {
		t.Fatal("expected rejection")
	}
	if SafeName(`../../bad.exe`) != "bad.exe" {
		t.Fatal(SafeName(`../../bad.exe`))
	}
}
func TestJoinText(t *testing.T) {
	r, e := Run("join-text", []Input{{"one.txt", []byte("first")}, {"two.md", []byte("second")}})
	if e != nil || !bytes.Contains(r.Data, []byte("second")) {
		t.Fatal(e)
	}
}
func TestDOCX(t *testing.T) {
	var b bytes.Buffer
	z := zip.NewWriter(&b)
	w, _ := z.Create("word/document.xml")
	_, _ = w.Write([]byte(`<w:document xmlns:w="x"><w:body><w:p><w:r><w:t>Hello</w:t></w:r></w:p><w:p><w:r><w:t>World</w:t></w:r></w:p></w:body></w:document>`))
	_ = z.Close()
	r, e := Run("docx-txt", []Input{{"doc.docx", b.Bytes()}})
	if e != nil {
		t.Fatal(e)
	}
	if string(r.Data) != "Hello\nWorld\n" {
		t.Fatal(string(r.Data))
	}
}
