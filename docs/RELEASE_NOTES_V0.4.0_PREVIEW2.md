# OpenDesk Free v0.4.0 Preview 2 — Scanned PDF OCR

- Added **scanned PDF → searchable PDF** via optional local OCRmyPDF (not bundled). Tesseract and language data also required.
- Input: exactly 1 valid PDF, maximum 100 pages and 40 MB. Output PDF checked against original page count; maximum 80 MB.
- Uses --skip-text: pages that already have text are skipped, preserving existing text. Never overwrites original PDF.
- No shell, no remote API. Uses private local temporary files deleted after completion or error.
- English / Tiếng Việt UI, optional OCRmyPDF detection + status, no fake enabled buttons when dependencies are missing.
- Regression tests include explicit mocked command arguments, malformed input, failed processing, invalid outputs and temporary data removal.
- Does not include OCRmyPDF, Ghostscript, Tesseract binaries or models. Real-world OCR accuracy and Windows installer validation remain pending; unsigned Preview only.
