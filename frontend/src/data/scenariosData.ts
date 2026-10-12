import type { Scenario } from '@/types';

export const SCENARIOS: Scenario[] = [
  // ==================== CDN SCENARIOS ====================
  {
    id: 'c1',
    group: 'cdn',
    priority: 'P0',
    title: 'C1: Onboard site lên Multi-CDN',
    goal: 'Đăng ký cấu hình domain lên Multi-CDN CMC, trỏ CNAME và kiểm tra luồng phân giải qua header Server và X-Cache-Status.',
    portalSteps: [
      'Đăng nhập CMC Cloud Portal > Chọn dịch vụ CDN > Sites',
      'Bấm "Add Site", nhập tên miền phân phối và IP/Domain Origin (port 80/443)',
      'Lấy bản ghi CNAME được cấp và cập nhật vào DNS quản lý tên miền',
      'Chờ phân giải DNS và vào tab Sites > Overview kiểm tra trạng thái Active',
    ],
    requests: [
      {
        method: 'GET',
        path: '/api/health',
        runnable: true,
        curl: 'curl -i https://cdn.labshop.local/api/health',
      },
    ],
    expect: [
      {
        condition: 'Origin trực tiếp',
        expectedResult: 'Header Server có thể là nginx/Go, không có X-Cache-Status hoặc Via từ CDN.',
      },
      {
        condition: 'Qua Multi-CDN',
        expectedResult: 'Xuất hiện header X-Cache-Status (MISS hoặc HIT) và header Via/Server của CDN provider.',
      },
    ],
    notes: 'Kiểm tra độ trễ (latency ms) giảm đáng kể ở các PoP gần người dùng sau khi cache ấm.',
    docsUrl: 'https://cloud.cmctelecom.vn/docs/cdn',
  },
  {
    id: 'c2',
    group: 'cdn',
    priority: 'P0',
    title: 'C2: Cache mặc định (Default 30 phút)',
    goal: 'Quan sát hiện tượng CDN tự động cache mọi phản hồi HTTP 200 trong 30 phút nếu không có chỉ thị ngược lại, dẫn đến API động bị "đóng băng" dữ liệu.',
    portalSteps: [
      'Gọi endpoint /api/time 3 lần liên tiếp qua CDN và quan sát timestamp',
      'Vào CMC Portal > CDN > Sites > Cache Rules',
      'Thêm quy tắc "No Cache" với Path pattern kiểu Directory: /api/*',
      'Lưu cấu hình, đợi 30s và gọi lại /api/time để xác nhận timestamp thay đổi theo từng request',
    ],
    requests: [
      {
        method: 'GET',
        path: '/api/time',
        runnable: true,
        curl: 'curl -i https://cdn.labshop.local/api/time',
      },
    ],
    expect: [
      {
        condition: 'Trước khi cấu hình Cache Rule',
        expectedResult: 'Lần 1 MISS (tạo mới timestamp), lần 2 và 3 HIT (timestamp và request_id giữ nguyên y hệt, Age tăng dần).',
      },
      {
        condition: 'Sau khi cấu hình No Cache /api/*',
        expectedResult: 'Mỗi lần gọi đều trả timestamp mới từ origin, X-Cache-Status là BYPASS hoặc MISS.',
      },
    ],
    notes: 'Đây là lỗi phổ biến nhất khi mới tích hợp CDN vào ứng dụng có cả API và Web.',
    docsUrl: 'https://cloud.cmctelecom.vn/docs/cdn/cache-rules',
  },
  {
    id: 'c2b',
    group: 'cdn',
    priority: 'P0',
    title: 'C2b: Kiểm tra rò rỉ dữ liệu cá nhân do CDN Cache nhầm API',
    goal: 'Chứng minh nguy cơ bảo mật nghiêm trọng khi CDN cache endpoint trả thông tin riêng tư (/api/auth/me) dẫn đến User B nhận dữ liệu của User A.',
    portalSteps: [
      'Đăng nhập với User 1, gửi GET /api/auth/me qua CDN',
      'Đăng nhập với User 2 ở tab ẩn danh khác, gửi GET /api/auth/me với token của User 2',
      'Quan sát xem User 2 có nhận nhầm payload chứa email của User 1 do CDN Cache HIT không',
      'Khắc phục triệt để bằng cách thiết lập Cache Rule: Bypass Cache cho toàn bộ /api/*',
    ],
    requests: [
      {
        method: 'GET',
        path: '/api/auth/me',
        runnable: true,
        curl: 'curl -i -H "Authorization: Bearer <TOKEN_USER>" https://cdn.labshop.local/api/auth/me',
      },
    ],
    expect: [
      {
        condition: 'Khi CDN vô tình cache /api/auth/me',
        expectedResult: 'User 2 nhận thông tin User 1 kèm X-Cache-Status: HIT (Rò rỉ thông tin cực kỳ nguy hiểm).',
      },
      {
        condition: 'Sau khi cấu hình Bypass/No Cache',
        expectedResult: 'Mỗi User nhận đúng thông tin cá nhân của mình, X-Cache-Status: BYPASS/MISS.',
      },
    ],
    notes: 'Khuyến nghị tuân thủ quy tắc: Không bao giờ cache endpoint có thông tin phân quyền theo Authorization.',
    docsUrl: 'https://cloud.cmctelecom.vn/docs/cdn/security',
  },
  {
    id: 'c3',
    group: 'cdn',
    priority: 'P0',
    title: 'C3: Caching Level (Query String Sensitivity)',
    goal: 'Khảo sát cách CDN phân biệt khóa cache (Cache Key) dựa trên tham số query string (?v=1 vs ?v=2).',
    portalSteps: [
      'Vào CMC Portal > CDN > Caching > Caching Level',
      'Thử nghiệm chế độ Standard (mỗi query string là 1 cache key riêng)',
      'Thử nghiệm chế độ Ignore Query String (bỏ qua query, ?v=1 và ?v=2 dùng chung cache)',
      'Gọi /api/cache/public?v=1 rồi gọi /api/cache/public?v=2 để kiểm chứng',
    ],
    requests: [
      {
        method: 'GET',
        path: '/api/cache/public?v=1',
        runnable: true,
        curl: 'curl -i "https://cdn.labshop.local/api/cache/public?v=1"',
      },
      {
        method: 'GET',
        path: '/api/cache/public?v=2',
        runnable: true,
        curl: 'curl -i "https://cdn.labshop.local/api/cache/public?v=2"',
      },
    ],
    expect: [
      {
        condition: 'Standard Mode',
        expectedResult: '?v=1 và ?v=2 lưu 2 bản cache độc lập; v=1 HIT không làm v=2 HIT.',
      },
      {
        condition: 'Ignore Query String',
        expectedResult: 'Gọi v=1 lần đầu MISS, gọi v=2 ngay sau đó trả kết quả HIT của v=1.',
      },
    ],
    notes: 'Dùng Ignore Query String cẩn thận khi ứng dụng dùng tham số lọc (filter/pagination).',
    docsUrl: 'https://cloud.cmctelecom.vn/docs/cdn/caching-level',
  },
  {
    id: 'c4',
    group: 'cdn',
    priority: 'P0',
    title: 'C4: Cache Rule theo File extension & Directory',
    goal: 'Cấu hình thời gian lưu cache chi tiết theo đuôi file tĩnh (.jpg, .png, .css, .js) và theo thư mục (/static/img/).',
    portalSteps: [
      'Vào CDN > Sites > Cache Rules > Add Rule',
      'Rule 1: File extension (png|jpg|webp) -> Edge TTL 7 ngày',
      'Rule 2: Directory (/static/img/) -> Edge TTL 30 ngày',
      'Kiểm tra gọi /static/img/products/p01.jpg',
    ],
    requests: [
      {
        method: 'GET',
        path: '/static/img/products/p01.jpg',
        runnable: true,
        curl: 'curl -i https://cdn.labshop.local/static/img/products/p01.jpg',
      },
    ],
    expect: [
      {
        condition: 'Đúng quy tắc',
        expectedResult: 'Ảnh sản phẩm được lưu cache tại edge, lần thứ hai trả X-Cache-Status: HIT với độ trễ < 15ms.',
      },
    ],
    notes: 'Quy tắc Directory cần có dấu gạch chéo đầu và cuối (/static/img/) theo tài liệu CMC.',
    docsUrl: 'https://cloud.cmctelecom.vn/docs/cdn/cache-rules',
  },
  {
    id: 'c5',
    group: 'cdn',
    priority: 'P0',
    title: 'C5: Tôn trọng Origin Cache-Control',
    goal: 'Kiểm tra CDN có tôn trọng các chỉ thị Cache-Control từ origin (public, no-store, etag, vary) hay bị ghi đè bởi cấu hình portal.',
    portalSteps: [
      'Vào CDN > Caching > Origin Cache-Control: Bật "Respect Origin Headers"',
      'Gọi /api/cache/no-store -> CDN không được lưu cache',
      'Gọi /api/cache/etag kèm If-None-Match -> mong đợi HTTP 304 Not Modified',
      'Gọi /api/cache/set-cookie -> kiểm tra CDN có bỏ qua cookie hay lưu cả Set-Cookie',
    ],
    requests: [
      {
        method: 'GET',
        path: '/api/cache/no-store',
        runnable: true,
        curl: 'curl -i https://cdn.labshop.local/api/cache/no-store',
      },
      {
        method: 'GET',
        path: '/api/cache/etag',
        runnable: true,
        curl: 'curl -i https://cdn.labshop.local/api/cache/etag',
      },
    ],
    expect: [
      {
        condition: 'Origin gửi no-store',
        expectedResult: 'CDN luôn chuyển request về origin, X-Cache-Status: BYPASS/MISS, không lưu đệm.',
      },
      {
        condition: 'Origin gửi ETag',
        expectedResult: 'Edge phản hồi 304 khi client gửi đúng If-None-Match.',
      },
    ],
    notes: 'Nếu portal đặt Edge Cache TTL cố định mà không bật Respect Origin, CDN có thể bỏ qua no-store.',
    docsUrl: 'https://cloud.cmctelecom.vn/docs/cdn/caching',
  },
  {
    id: 'c6',
    group: 'cdn',
    priority: 'P0',
    title: 'C6: Purge Cache (Custom URL & Purge All)',
    goal: 'Xóa tức thời cache tại các PoP khi cập nhật tài nguyên, theo dõi thống kê số request về origin tại /api/stats.',
    portalSteps: [
      'Gọi /api/time để CDN lưu cache HIT',
      'Ghi nhận số hit origin tại /api/stats',
      'Vào Portal > CDN > Purge Cache > Custom Purge > Nhập URL /api/time',
      'Gửi Purge, kiểm tra tab History trạng thái Completed',
      'Gọi lại /api/time -> Quan sát X-Cache-Status chuyển về MISS và origin hit tăng thêm 1',
    ],
    requests: [
      {
        method: 'GET',
        path: '/api/time',
        runnable: true,
        curl: 'curl -i https://cdn.labshop.local/api/time',
      },
      {
        method: 'GET',
        path: '/api/stats',
        runnable: true,
        curl: 'curl -i https://cdn.labshop.local/api/stats',
      },
    ],
    expect: [
      {
        condition: 'Sau khi Purge thành công',
        expectedResult: 'Request tiếp theo bắt buộc về Origin lấy dữ liệu mới (X-Cache-Status: MISS).',
      },
    ],
    notes: 'Custom Purge hỗ trợ tối đa 30 URL mỗi lần, không chấp nhận ký tự đại diện wildcard (*).',
    docsUrl: 'https://cloud.cmctelecom.vn/docs/cdn/purge',
  },
  {
    id: 'c10',
    group: 'cdn',
    priority: 'P0',
    title: 'C10: Nén dữ liệu tại Edge (Gzip / Brotli) & Auto Minify',
    goal: 'Tận dụng khả năng nén tại PoP CDN để giảm kích thước tải về máy khách mà không tốn CPU của máy chủ Origin.',
    portalSteps: [
      'Vào CDN > Optimization > Compression: Bật Gzip & Brotli',
      'Gọi /api/payload?kb=512 với header Accept-Encoding: gzip, br',
      'Kiểm tra Content-Encoding ở header phản hồi và dung lượng Content-Length thực tế',
    ],
    requests: [
      {
        method: 'GET',
        path: '/api/payload?kb=512',
        runnable: true,
        curl: 'curl -i -H "Accept-Encoding: gzip, br" "https://cdn.labshop.local/api/payload?kb=512"',
      },
    ],
    expect: [
      {
        condition: 'Nén thành công',
        expectedResult: 'Content-Encoding là "gzip" hoặc "br", dung lượng giảm từ 512KB xuống dưới 15KB nhờ văn bản lặp lại.',
      },
    ],
    notes: 'Khi ORIGIN_GZIP=off trên origin, việc nén hoàn toàn do Edge PoP đảm nhận.',
    docsUrl: 'https://cloud.cmctelecom.vn/docs/cdn/compression',
  },
  {
    id: 'c17',
    group: 'cdn',
    priority: 'P0',
    title: 'C17: Hashed Assets & Immutable Cache',
    goal: 'Hiểu chiến lược đặt tên file có hash (/assets/app-[hash].js) kết hợp Cache-Control immutable 1 năm để tăng tốc tối đa mà không lo cache cũ.',
    portalSteps: [
      'Xem mã nguồn index.html để tìm tên file bundle dạng /assets/index-[hash].js',
      'Kiểm tra header Cache-Control: public, max-age=31536000, immutable',
      'So sánh với file tĩnh không có hash (/static/img/products/p01.jpg)',
      'Thực hiện Purge thử nghiệm khi có phiên bản build mới',
    ],
    requests: [
      {
        method: 'GET',
        path: '/index.html',
        runnable: true,
        curl: 'curl -i https://cdn.labshop.local/index.html',
      },
    ],
    expect: [
      {
        condition: 'Hashed assets',
        expectedResult: 'Có thể lưu cache vĩnh viễn ở trình duyệt và CDN; chỉ cần cập nhật index.html khi deploy.',
      },
    ],
    notes: 'Kịch bản bổ sung theo tiêu chuẩn Phase 3 / Vite SPA.',
    docsUrl: 'https://cloud.cmctelecom.vn/docs/cdn/best-practices',
  },
  {
    id: 'c18',
    group: 'cdn',
    priority: 'P0',
    title: 'C18: Deploy bản mới nhưng CDN vẫn giữ index.html cũ',
    goal: 'Xử lý tình huống éo le khi deploy mã nguồn mới nhưng người dùng vẫn thấy phiên bản cũ do CDN cache index.html 30 phút.',
    portalSteps: [
      'Vào CDN > Cache Rules > Thêm quy tắc No Cache cho path: /index.html và /',
      'Đặt Browser Cache TTL cho index.html = 0 (no-cache)',
      'Thực hiện Custom Purge URL https://cdn.labshop.local/index.html',
      'Tải lại trang và kiểm tra file bundle hash mới nhất được nạp',
    ],
    requests: [
      {
        method: 'GET',
        path: '/index.html',
        runnable: true,
        curl: 'curl -i https://cdn.labshop.local/index.html',
      },
    ],
    expect: [
      {
        condition: 'Đã áp dụng No Cache cho index.html',
        expectedResult: 'Trình duyệt luôn lấy bản index.html mới nhất, từ đó trỏ đúng các bundle hash mới.',
      },
    ],
    notes: 'index.html luôn phải no-cache để đảm bảo tính sẵn sàng khi deploy liên tục.',
    docsUrl: 'https://cloud.cmctelecom.vn/docs/cdn/troubleshooting',
  },

  // ==================== WAAP SCENARIOS ====================
  {
    id: 'w1',
    group: 'waap',
    priority: 'P0',
    title: 'W1: Onboard ứng dụng lên CMC WAAP',
    goal: 'Cấu hình bảo vệ trang web qua CMC WAAP ở chế độ Detect (giám sát), xác thực kết nối và gửi chuỗi kiểm tra.',
    portalSteps: [
      'Vào CMC Portal > WAAP > Applications > Add Application',
      'Nhập Domain, chọn Mode ban đầu là "Detect", Endpoint Type IP/Domain Origin',
      'Lấy CNAME từ hệ thống WAAP và cấu hình DNS',
      'Vào tab Verify để xác nhận WAAP đã nhận diện được traffic hợp lệ',
    ],
    requests: [
      {
        method: 'GET',
        path: '/api/health',
        runnable: true,
        curl: 'curl -i https://waap.labshop.local/api/health',
      },
    ],
    expect: [
      {
        condition: 'Chế độ Detect',
        expectedResult: 'Traffic đi qua WAAP bình thường, HTTP 200 OK.',
      },
    ],
    notes: 'Luôn khởi đầu bằng Detect trong 24-48 giờ để phát hiện và xử lý False Positive trước khi chuyển sang Block.',
    docsUrl: 'https://cloud.cmctelecom.vn/docs/waap/onboard',
  },
  {
    id: 'w2',
    group: 'waap',
    priority: 'P0',
    title: 'W2: Chế độ Detect so với Block (View Logs & Block ID)',
    goal: 'Quan sát sự khác biệt khi gửi mẫu tấn công ở chế độ Detect (ghi log nhưng cho qua) so với Block (chặn ngay với mã 403 Forbidden và Block ID).',
    portalSteps: [
      'Gửi chuỗi probe XSS qua tham số URL /api/products?q=<script>alert(1)</script>',
      'Ở Mode Detect: Request vẫn tới Origin, vào Portal > WAAP > Logs thấy bản ghi MATCH nhưng Action là LOG',
      'Chuyển Mode sang Block: Gửi lại chuỗi trên -> Nhận HTTP 403 Forbidden từ WAAP',
      'Đối chiếu Block ID trong log portal với thông tin phản hồi của WAAP',
    ],
    requests: [
      {
        method: 'GET',
        path: '/api/products?q=%3Cscript%3Ealert(1)%3C%2Fscript%3E',
        runnable: true,
        isProbe: true,
        curl: "curl -i 'https://waap.labshop.local/api/products?q=<script>alert(1)</script>'",
      },
    ],
    expect: [
      {
        condition: 'Mode Detect',
        expectedResult: 'HTTP 200, phản hồi từ Origin (có X-Origin-Instance), log trên portal ghi nhận Detect.',
      },
      {
        condition: 'Mode Block',
        expectedResult: 'HTTP 403 Forbidden, không có header X-Origin-Instance (bị chặn tại Edge WAAP).',
      },
    ],
    notes: 'Chỉ chạy kịch bản thử nghiệm trên hệ thống lab của bạn.',
    docsUrl: 'https://cloud.cmctelecom.vn/docs/waap/modes',
  },
  {
    id: 'w3',
    group: 'waap',
    priority: 'P0',
    title: 'W3: Thử nghiệm WAF Rules theo các Vector tấn công',
    goal: 'Kiểm tra khả năng bảo vệ của các bộ luật WAF Core Rules trên nhiều thành phần HTTP khác nhau: GET args, Request Body, Header, Path traversal.',
    portalSteps: [
      'Vector 1 (GET Query SQLi): /api/products?q=1%27%20OR%20%271%27=%271',
      'Vector 2 (Path Traversal): /api/echo/..%2f..%2fetc%2fpasswd',
      'Vector 3 (Body XSS): POST review chứa thẻ img với onerror',
      'Xem báo cáo chi tiết sự kiện tấn công theo từng Rule ID trong tab Logs của CMC WAAP',
    ],
    requests: [
      {
        method: 'GET',
        path: "/api/products?q=1'%20OR%20'1'='1",
        runnable: true,
        isProbe: true,
        curl: "curl -i \"https://waap.labshop.local/api/products?q=1' OR '1'='1\"",
      },
      {
        method: 'GET',
        path: '/api/echo/..%2f..%2fetc%2fpasswd',
        runnable: true,
        isProbe: true,
        curl: 'curl -i "https://waap.labshop.local/api/echo/..%2f..%2fetc%2fpasswd"',
      },
    ],
    expect: [
      {
        condition: 'WAF Block kích hoạt',
        expectedResult: 'Các yêu cầu đều bị chặn với HTTP 403/400 ngay tại Edge, không chạm tới máy chủ Go.',
      },
    ],
    notes: 'Bộ luật Core Rules tự động chuẩn hóa URL decode trước khi kiểm tra chữ ký tấn công.',
    docsUrl: 'https://cloud.cmctelecom.vn/docs/waap/core-rules',
  },
  {
    id: 'w4',
    group: 'waap',
    priority: 'P0',
    title: 'W4: Custom Rule (Chặn từ khóa nhạy cảm)',
    goal: 'Tạo quy tắc tùy biến trên portal để chặn các chuỗi từ khóa cụ thể trong nội dung body hoặc header.',
    portalSteps: [
      'Vào WAAP > Applications > Chọn site > Custom Rules > Add Rule',
      'Đặt tên: "Block Malicious Keyword"',
      'Match Condition: Request Body chứa từ khóa "eval(" hoặc "drop table"',
      'Action: BLOCK (Trả về 403 Forbidden)',
      'Gửi yêu cầu thử nghiệm chứa từ khóa để xác nhận luật hoạt động',
    ],
    requests: [
      {
        method: 'POST',
        path: '/api/echo',
        body: JSON.stringify({ comment: 'Please drop table users' }),
        runnable: true,
        isProbe: true,
        curl: 'curl -i -X POST -H "Content-Type: application/json" -d \'{"comment":"Please drop table users"}\' https://waap.labshop.local/api/echo',
      },
    ],
    expect: [
      {
        condition: 'Khớp Custom Rule',
        expectedResult: 'WAAP phản hồi 403 Forbidden ngay lập tức; log portal ghi nhận Custom Rule ID đã kích hoạt.',
      },
    ],
    notes: 'Custom Rules có thứ tự ưu tiên cao hơn Core Rules thông thường.',
    docsUrl: 'https://cloud.cmctelecom.vn/docs/waap/custom-rules',
  },
  {
    id: 'w5',
    group: 'waap',
    priority: 'P0',
    title: 'W5: Xử lý False Positive & Cấu hình Whitelist',
    goal: 'Thực hành giải quyết trường hợp người dùng gửi nội dung hợp lệ nhưng chứa từ khóa giống mã nguồn/SQL nên bị WAF chặn nhầm.',
    portalSteps: [
      'Gửi câu hỗ trợ kỹ thuật: "Please SELECT the best option FROM the list; my code: if (a < b) { return \"<div>\"; }"',
      'Quan sát yêu cầu bị chặn 403 bởi Rule SQLi / XSS của WAF',
      'Vào Portal > Logs > Tìm log bị chặn > Copy Rule ID gây ra False Positive',
      'Vào WAAP > Whitelist / Exception Rules > Thêm ngoại lệ cho Rule ID đó trên URL /api/support/*',
      'Gửi lại câu hỗ trợ để xác nhận yêu cầu đã đi qua thành công',
    ],
    requests: [
      {
        method: 'POST',
        path: '/api/echo',
        body: JSON.stringify({ message: 'Please SELECT the best option FROM the list; if (a < b) { return "<div>"; }' }),
        runnable: true,
        isProbe: true,
        curl: 'curl -i -X POST -H "Content-Type: application/json" -d \'{"message":"Please SELECT the best option FROM the list"}\' https://waap.labshop.local/api/echo',
      },
    ],
    expect: [
      {
        condition: 'Trước khi Whitelist',
        expectedResult: 'HTTP 403 Forbidden do dính chữ ký SELECT ... FROM.',
      },
      {
        condition: 'Sau khi thêm Whitelist ngoại lệ',
        expectedResult: 'HTTP 200 OK, Origin xử lý bình thường.',
      },
    ],
    notes: 'Kỹ năng quan trọng nhất của quản trị viên WAF để không làm gián đoạn trải nghiệm người dùng thật.',
    docsUrl: 'https://cloud.cmctelecom.vn/docs/waap/exceptions',
  },
  {
    id: 'w6',
    group: 'waap',
    priority: 'P0',
    title: 'W6: API Protection bằng OpenAPI Schema',
    goal: 'Bảo vệ API bằng cách nhập file đặc tả OpenAPI (Swagger) để WAAP tự động chặn mọi endpoint hoặc method không được khai báo.',
    portalSteps: [
      'Vào WAAP > API Protection > Import Schema',
      'Tải lên file https://waap.labshop.local/openapi.json',
      'Thiết lập Action: LOG cho tất cả, chuyển riêng DELETE /api/admin/products/{id} sang BLOCK',
      'Thử gọi endpoint không có trong schema (/api/fake-secret-path) để kiểm chứng',
    ],
    requests: [
      {
        method: 'GET',
        path: '/openapi.json',
        runnable: true,
        curl: 'curl -i https://waap.labshop.local/openapi.json',
      },
      {
        method: 'GET',
        path: '/api/unknown-ghost-endpoint',
        runnable: true,
        curl: 'curl -i https://waap.labshop.local/api/unknown-ghost-endpoint',
      },
    ],
    expect: [
      {
        condition: 'Endpoint nằm ngoài schema',
        expectedResult: 'WAAP chặn với mã 400 hoặc 404/403 trước khi đến router của Go.',
      },
    ],
    notes: 'Giúp ngăn chặn tấn công dò quét endpoint ẩn (shadow API).',
    docsUrl: 'https://cloud.cmctelecom.vn/docs/waap/api-protection',
  },
  {
    id: 'w7',
    group: 'waap',
    priority: 'P0',
    title: 'W7: Rate Limiting chống Brute-Force & DoS',
    goal: 'Thiết lập ngưỡng giới hạn số lượng request trên mỗi IP trong 1 phút để bảo vệ các endpoint nhạy cảm như đăng nhập (/api/auth/login).',
    portalSteps: [
      'Vào WAAP > Rate Limiting > Add Rule',
      'URL Path: /api/auth/login, Method: POST',
      'Threshold: 5 requests / 10 seconds, Action: BLOCK 60 seconds (HTTP 429 Too Many Requests)',
      'Sử dụng tính năng Burst Test trong Runner (gửi 15 request liên tiếp) để quan sát histogram phản hồi',
    ],
    requests: [
      {
        method: 'POST',
        path: '/api/auth/login',
        body: JSON.stringify({ email: 'test@labshop.local', password: 'wrong' }),
        runnable: true,
        curl: 'curl -i -X POST -H "Content-Type: application/json" -d \'{"email":"test@labshop.local","password":"wrong"}\' https://waap.labshop.local/api/auth/login',
      },
    ],
    expect: [
      {
        condition: '5 request đầu tiên',
        expectedResult: 'HTTP 401 Unauthorized từ Origin (mật khẩu sai).',
      },
      {
        condition: 'Từ request thứ 6 trở đi',
        expectedResult: 'HTTP 429 Too Many Requests do WAAP chặn tại Edge.',
      },
    ],
    notes: 'Có thể thiết lập Rate Limit theo IP hoặc theo Header/Token.',
    docsUrl: 'https://cloud.cmctelecom.vn/docs/waap/rate-limit',
  },
  {
    id: 'w8',
    group: 'waap',
    priority: 'P0',
    title: 'W8: Bot Management (Predefined & Custom User-Agent Rules)',
    goal: 'Ngăn chặn các công cụ quét tự động, scraper và bot độc hại dựa trên danh mục nhận diện chữ ký Bot của CMC WAAP.',
    portalSteps: [
      'Vào WAAP > Bot Management > Bật tính năng Predefined Bot Defense',
      'Chọn chặn các nhóm: Search Engines độc hại, Scrapers, Tool curl / python-requests / sqlmap',
      'Chạy lệnh cURL từ máy trạm với User-Agent giả lập bot',
      'Quan sát cURL bị từ chối kết nối hoặc nhận 403 Forbidden',
    ],
    requests: [
      {
        method: 'GET',
        path: '/api/products',
        runnable: false,
        nonRunnableReason: 'Trình duyệt cấm can thiệp User-Agent header (Forbidden header name). Bạn cần chạy lệnh qua Terminal/cURL bên ngoài.',
        curl: 'curl -i -A "sqlmap/1.7" https://waap.labshop.local/api/products',
      },
    ],
    expect: [
      {
        condition: 'User-Agent thuộc danh sách đen',
        expectedResult: 'HTTP 403 Forbidden từ WAAP Bot Protection.',
      },
    ],
    notes: 'Kịch bản đánh dấu runnable=false vì hạn chế bảo mật của trình duyệt web.',
    docsUrl: 'https://cloud.cmctelecom.vn/docs/waap/bot-management',
  },

  // ==================== SHOP SCENARIOS ====================
  {
    id: 'shop-flow',
    group: 'shop',
    priority: 'P0',
    title: 'Shop: Toàn trình Nghiệp vụ Thương mại điện tử',
    goal: 'Trải nghiệm luồng mua sắm end-to-end: Đăng ký -> Duyệt hàng -> Giỏ hàng -> Áp Coupon -> Thanh toán với Idempotency-Key -> Đánh giá.',
    portalSteps: [
      '1. Khách hàng duyệt danh mục thiết bị và chọn sản phẩm',
      '2. Thêm vào giỏ hàng, áp mã giảm giá WELCOME10',
      '3. Thực hiện Checkout: sinh khóa Idempotency-Key duy nhất',
      '4. Thanh toán đơn hàng qua Mock Gateway',
      '5. Đăng nhập quyền Quản trị (Admin) để cập nhật trạng thái Giao hàng (DELIVERED)',
      '6. Khách hàng quay lại đánh giá 5 sao cho sản phẩm vừa mua',
    ],
    requests: [
      {
        method: 'GET',
        path: '/api/products',
        runnable: true,
        curl: 'curl -i https://shop.labshop.local/api/products?page=1&page_size=8',
      },
      {
        method: 'GET',
        path: '/api/categories',
        runnable: true,
        curl: 'curl -i https://shop.labshop.local/api/categories',
      },
    ],
    expect: [
      {
        condition: 'Luồng mua hàng chuẩn',
        expectedResult: 'Tất cả các giao dịch dữ liệu đều được lưu trữ an toàn trong SQLite Origin, không thất thoát đơn.',
      },
    ],
    notes: 'Mọi thao tác đều được ghi lại trong Lab Console phía dưới màn hình.',
    docsUrl: 'https://cloud.cmctelecom.vn/docs/labshop/shop-flow',
  },
];

export const PROBES = [
  { category: 'SQL Injection', payload: "' OR '1'='1' --", description: 'Thử nghiệm vượt qua câu lệnh điều kiện SQL' },
  { category: 'SQL Injection', payload: '1 UNION SELECT NULL,NULL--', description: 'Dò tìm cấu trúc cột cơ sở dữ liệu' },
  { category: 'Cross-Site Scripting (XSS)', payload: '<script>alert(1)</script>', description: 'Chèn mã script cổ điển vào tham số' },
  { category: 'Cross-Site Scripting (XSS)', payload: '"><img src=x onerror=alert(1)>', description: 'Tấn công phá vỡ thuộc tính HTML' },
  { category: 'Path Traversal', payload: '../../etc/passwd', description: 'Truy cập tệp tin hệ thống nhạy cảm' },
  { category: 'Path Traversal', payload: '..%2f..%2fetc%2fpasswd', description: 'Mã hóa URL kép nhằm lách kiểm tra chuỗi' },
  { category: 'Command Injection', payload: '; cat /etc/passwd', description: 'Nối lệnh thực thi hệ điều hành' },
  { category: 'False Positive', payload: 'Please SELECT the best option FROM the list; my code: if (a < b) { return "<div>"; }', description: 'Nội dung người dùng hợp lệ dễ bị chặn nhầm' },
];
