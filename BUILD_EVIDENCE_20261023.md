# Build & test evidence — 2026-10-23

| Test | Outcome | What it validates |
|---|---|---|
| `go test -v ./...` | 6/6 PASS | engine file-organization collisions, CSV/XLSX strings and formula safety, SHA-256 duplicates, invalid input, text joining, DOCX text |
| `go test -race ./...` | PASS | Go unit tests under race detector |
| `go vet ./...` | PASS | Go static checks |
| Linux local HTTP smoke (binary) | PASS | `GET /`, CSRF 403, Host-header 403, XLSX output XML valid, ZIP organization and duplicate report |
| Chromium `page.set_content()` static UI exercise | PASS | Desktop 1280px & mobile 390px, VI/EN switch, file selection, button enabled, zero browser JS exceptions |
| Windows amd64 app cross-compile | PASS | PE32+ x86-64 native GUI executable |
| Windows amd64 setup cross-compile | PASS | PE32+ x86-64 native GUI executable embedding the compiled app |
| Full archive ZIP integrity | PASS | All ZIP entries CRC OK, binary SHA-256 manifest included |
| Chromium end-to-end over localhost | BLOCKED | Sandbox browser networking policy returned `ERR_BLOCKED_BY_ADMINISTRATOR`; not represented as PASS |
| Windows installation, UI launch, re-install and uninstall | NOT TESTED | Requires a real Windows 10/11 acceptance run |
| Defender / SmartScreen / digital signature | NOT TESTED | Executables currently unsigned |
| GitHub Actions remote run | NOT RUN | Requires publishing repo to GitHub and running Windows workflow |

**Release grade:** PREVIEW ONLY. Not eligible for stable release until platform acceptance tests pass. No external API calls in the app; code generation and preview were performed locally. No user documents were opened.
