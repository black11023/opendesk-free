# OpenDesk Free v0.6.4 Preview 1 — Reliable cancellation

- New **Cancel processing / Hủy xử lý** button: a live local request can be cancelled without closing the tab; the button resets afterwards, keeping the selected files and original documents unchanged.
- Optional local Office/Unicode OCR/Tesseract and OCRmyPDF subprocesses are started without a shell and equipped with process-tree cancellation: isolated process groups on Linux/macOS, taskkill /T /F fallback on Windows.
- Genuine Linux subprocess regression starts a child process, cancels its context and verifies the descendant is no longer running. Real Chromium UI tests cancellation and successful recovery.
- International README feature table is repaired so distinctions between built-in, optional, planned and external link tools display correctly.
- Three-gate GitHub release continues: Windows executable and installer, real Office/OCR engines, Chromium UI; verifies EXE SHA-256 before public release.
- **Preview, unsigned**. Descendant cleanup is best-effort; it is not a guaranteed hard system RAM/CPU cap. External binaries remain separately installed, and Windows real-life third-party process-tree cleanup requires further verification.
