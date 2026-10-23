# OpenDesk Free v0.3.0 Preview 1 — Tools Hub

**Development milestone, unsigned Preview, not Stable.**

## Working additions
- Images → PDF: combine up to 25 PNG/JPG images into a single PDF (one page per image), preserve order and original files; max 20 megapixels per image, 40 MB total.
- Bilingual professional tool-discovery section with search (Ctrl+K), 11 category filters and 44 curated third-party free/open-source project links.
- Explicit visual and documentary disclosure: all community directory entries are **external projects only**; not bundled, integrated or auto-installed.
- Existing PDF selected-page operations, CSV/Excel, DOCX text, hashing, organization and text combining remain.

## Build gates
- Go tests including image fixtures, PDF page count and invalid input checks.
- JS syntax, bilingual UI parity, 11 categories / external-link catalog tests.
- Windows Setup + Portable EXE compilation, SHA256SUMS and public pre-release only from the branch HEAD after successful CI.

## Not included
- OCR, external tool installations, LibreOffice/FFmpeg adapter and automated workflows: still roadmap.
- Installer run/uninstall acceptance on an actual end-user Windows 10/11 machine; executables are unsigned.

**Important:** source project identity, copyright and license remain the property of each upstream owner.
