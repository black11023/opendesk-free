# OpenDesk Free 🧰

**Free, offline-first document utilities for everyone. / Công cụ xử lý tài liệu miễn phí dành cho mọi người.**

![OpenDesk Free desktop preview](assets/desktop-preview.png)

> **Status: v0.4.0-preview.1 development preview.** The app builds for Windows; Windows runtime acceptance testing is still required before labeling it a stable release.

## Features in v0.2 preview / Các chức năng hiện có

- Organize uploaded documents into a ZIP with folders by type and SHA-256 manifest. No original file is changed.
- CSV to Excel XLSX (one input CSV per worksheet; every imported cell is saved as text to preserve zeros and prevent formula execution).
- DOCX to plain UTF-8 text (document body, without preserving formatting or images).
- Detect byte-identical file duplicates by SHA-256 and download a CSV report.
- Combine UTF-8 TXT/Markdown documents with file-name separators.
- Vietnamese/English UI, local HTTP server limited to `127.0.0.1`, randomized port, per-session protection token, no telemetry, no paid API.

**New in v0.2 preview:** English is the default language; a dropdown switches to Vietnamese and saves your preference locally. PDF merge, split to pages and rotate 90° clockwise are added using Apache-2.0-licensed pdfcpu. **Not yet supported:** OCR, searchable scanned PDFs, editing PDF text, Excel editing and workflow automation. This is not a stable release.

## System requirements

- Windows 10/11 64-bit for supplied EXEs (no Python, Docker or Node required).
- A modern browser such as Edge, Chrome or Firefox.
- Max 50 files and 40 MB total per operation for this preview.
- Browser/PC access only. No file contents are transmitted to internet servers.

## Download

After the first build, open the repository's **Actions > Build and test Windows installer** and download `OpenDeskFree-Windows-v0.1.0.zip` from run artifacts. Extract the ZIP and run `OpenDeskFree_Setup_0.1.0.exe` or use `OpenDeskFree_Portable_0.1.0.exe`. When published as a GitHub Release, the `.exe` files appear on its release page as assets.

**Unsigned binary:** Windows SmartScreen may warn about an unfamiliar publisher. We recommend validating checksums and building from source; never bypass security warnings blindly.

## Compile from source

Go >=1.26, pdfcpu v0.16.3 Apache-2.0 dependency (downloaded during build):

```powershell
go test ./...
go build -o OpenDeskFree_Portable.exe ./cmd/opendesk
go build -ldflags "-H=windowsgui" -o cmd/setup/payload/OpenDeskFree.exe ./cmd/opendesk
go build -ldflags "-H=windowsgui" -o OpenDeskFree_Setup.exe ./cmd/setup
```

The installer sources are Windows-specific. On Linux, cross-compile using `GOOS=windows GOARCH=amd64 CGO_ENABLED=0` for both binaries.

## Privacy and security

- Runs solely on loopback (`127.0.0.1`). Not available to the internet or other computers.
- Most operations run in memory; PDF splitting may use a temporary local directory deleted after processing. There is no persistent document database or external upload.
- POST operations require a secret token and Origin check.
- No automated file deletion, renaming or modification of original uploads.
- An operation may consume up to tens of megabytes of RAM and processing time depending on input.
- Review files before using them in other applications. This preview is not a substitute for antivirus or document validation.

## Contribute

Please use Issues to report bugs, including Windows version, input format (do not include confidential documents), steps to reproduce, expected and actual result. See `CONTRIBUTING.md` and `docs/ROADMAP.md`.

## License

MIT. Copyright © 2026 OpenDesk Free contributors.

### International users / Người dùng quốc tế
English is selected by default for new users. Switch to **Tiếng Việt** in the Language dropdown. The selected language is stored in a small localhost browser cookie, shared across random local app ports in the same browser profile. No account required.

## Public preview releases

After tests and Windows EXE builds succeed on `main`, the workflow creates the **v0.4.0-preview.1** public GitHub pre-release once. Subsequent development commits continue to produce Actions artifacts but do not silently replace the release; a future release requires a deliberate version bump. Executables are unsigned and desktop install/uninstall on an end-user Windows PC still needs final acceptance testing.

**Preview 2:** language cookie contains only `en` or `vi`. No document contents or account identifiers are stored in it.

### Select PDF pages / Chọn trang PDF

For **Extract PDF pages**, enter `2`, `1,5,9`, `10-15`, or even `7,2,5`; the custom order is honored. Leave blank to select all pages (limited to 300 output pages). Download one combined PDF or individual PDFs inside a ZIP. For **Merge PDFs**, choose a page range for each file and use the arrows to change file order. Supported input: up to 5,000 pages/document under the existing 40 MB total upload limit; max 300 selected output pages; ZIP mode up to 100 pages. Page indexes start from **1** and out-of-bounds selection is rejected. Original source files are never overwritten.

**v0.2 Preview 4:** Added per-PDF page selection and input order controls to the bilingual app. Download selected pages into one PDF or as separate PDF files in a ZIP. Build and regression tests must pass before public prerelease. [Latest preview](https://github.com/black11023/opendesk-free/releases/tag/v0.4.0-preview.1).

## OpenDesk Tools Hub — 0.3 preview

A professional local **tools hub** with search, category filters, English/Vietnamese labels and a curated catalog of 44 independent third-party free/open-source projects. The catalog is stored in the app, without telemetry or background downloads. **Catalog entries are links, not embedded tools or guaranteed integrations.** Upstream tools may need separate installation, resources and licensing review.

**New working native feature:** PNG/JPG images to one PDF (one page per image) using existing pdfcpu library, max 25 images, 20 MP per image, 40 MB combined upload. Originals remain untouched. See [integration blueprint](docs/INTEGRATION_BLUEPRINT.md), [landscape](docs/SOFTWARE_LANDSCAPE_2026.md) and [legal checks](docs/LICENSES_AND_INTEGRATIONS.md).

## V0.4 — Offline OCR (optional local Tesseract adapter)

The app can use a **separately installed** Tesseract OCR executable (Apache-2.0). It is **not bundled** in the Setup EXE. OpenDesk detects whether the engine and the English (`eng`) and/or Vietnamese (`vie`) traineddata are installed; a missing engine or language pack is shown as an actionable BLOCKED status. No automatic downloads.

Select up to 10 PNG/JPG image scans and choose English, Vietnamese or combined language mode. Download recognized UTF-8 TXT or a newly generated searchable PDF with a text layer. Each job runs on the local PC using standard input; document contents are not sent to OpenAI or third-party APIs. Input validation checks extension, actual format and dimensions (20 MP/image); file upload remains max 40 MB total.

**Not yet available:** rasterizing scanned source PDF pages for OCR, automatic Tesseract installation, bundled traineddata, document-layout fidelity guarantees, OCR confidence/proofreading and full Windows end-user acceptance. Use [setup instructions](docs/OCR_LOCAL_SETUP.md).
