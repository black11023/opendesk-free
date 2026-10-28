# OpenDesk PreserveTranslate + FormGuard — SRS V1.0 (PROPOSED)
Date: 2026-10-28 | Repository baseline: `400589dfd425b5a8a2289372bdeaea855a3feea4` (V0.6.6 Preview 1)
Status: **ROADMAP / prototype-in-progress — no end-user translation feature shipped**
Goal: free, privacy-first, layout-aware multilingual document translator with measurable preservation of document structure and critical values. This is NOT a claim that format-preserving translation is unique: DeepL, Google and Microsoft already offer document translation with some formatting retention.

## 1. User problem and differentiation
Users need accurate translations of Word/Excel/PowerPoint/PDF documents that remain editable and usable, without breaking formulas, amounts, table layouts, financial quantities, IDs or organization templates. People on slow PCs and small budgets need optional offline engines and reliable failure reports. Our differentiation is **local processing + per-field immutability + visible changes/risks + user-confirmed review**, rather than "PDF translate" alone.

Reference capabilities already available: DeepL file translation, Microsoft Translator Document Translation, Google Translate document upload, Argos/LibreTranslate offline translation, OCRmyPDF/Tesseract. Market positioning must not assert competing software lacks an already documented feature.

## 2. Non-negotiable invariants (P0)
- Always retain the exact original bytes and report its SHA-256. Generate a new file; never overwrite or delete any source.
- Show exactly which content was translated and which was intentionally protected/untranslated; do not invent missing text or quantities.
- Never alter workbook formulas, numeric and date cells, defined names, relationships, merges, hyperlinks, hidden rows/columns, sheet protection, styles, print areas, page settings or embedded objects without an explicitly verified operation contract. On unsupported features report BLOCKED rather than silently mutate.
- Use binary-exact comparisons of unaffected OOXML ZIP **entry contents** where feasible. Repacking ZIP may change compressed bytes even when document content is unchanged.
- DOCX: preserve paragraph/table structure, section settings, headers/footers, numbering/styles, bookmarks, relationships and anchors; translating rich text spanning multiple runs requires a reversible mapping. Unsupported smart fields, tracked changes, forms, drawings and content controls must be detected and warned/blocked.
- XLSX: translate explicitly selected human-language text cells only; exclude formulas, cached results, numbers, dates, IDs, cell references, sheet names (unless user opts in with full reference rewrite), validations and calculation metadata.
- PPTX: handle shape text while leaving geometry, animations, masters, charts, notes and relationships unchanged unless supported explicitly.
- PDF with selectable text: preserve pages and visual positions where achievable; measured rendering/overflow detection required. Scanned PDFs need OCR bounding boxes and independent quality checks. PDF may require reconstructed output, not pixel-identical pages.
- Digitally signed source files: never claim the translated copy retains original cryptographic signatures. Preserve signed original unchanged; generate separately named translated *reference* output and warn that signatures/official validity require independent sign-off.
- Absolute or pixel-perfect layout equivalence cannot be promised for all languages. Longer text, font substitutions, embedded images, RTL scripts, rotated text and scan artifacts can force reflow. Fail visibly if tolerance is exceeded.
- No automatic cloud upload, paid API key, subscription, silent model download, AV exclusion or third-party program install. Model packages are optional, user-authorized, checksum/license verified.

## 3. User experience
1. Select 1–N local files and source/target languages; show explicit model availability (offline/optional) and expected resources.
2. PRECHECK: inventory format, encryption/signatures/macros, tables, formulas, numeric literal text, images, scan/OCR, fonts and protected fields. Show GREEN/AMBER/BLOCKED readiness before translation.
3. Select protected tokens, glossary, domain presets (general; education; business; construction/tender optional); allow per-sheet/range and paragraph inclusion.
4. Preview side-by-side original/translated layout, highlight overflow and unchanged regions, glossary substitutions and uncertain OCR. No AI changes to locked values.
5. Translate with progress, cancel, bounded memory and a per-file PASS/PARTIAL/BLOCKED result; provide localized actionable errors.
6. Download editable native document where feasible + optional bilingual comparison + full audit (CSV/XLSX/PDF) with source file, sheet/cell or paragraph/shape or page/box, source/translated text, locks, validation flags, SHA and engine/model version. Export **all rows**, no 50-row cap.
7. For high-risk rows, require explicit user confirmation before exporting a final document; retain drafts and support retry without mutating the original.

## 4. Translation backend (pluggable, not bundled without a legal/security review)
- Keep existing Go 1.26 loopback-only app and CSRF/Host/Origin protections as the only UI/HTTP gateway.
- Parse/patch DOCX/XLSX/PPTX OOXML in Go, with a package-map/anchor preservation layer and independent no-change oracle.
- Optional local engine adapter: Argos Translate (Python, MIT/CC0 library; independently validate each language-model package license) or another tested license-compatible engine. Official index includes en->vi and vi->en packages. First supported pair **English ↔ Vietnamese**, with English/Chinese/Japanese/Korean/French after model/fixture acceptance. Adding models does not change the base EXE or require admin privileges.
- Prefer local subprocess started using fixed allowlisted args and isolated temporary folders, timeout/cancel and process-tree kill; don't use shell command construction or external translation sites.
- Optional device acceleration after CPU baseline verified. RAM/performance targets must be measured on actual low-end Windows PCs before publication. No invented SLA or language quality scores.
- Avoid shipping Meta NLLB-200 model by default: its model card license is CC-BY-NC-4.0 and the free-for-everyone project's lawful downstream uses should not be restricted by accident. Don't bundle AGPL LibreTranslate server in MIT package without obligations review.

## 5. Phased delivery and scope
- **G0 / foundational prototype (FIRST):** test-driven numeric/identifier placeholder protection `internal/preservetranslate`, with fail-closed restore; OOXML inventory/readiness plan; documentation, no UI "translated" badge.
- **G1 / first real translator:** DOCX English<->Vietnamese; offline Argos optional engine; preserve OOXML structures; glossary, compare text, audit and failure diagnostics.
- **G2 / XLSX:** per-sheet/range selection; rich/shared strings; strict formula/numeric/merged/style/print rules; specialized Excel audit. Do not modify `.xlsm` or encrypted workbooks.
- **G3 / PPTX:** slide shape text/notes where proven; layout overflow check; original slide count.
- **G4 / PDFs:** selectable text first; scan/OCR later; page visual diff, font/overflow warnings; signed documents get a separate, explicitly unofficial translation.
- **G5 / multi-language & batch:** glossary/translation-memory packs, optional language additions, bulk tasks and resumable jobs.
- Existing basic PDF tools remain available but are demoted from product differentiation; do not delete features without explicit approval.

## 6. Acceptance matrix (must be automated and manually reproducible)
- P0 tokens: missing/duplicate/reordered/modified marker fails closed. Source bytes, protected numeric/code literals remain unchanged.
- DOCX: paragraphs/runs, mixed bold/italic, tables, section breaks, headers/footers, Unicode Vietnamese, images and links; original file SHA unchanged. Strict OOXML relationships/inventory parity, visual diff with page screenshots when layout engine is available.
- XLSX: 20-sheet fixture, 300 formulas, hidden cells, merged ranges, data validations, print areas, numbers/dates/currency and source comments; verify exact formula lists and protected value hashes, no Excel repair prompt.
- PPTX: slides/masters/images/shapes and geometry, table and text box overflow, slide-count invariant.
- PDF: native 1-column, 2-column, complex tables, scan with OCR errors, rotated pages and digital signatures. Error/partial labels required where layout cannot be faithfully maintained.
- Security: ZIP bomb/XXE/path traversal protections, limits on XML size, no formula injection in CSV reports, malicious macros blocked and source documents never executed, keep process within user privilege.
- Evidence: per-test PASS/FAIL, hashes, source-to-output map, layout screenshots where supported, round-trip open in real LibreOffice/Windows Office fixtures; unsupported cases BLOCKED.
- Release: Node/Go/Windows/Chromium/Office/OCR CI unaffected; new translation test gates PASS; malware-positive-control and exact EXE scans PASS; real user-PC smoke and licensing check before marking translation as BUILT-IN.

## 7. Other innovations worth separate epics (do not dilute P0)
1. **FormGuard:** automatically detect critical fields (amounts, dates, IDs, formulas, placeholders), then audit every change and flag unsupported form transformations.
2. **Smart bilingual forms:** fill existing blank templates using a user-reviewed field mapping; never fabricate or submit signatures/data.
3. **Bilingual Diff & Provenance:** linked side-by-side passages with exact locations and change/confidence reasons, human edits and exportable audit.
4. **Glossary + Translation Memory:** user-owned packs, priority rules and domain dictionaries applied consistently across batches, offline.
5. **Layout-preserving OCR reconstruction:** recover structured paragraphs and tables from scans into editable drafts, explicitly showing uncertain areas.
6. **Privacy redaction:** true PDF removal of personal data (not merely black rectangles), with read-back proof and an audit log.
7. **Trusted table extraction:** OCR/source cell provenance to Excel, explicit human signoff of ambiguous rows.
8. **Workflow recipes:** repeatable document pipelines, dry run and cancellation, no destructive action.

## 8. External evidence (verified 2026-10-28)
- DeepL document translation and known layout shifts: https://www.deepl.com/en/features/document-translation , https://support.deepl.com/hc/en-us/articles/21315109374748-Edit-your-file-translation
- Microsoft Document Translation / original formatting, glossary: https://www.microsoft.com/en-us/translator/business/document-translation/
- Google Translate document formats and scan limitation: https://support.google.com/translate/answer/2534559
- Google Cloud complex PDF layout warning: https://docs.cloud.google.com/translate/docs/advanced/translate-documents
- Argos offline Python library and license: https://github.com/argosopentech/argos-translate
- Argos model catalog (en_vi/vi_en): https://data.argosopentech.com/argospm/v1/
- NLLB model license (not default): https://huggingface.co/facebook/nllb-200-distilled-600M
- LibreOffice import/export fidelity warning: https://help.libreoffice.org/latest/en-GB/text/shared/guide/ms_import_export_limitations.html
