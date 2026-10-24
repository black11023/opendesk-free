# OpenDesk Free — free, private tools for everyone 🧰

**English · Tiếng Việt** | **No subscription · No account · Offline-first**

![OpenDesk Free preview](assets/desktop-preview.png)

[Download for Windows](https://github.com/black11023/opendesk-free/releases) · [Report a bug](https://github.com/black11023/opendesk-free/issues) · [Contribute](CONTRIBUTING.md) · [Roadmap](docs/ROADMAP.md)

> **v0.5.1-preview.1 — Windows 10/11 x64 preview.** Automated tests, Windows executable builds and a local HTTP runtime smoke test run on GitHub Actions. End-user Windows installation and third-party converter accuracy are not yet certified. Executables are unsigned.

OpenDesk Free provides a friendly desktop launcher for essential document operations without paid cloud APIs. English is the default; switch to **Tiếng Việt** at any time. The application runs on `127.0.0.1` and does not upload your files.

## Features — truthful availability

| Feature | State | Details |
|---|---|---|
| PDF merge, select/split/reorder pages, rotate | **Built-in** | Choose page ranges |
| PNG/JPG → PDF | **Built-in** | Max 25 images |
| CSV → XLSX, DOCX → plain text | **Built-in** | Original files stay unchanged |
| Organize ZIP, SHA-256 duplicate detection, text merge | **Built-in** | Download new results |
| Image OCR → TXT/searchable PDF | **Optional local adapter** | Install Tesseract and language data |
| Scan-PDF → searchable PDF | **Optional local adapter** | Install OCRmyPDF + Tesseract |
| Word/Excel/PowerPoint → PDF | **Optional local adapter (V0.5)** | Install LibreOffice separately |
| 44 free software links, 11 categories | **Directory only** | External projects, not integrated |
| FFmpeg, workflow automation, more platforms | **Planned** | Not yet available |

### Install and use

1. Open [Releases](https://github.com/black11023/opendesk-free/releases). Download the newest `OpenDeskFree_Setup_*.exe` or the Portable EXE.
2. Check `SHA256SUMS.txt`. These EXEs are unsigned; do not blindly bypass system security warnings.
3. Launch OpenDesk, select English or Vietnamese, choose an operation and drop your files into the window.
4. To enable **Office → PDF**, install [LibreOffice](https://www.libreoffice.org/) separately and follow [Office setup](docs/OFFICE_LOCAL_SETUP.md). To enable OCR use [OCR setup](docs/OCR_LOCAL_SETUP.md).

### Safety and limits

Built-in operations work offline without registration. Optional adapters run installed third-party programs locally, never download binaries for you, and store data in a private temporary workspace cleaned after processing.

- Typical maximum: 50 input files / 40 MB combined, with smaller limits per tool.
- Office conversion: **1–8 files**, 40 MB input, 250 pages per generated PDF, 500 pages and 80 MB of combined output; 1 input → 1 PDF, multiple inputs → ZIP.
- DOCX/Excel/PowerPoint rendering can change fonts, layout and pagination. Always inspect resulting PDFs before submitting important documents.
- Local server binds only `127.0.0.1`, with per-session token and origin checks, no telemetry, no cloud storage. External tools are **not** a hardened sandbox: avoid untrusted documents.

### Build and contribute

Go 1.26+; first build needs the Go dependencies. Windows:

```powershell
go mod tidy
go test ./...
go vet ./...
go build -o OpenDeskFree_Portable.exe ./cmd/opendesk
go build -ldflags "-H=windowsgui" -o cmd/setup/payload/OpenDeskFree.exe ./cmd/opendesk
go build -ldflags "-H=windowsgui" -o OpenDeskFree_Setup.exe ./cmd/setup
```

[Build workflow](.github/workflows/build-windows.yml) · [Contributing guidelines](CONTRIBUTING.md) · [Project roadmap](docs/ROADMAP.md) · [External integrations and licenses](docs/LICENSES_AND_INTEGRATIONS.md)

Community contributions, translations and test reports are welcome. Never attach personal documents to public issues. A successful GitHub Actions build does not prove third-party conversion accuracy on a user's PC.

**License:** MIT for OpenDesk's original code. Independent tools retain their own licenses and trademarks.

### Runtime and reliability hardening (V0.5.1)

- The actual compiled Windows Portable EXE is now started in CI, with its browser launch disabled only for testing. CI probes the live local HTTP server and exercises an authorized file-processing request plus forbidden/malformed requests.
- Two document operations can run at the same time; additional calls receive HTTP 429 instead of accumulating unbounded jobs in memory.
- Built-in PDF/image operations now receive the client request context, allowing cancellation to propagate where supported by pdfcpu.
- Download filenames use standards-aware Unicode content disposition and client parsing. These improvements do not change the no-cloud architecture.

### Stateless PDF engine

OpenDesk disables pdfcpu's persistent configuration directory at process initialization. Document operations use pdfcpu's built-in default PDF configuration and no longer create or update a shared per-user `pdfcpu/config.yml`. This avoids first-run config write contention on Windows and does not touch existing pdfcpu settings. Custom user-installed PDF fonts/configuration from that directory are deliberately not used by these built-in operations.
