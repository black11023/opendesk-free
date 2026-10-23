# OpenDesk Free v0.1.0 — Initial Preview

The first free and offline-first Windows preview of OpenDesk Free.

## Included

- 5 working tools: organize ZIP, CSV to XLSX, DOCX to text, SHA-256 duplicate report, combine text.
- Windows x64 portable executable and a per-user Windows installer.
- Vietnamese/English interface and built-in local web UI.
- Go unit tests, checksum manifest, GitHub Actions Windows build workflow.
- Public MIT source code, roadmap, privacy/security notes and publishing guide.

## Known limitations

- Windows installer compiled on Linux; no Windows GUI/install/uninstall acceptance run has been performed in this build environment.
- Executables are unsigned and may trigger Windows SmartScreen warnings.
- Documents processed in browser memory. Maximum 50 files and total 40 MB per operation.
- Only UTF-8 CSV and TXT/Markdown supported; CSV values saved as strings, not inferred numeric types.
- DOCX extraction handles document body text only, not headers, footnotes, layout or embedded images.
- PDF/OCR, Excel analysis and AI automation are not yet implemented.

This is a **preview**, not a stable or security-audited release.
