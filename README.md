# OpenDesk Free

**Free, privacy-first document tools — multilingual translation and layout preservation are our development priorities.**

[Windows downloads](https://github.com/black11023/opendesk-free/releases/tag/v0.6.7-preview.1) · [Roadmap](docs/PRESERVETRANSLATE_SRS_V1.md) · [Report an issue](https://github.com/black11023/opendesk-free/issues) · [Contributing](CONTRIBUTING.md) · [MIT License](LICENSE)

## Latest interface — v0.6.7 Preview 1

![OpenDesk Free v0.6.7 current English-language interface](assets/desktop-preview.png)

*Actual English-language Chromium screenshot from the compiled application, captured in GitHub Actions. The workflow refreshes this image when the tested UI changes on `main`. Screenshot evidence is not a claim that unfinished functions are available.*

The v0.6.7 workbench is deliberately focused: **FormGuard**, image OCR, and scanned-PDF OCR. Generic PDF merge/split/rotate, image-to-PDF, miscellaneous conversions and the external-app directory are **no longer shown in the main interface**. Older processing routes remain temporarily for compatibility.

### What works today?

| Feature | Availability | What it does |
| --- | --- | --- |
| **FormGuard — document preflight** | **Built-in** | Read-only structural checks for DOCX, XLSX and PPTX; produces a per-file CSV report with original SHA-256, formula/number/merged-range estimates and risk indicators |
| **OCR for scanned images** | Optional local engine | PNG/JPG to TXT or searchable PDF; requires independently installed Tesseract and the relevant language data |
| **OCR for scanned PDFs** | Optional local engines | Scanned PDF to searchable PDF; requires independently installed OCRmyPDF and Tesseract |
| **PreserveTranslate — multilingual document translation** | **In development; not available yet** | Planned translation of Word, Excel, PowerPoint and PDF with original-format checks, protected values/formulas and reviewable differences |

**Important:** The 46 candidate language tags in the source code are not 46 working translation models. No translation models or document-translation workflow are bundled with v0.6.7. FormGuard checks *structure*; it does **not translate** documents or certify layout fidelity.

### Download and use

1. Visit the [v0.6.7 Preview 1 release](https://github.com/black11023/opendesk-free/releases/tag/v0.6.7-preview.1) and choose the Windows x64 **Setup** or **Portable** executable.
2. Compare its SHA-256 against the release's `SHA256SUMS.txt`.
3. Start OpenDesk. The interface defaults to English; you may switch it to Vietnamese.
4. Select **FormGuard** to inspect a Word, Excel or PowerPoint document, or use OCR after installing the optional local tools described in the [OCR setup guide](docs/OCR_LOCAL_SETUP.md).

Files are processed on the user's computer through a service bound to `127.0.0.1`. The built-in checker does not send documents to a translation cloud or change source files. OCR engines are external optional installations, not included in the Setup. FormGuard supports up to eight Office files in one preflight task; combined upload and per-part resource limits also apply.

### Windows download security

> **Unsigned preview — antivirus incident still open.** Chrome previously blocked an older `0.6.3` Portable download as a suspected virus. See [security issue #14](https://github.com/black11023/opendesk-free/issues/14). The v0.6.7 CI checks executable hashes and scans both exact release EXEs with current ClamAV definitions, including an EICAR-positive-control check, but that does **not** guarantee acceptance by Chrome or Microsoft Defender. Do not disable antivirus or bypass a security warning to run this software. Publisher signing is tracked in [issue #17](https://github.com/black11023/opendesk-free/issues/17).

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
