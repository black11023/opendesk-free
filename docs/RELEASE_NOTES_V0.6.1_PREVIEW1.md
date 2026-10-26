# OpenDesk Free v0.6.1 Preview 1 — Real PowerPoint and browser acceptance

- Strengthened real-dependency CI with a generated, two-slide PPTX including text, shapes and an image. Assert actual LibreOffice PDF output is 2 pages with both titles; run an actual DOCX+XLSX+PPTX batch conversion.
- New Chromium end-to-end browser CI on the compiled local Go app: English/Vietnamese language switch, tool searching and filtering, PDF page selection and actual download, batch PDF rotation ZIP. Screenshots saved as CI artifacts.
- Added measured Go heap retention after repeated multi-file PDF rotation and sample app RSS measurement on Linux. Also reject aggregate oversized inputs before copying them into Go memory.
- Preserves all Windows build and real Setup install/upgrade/uninstall tests. Core functions remain free and local; optional Office/OCR binaries are not bundled.
- **Unsigned Preview.** Small-fixture RSS and binary output tests cannot guarantee a maximum RAM footprint, perfect slide layout, or full Windows GUI acceptance.
