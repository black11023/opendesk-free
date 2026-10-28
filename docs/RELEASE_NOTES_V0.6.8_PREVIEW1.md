# OpenDesk Free v0.6.8 Preview 1 — Direct Offline DOCX Translation Beta

New optional **Translate DOCX** operation. It calls an already installed local Python Argos Translate package and a direct language-pair model. No cloud, model download or paid API. The user can choose only direct language pairs found on their PC.

Strict conditions: one .docx up to 12 MB, no rich-run paragraph splitting, fields, changes, headers/footers or unsupported XML, no source overwrite. FormGuard numeric/code placeholders are restored and checked. The output ZIP includes DOCX, full paragraph audit and SHA-256 manifest. **Manual page layout and language review is mandatory.**

Not included: ready-made language model installers, rich-format documents, XLSX/PPTX/PDF translation, automatic language detection and guaranteed same-layout export. Existing FormGuard and OCR remain unchanged.