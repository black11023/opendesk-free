# OpenDesk Free v0.2.0 Preview 3

## Page-range selection for PDF documents

- Extract pages by number (`2`), list (`1,5,9`) or range (`10-15`), including explicit order (`7,2,5`) and descending ranges.
- Extract to a single combined PDF, or separate PDFs packed in a ZIP. No automatic extraction of hundreds of pages when the user only selects 1–2.
- Merge specific pages from each selected PDF, with user-controlled file order (up/down arrows).
- Input may contain up to 5,000 pages per PDF, with a maximum of 300 selected output pages; separate ZIP limited to 100 pages; upload remains 40 MB total.
- Reject invalid or out-of-range page numbers and preserve original files.
- English and Vietnamese interface; Windows CI and PDF regression tests required.
- OCR still not implemented. Unsigned preview, pending manual Windows install/uninstall validation.
