# OpenDesk Free v0.4.0 Preview 1 — Offline OCR adapter

- Add an optional Tesseract engine adapter for local OCR without paid services.
- Supports scanned PNG/JPG to UTF-8 text and searchable PDF, English and Vietnamese traineddata checks.
- Built-in engine status detection and dedicated bilingual UI; fails closed and explains missing installation.
- Add tests for text, PDF page count, language-pack validation, invalid and mislabeled images, failure handling.
- No executable or language packs downloaded or bundled; user must separately install Tesseract.
- Existing tools remain available offline without the OCR engine.
- **Not implemented:** source scanned PDF → OCR, turnkey OCR setup, confidence benchmarks. Unsigned Windows preview requires end-user install and runtime acceptance.

References: https://tesseract-ocr.github.io/tessdoc/Command-Line-Usage.html
