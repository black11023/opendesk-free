# OpenDesk Free v0.6.3 Preview 1 — Office batch reliability

- New optional Office batch → PDF + report (2–8 files). One ZIP contains successfully converted PDFs and a CSV with each file's PASS/FAIL status, original bytes, SHA-256 and error description. An invalid file does not discard successful conversions; a fully failed batch provides a report-only ZIP.
- Existing standard Office conversion remains all-or-nothing. Third-party LibreOffice must still be installed separately.
- External LibreOffice / OCRmyPDF / Tesseract tasks share one heavy-conversion admission slot. Concurrent conflicting requests receive bilingual HTTP 429 with Retry-After; separate Go built-in tasks remain usable.
- The GitHub public release is now created only **after Windows, actual Office/OCR, and Chromium workflows all PASS on the exact latest main commit**; executable SHA-256 is checked before publishing. Previous release logic did not wait for all three.
- Windows EXE/Setup, real Linux LibreOffice/OCR, Chromium interaction and Go tests must pass. Unsigned Preview; no guarantee of a hard OS-wide RAM limit or faithful rendering of all documents.
