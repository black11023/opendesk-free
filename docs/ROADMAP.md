# OpenDesk Free roadmap — adopted 2026-10-23

This is a proposed plan, not an assurance of delivery dates. Each gate requires practical testing and evidence; no stage is marked done in advance.

| Stage | Priority | Scope | Acceptance gate |
|---|---|---|---|
| V0.1 — initial preview | P0 | Bilingual desktop GUI, offline operations (organize, CSV→XLSX, DOCX→TXT, hashes/duplicates, combine text), Windows EXE + installer, README, MIT | Go tests PASS, Windows EXE builds, malware scan, clean-install/run/uninstall manual Windows test; no claim of stability without that check |
| V0.2 — PDF/OCR | P0 | PDF selected-page merge / extract / separate-ZIP split and rotate implemented in preview; OCR Vietnamese/English, searchable PDF and richer progress still pending | Real PDF scanned/digital test set; document originals unchanged; text accuracy and page count reviewed |
| V0.3 — Excel/Word automation | P1 | Excel matching via explicitly selected keys, comparison report, highlight mismatches; DOCX table extraction; batch transforms | Fixtures with quantities, encodings, missing values and formulas; no silent number coercion |
| V0.4 — Workflow builder | P1 | Saved local recipes, previews, rollback-friendly outputs, per-step audit/logs | End-to-end recipes on multiple PCs; no destructive automation without confirmation |
| V0.5 — Accessibility + portability | P1 | Screen reader, full keyboard UX, low-RAM profiling, Linux/macOS builds | Usability tests and platform-specific installers |
| V1.0 — stable | P0 | Signed builds when feasible, versioned release notes, maintained security updates, community translations and contributions | Repeatable CI, release SHA256, zero known release blockers, real-world acceptance tests |

## Release policy

- All basic functions remain free, offline-first and usable without registration.
- No paid API keys or internet required for core workflows.
- Windows binaries distributed through GitHub Releases (not in git history).
- Document all limitations; never call a roadmap item implemented.
- Keep updates reversible and preserve users' original documents.

### PDF page selection acceptance (Preview 4)

- For a source PDF with hundreds of pages, select e.g. page `2` only; no unselected pages may enter the output.
- `2,5,10-12` and `7,2,5` must preserve requested page order in a single extracted PDF.
- Merge allows choosing pages from each source file and reordering input files with arrows.
- Single-PDF extraction and per-page ZIP extraction are separate user-selectable outputs; invalid page indices must fail with a visible error.

## OpenDesk Super Suite expansion (approved direction, NOT all implemented)

- **V0.3 P0:** professional searchable Free Tools Hub; real image-to-PDF module; module manifest/registry and user-visible integration status. Source-based CI gates.
- **V0.4 P0:** OCR for Vietnamese/English scanned images and PDFs, using Tesseract language packs with documented optional local installation; preserve page layout wherever possible. No pretend OCR.
- **V0.5 P1:** LibreOffice command-line adapters for Office PDF export if separately installed, constrained file paths and process limits; no automatic binary download.
- **V0.6 P1:** accessible image conversions, audio/video conversion using optional FFmpeg, supported secure adapters; resource budgeting.
- **V0.7 P1:** local workflow builder, saved recipes, audit, preview/undo workflows, no destructive action by default.
- **V1.0:** Windows install/uninstall acceptance tests, multilingual accessibility, project licensing/SBOM, release signatures when feasible; Linux/macOS packages after real platform QA.

**Status definitions:** BUILT-IN (available now), OPTIONAL ADAPTER (implemented but requires separately installed provider), DIRECTORY (external link only), ROADMAP (not implemented). Never represent DIRECTORY as installed or integrated.

### V0.4 OCR gate status

- Optional Tesseract installed-locally adapter, eng/vie status detection, image → TXT and searchable PDF, fail-closed for missing engine/language.
- Tesseract executable + traineddata are **not bundled** in installer. A preinstalled Tesseract is required.
- Scanned source-PDF OCR, multilingual accuracy benchmark, Windows end-user acceptance and single-click optional install are pending. Do not label OCR complete across all formats.
