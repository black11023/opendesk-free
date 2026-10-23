# OpenDesk Free v0.5.0 Preview 1 — Office → PDF

- New **optional** local LibreOffice adapter: DOC/DOCX, XLS/XLSX, PPT/PPTX, ODT/ODS/ODP.
- One input produces one PDF; up to eight inputs produce a ZIP of PDFs.
- Browser interface detects whether LibreOffice is available and disables the operation otherwise.
- Source files are not altered. Processing uses isolated temp folders, fixed subprocess arguments, timeouts and PDF page validation.
- Limits: 8 files, 40MB input, 250 pages/PDF, 500 pages total, 80MB result.
- English-first international README and updated Office installation guide.
- CI tests verify the contract with a mock LibreOffice runner and build unsigned Windows executables.
- **LibreOffice is NOT bundled**. Real-world conversion fidelity, SmartScreen reputation and native installation/uninstallation need real PC acceptance tests. Preview, not Stable.
