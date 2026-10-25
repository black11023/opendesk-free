# OpenDesk Free v0.5.2 Preview 1 — Windows installer acceptance

- GitHub Actions now tests the **real Windows installer executable** for first install, upgrade over an existing installation, Start Menu shortcut, HKCU Installed Apps registration, launch of the installed EXE, uninstallation and user data preservation.
- The automated test uses isolated current-user AppData directories on a disposable Windows GitHub runner. The installer never asks for administrator access.
- Added explicit noninteractive `--quiet-install` and `--quiet-uninstall` commands for reproducible tests; interactive user confirmation is preserved.
- Uninstallation removes only OpenDesk-owned binaries, shortcut and registry registration; it never recursively removes user-created files.
- Build/checks still cover bilingual UI, PDF/OCR contracts, Go tests/static checks, real Portable EXE runtime and Windows installer acceptance.
- **Preview / unsigned:** a GitHub-hosted Windows test is not a real consumer-PC acceptance test. Desktop dialog accessibility, OS-reputation warnings and Office/OCR fidelity remain open items.
