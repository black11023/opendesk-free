# OpenDesk Free v0.5.1 Preview 1 — Runtime and safety

- **Windows runtime CI:** launch the compiled GUI Portable EXE on localhost and run a live HTTP smoke test for health, CSRF, Origin, file processing, invalid PDF and shutdown.
- **Resource protection:** at most 2 simultaneous document-processing requests. Additional requests receive HTTP 429 with a clear bilingual message.
- **Cancellation:** built-in Go/PDF/image tasks receive request cancellation when the client disconnects; source documents remain untouched.
- **International filenames:** use RFC-compatible Unicode download filenames for Office documents with accents and other scripts.
- No bundled LibreOffice/OCRmyPDF/Tesseract; these remain optional external adapters.
- Automated Windows build/Go tests/static analysis required before release; manual installer install/uninstall and real OCR/LibreOffice rendering still pending. Unsigned Preview, not Stable.
