# OpenDesk Free — đa ngôn ngữ bắt buộc ngay từ nền kiến trúc

**Trạng thái (28/10/2026):** G0 — đã có danh mục ứng viên và bộ định tuyến an toàn; chưa có chức năng dịch tài liệu thực tế.

- Người dùng có thể lựa chọn ngôn ngữ nguồn và **nhiều ngôn ngữ đích** trong cùng một tác vụ. Giới hạn 16 ngôn ngữ đích mỗi lượt để tránh quá tải; mỗi đích tạo bản dịch tài liệu riêng. Đây là yêu cầu kiến trúc, chưa phải chức năng UI hoàn chỉnh.
- Danh mục ban đầu gồm **46 ứng viên ngôn ngữ/biến thể** như Việt, Anh, Trung giản thể/phồn thể, Nhật, Hàn, Pháp, Đức, Nga, Tây Ban Nha, Bồ Đào Nha, Ả Rập, Thái, Hindi, Malay… Có thể mở rộng; chỉ quảng cáo một cặp là **khả dụng** khi có mô hình đã cài, xác thực hash, rà soát giấy phép và kiểm thử thực tế. Không hứa 46 cặp đã chạy được.
- Dùng mã BCP 47 (VD `vi`, `en`, `zh-Hans`, `zh-Hant`, `pt-BR`). Phân biệt chính xác bản chữ Trung giản thể/phồn thể và biến thể ngôn ngữ khi chọn mô hình.
- Ưu tiên cặp **trực tiếp**. Chỉ dịch qua **Anh** làm trung gian nếu user đồng ý; ghi đầy đủ tuyến dịch, cảnh báo có thể mất nghĩa và yêu cầu duyệt kết quả. Không đi qua >1 nút trung gian hay âm thầm chuyển dịch vụ cloud.
- Tài liệu có nhiều ngôn ngữ: phân đoạn và xác nhận nguồn từng đoạn/ô/khung trước khi dịch; vùng đã đúng ngôn ngữ đích **giữ nguyên**. Tự nhận diện ngôn ngữ là tính năng phải có sau khi tích hợp detector thực tế; ở G0, `auto` trả về lỗi, không đoán.
- Word DOCX, Excel XLSX, PowerPoint PPTX, PDF (text/PDF quét) phải dùng **chung kiến trúc ngôn ngữ**, nhưng bộ xuất Word/Excel/PDF, kiểm tra font/RTL/bố cục và OCR là các cổng nghiệm thu riêng.
- FormGuard giữ nguyên số liệu, ngày, mã định danh, công thức, bảng, trường mẫu, chữ ký gốc; nếu không đảm bảo cấu trúc cần BLOCKED và báo cáo từng vị trí. Một nguồn → nhiều bản đích không làm sửa file gốc, cùng lưu SHA-256 và báo cáo.
- Nền chính miễn phí, không API trả phí, không bắt tạo tài khoản, không tự upload tài liệu/tải model. Mô hình dịch offline tùy chọn, cài theo từng gói có sự đồng ý và hash/giấy phép hợp lệ.
- Mô-đun `internal/preservetranslate/languages.go` chỉ **quy hoạch đường dịch từ inventory mô hình đã kiểm chứng**, chưa hề dịch chữ hoặc xử lý file. G1 phải triển khai provider thật và các cặp ngôn ngữ kiểm thử (Anh ↔ Việt, Anh ↔ Pháp, Anh ↔ Nhật là các ví dụ thử nghiệm, không phải giới hạn phần mềm). G2/G3/G4 áp dụng cho XLSX/PPTX/PDF, không lùi lại yêu cầu đa ngôn ngữ đến G5.

**Kiểm thử tối thiểu G0:** chọn nhiều đích theo thứ tự; ưu tiên trực tiếp; không tự pivot; từ chối model thiếu hash/license/installation; phân biệt `zh-Hans`/`zh-Hant`; mixed-source segments không tái dịch đoạn đúng đích; batch lỗi thì không trả kết quả từng phần.

**Nguồn:** https://argos-translate.readthedocs.io/en/stable/ ; https://docs.libretranslate.com/guides/supported_languages/ ; https://www.w3.org/International/articles/language-tags/Overview.en.php
