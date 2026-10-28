# OpenDesk Free V0.7.0 Preview 1 — Native Excel label translation beta

This release adds **Translate Excel labels** as an actual task, not a teaser.

- Supports plain shared-string labels and plain inline text in XLSX. Data and formulas are never passed to Argos.
- All untouched ZIP/OOXML parts, including styles and workbook relationships, are copied without modification.
- Rejects rich text, charts, pivot/tables, external data, unsupported text-bearing sheet parts, macros and document signatures.
- Protects numeric/code spans with FormGuard. Any missing or altered marker rejects the job, without partial output.
- Reviewable ZIP: editable XLSX, full per-label/cell audit CSV and SHA-256 manifest, without changing the input file.
- Native Argos model installation remains optional and requires separate Python+Argos. Model pack catalog is not a claim of 46 installed languages.
- This preview does not guarantee same page/print layout or machine-translation quality. Manually review Excel and DOCX results before operational use. EXEs remain unsigned; antivirus warning must not be bypassed.

## Acceptance requirements

Go safety unit tests, direct real EN→ES Argos XLSX translation, real LibreOffice opened/converted workbook, Windows Setup and Portable acceptance, Chromium bilingual UI, ClamAV with EICAR control. No merge or public release with mandatory FAIL or BLOCKED.
