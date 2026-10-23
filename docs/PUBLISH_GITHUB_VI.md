# HƯỚNG DẪN ĐƯA OPENDESK FREE LÊN GITHUB — CHO NGƯỜI MỚI

## A. Tạo repository công khai

1. Truy cập https://github.com/new, đăng nhập tài khoản GitHub.
2. **Owner**: tài khoản của bạn. **Repository name**: `opendesk-free`.
3. **Description**: `Free, private, offline-first document tools for everyone. No subscriptions or paid API required.`
4. **Visibility**: chọn **Public** ngay lúc tạo.
5. **KHÔNG chọn** thêm README, .gitignore hoặc license từ GitHub, vì gói mã nguồn đã chứa cả ba.
6. Nhấn **Create repository**. Repo sẽ có URL dạng `https://github.com/black11023/opendesk-free`.

## B. Đăng đầy đủ mã nguồn bằng GitHub Desktop (dễ nhất)

1. Cài và đăng nhập **GitHub Desktop** tại https://desktop.github.com/.
2. Vào **File > Clone repository** > tab **GitHub.com**; chọn `black11023/opendesk-free` mà bạn vừa tạo, chọn đường dẫn máy, ví dụ `C:\OpenDeskFreeGit` > **Clone**.
3. Mở thư mục `C:\OpenDeskFreeGit` bằng Windows Explorer. Sao chép **toàn bộ nội dung** của thư mục `SOURCE` trong gói FULL vào đây, bao gồm `.github`, `.gitignore`, README.md, LICENSE và các thư mục `cmd`, `internal`, `docs`, `tests`. Không sao chép riêng bộ cài `.exe` vào git.
4. Quay lại GitHub Desktop, bạn sẽ thấy danh sách file thay đổi. Nhập **Summary**: `Initial public preview v0.1.0`, nhấn **Commit to main**.
5. Nhấn **Push origin**. Chờ push hoàn thành, mở trang repo và xác nhận có README.md, LICENSE, `.github/workflows/build-windows.yml`.

**Cách thay thế bằng Git CLI (PowerShell) — sau khi đã tạo repository GitHub rỗng:**

```powershell
cd C:\DUONG_DAN\SOURCE
git init
git branch -M main
git add .
git commit -m "Initial public preview v0.1.0"
git remote add origin https://github.com/black11023/opendesk-free.git
git push -u origin main
```

Nếu Git hỏi danh tính tác giả commit, chạy `git config --global user.name "Your Name"` và `git config --global user.email "your@email.com"`. Xác thực bằng trình duyệt/Git Credential Manager, không lưu password/token vào mã nguồn.

## C. Build bộ cài trên Windows bằng GitHub Actions

1. Mở repo trên GitHub, vào tab **Actions**.
2. Nếu có nút **I understand my workflows, go ahead and enable them** / **Enable workflows**, nhấn bật.
3. Bấm workflow **Build and test Windows installer** > **Run workflow** > **Run workflow** (hoặc bản push lên `main` sẽ tự chạy).
4. Chỉ xem là thành công nếu toàn bộ job `windows-release` màu xanh và step `Run automated tests` đạt.
5. Trong workflow run, cuối trang `Artifacts`, tải file ZIP `OpenDeskFree-Windows-v0.1.0` (mặc định artifact có thể chỉ được giữ trong 30 ngày).
6. Giải nén, sẽ có `OpenDeskFree_Setup_0.1.0.exe`, `OpenDeskFree_Portable_0.1.0.exe` và `SHA256SUMS.txt`.
7. Cài thử trên Windows 10/11: chạy Setup, mở từ Start Menu, thử 5 tác vụ trên file mẫu, gỡ cài đặt qua Settings > Apps. Đối chiếu SHA256. Không công bố là bản stable nếu chưa nghiệm thu.

## D. Công bố bản tải xuống cho mọi người qua GitHub Releases

1. Repo > **Releases** ở cột phải > **Create a new release**.
2. Nhấn **Choose a tag**, gõ `v0.1.0`, chọn **Create new tag** dựa trên `main`.
3. **Release title**: `OpenDesk Free v0.1.0 — Initial Preview (Windows)`.
4. Nội dung: tính năng thực có, giới hạn dung lượng 50 files / 40 MB, yêu cầu Windows 10/11 x64, lưu ý bộ cài unsigned.
5. Trong **Attach binaries**, kéo 2 file `.exe` và `SHA256SUMS.txt` tải được từ Actions (không upload file ZIP source vào repo history vì GitHub tự cấp source archive).
6. Chọn **Set as a pre-release** vì chưa kiểm thử Windows thực tế; nhấn **Publish release**.
7. Link chia sẻ dạng: `https://github.com/black11023/opendesk-free/releases/latest` khi có bản stable; với preview nên gửi link trực tiếp `/releases/tag/v0.1.0`.

## E. Kiểm tra công khai

1. Mở cửa sổ ẩn danh/private window (không đăng nhập GitHub).
2. Truy cập URL repo: phải xem được README và mã nguồn.
3. Truy cập Releases/tag/v0.1.0: phải tải được `.exe` mà không cần đăng nhập.
4. Đảm bảo Settings > General > Danger Zone > Change repository visibility hiện trạng thái **Public**. Không đăng khóa API, file cá nhân hay tài liệu nội bộ lên repository công khai.
5. Nếu repository đã tạo Private: Settings > General > Danger Zone > Change repository visibility > **Make public**, đọc kỹ các cảnh báo rồi xác nhận.

**Lưu ý:** source ZIP chứa mã nguồn và tài liệu; bộ cài EXE nằm trong gói FULL và cũng nên đính kèm riêng ở Releases để người sử dụng dễ tải. File Setup hiện chưa được ký số. Windows SmartScreen có thể cảnh báo; người dùng nên kiểm tra nguồn và mã SHA256, không bỏ qua cảnh báo bảo mật một cách mù quáng.
