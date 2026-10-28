# OpenDesk Free

**Free, privacy-first document tools — multilingual translation and layout preservation are our development priorities.**

[Windows downloads](https://github.com/black11023/opendesk-free/releases/tag/v0.7.0-preview.1) · [Roadmap](docs/PRESERVETRANSLATE_SRS_V1.md) · [Report an issue](https://github.com/black11023/opendesk-free/issues) · [Contributing](CONTRIBUTING.md) · [MIT License](LICENSE)

## Latest interface — v0.7.0 Preview 1

![OpenDesk Free v0.7.0 current English-language interface](assets/desktop-preview.png)

*Actual English-language Chromium screenshot from the compiled application, captured in GitHub Actions. The workflow refreshes this image when the tested UI changes on `main`. Screenshot evidence is not a claim that unfinished functions are available.*

The v0.7.0 workbench focuses on **FormGuard**, **limited DOCX offline translation**, image OCR, and scanned-PDF OCR. Generic PDF merge/split/rotate, image-to-PDF, miscellaneous conversions and the external-app directory are **no longer shown in the main interface**. Older processing routes remain temporarily for compatibility.

### What works today?

| Feature | Availability | What it does |
| --- | --- | --- |
| **FormGuard — document preflight** | **Built-in** | Read-only structural checks for DOCX, XLSX and PPTX; produces a per-file CSV report with original SHA-256, formula/number/merged-range estimates and risk indicators |
| **OCR for scanned images** | Optional local engine | PNG/JPG to TXT or searchable PDF; requires independently installed Tesseract and the relevant language data |
| **OCR for scanned PDFs** | Optional local engines | Scanned PDF to searchable PDF; requires independently installed OCRmyPDF and Tesseract |
| **Translate DOCX — direct offline model beta** | **Optional, limited** | If Python + Argos Translate and a direct language model are installed: translate simple single-run DOCX paragraphs and export an auditable ZIP; complex formatting is BLOCKED, no layout guarantee |
| **Translate XLSX — plain text cells beta** | **Optional, limited** | With installed direct Argos model: translate plain shared/inline Excel strings, without changing formulas, numbers, styles, merges or workbook references; audit every edited string and cell address, fail-closed on complex structures |
| **PreserveTranslate — full multilingual translation** | **In development** | DOCX rich formatting, complex Excel, PowerPoint, PDF and more complete preservation/automated quality checks are not yet available |

**Important:** The 46 candidate language tags in the source code are not 46 working translation models. No language models or Python runtime are bundled. Limited DOCX/XLSX betas run with a separately installed Python + Argos Translate engine. A new in-app manager can explicitly download/import one pinned, verified English → Spanish language pack on PCs that already have that engine; it never downloads anything in the background. FormGuard checks *structure*; it does **not translate** documents or certify layout fidelity.

### Download and use

1. Visit the [v0.7.0 Preview 1 release](https://github.com/black11023/opendesk-free/releases/tag/v0.7.0-preview.1) and choose the Windows x64 **Setup** or **Portable** executable.
2. Compare its SHA-256 against the release's `SHA256SUMS.txt`.
3. Start OpenDesk. The interface defaults to English; you may switch it to Vietnamese.
4. Select **FormGuard** to inspect a Word, Excel or PowerPoint document, or use OCR after installing the optional local tools described in the [OCR setup guide](docs/OCR_LOCAL_SETUP.md).

Files are processed on the user's computer through a service bound to `127.0.0.1`. The built-in checker does not send documents to a translation cloud or change source files. OCR engines are external optional installations, not included in the Setup. FormGuard supports up to eight Office files in one preflight task; combined upload and per-part resource limits also apply.

### Windows download security

> **Unsigned preview — antivirus incident still open.** Chrome previously blocked an older `0.6.3` Portable download as a suspected virus. See [security issue #14](https://github.com/black11023/opendesk-free/issues/14). The v0.7.0 CI checks executable hashes and scans both exact release EXEs with current ClamAV definitions, including an EICAR-positive-control check, but that does **not** guarantee acceptance by Chrome or Microsoft Defender. Do not disable antivirus or bypass a security warning to run this software. Publisher signing is tracked in [issue #17](https://github.com/black11023/opendesk-free/issues/17).

**Automated evidence:** [Windows installer + antivirus](https://github.com/black11023/opendesk-free/actions/runs/40630593417) · [Real Chromium UI + screenshots](https://github.com/black11023/opendesk-free/actions/runs/40630593481) · [Office/OCR integration](https://github.com/black11023/opendesk-free/actions/runs/40630593497). These are automated checks on the release source; an individual user's Windows PC remains a separate acceptance environment.

### Project direction

OpenDesk's primary goal is a **free, offline-capable, multilingual document workflow** that reduces the risk of damaging original templates. The development sequence is:

1. **FormGuard:** preserve documents, inventory OOXML structures, detect risky fields and export clear reports — structural preflight is implemented.
2. **PreserveTranslate:** connect tested local language models, detect language, translate actual DOCX contents, protect quantities/identifiers and verify output.
3. **Excel, PowerPoint and PDF:** add per-format fidelity checks, side-by-side review, OCR and layout-overflow warnings; never silently change formulas, amounts or signatures.
4. **Language packs and automation:** multiple destinations, glossaries, translation memory and auditable batches only after actual model and document tests succeed.

See the [technical specification](docs/PRESERVETRANSLATE_SRS_V1.md), [multilingual specification](docs/MULTILINGUAL_SPEC_V1.md) and [development tracker](https://github.com/black11023/opendesk-free/issues/19). Roadmap features must not appear as working buttons until genuinely available.

### For developers

The application uses Go. Tests and Windows packaging run on GitHub Actions.

```bash
go test ./...
go vet ./...
go build -o OpenDeskFree ./cmd/opendesk
```

Windows Setup also builds from `./cmd/setup`. See [Windows CI](.github/workflows/build-windows.yml), [contribution guidance](CONTRIBUTING.md), and [third-party license notes](docs/LICENSES_AND_INTEGRATIONS.md). Don't include private documents in public bug reports.

**License:** MIT for original OpenDesk Free code. Optional third-party software and language-model packages retain their individual license terms.

### V0.7.0 Preview — optional local DOCX translation beta

- One DOCX per task, up to 12 MB. Only paragraphs with a single ordinary text run can be translated without losing bold/italic markup; complex runs, fields, headers/footers and unsupported document structures are blocked.
- User-selected source/target pair must be installed **directly** in local Argos Translate; no silent pivot via English, cloud API, model download or signup.
- Numeric/date/material-code literals are protected by FormGuard placeholders. If a language model changes/omits a placeholder, the job fails without returning a partial document.
- Output ZIP contains translated editable DOCX, every translated paragraph in a full CSV audit and SHA-256 of source/output. Original bytes are never modified.
- **Limitations:** Argos and language model packages are optional and require separate installation. Model-package integrity/licensing and actual translation quality need independent review; page layout, line wrapping, rich formatting and official translations are not certified. Do not use output as an approved contract without review.

See [offline translation setup](docs/ARGOS_OFFLINE_TRANSLATION.md).

### V0.7.0 Preview — verified local language-pack installer (partial #29)

- English/Vietnamese settings now display a strictly approved direct **English→Spanish Argos model v1.0** (87,503,191 bytes), its vendor URL and pinned SHA-256. User consent and explicit button are required before any network access.
- Two real workflows: **Download & install** the exact approved model or **Verify & import** an existing offline `.argosmodel`. Both use the same signed-off digest, safe ZIP inspection, archive/path/size checks, metadata language match and actual post-install Argos translation probe.
- Model download is initiated by the user only; no cloud document uploads. The original DOCX and installed models are not removed during app updates. Failed verification never installs a model.
- **Partial usability**: The Argos inference engine still requires the user's existing Python/Argos installation; installing that engine without Python and model licensing/redistribution research remain issue #29. This release does **not** bundle models, auto-install Python, certify third-party model licensing, support every listed language, or preserve complex document layouts. Do not disable antivirus warnings to use unsigned previews.

### V0.7.0 Preview — limited native Excel translation

- Adds a new **Translate Excel labels** task alongside Word. It translates plain shared-string / inline-string worksheet labels in editable XLSX, never source workbook in-place.
- Native workbook formulas, raw numbers, cell addresses, merge ranges, styles and untouched ZIP entries remain unchanged. Rich text, charts, hyperlinks, macros, external links, unsupported sheet controls and complex metadata cause a safe BLOCKED response rather than partial export.
- Export ZIP contains an editable translated XLSX, a complete cell-reference audit CSV and source/output SHA-256. Source workbook preserved; translation/visual overflow still needs manual review. At most 600 plain text labels / 125,000 characters / 12 MB input per task.
- Uses existing direct offline Argos language models and the opt-in pack installer introduced in V0.6.9. No paid API or cloud document upload, but the inference runtime currently requires Python+Argos installed separately. This is not pixel-perfect spreadsheet translation or a fully standalone runtime.
