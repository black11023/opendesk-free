# OpenDesk Free v0.6.0 Preview 1 — Professional tools and real conversion tests

- Professional English/Vietnamese workbench: instant search for operations, category filters, visible active state, progress feedback and responsive mobile layout.
- Batch-rotate 2–10 PDFs 90° clockwise and download a ZIP. Limit 300 pages total, 80 MB output. Single-PDF rotation stays available.
- Memory protection: Go soft heap limit at 512 MiB, reject new processing at 300 MiB observed Go HeapAlloc (HTTP 503 + retry). These are not physical RAM limits and do not cap third-party Office/OCR engines.
- New **actual real-dependency Linux CI**: render generated DOCX and XLSX with installed LibreOffice; confirm text content using pdftotext. Scan generated PNG with Tesseract and render scanned PDF with OCRmyPDF; check OCR text exists in searchable output.
- All prior Windows EXE local HTTP and Windows Setup install-upgrade-uninstall gates retained.
- OpenDesk Free remains entirely free; external LibreOffice/Tesseract/OCRmyPDF packages are optional local installs, not bundled, and no account/paid API is needed.
- Unsigned Preview. Visual typography/layout fidelity, memory stress on low-end machines and interactive Windows acceptance remain unfinished.
