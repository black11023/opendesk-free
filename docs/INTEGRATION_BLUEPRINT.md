# OpenDesk Free — Integration blueprint / Kiến trúc Super Suite

**Date:** 2026-10-23. **Principle:** one lightweight local desktop shell, many independently versioned tools. No promise to merge all software in existence.

## Four integration states

- **BUILT-IN:** code shipped inside the official OpenDesk executable; tests pass.
- **OPTIONAL ADAPTER:** tested command/API adapter invoking independently installed software; clear dependency check, explicit consent and user-controlled downloads.
- **DIRECTORY:** reviewed entry pointing to an independent upstream project; **not integrated**, not installed and no affiliation implied.
- **ROADMAP:** proposed and not usable yet. Never present a roadmap card as functional.

## Architecture contracts

1. Core host (Go): HTTP bound to 127.0.0.1, random per-run CSRF token, Host/Origin checks, no cloud by default.
2. Modules: capability registry with typed input/output formats, max file count/page sizes, expected dependencies, failure mode, license metadata and accessibility labels.
3. Execution: Go built-ins in-process; other vendors only via explicit opt-in, safe command argument arrays, isolated temporary workspaces and defined timeouts.
4. Workspace: user files remain unchanged; output is a new download; temporary data must be cleaned and errors must be visible and retained until acknowledged.
5. UI: English default + Vietnamese, keyboard navigation, accessible labels and visible distinction between functional tools and external directory links.
6. Safety: no arbitrary shell commands from user data, no silent install, no automatic upload or credential exfiltration, no privileged execution.
7. Release gate: automated Go tests, bilingual catalog checks, Windows EXE build, SHA-256; separate real Windows smoke/install/uninstall signoff before STABLE.

## Proposed adapters

- **Documents/OCR:** Tesseract, OCRmyPDF and LibreOffice; external dependencies, packages and licenses assessed individually.
- **Images:** in-app PNG/JPG → PDF (built-in); GIMP/Inkscape/Krita as DIRECTORY until reviewed adapters exist.
- **Media:** FFmpeg as OPTIONAL candidate for offline conversions; OBS/Kdenlive/Audacity as DIRECTORY.
- **Automation:** Node-RED and Activepieces as DIRECTORY; direct embedding requires separate deployment and permission reviews.
- **Privacy, dev, learning, finance, 3D:** external community discovery first, capabilities only after explicit independently tested module implementation.

## Decisions

- Do not embed external projects or copy third-party code from online sources solely based on GitHub availability.
- No cryptocurrency trading, banking access or destructive computer control in the core suite.
- Catalog requires no network to browse; following an upstream external hyperlink intentionally uses the network.
