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
