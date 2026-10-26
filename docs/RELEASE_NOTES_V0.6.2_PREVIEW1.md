# OpenDesk Free v0.6.2 Preview 1 — PDF batch auditing

- New fully built-in **PDF batch audit → CSV** operation: inspect 1–50 PDF files, up to 40 MB combined, entirely locally and without changing originals.
- One report row per source file: index, sanitized name, PASS/FAIL, PDF page count, original bytes, SHA-256 and clear invalid-file reason. Invalid files appear in the report instead of disappearing.
- Excel-compatible UTF-8 CSV with BOM; neutralize dangerous formula-like filenames.
- Browser end-to-end test verifies authentic PDF audit CSV downloads with both valid and malformed inputs.
- Carries forward actual real LibreOffice PowerPoint/Word/Excel + OCR integration, Chromium language/PDF workflows, RSS and Go memory stress reporting, full Windows Setup installation/upgrade/uninstallation regression tests.
- Unsigned Preview: audit does not check malware, electronic signatures, document layout fidelity or source authenticity.
