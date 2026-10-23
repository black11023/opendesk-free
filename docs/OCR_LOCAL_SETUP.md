# Offline OCR local setup — English / Tiếng Việt

OpenDesk Free includes an **optional adapter** to Tesseract; the Tesseract engine is **not bundled**. The OCR feature remains disabled until you separately install a local Tesseract executable and appropriate language packs. This is not a paid API and OpenDesk does not download any third-party executable automatically.

## Windows 10/11

1. Read the upstream [Tesseract installation instructions](https://tesseract-ocr.github.io/tessdoc/Installation.html). Its Windows guidance links to the UB Mannheim builds of Tesseract 5.
2. Install Tesseract from a source you trust. If installed to `C:\Program Files\Tesseract-OCR\tesseract.exe`, OpenDesk can discover it even when not added to PATH.
3. Select English language `eng` and Vietnamese language `vie` training data when installing, or place the corresponding trusted `eng.traineddata` and `vie.traineddata` files into the installation `tessdata` directory.
4. In a new Windows terminal, verify `tesseract --list-langs` (if executable is on PATH). Look for the exact codes `eng` and `vie`; the language menu supports English, Vietnamese and combined English+Vietnamese.
5. Restart OpenDesk if installed while OpenDesk was already running; select **OCR scanned images / Nhận dạng chữ từ ảnh** and click **Recheck Tesseract / Kiểm tra lại**.

## Offline behavior and limits

- Tesseract works entirely on the user's PC. The app runs `tesseract stdin stdout -l eng` (plus `pdf` when a searchable PDF is requested); image bytes flow through local stdin. It does not invoke a shell, execute input filenames, upload content or download new binaries.
- OCR scanned PNG/JPG only, max 10 images, 20 megapixels per image, 40 MB total. Text output combines all images with per-file headings; PDF output creates one searchable page per image.
- Misrecognized or blank text is possible. Verify important facts, amounts, names and Vietnamese accents manually.
- OCR of already existing scanned PDFs is **not implemented**. Convert PDF pages to images with a trusted tool first if necessary.
- Windows installer is not code-signed and has not completed real end-user Windows install/uninstall testing.

Upstream documentation:
- https://tesseract-ocr.github.io/tessdoc/Command-Line-Usage.html
- https://github.com/tesseract-ocr/tessdoc/blob/main/Installation.md
- https://github.com/tesseract-ocr/tessdata_fast
