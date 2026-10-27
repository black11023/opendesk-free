# OpenDesk Free v0.6.5 Preview 1 — Easier for first-time users

- **English & Tiếng Việt**: every tool now displays a simple three-step workflow and its expected output, fully offline.
- File picker only suggests accepted extensions for each ordinary conversion. When users drag or choose an unsupported extension, they see an inline bilingual warning and the Process button is disabled.
- **Exceptions for audits:** PDF batch audit and resilient Office batch still accept invalid files and report each as FAIL in their generated CSV.
- Office batch is listed under the Office filter, not generic Files.
- Security/privacy: same local-only processing, no sign-up, no payment, no telemetry, source documents unchanged.
- Tests: Windows installer/upgrade/uninstall, real third-party Office/OCR tests and actual Chromium bilingual wrong-file/recovery exercises required for release.

**Preview / unsigned.** Core tools work offline. LibreOffice, Tesseract and OCRmyPDF are independent optional installations, not bundled; Windows 10/11 end-user visual acceptance and accuracy audits remain necessary.
