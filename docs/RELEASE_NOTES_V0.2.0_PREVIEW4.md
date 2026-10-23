# OpenDesk Free v0.2.0 Preview 4 — Selected PDF pages

**Recommended preview for Windows x64** (supersedes v0.2.0-preview.3).

### New features
- Enter PDF page selections such as `2`, `2,5,10-12`, `7,2,5` or descending `6-4`. Pages use 1-based numbering.
- Extract selected pages into **one PDF** (default) or into **separate PDFs in a ZIP**.
- Merge PDFs by choosing pages **per source file**, using **↑/↓** to reorder source files.
- English and Vietnamese UI; both languages explain page ranges and output mode.
- Several-hundred-page input PDFs supported, provided the combined uploads fit 40 MB and the selected output is at most 300 pages (100 for separate ZIP).
- Invalid/out-of-range pages are rejected without changing originals.

### Test gates
- Bilingual UI translation checks, JavaScript syntax, Go unit and PDF fixture tests (including selected content and page order), go vet, Windows Setup and Portable EXE builds.
- OCR remains pending. Preview EXE unsigned; Windows install/uninstall still needs manual user acceptance testing.
