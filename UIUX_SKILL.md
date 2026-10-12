# LabShop UI/UX

## 1. Định hướng

- Một hệ thống thiết kế, hai "khu": Shop (thân thiện, ảnh sản phẩm lớn, nhiều khoảng thở) và Lab/Tutorial/Admin
  (mật độ cao, kiểu công cụ cho developer). Dùng chung token, khác độ dày thông tin.
- Ba mục tiêu theo thứ tự: dùng được -> quan sát được (lab phải cho thấy HTTP đang xảy ra gì) -> đẹp.
  Cái đẹp đến từ tiết chế: lưới đều, chữ rõ, một màu nhấn, bóng nhẹ, chuyển động ngắn.
- Tránh: gradient tím/xanh mặc định, glassmorphism, emoji làm icon, animation > 250ms, ảnh stock,
  chữ xám nhạt khó đọc, bo góc mỗi nơi một kiểu.

## 2. Design tokens (khai báo một lần trong Tailwind theme + CSS variables, có cả light và dark)

- Neutral: slate. Brand: teal (primary teal-600, hover teal-700, nền nhẹ teal-50/950). Chỉ một màu nhấn.
- Semantic: success, warning, danger, info. Màu riêng cho lab (luôn đi kèm icon + chữ, không chỉ dựa vào màu):
  origin = xanh dương, cache HIT = teal, cache MISS = amber, bị chặn ở edge = đỏ, nghi cache cũ = cam.
- Chữ: Inter (variable) cho UI, JetBrains Mono cho code/header/curl; self-host woff2 qua @fontsource,
  `font-display: swap`. Thang chữ 12/14/16/18/20/24/30/36; line-height thoáng ở đoạn văn, chặt ở bảng.
- Spacing lưới 4px; radius 8 (control) và 12 (card); shadow 2 mức; focus ring 2px offset 2px, thấy rõ ở cả hai theme.
- Tương phản tối thiểu WCAG AA ở cả light và dark. Theme theo `prefers-color-scheme`, cho đổi tay và nhớ lựa chọn
  (localStorage bọc try/catch, không bắt buộc).

## 3. Bố cục

- App shell: navbar (logo, tìm sản phẩm, giỏ hàng có badge số lượng, menu tài khoản, đổi theme), nội dung max-w ~1200px, footer.
  Admin dùng sidebar thu gọn được. Tutorial dùng 3 cột: danh sách kịch bản | nội dung | runner (mobile gộp thành tab).
- Lab Console là drawer cố định dưới đáy, mở bằng nút nổi hoặc Ctrl+Shift+L, có ở mọi trang (xem mục 7).

## 4. Trang và hành vi chính

- Home: hero ngắn, danh mục, sản phẩm bán chạy.
- Danh sách sản phẩm: bộ lọc (q, danh mục, khoảng giá, sắp xếp) đồng bộ với URL; mobile đưa bộ lọc vào Sheet;
  phân trang; kết quả rỗng có gợi ý bỏ bộ lọc.
- Chi tiết sản phẩm: gallery, giá, trạng thái kho (còn hàng / sắp hết / hết), QuantityStepper 1..10, thêm vào giỏ,
  phân bố sao, danh sách review, form review. Chưa đủ điều kiện review thì vô hiệu hóa form và giải thích lý do.
- Giỏ hàng: dòng sản phẩm, đổi số lượng lạc quan (optimistic) có rollback khi lỗi, ô coupon báo lỗi theo mã
  (hết hạn, chưa đủ min, đã dùng), khối tóm tắt tiền cố định khi cuộn.
- Checkout: một trang 3 phần (giao hàng, thanh toán, xác nhận). Mỗi lần đặt hàng sinh một `Idempotency-Key` mới,
  giữ nguyên khi retry cùng payload. Khóa nút khi đang gửi. Trang kết quả có mã đơn và nút xem đơn.
- Đơn hàng: danh sách + chi tiết với timeline trạng thái (PENDING, PAID, SHIPPED, DELIVERED, CANCELLED).
  Hủy đơn qua Dialog xác nhận.
- Hỗ trợ: danh sách ticket, thread dạng hội thoại, vùng kéo thả đính kèm (kiểm tra đuôi/dung lượng ở client trước;
  server vẫn là nguồn sự thật).
- Đăng nhập/đăng ký: hiện/ẩn mật khẩu, quy tắc mật khẩu hiển thị trực tiếp, lỗi map theo `error.details[].field`.
- Admin: thẻ số liệu + biểu đồ doanh thu (lazy-load Recharts), bảng có sắp xếp/lọc/phân trang,
  đổi trạng thái đơn chỉ cho các bước hợp lệ.
- Tutorial và /docs: xem mục 7.

## 5. Component

- Dùng primitive của shadcn/ui: Button, Input, Textarea, Select, Checkbox, RadioGroup, Dialog, Sheet, DropdownMenu,
  Tabs, Table, Badge, Card, Tooltip, Accordion, Skeleton, sonner.
- Component nghiệp vụ: ProductCard, PriceTag (Intl.NumberFormat vi-VN, VND), RatingStars (có text thay thế),
  QuantityStepper, CouponForm, OrderStatusBadge, OrderTimeline, EmptyState, ErrorState, FormField (label + hint + lỗi).
- Component lab: ResponseBadges, HeadersTable, CurlBlock (nút copy), CodeBlock (nhẹ, không nạp thư viện highlight nặng),
  RunnerPanel, StatusHistogram, LatencySparkline (SVG tự vẽ).
- Mỗi component có đủ trạng thái: default, hover, focus-visible, active, disabled, loading, error.

## 6. Trạng thái dữ liệu, form, lỗi

- Mọi màn hình lấy dữ liệu có 4 trạng thái: loading (Skeleton đúng hình dạng nội dung), empty,
  error (kèm `request_id`, nút Thử lại), success.
- Form: validate bằng Zod phía client, báo lỗi tại chỗ, focus vào ô lỗi đầu tiên, giữ dữ liệu khi lỗi, chống gửi trùng.
- Bảng map `error.code` -> thông điệp tiếng Việt nằm trong một file duy nhất (`src/lib/messages.ts`).
  Không hiển thị stack trace hay thông báo thô từ server.
- Toast chỉ cho phản hồi nhanh; lỗi cần hành động thì đặt ngay trên giao diện.

## 7. UX riêng cho lab WAAP/CDN

- Phân loại phản hồi trong MỘT fetch wrapper duy nhất (mọi request của app đều đi qua):
  có header `X-Origin-Instance` -> "Origin"; không có và (status >= 400 hoặc body không phải JSON) ->
  "Nghi bị chặn ở edge/WAAP" (ghi rõ đây là suy đoán, kèm bằng chứng đã quan sát);
  `X-Cache-Status` HIT/MISS và `Age` -> badge cache; `Content-Encoding` -> badge nén.
- Lab Console: lịch sử 50 request gần nhất (method, path, status, ms, badges, request_id), bấm để xem toàn bộ header,
  lọc theo lỗi/cache, xuất JSON, copy dưới dạng curl, nút xóa. Trạng thái bật/tắt nhớ qua localStorage (try/catch).
- Environment bar: hiện hostname đang truy cập và nhãn đường đi (direct, waap, cdn, full) lấy từ cài đặt người dùng
  (nhập 4 hostname, lưu localStorage); host lạ thì cảnh báo nhẹ nhàng.
- Runner: Run once, Run N (1–20), Burst (tối đa 200 request, concurrency tối đa 20, có Dialog xác nhận,
  nút Dừng bằng AbortController). Bảng kết quả cập nhật trực tiếp; với /api/time tô cùng màu các dòng có
  `request_id` lặp lại để thấy cache HIT; Burst có histogram theo mã trạng thái.
- Trang kịch bản: stepper "Chuẩn bị -> Cấu hình trên portal -> Chạy thử -> Quan sát -> Dọn dẹp", checklist nhớ tiến độ,
  bảng `expect` (cấu hình nào thì kết quả kỳ vọng nào), link docs CMC mở tab mới với `rel="noopener noreferrer"`.
  Kịch bản `runnable=false` chỉ hiện lệnh curl kèm lý do.
- Chuỗi thử WAF (probes) có nhãn riêng, chạy phải qua xác nhận một lần mỗi phiên, kèm dòng nhắc:
  chỉ chạy trên hệ thống của bạn.
- Tài khoản demo chỉ hiện khi API báo bật; không bao giờ hiện thông tin admin.

## 8. Accessibility (WCAG 2.2 AA)

- HTML ngữ nghĩa và landmark (header, nav, main, footer); skip link; mọi control dùng được bằng bàn phím;
  thứ tự Tab hợp lý; Dialog/Sheet bẫy focus và trả focus khi đóng (Radix lo phần lớn, vẫn phải kiểm tra).
- Label đầy đủ, lỗi gắn `aria-describedby`; cập nhật giỏ hàng và kết quả runner dùng `aria-live="polite"`;
  badge luôn có chữ, không chỉ màu.
- Mục tiêu bấm >= 44px cho hành động chính; tôn trọng `prefers-reduced-motion`;
  ảnh có `alt` có nghĩa (ảnh trang trí để alt rỗng).

## 9. Responsive và hiệu năng

- Mobile-first, kiểm tra ở 360, 768, 1280. Không cuộn ngang. Bảng chuyển thành card ở mobile.
- Tách code theo route (lazy); Recharts chỉ nạp ở admin. Ảnh có width/height, `loading="lazy"` (trừ ảnh đầu trang).
  Mục tiêu JS ban đầu <= 250KB gzip, đo bằng báo cáo build và ghi vào docs/ui/README.md.
- Không nạp font, script, ảnh từ domain ngoài.

## 10. Microcopy

- Tiếng Việt, giọng thân thiện và gọn. Tiền `1.250.000 ₫`; ngày giờ theo Asia/Ho_Chi_Minh.
  Thuật ngữ kỹ thuật giữ tiếng Anh như trên portal (WAAP, Cache Rule, Purge, Whitelist…).
- Gom chuỗi giao diện vào một nơi để sau này thêm i18n, nhưng chưa cần thư viện i18n lúc này.

## 11. Quy trình và nghiệm thu UI

- Thứ tự làm: tokens + app shell -> component nền -> storefront -> admin -> tutorial + Lab Console.
- Playwright chụp ảnh các trang chính ở 360/768/1280, light và dark, lưu vào docs/ui/; chạy axe-core,
  không còn lỗi mức critical/serious.
- Luồng bàn phím hoàn chỉnh cho: tìm sản phẩm -> thêm giỏ -> coupon -> checkout -> xem đơn.
- Khi phải chọn giữa "đẹp hơn" và "rõ hơn / dễ quan sát hơn" thì chọn rõ hơn.
  Mọi lựa chọn thiết kế đáng kể ghi vào docs/DECISIONS.md.
