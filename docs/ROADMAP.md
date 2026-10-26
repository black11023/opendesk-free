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

### V0.4 Preview 2 — OCR scanned PDF

- Optional external OCRmyPDF adapter discovered on PATH; requires separately installed Tesseract with selected traineddata and OCRmyPDF upstream dependencies.
- Actual PDF input: preserve original, fail closed on malformed PDF, no shell, private temporary files removed, check page count of final result, limit pages/size/time.
- Automated tests use an injected fake runner to verify arguments, rejection, output verification and temp cleanup; **this is not OCR recognition accuracy validation**. End-user PDF OCR requires live Windows acceptance with OCRmyPDF installed.

### V0.5 Preview 1 — LibreOffice PDF adapter

- Optional local conversion for Word, Excel and PowerPoint families. Convert files separately in a private temporary profile; original files remain untouched.
- Verified output structure and page counts with strict resource limits; mock-converter unit tests cover errors and cleanup, not real-world layout fidelity.
- Accessibility and English-first onboarding improved, README rewritten, transparent integrated/optional/external/planned statuses.
- Next: real native Windows install/uninstall and LibreOffice fixture rendering, error tracing and stronger multilingual docs.

### V0.5.1 Runtime verification

- Build Windows GUI EXE then run it from CI on Windows (binds only 127.0.0.1), test /health, authorized local POST, denial of unauthenticated and cross-origin POST, then clean shutdown.
- Cap concurrent document processing to 2 and return 429 for extra jobs; propagate HTTP cancellation into built-in PDF/image processing. Add unit/negative tests.
- Check standards-compliant Unicode output filenames and client-side parsing.
- Remaining: actual Windows installer install/upgrade/uninstall on a user machine; actual LibreOffice/OCRmyPDF rendering and visual UX review. Not STABLE.

### v0.5.2 Windows installer acceptance

- GitHub-hosted Windows smoke: run actual Setup EXE quiet installation and upgrade using isolated LOCALAPPDATA/APPDATA; assert files, shortcut and HKCU Uninstall registration.
- Launch installed EXE and request local health/quit; quiet uninstall; confirm registry/shortcut/executable removal and preservation of independent downloads and user-created files.
- Failure of any assertion blocks Release. Installer does not recursively delete the installation directory.
- Remaining gate for Stable: user-visible dialog acceptance and Windows 10/11 real-machine installation/upgrades, security and antivirus reputation review, and verification of optional external engines.

### V0.6 Preview — professional workspace, real document tests, memory

- Bilingual searchable built-in tools and responsive category filters; progress/busy state and keyboard accessibility.
- Batch rotate up to 10 PDFs into a single ZIP, with maximum 300 total pages and 80 MB output. Reject corrupt documents.
- Soft Go heap budget 512 MiB and reject memory-intensive requests if Go HeapAlloc exceeds 300 MiB; still requires real high-volume profiling and external-tool memory controls.
- Linux real LibreOffice DOCX/XLSX-to-PDF content checks and Tesseract/OCRmyPDF text extraction from actual generated scanned images. These are real dependency tests, not mocks.
- Windows EXE and Setup install-upgrade-uninstall CI retained. Platform behavior remains Preview until real end-user installations and rendering reviews.

### V0.6.1 Preview 1 — browser and real PowerPoint gate

- Linux Chromium automation: English/Vietnamese interaction, operation search and categories, page-specific PDF download, multi-file PDF ZIP, screenshots and per-run process RSS report.
- Real two-slide PPTX rendering in LibreOffice with strict slide count and extracted text checks; three-format DOCX/XLSX/PPTX batch acceptance.
- Repeated batch PDF memory regression test (retained Go heap after GC); reject combined upload size before all files are buffered.
- Still missing: real Windows interactive visual walkthrough, complicated fonts/charts/animations, low-memory-PC and large-file benchmarking, external OCR/Office subprocess RSS enforcement.

### V0.6.2 Preview — read-only PDF batch audits

- Per-file CSV evidence for up to 50 inputs / 40 MB combined: source index, safe filename, PASS/FAIL, page count, original bytes, SHA-256, explanation for invalid structures. No source is altered.
- Security: protect CSV output against spreadsheet-formula injection; retain UTF-8 BOM for Excel and exact input order.
- Go negative unit tests and real Chromium download test are release gates, in addition to Windows Setup and real Office/OCR integration.
- Not a virus scan, legal authenticity audit or certification of an entire PDF's visible content.

### V0.6.3 Preview 1 — Office batch status and strict release gate

- External Office/OCR jobs limited to one active at a time (HTTP 429 on conflict); retains two overall Go processing slots. Not a hard memory cap; no claim of taming LibreOffice child-process RSS on Windows.
- Resilient 2–8 source Office conversion with per-file CSV PASS/FAIL and SHA-256 in one ZIP, alongside successful PDFs; existing ordinary Office conversion contract is preserved.
- Dedicated workflow_run publisher requires three independent **push-to-main** workflows all successful on exactly the current main SHA; verifies binary hashes before release. Pull-request runs cannot satisfy the release gate.
- Remaining: native OS process groups / child-tree termination and hard per-process RAM caps, user-PC validation, portable desktop UX polish, full PDF metadata report.

### V0.6.4 — cancellable renderers and real browser cancellation tests

- A local Cancel button aborts browser processing without closing OpenDesk or overwriting source documents; request cancellation is carried to backend command contexts.
- Optional Office/OCR subprocesses launch in dedicated process groups. Linux/macOS cancels an entire process group; Windows attempts taskkill /T /F of the process tree, falling back to killing the parent. This is **best effort**, not a hardened sandbox or guaranteed OS-level memory cap.
- A Linux regression starts a real shell and background worker, cancels and checks that the worker stops. Chromium UI exercises request cancellation and responsive reset, besides earlier PDF and bilingual operations. Windows compile/runtime installer checks retained.
- Fixed malformed Markdown feature-table line breaks in README, plus more explicit limits and documentation.
- Future: Windows real Office/Tesseract child-process exit verification on actual PC and hard memory-budgeting through OS facilities where safe.
