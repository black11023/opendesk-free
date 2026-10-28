# OpenDesk Free V0.6.9 Preview 1 — Verified optional language-pack installer

## Working changes

- An explicit user-consent model-manager section in the offline DOCX translation panel, in English and Vietnamese.
- Direct model **English → Spanish, v1.0** only in the in-app approved catalog, exact model size 87,503,191 bytes and SHA-256 `d698d0ef87ad70d5d184b7fa6965905bf4368f09a2bb9ffb165a79bac96af0c4` from real integration evidence. Same checks apply whether downloading from the official Argos vendor or importing a local .argosmodel.
- Every archive is checked for source integrity, exact size, ZIP CRC, path traversal, duplicate files, compression bombs, and actual language metadata before Argos installation and a local live translation probe.
- Actions are authenticated, local-only, cancellable, and never execute a shell, pip, or silent network/model installation. Existing core FormGuard, OCR, translation safety and installer gates remain.

## Limitations / release honesty

- This is **NOT a finished one-click language runtime manager**. Python and the Argos Translate inference package must already be installed independently; no Python installation, model redistribution, commercial license certification, or full document layout guarantee is included.
- Model-package third-party licensing requires independent review. OpenDesk does not bundle or redistribute the model. EN→VI and other routes are not auto-downloadable until their own verified release checksum/license are approved.
- The Windows EXEs remain unsigned previews. Do not bypass Chrome/Defender warnings; SHA + ClamAV are not code signing.
