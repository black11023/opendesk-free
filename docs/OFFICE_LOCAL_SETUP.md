# Convert Office to PDF / Chuyển văn bản Office sang PDF

**Optional integration only:** LibreOffice is NOT included in the OpenDesk installer.

## Windows installation

1. Download LibreOffice from the official website: https://www.libreoffice.org/download/download/
2. Install it separately. A default Windows installation is usually found at `C:\Program Files\LibreOffice\program\soffice.exe`, or the program must be available on PATH.
3. Restart OpenDesk, select **Office → PDF** and click **Recheck LibreOffice**.
4. Choose DOC/DOCX, XLS/XLSX, PPT/PPTX, ODT/ODS/ODP documents. One input creates one PDF, two to eight create a ZIP of PDFs.

Input limits: 8 files / 40 MB. Outputs: up to 250 pages per document, 500 total pages, maximum 80 MB combined.

OpenDesk runs LibreOffice with a private temporary profile and fixed headless arguments; it checks the generated PDF and removes temporary files afterwards. The source documents remain unchanged.

**Important:** Font availability, spreadsheet print ranges, embedded fonts and complex slide effects can change document layout. You must review and compare generated PDFs before using them for legal, financial or official submissions.

Detection only checks whether LibreOffice can start. Actual conversion compatibility still needs manual verification on Windows; macros and malicious Office documents remain third-party risks. No auto-installation and no cloud services.

Official CLI reference: https://help.libreoffice.org/latest/en-US/text/shared/guide/pdf_params.html

## Tiếng Việt

Cài LibreOffice riêng từ trang chính thức. Khởi động lại OpenDesk, chọn **Văn phòng → PDF**, bấm **Kiểm tra lại LibreOffice**. Chọn tối đa 8 tệp Word/Excel/PowerPoint. Kết quả tải về dưới dạng PDF hoặc ZIP; không ghi đè bản gốc. Cần đối chiếu lại phông chữ, khổ giấy, bảng tính, số trang và hình ảnh.
