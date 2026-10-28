package main

import (
	"context"
	"crypto/rand"
 "flag"
	"encoding/json"
	"embed"
	"encoding/hex"
	"fmt"
	"html"
	"io"
 "mime"
	"log"
	"net"
	"net/http"
	"runtime"
 "runtime/debug"
	"strings"
	"time"

	"opendesk-free/internal/engine"
	"opendesk-free/internal/ocr"
 "opendesk-free/internal/office"
 "opendesk-free/internal/preservetranslate"
)

//go:embed web/index.html
var web embed.FS

const Version = "0.7.0-preview.1"

func main() {
 // This is a soft Go heap budget, not a cap on external application processes.
 debug.SetMemoryLimit(512 << 20)
 port := flag.Int("port",0,"local-only listen port (0 chooses a random port)")
 noBrowser := flag.Bool("no-browser",false,"do not open the default browser")
 flag.Parse()
 if *port<0 || *port>65535 { critical(fmt.Errorf("invalid local port %d",*port));return }
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		critical(err)
		return
	}
	token := hex.EncodeToString(secretBytes)
	ln, err := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d",*port))
	if err != nil {
		critical(err)
		return
	}
	mux := http.NewServeMux()
	expectedHost := ln.Addr().String()
	// Reject DNS rebinding and alternate Host headers, even for requests on loopback.
	sameHost := func(w http.ResponseWriter, r *http.Request) bool {
		if r.Host != expectedHost {
			http.Error(w, "Forbidden Host", http.StatusForbidden)
			return false
		}
		return true
	}
	htmlBytes, err := web.ReadFile("web/index.html")
	if err != nil {
		critical(err)
		return
	}
	page := strings.ReplaceAll(string(htmlBytes), "__TOKEN__", html.EscapeString(token))
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if !sameHost(w, r) {
			return
		}
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Security-Policy", `default-src 'none'; style-src 'unsafe-inline'; script-src 'self' 'unsafe-inline'; connect-src 'self'; img-src 'self' data:; form-action 'self'; base-uri 'none'; frame-ancestors 'none'`)
		_, _ = io.WriteString(w, page)
	})
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if !sameHost(w, r) {
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		_, _ = io.WriteString(w, "OpenDesk Free "+Version+" OK")
	})
	auth := func(w http.ResponseWriter, r *http.Request) bool {
		// A random per-process token prevents third-party websites from invoking local file operations.
		if !sameHost(w, r) {
			return false
		}
		if r.Header.Get("X-OpenDesk-Token") != token {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return false
		}
		if o := r.Header.Get("Origin"); o != "" && o != "http://"+r.Host {
			http.Error(w, "Invalid Origin", http.StatusForbidden)
			return false
		}
		return true
	}
	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second, IdleTimeout: 2 * time.Minute}
 processGate := newProcessGate(2)
 // Optional external renderers (LibreOffice, Tesseract, OCRmyPDF) share
 // one slot so multiple subprocess-heavy jobs never overlap in this app.
 externalGate := newProcessGate(1)
	mux.HandleFunc("GET /model-packs",func(w http.ResponseWriter,r *http.Request){
  if !auth(w,r){return}
  w.Header().Set("Content-Type","application/json; charset=utf-8")
  w.Header().Set("Cache-Control","no-store")
  w.Header().Set("X-Content-Type-Options","nosniff")
  _=json.NewEncoder(w).Encode(preservetranslate.ApprovedPacks())
 })
 mux.HandleFunc("POST /model-packs/install",func(w http.ResponseWriter,r *http.Request){
  if !auth(w,r){return}
  if !processGate.TryEnter(){http.Error(w,"Another task is running",http.StatusTooManyRequests);return}
  defer processGate.Leave()
  if !externalGate.TryEnter(){http.Error(w,"A local document engine is busy",http.StatusTooManyRequests);return}
  defer externalGate.Leave()
  // Uploads are streamed into a temporary file, never entirely buffered in
  // memory. No model is installed until SHA-256 + ZIP/metadata checks pass.
  r.Body=http.MaxBytesReader(w,r.Body,96<<20)
  if e:=r.ParseMultipartForm(1<<20);e!=nil{
   http.Error(w,"Invalid or oversized model request",http.StatusRequestEntityTooLarge);return
  }
  if r.MultipartForm!=nil{defer r.MultipartForm.RemoveAll()}
  id,mode:=r.FormValue("model_id"),r.FormValue("mode")
  ctx,cancel:=context.WithTimeout(r.Context(),14*time.Minute);defer cancel()
  var receipt preservetranslate.PackReceipt
  var err error
  switch mode {
  case "download":
   if len(r.MultipartForm.File)!=0{http.Error(w,"Unexpected uploaded file",400);return}
   receipt,err=preservetranslate.DownloadApprovedModel(ctx,id)
  case "import":
   upload,header,e:=r.FormFile("model_file")
   if e!=nil{http.Error(w,"Select one verified .argosmodel file",400);return}
   defer upload.Close()
   if header.Size>96<<20||header.Size<0{http.Error(w,"Unsupported model size",413);return}
   receipt,err=preservetranslate.InstallApprovedModel(ctx,id,upload)
  default:
   http.Error(w,"Invalid model installation mode",400);return
  }
  if err!=nil{
   http.Error(w,"Model installation BLOCKED: "+err.Error(),http.StatusUnprocessableEntity);return
  }
  w.Header().Set("Content-Type","application/json; charset=utf-8")
  w.Header().Set("Cache-Control","no-store")
  w.Header().Set("X-Content-Type-Options","nosniff")
  _=json.NewEncoder(w).Encode(receipt)
 })
	mux.HandleFunc("GET /translate-status",func(w http.ResponseWriter,r *http.Request){
  if !auth(w,r){return}
  w.Header().Set("Content-Type","application/json; charset=utf-8")
  w.Header().Set("Cache-Control","no-store")
  w.Header().Set("X-Content-Type-Options","nosniff")
  _=json.NewEncoder(w).Encode(preservetranslate.ProbeArgos(r.Context()))
 })
	mux.HandleFunc("GET /ocr-status",func(w http.ResponseWriter,r *http.Request){
  if !auth(w,r){return}
  w.Header().Set("Content-Type","application/json; charset=utf-8")
  w.Header().Set("Cache-Control","no-store")
  w.Header().Set("X-Content-Type-Options","nosniff")
  _=json.NewEncoder(w).Encode(ocr.Probe(r.Context()))
 })
	mux.HandleFunc("GET /pdf-ocr-status",func(w http.ResponseWriter,r *http.Request){
  if !auth(w,r){return}
  w.Header().Set("Content-Type","application/json; charset=utf-8")
  w.Header().Set("Cache-Control","no-store")
  w.Header().Set("X-Content-Type-Options","nosniff")
  _=json.NewEncoder(w).Encode(ocr.PDFProbe(r.Context()))
 })
	mux.HandleFunc("GET /office-status",func(w http.ResponseWriter,r *http.Request){
  if !auth(w,r){return}
  w.Header().Set("Content-Type","application/json; charset=utf-8")
  w.Header().Set("Cache-Control","no-store")
  w.Header().Set("X-Content-Type-Options","nosniff")
  _=json.NewEncoder(w).Encode(office.Probe(r.Context()))
 })
	mux.HandleFunc("POST /process", func(w http.ResponseWriter, r *http.Request) {
		if !auth(w, r) {
			return
		}
        if !processGate.TryEnter(){
          w.Header().Set("Retry-After","2")
          http.Error(w,localized(r,"OpenDesk is busy with other jobs. Try again shortly.","OpenDesk đang xử lý tác vụ khác. Vui lòng thử lại sau."),http.StatusTooManyRequests)
          return
        }
        defer processGate.Leave()
        var mem runtime.MemStats
        runtime.ReadMemStats(&mem)
        if mem.HeapAlloc>300<<20{
          w.Header().Set("Retry-After","5")
          http.Error(w,localized(r,"Memory usage is high; try again later.","Bộ nhớ đang cao; vui lòng thử lại sau."),http.StatusServiceUnavailable)
          return
        }
		r.Body = http.MaxBytesReader(w, r.Body, engine.MaxBytes+(3<<20))
		if err := r.ParseMultipartForm(engine.MaxBytes + (1 << 20)); err != nil {
			http.Error(w, localized(r,"Invalid or oversized upload","Tệp không hợp lệ hoặc quá lớn"), http.StatusBadRequest)
			return
		}
		if r.MultipartForm != nil { defer r.MultipartForm.RemoveAll() }
		uploads := r.MultipartForm.File["files"]
		if len(uploads) == 0 || len(uploads) > engine.MaxFiles {
			http.Error(w, localized(r,"Select 1–50 files","Chọn 1–50 tệp"), http.StatusBadRequest)
			return
		}
        // Enforce an aggregate input budget before copying uploads into
        // Go byte slices (the multipart parser may already have disk-spooled).
        var combined int64
        for _,header:=range uploads {
          if header.Size<0||header.Size>engine.MaxBytes-combined {
            http.Error(w,localized(r,"Combined upload exceeds 40 MB.","Tổng dung lượng tệp vượt 40 MB."),http.StatusRequestEntityTooLarge)
            return
          }
          combined+=header.Size
        }
		op := r.FormValue("operation")
        if isExternalOperation(op)||op=="translate-docx"||op=="translate-xlsx"{
          if !externalGate.TryEnter(){
            w.Header().Set("Retry-After","3")
            http.Error(w,localized(r,
              "Office/OCR engine busy. Try again after the current conversion.",
              "Công cụ Office/OCR đang bận. Vui lòng thử lại sau khi tác vụ hiện tại hoàn tất."),
              http.StatusTooManyRequests)
            return
          }
          defer externalGate.Leave()
        }
		pageSelections := r.MultipartForm.Value["page_ranges"]
		if op == "pdf-merge" {
			if len(pageSelections) != len(uploads) {
				http.Error(w, localized(r,"Missing PDF page selections","Thiếu lựa chọn trang PDF"),400)
				return
			}
		} else if op == "pdf-split" || op == "pdf-extract" {
			pageSelections = []string{r.FormValue("pages")}
		} else {
			pageSelections = nil
		}
		files := make([]engine.Input,0,len(uploads))
		for _, f := range uploads {
			if f.Size > engine.MaxBytes {
				http.Error(w, localized(r,"File too large","Tệp quá lớn"), http.StatusBadRequest)
				return
			}
			src, e := f.Open()
			if e != nil {
				http.Error(w, localized(r,"Cannot read file","Không đọc được tệp"), 400)
				return
			}
			data, e := io.ReadAll(io.LimitReader(src, engine.MaxBytes+1))
			_ = src.Close()
			if e != nil || len(data) > engine.MaxBytes {
				http.Error(w, localized(r,"Cannot read / oversized file","Không đọc được tệp hoặc tệp quá lớn"), 400)
				return
			}
			files = append(files, engine.Input{Name: f.Filename, Data: data})
		}
		var result engine.Result
        var e error
        if op=="translate-xlsx"{
           ctx,cancel:=context.WithTimeout(r.Context(),3*time.Minute)
           defer cancel()
           result,e=preservetranslate.TranslateXLSX(ctx,files,r.FormValue("source_lang"),r.FormValue("target_lang"),preservetranslate.ArgosProvider{})
         }else if op=="translate-docx"{
           ctx,cancel:=context.WithTimeout(r.Context(),3*time.Minute)
           defer cancel()
           result,e=preservetranslate.TranslateDOCX(ctx,files,r.FormValue("source_lang"),r.FormValue("target_lang"),preservetranslate.ArgosProvider{})
         }else if op=="formguard-audit"{
           result,e=preservetranslate.Preflight(r.Context(),files)
         }else if op=="office-pdf-batch"{
          ctx,cancel:=context.WithTimeout(r.Context(),4*time.Minute)
          defer cancel()
          result,e=office.RunBatch(ctx,files)
        }else if op=="office-pdf"{
          ctx,cancel:=context.WithTimeout(r.Context(),4*time.Minute)
          defer cancel()
          result,e=office.Run(ctx,files)
        }else if op=="ocr-pdf"{
           ctx,cancel:=context.WithTimeout(r.Context(),4*time.Minute)
           defer cancel()
           result,e=ocr.RunPDF(ctx,files,r.FormValue("ocr_lang"))
        }else if op=="ocr-images"{
           ctx,cancel:=context.WithTimeout(r.Context(),3*time.Minute)
           defer cancel()
           result,e=ocr.RunImages(ctx,files,r.FormValue("ocr_lang"),r.FormValue("ocr_format"))
        }else{
           result,e=engine.RunContext(r.Context(),op, files, pageSelections...)
        }
		if e != nil {
			http.Error(w, localized(r,"Processing failed: "+e.Error(),"Xử lý thất bại: "+e.Error()), http.StatusUnprocessableEntity)
			return
		}
		w.Header().Set("Content-Type", result.ContentType)
		w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename":result.Filename}))
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		_, _ = w.Write(result.Data)
	})
	mux.HandleFunc("POST /quit", func(w http.ResponseWriter, r *http.Request) {
		if !auth(w, r) {
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte("bye"))
		go func() { time.Sleep(200 * time.Millisecond); _ = srv.Close() }()
	})
	url := "http://" + ln.Addr().String()
	log.Printf("OpenDesk Free v%s | local only: %s", Version, url)
	if !*noBrowser { go func() { time.Sleep(350 * time.Millisecond); _ = openBrowser(url) }() }
	if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
		critical(err)
	}
}
func openBrowser(url string) error { return openSystemBrowser(url) }
func critical(err error) {
    log.Printf("OpenDesk Free error: %v",err)
    showStartupError(err)
}

func localized(r *http.Request, en, vi string) string {if r.Header.Get("X-OpenDesk-Lang")=="vi" {return vi};return en}
