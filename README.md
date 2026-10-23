# OpenDesk Free 🧰

**Free, offline-first document utilities for everyone. / Công cụ xử lý tài liệu miễn phí dành cho mọi người.**

![OpenDesk Free desktop preview](assets/desktop-preview.png)

> **Status: v0.2.0-preview.1 development preview.** The app builds for Windows; Windows runtime acceptance testing is still required before labeling it a stable release.

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

Go >=1.25, pdfcpu v0.16.3 Apache-2.0 dependency (downloaded during build):

```powershell
go test ./...
go build -o OpenDeskFree_Portable.exe ./cmd/opendesk
go build -ldflags "-H=windowsgui" -o cmd/setup/payload/OpenDeskFree.exe ./cmd/opendesk
go build -ldflags "-H=windowsgui" -o OpenDeskFree_Setup.exe ./cmd/setup
```

The installer sources are Windows-specific. On Linux, cross-compile using `GOOS=windows GOARCH=amd64 CGO_ENABLED=0` for both binaries.

## Privacy and security

- Runs solely on loopback (`127.0.0.1`). Not available to the internet or other computers.
- Files are processed in memory, with no persistent database and no upload to external providers.
- POST operations require a secret token and Origin check.
- No automated file deletion, renaming or modification of original uploads.
- An operation may consume up to tens of megabytes of RAM and processing time depending on input.
- Review files before using them in other applications. This preview is not a substitute for antivirus or document validation.

## Contribute

Please use Issues to report bugs, including Windows version, input format (do not include confidential documents), steps to reproduce, expected and actual result. See `CONTRIBUTING.md` and `docs/ROADMAP.md`.

## License

MIT. Copyright © 2026 OpenDesk Free contributors.

### International users / Người dùng quốc tế
English is selected by default for new users. Switch to **Tiếng Việt** in the Language dropdown. Each browser saves its own preference on this PC. No account required.
