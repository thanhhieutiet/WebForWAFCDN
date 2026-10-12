# PROMPT v2 — LabShop: web Go (Clean Architecture) làm origin cho lab WAAP + CDN của CMC Telecom

(Thay thế hoàn toàn bản plan trước. Đọc hết trước khi code.)

## 0. Bối cảnh và quy tắc làm việc

- Thư mục làm việc: "C:\CMC\WAF\WebForWAF&CDN" (Windows). Ký tự `&` là ký tự đặc biệt của cmd/PowerShell:
  LUÔN đặt đường dẫn trong dấu nháy kép, không dùng tên thư mục làm tên module/image/tag, mọi script phải chịu được
  đường dẫn có `&` và dấu cách.
- Mục tiêu: một web Go có nghiệp vụ thật + API thật (không phải HTML tĩnh) để thực hành WAAP và Multi-CDN của CMC
  (docs: https://docs.cmctelecom.io/). Web KHÔNG được cố ý có lỗ hổng: mọi request bị chặn phải đến từ WAAP, không phải từ app.
- Luồng triển khai: build bằng Docker, push GitHub, trên EC2 `git pull` rồi `docker compose up -d --build`.
- Cách làm: đi theo phase ở mục 11. Mỗi phase gồm code + test + chạy thử (`go build ./...`, `go vet ./...`,
  `go test -race ./...`, `docker compose build`) rồi commit riêng. Không hỏi lại trừ khi bị chặn; quyết định nhỏ ghi vào
  `docs/DECISIONS.md`. Không thêm tính năng ngoài prompt, không bỏ bớt. Không commit secret. Không nới lỏng bảo mật cho tiện.
  Cuối mỗi phase báo: đã làm gì, lệnh kiểm tra + kết quả, điểm nào lệch prompt và vì sao.
- Dev trên Windows, chạy thật trên Linux:
  - `.gitattributes` ép LF cho `*.sh`, `*.go`, `Dockerfile*`, `*.conf`, `*.template`, `*.yml`, `*.yaml`, `*.sql`, `*.json`.
  - Script dev viết PowerShell (`scripts/*.ps1`), script deploy viết bash (`scripts/deploy.sh`).
  - Smoke test viết bằng Go (`cmd/smoke`) để chạy được ở cả hai môi trường.

## 1. Stack

- Go >= 1.23 (dùng bản stable mới nhất có trên máy), module name `labshop`.
- Router `go-chi/chi/v5` (hoặc net/http ServeMux 1.22+, ghi DECISIONS). SQLite thuần Go `modernc.org/sqlite`
  (CGO_ENABLED=0, bật WAL + busy_timeout). Migrations là file SQL nhúng bằng `embed` + runner tối giản (bảng `schema_migrations`).
- JWT `golang-jwt/jwt/v5` (HS256, TTL 60 phút). Hash mật khẩu argon2id (`golang.org/x/crypto/argon2`). UUID `google/uuid`.
  WebSocket `coder/websocket`. Log `log/slog` (JSON). Config từ env, validate và fail-fast.
- Spec-first OpenAPI 3.0.3: `oapi-codegen` (sinh types + server interface) và `kin-openapi` (validate trong test).
  Nếu oapi-codegen vướng thì viết tay spec 3.0.3 + contract test (mục 6).
- Test: `testing` + `testify`. Lint: golangci-lint (bật gofmt, govet, staticcheck, errcheck, depguard).
- Frontend: `html/template` + `embed`, vanilla JS/CSS, KHÔNG npm, KHÔNG bundler, KHÔNG framework.
- Ít dependency nhất có thể. Muốn đổi thư viện thì ghi lý do vào DECISIONS.md.

## 2. Clean Architecture (bắt buộc)

Cấu trúc thư mục:

```text
cmd/
  server/main.go            # composition root: dựng dependency, KHÔNG chứa logic
  genassets/main.go         # sinh file tĩnh (ảnh, bin, pdf, zip...) lúc build
  smoke/main.go             # smoke test chạy từ xa theo BASE_URL
internal/
  domain/                   # entity, value object, rule nghiệp vụ, domain error. CHỈ import stdlib
    user/ catalog/ cart/ coupon/ order/ review/ support/ money/
  usecase/                  # application service. Định nghĩa PORT (interface) mà nó cần
    auth/ catalog/ cart/ order/ review/ support/ admin/ lab/ tutorial/
    port/                   # repository, TxManager, Clock, IDGenerator, PasswordHasher, TokenIssuer,
                            # PaymentGateway, HitCounter, ScenarioRepository
  adapter/
    http/                   # router, middleware, handler mỏng, DTO, presenter (map error -> status), oapi gen
    web/                    # html/template + embed: storefront, tutorial, docs page
    repository/sqlite/      # cài đặt các repository port + TxManager
    security/               # argon2, jwt
    payment/mock/           # PaymentGateway giả lập
    counter/memory/         # HitCounter trong bộ nhớ
    scenario/embedded/      # đọc kịch bản tutorial từ JSON nhúng
  infrastructure/           # config, logger, mở DB, migrations, seed, bootstrap server
api/openapi.yaml            # nguồn sự thật của spec (OpenAPI 3.0.3)
web/static/                 # css/js viết tay (nginx phục vụ), KHÔNG embed vào binary
nginx/                      # nginx.conf.template
scripts/ docs/ test/
```

Dependency rule (hướng vào trong):

- domain: không import gì ngoài stdlib (không net/http, không database/sql, không tag json trên entity).
- usecase: chỉ import domain + stdlib. Mọi thứ bên ngoài (DB, JWT, thời gian, sinh ID, thanh toán) đi qua port do usecase định nghĩa.
- adapter: cài port hoặc gọi usecase; adapter KHÔNG import adapter khác (http không import sqlite).
- infrastructure và cmd: chỉ nối dây.
- Viết `internal/archtest` (parse import bằng go/parser hoặc `go list -deps`) làm test fail khi vi phạm; bật thêm depguard trong golangci-lint.

Quy ước:

- Domain error có kiểu rõ ràng: NotFound, Conflict, Validation, Unauthorized, Forbidden, BusinessRule(code).
  Presenter map sang HTTP: 400 JSON/validation sai, 401, 403 sai quyền, 404 không có hoặc không phải chủ sở hữu,
  409 trùng/xung đột (email, slug, review đã có, hết hàng), 422 vi phạm nghiệp vụ (coupon sai, chuyển trạng thái sai).
- Handler chỉ: parse + validate DTO -> gọi usecase -> present. Không có logic nghiệp vụ trong handler.
- Giao dịch qua port `TxManager.WithTx(ctx, func(ctx) error)`. Thời gian qua `Clock`, ID qua `IDGenerator`.
- Tiền là int64 VND (value object `money.VND`). Thời gian RFC3339 UTC. JSON snake_case.
  Từ chối field lạ (DisallowUnknownFields), giới hạn body bằng http.MaxBytesReader (mặc định 1MB, upload riêng).
- Lỗi trả về dạng: {"error":{"code","message","details":[{"field","issue"}],"request_id"}}.
- Phân trang: ?page=1&page_size=20 (max 100). Trả {items,page,page_size,total}.

Chiến lược test:

- Unit test domain + usecase bằng fake repository (không cần DB). Mục tiêu coverage domain+usecase >= 80%.
- Integration test cho repository sqlite (DB tạm).
- Handler test bằng httptest.
- archtest, contract test OpenAPI, payload-hygiene test (mục 9).

## 3. Nghiệp vụ LabShop (shop đồ công nghệ)

Vai trò: customer, admin.

- Account: đăng ký (email duy nhất, mật khẩu >= 10 ký tự), đăng nhập, xem/sửa hồ sơ (tên, sđt, địa chỉ), đổi mật khẩu.
  Không khóa tài khoản ở tầng app.
- Catalog: category, product (name, slug duy nhất, mô tả, giá > 0, stock >= 0, image path, status active/archived,
  rating_avg, rating_count). Tìm kiếm ?q=, lọc category/min_price/max_price, sắp xếp price|newest|rating.
- Cart (lưu server theo user): quantity mỗi dòng 1..10, tối đa 20 dòng, không vượt stock, không thêm product archived.
  Giá chỉ chốt khi checkout (không tin giá từ client).
- Coupon: percent (1..50, có trần giảm tối đa) hoặc fixed; min_subtotal; valid_from/valid_to; max_uses tổng;
  mỗi user dùng 1 lần. Áp 1 coupon/giỏ.
  Seed: WELCOME10 (10%, trần 100.000, min 200.000), FREESHIP (miễn phí ship), EXPIRED5 (đã hết hạn), LIMIT1 (max_uses=1).
- Shipping: phí cố định 30.000đ, miễn phí khi subtotal >= 500.000đ (cấu hình được).
- Order:
  - Tạo từ giỏ hàng, có header `Idempotency-Key` (cùng key + cùng payload trả lại cùng đơn; cùng key khác payload trả 422 `idempotency_conflict`).
  - Tính tổng ở server: subtotal - discount + shipping.
  - Trừ kho nguyên tử trong transaction (`UPDATE ... SET stock = stock - ? WHERE id=? AND stock >= ?`, kiểm tra rows affected). Hủy thì hoàn kho.
  - Trạng thái: PENDING -> {PAID, CANCELLED}; PAID -> {SHIPPED, CANCELLED}; SHIPPED -> {DELIVERED}; DELIVERED và CANCELLED là terminal.
    Bảng chuyển trạng thái nằm trong domain.
  - payment_method: `mock_gateway` hoặc `cod`. Không thu số thẻ, không có trường thẻ. `POST /orders/{id}/pay` gọi PaymentGateway mock
    (độ trễ 200–800ms, env MOCK_PAYMENT_FAIL_RATE mặc định 0).
  - Customer chỉ xem/hủy được đơn của mình. Admin đổi trạng thái theo bảng chuyển.
- Review: chỉ user có đơn DELIVERED chứa sản phẩm mới được review; 1 review/(user, product); rating 1..5; comment 1..2000 ký tự
  (text tự do, cố ý để test false positive). Cập nhật rating_avg/rating_count. Tác giả xóa được review của mình.
- Support ticket: OPEN -> ANSWERED -> CLOSED; customer thêm message khi chưa CLOSED; admin trả lời/đóng; attachment tối đa 3 file/ticket,
  mỗi file <= 2MB, đuôi cho phép: png, jpg, pdf, txt, log. Mặc định KHÔNG lưu nội dung file (`STORE_UPLOADS=false`): chỉ ghi metadata
  (tên đã sanitize, size, sha256, content-type) và không có endpoint tải lại.
- Admin: CRUD product (xóa = archive), xem đơn/ticket/user, đổi trạng thái đơn, dashboard (doanh thu theo ngày, top sản phẩm, số đơn theo trạng thái).
- App KHÔNG tự rate limit (`APP_RATE_LIMIT_ENABLED=false` mặc định) để WAAP là bên giới hạn.

## 4. API nghiệp vụ (tiền tố /api, JSON)

Auth:

- POST /auth/register
- POST /auth/login
- GET,PATCH /auth/me
- POST /auth/change-password

Catalog:

- GET /categories
- GET /products (q, category, min_price, max_price, sort, page, page_size)
- GET /products/{slug}

Cart:

- GET /cart
- PUT /cart/items/{product_id}
- DELETE /cart/items/{product_id}
- DELETE /cart
- POST,DELETE /cart/coupon

Orders:

- POST /orders (Idempotency-Key)
- GET /orders
- GET /orders/{id}
- POST /orders/{id}/pay
- POST /orders/{id}/cancel

Reviews:

- GET,POST /products/{slug}/reviews
- DELETE /reviews/{id}

Support:

- POST,GET /support/tickets
- GET /support/tickets/{id}
- POST /support/tickets/{id}/messages
- POST /support/tickets/{id}/attachments (multipart)

Admin:

- GET,POST /admin/products
- PUT,PATCH,DELETE /admin/products/{id}
- GET /admin/orders
- PATCH /admin/orders/{id}/status
- GET /admin/tickets
- POST /admin/tickets/{id}/reply
- GET /admin/users
- GET /admin/stats

Tutorial:

- GET /tutorial/scenarios

Trang HTML (tối giản, phase 3):

- / (trang chủ)
- /products
- /products/{slug}
- /cart
- /checkout
- /orders
- /login
- /register
- /support
- /admin
- /tutorial
- /tutorial/{id}
- /docs.

## 5. Lab endpoints (giữ nguyên đường dẫn đã chốt, đặt trong usecase/lab + adapter)

- GET /api/health {status:"ok"}, nhẹ

- GET /api/time {generated_at, request_id}. KHÔNG gửi Cache-Control/Expires/ETag
  (để quan sát cache mặc định 30 phút của CDN)

- GET /api/cache/{mode}?v= mode: public (max-age=60) | no-store | etag (hỗ trợ If-None-Match -> 304) | vary (Vary: Accept-Encoding, Accept-Language)
  | set-cookie (kèm Set-Cookie). Luôn trả {mode, generated_at, request_id, v}

- GET /api/headers echo header nhận được (che Authorization, Cookie), IP kết nối, chuỗi X-Forwarded-For

- GET|POST /api/echo, /api/echo/{path} phản chiếu query, JSON body, form, header, cookie dưới dạng JSON (mọi giá trị là chuỗi đã escape bởi encoder)

- GET /api/status/{code} trả đúng mã trong {200,204,301,302,304,400,401,403,404,429,500,502,503,504}

- GET /api/slow?ms= sleep theo context, giới hạn MAX_SLOW_MS (mặc định 60000), tối đa 20 request đồng thời (quá thì 503 {error busy})

- GET /api/payload?kb= JSON/text nén được, tối đa MAX_PAYLOAD_KB (mặc định 5120)

- GET /api/stats đếm request tới origin theo path kể từ lúc khởi động (port HitCounter, middleware tăng bộ đếm)

- POST /api/stats/reset yêu cầu header X-Admin-Token

- POST /api/upload multipart, giới hạn MAX_UPLOAD_MB, trả sha256/size/tên đã sanitize/content-type, không lưu file

- WS /ws/echo WebSocket echo (giới hạn kích thước message, ping/pong, timeout)

- GET /openapi.json spec OpenAPI 3.0.x đầy đủ

- GET /robots.txt cho phép tất cả

- GET /sitemap.xml liệt kê trang + sản phẩm active (cache public)

## 6. OpenAPI 3.0 (WAAP > API Protection > Upload chỉ nhận OpenAPI/Swagger 3.0 .json, tối đa 5MB, cần đủ endpoint/method/request/response)

- `api/openapi.yaml` là OpenAPI **3.0.3** (không phải 3.1). Khai báo đầy đủ mọi route JSON (business + lab, trừ WS, static, HTML):
  parameters, requestBody schema, response schema cho 2xx/4xx, securitySchemes bearer JWT, tags.
- Phục vụ tại /openapi.json (chuyển YAML sang JSON lúc khởi động hoặc lúc build). Kích thước < 1MB.
- Contract test: (1) spec load + Validate bằng kin-openapi; (2) mọi route đã đăng ký trong router đều có trong spec và ngược lại
  (trừ danh sách loại trừ: /ws/_, /static/_, /contents/\*, trang HTML, /docs, /openapi.json); (3) vài response mẫu khớp schema.
- /docs: trang HTML tự render danh sách endpoint từ /openapi.json bằng JS thuần (không Swagger UI, không CDN ngoài). Bật/tắt bằng ENABLE_DOCS.

## 7. File tĩnh và nginx

`cmd/genassets` sinh vào thư mục đầu ra (gitignore), không dùng ảnh có bản quyền:

- img/products/p01..p40.jpg (800x600, nhiều kích cỡ 150–400KB), img/banner-large.jpg (~2.5MB, nhiễu để JPEG khó nén),
  img/sample.png, sample.gif, sample.svg (webp: bỏ qua nếu stdlib không encode được).
- docs/sample.pdf (PDF tối giản tự viết), files/sample.zip, files/1mb.bin, 10mb.bin, 50mb.bin (math/rand seed cố định),
  data/sample.json, fonts/dummy.woff2 (byte giả, ghi rõ KHÔNG phải font thật), private/secret-1.txt, secret-2.jpg.
  Viết tay trong web/static: css/app.css + app.min.css, js/app.js + app.min.js (bản không minify để test Auto Minify; bản .min tạo bằng
  genassets bằng cách xóa comment và khoảng trắng đơn giản), js/runner.js, js/tutorial.js, js/storefront.js.

nginx (dùng cơ chế `/etc/nginx/templates/*.template` của image nginx chính thức):

- server*name * (nhận mọi Host), server_tokens off.
- /static/ phục vụ từ /srv/static (ETag, Last-Modified, Accept-Ranges); /contents/ là alias cùng thư mục (để test Origin Path).
  Đuôi lạ (woff2, webmanifest, m3u8...) có mime đúng.
- Còn lại proxy_pass http://api:8080; riêng /ws/ có upgrade headers.
  Header: Host $host, X-Forwarded-For $proxy_add_x_forwarded_for, X-Forwarded-Proto $scheme.
  proxy_read_timeout 90s (đủ cho /api/slow).
- client_max_body_size = ${MAX_UPLOAD_MB}m. gzip theo ${ORIGIN_GZIP} (on|off, mặc định off để thấy CDN tự nén).
  Không tin X-Forwarded-For cho quyết định nào.
- access_log dạng JSON ra stdout (remote_addr, xff, host, ua, status, request_time, request_uri).
- Go app cũng phục vụ /static khi SERVE_STATIC=true (chỉ dùng khi chạy local không có nginx).

## 8. Trang Tutorial

Mục đích: hướng dẫn thực hành WAAP và Multi-CDN của CMC trên chính web này. Nội dung tách khỏi code, nằm trong
`internal/adapter/scenario/embedded/scenarios.json`, nạp qua usecase `tutorial.ListScenarios` (port `ScenarioRepository`).
Mỗi kịch bản có: id, group (cdn|waap|shop), title, goal, portal_steps[] (đường menu trên portal CMC đúng theo docs),
requests[] (method, path, headers, body, curl, runnable bool), expect[] (kết quả kỳ vọng theo từng cấu hình), notes, docs_url, priority (P0|P1).

Trang:

- /tutorial: tổng quan, sơ đồ 4 đường đi (text/SVG tĩnh) và danh sách kịch bản theo nhóm.
  Bốn hostname: `direct` (client -> EC2), `waap` (client -> WAAP -> EC2), `cdn` (client -> CDN -> EC2),
  `full` (client -> CDN -> WAAP -> EC2; đây là suy luận, docs chưa mô tả cách nối). Hiển thị host đang truy cập.
- /tutorial/{id}: một kịch bản, SSR, đọc được khi tắt JS. JS chỉ thêm nút chạy.
- Runner (js/runner.js, same-origin, render bằng textContent):
  - Run once: status, thời gian, các header đáng chú ý (X-Cache-Status, Age, Server, Via, Content-Encoding, Cache-Control, Content-Length,
    X-Request-Id, X-Origin-Instance, X-Origin-Time) và preview body.
  - Run N lần (1–20, tuần tự) -> bảng #, status, ms, X-Cache-Status, Age, generated_at/request_id để thấy HIT/MISS.
  - Burst (tối đa 200 request, concurrency tối đa 20) -> thống kê số lượng theo status code (dùng cho Rate limit).
  - Copy as curl. Với kịch bản cần User-Agent/Host/Cookie tùy ý (Bot Management...) đánh dấu `runnable=false` và chỉ hiện curl,
    vì trình duyệt không cho đặt các header đó.
- Tài khoản demo chỉ hiện khi TUTORIAL_SHOW_DEMO_CREDS=true (chỉ customer, không bao giờ hiện admin).
  Mọi chỗ cảnh báo dùng tài khoản/dữ liệu demo.

Kịch bản P0:
CDN
C1 Onboard site: CDN > Sites > Add Site (origin domain/IP, protocol, port), CNAME về Multi CDN URL, kiểm tra Server và X-Cache-Status (HIT/MISS).
C2 Cache mặc định (site mới cache tất cả 30 phút): gọi /api/time 3 lần; thêm Cache Rule No Cache kiểu Directory `/api/`; gọi lại.
C2b Kiểm tra rò rỉ dữ liệu do cache: gọi /api/auth/me bằng 2 token khác nhau qua CDN, đối chiếu; sửa bằng No Cache cho `/api/`.
C3 Caching Level (Standard / Ignore Query String / No Query String) với /api/cache/public?v=1, ?v=2.
C4 Cache Rule theo File extension (png|jpg|css|js), Directory (`/static/img/`, có `/` đầu và cuối), Advanced (wildcard, lưu ý bảng provider).
C5 Origin Cache-Control: /api/cache/{public,no-store,etag,vary,set-cookie}, quan sát CDN có tôn trọng không.
C6 Purge: Custom Purge (tối đa 30 URL, không wildcard), Purge everything, tab History; đối chiếu bằng /api/stats (số request tới origin).
C7 Browser Cache TTL (mặc định 30 phút) và Development Mode.
C8 Page rules: Modify Request Header (xem /api/headers), Modify Response Header, Deny cho `/static/private/`, Set secure token
(thuật toán md5 + expires theo docs, có ví dụ tạo URL ký bằng Python và Go).
C9 Origin: Set host header, Origin Path (`/contents/`), Max fails và timeout (mặc định 5 lần, 30s) bằng /api/slow?ms=35000 và /api/status/502.
C10 Nén: /api/payload?kb=512 khi ORIGIN_GZIP=off, kiểm tra Content-Encoding (gzip/br). Auto Minify: so /static/js/app.js (Optimization > Content optimization).
C16 Analytics: so Origin Request và HIT ratio trong Analytics với /api/stats.
WAAP
W1 Onboard: WAAP > Application > Add (Domain, Mode Detect, Endpoint Type Domain/IP, Protocol, Port), CNAME về WAAP Domain, tab Verify.
Kiểm tra bằng /api/health và một chuỗi thử (docs dùng chuỗi XSS-like, mong đợi 403 khi ở Block).
W2 Detect vs Block: cùng một request thử, xem View Logs (IP, BLOCK URI, Match, Block ID, Description, Date).
W3 Rules theo từng vector: GET args (/api/products?q=), request body (POST /api/products/{slug}/reviews), HTTP header, cookie, URL path,
filename (upload), Content-Type. Dùng probes.json (bên dưới).
W4 Custom Rule (Message, Detect, Action LOG|BLOCK, Match): ví dụ chặn một từ khóa trong body của /api/products/{slug}/reviews.
W5 False positive + Whitelist: gửi review/ticket hợp lệ nhưng giống tấn công (mẫu bên dưới) -> bị chặn -> lấy Block ID ở Logs -> thêm Whitelist
(ruleID + Match) -> gửi lại thành công.
W6 API Protection: upload /openapi.json (3.0, <= 5MB), rule mặc định LOG, đổi sang BLOCK cho `DELETE /api/admin/products/{id}`,
thêm parameter condition.
W7 Rate limit (cần admin CMC kích hoạt): Entire Domain vs Specific URL (`/api/auth/login`), Object Source IP vs Custom Header
(Authorization hoặc X-Api-Key), ngưỡng theo giây/phút, Action BLOCK; dùng Burst.
W8 Bot Management (cần kích hoạt): Predefined (curl, python-requests, GPTBot, ClaudeBot) và Custom Bot Rules (UA/header/URL), chạy bằng curl -A.
Shop (hướng dẫn dùng nghiệp vụ, P0 ngắn): đăng ký -> duyệt sản phẩm -> giỏ -> coupon -> checkout (Idempotency-Key) -> pay -> admin ship/deliver -> review -> ticket,
mỗi bước có curl.
Kịch bản P1 (viết sau khi đọc docs, mục 12): W9 Geo Access Control, W10 IP List, W11 Send file (giới hạn body 1–500MB) với /api/upload,
W12 SSL/TLS phía origin, W13 Load Balancer cho WAAP (cần origin thứ 2), C11 Image optimization (`/thumb/200x100/...`, WebP chỉ HTTP),
C12 Range với files/10mb.bin, C13 HTTP/2 (chỉ HTTPS) và WebSocket (khác nhau theo provider), C14 CDN Firewall rule geo, C15 Multi-CDN và Routing strategy.

probes.json (chuỗi kiểm thử kinh điển; chỉ gửi tới chính hệ thống của chủ lab):
sqli_like: ' OR '1'='1' -- | 1 UNION SELECT NULL,NULL--
xss_like: <script>alert(1)</script> | "><img src=x onerror=alert(1)>
traversal_like: ../../etc/passwd | ..%2f..%2fetc%2fpasswd
cmd_like: ; cat /etc/passwd | $(id)
scanner_ua: sqlmap/1.7 | Nikto
false_positive (hợp lệ nhưng dễ bị chặn nhầm):
"Please SELECT the best option FROM the list; my code: if (a < b) { return \"<div>\"; }" | "O'Brien & Sons -- great service"

## 9. Bảo mật mặc định, header, cache policy, log

Bảo mật:

- Chỉ dùng query tham số hóa. Không eval/exec/subprocess, không dựng đường dẫn từ input. Validate bằng struct + rule trong domain.
- Dữ liệu user chỉ ra JSON qua encoder chuẩn. HTML dùng html/template (auto-escape), JS dùng textContent, không inline script/style.
- Upload: allowlist đuôi + kiểm tra size + sanitize tên, không serve lại.
- Ownership kiểm tra trong usecase (user chỉ thấy tài nguyên của mình -> 404). Admin route kiểm tra role ở usecase, sai quyền -> 403.
- Secret từ env (JWT_SECRET >= 32 ký tự bắt buộc, fail-fast). CORS cấu hình qua env, mặc định đóng.
- Container chạy non-root. Server timeouts (ReadHeader 10s, Write 120s, Idle 120s), graceful shutdown 15s, recover panic.

Header trên mọi response: X-Request-Id (luôn tự sinh, không tin giá trị từ client), X-Origin-Instance (INSTANCE_ID hoặc hostname),
X-Origin-Time (RFC3339 ms), X-Content-Type-Options: nosniff, Referrer-Policy: no-referrer.
HTML thêm CSP `default-src 'self'; img-src 'self' data:; frame-ancestors 'none'` và X-Frame-Options: DENY.
403 do app luôn là JSON kèm X-Origin-Instance (khác trang chặn của WAAP).

Cache policy của origin (CATALOG_CACHE_SECONDS, mặc định 60):

- Catalog public (/api/categories, /api/products\*, reviews GET, /api/tutorial/scenarios, sitemap.xml): `public, max-age=<giây>`, có ETag, hỗ trợ 304.
- /api/time: không có header cache. /api/cache/\*: theo mode.
- Mọi endpoint theo user/admin và mọi POST/PUT/PATCH/DELETE: `private, no-store`.
- Trang tutorial: `public, max-age=60`. Trang storefront: `no-cache`.

Log: slog JSON ra stdout gồm ts, level, request_id, method, path, query (cắt 512), status, bytes, duration_ms, remote_addr, xff, ua, host, proto.
Không log body, Authorization, Cookie.

Payload-hygiene test (bắt buộc): gửi toàn bộ probes.json vào MỌI vector nhận input của origin (query, path, JSON body, form, header, cookie, filename upload)
và khẳng định: không có 5xx, không có thông báo lỗi SQL, không phản chiếu chuỗi thô vào HTML, response luôn là JSON hoặc đã escape, bảng DB còn nguyên.

## 10. Docker, env, deploy

- Dockerfile multi-stage: stage build (golang, CGO_ENABLED=0, chạy genassets) -> target `api` (distroless/static hoặc alpine, non-root, HEALTHCHECK gọi /api/health)
  và target `nginx` (nginx chính thức + web/static + file đã sinh vào /srv/static).
- docker-compose.yml: service `nginx` (80, chừa sẵn 443) và `api` (chỉ nội bộ 8080), volume `/data` cho SQLite, restart unless-stopped, healthcheck, giới hạn bộ nhớ hợp lý cho EC2 nhỏ.
- `.env.example` (không commit `.env`): APP_ENV, HTTP_ADDR, DB_PATH, JWT_SECRET, ADMIN_TOKEN, SEED, SEED_ADMIN_EMAIL/PASSWORD,
  SEED_CUSTOMER_EMAIL/PASSWORD, TUTORIAL_SHOW_DEMO_CREDS, MAX_UPLOAD_MB, MAX_SLOW_MS, MAX_PAYLOAD_KB, CATALOG_CACHE_SECONDS,
  ENABLE_DOCS, ENABLE_WS, CORS_ALLOWED_ORIGINS, SERVE_STATIC, STATIC_DIR, STORE_UPLOADS, APP_RATE_LIMIT_ENABLED, MOCK_PAYMENT_FAIL_RATE,
  ORIGIN_GZIP, INSTANCE_ID, LOG_LEVEL.
- Seed khi DB rỗng và SEED=true: 5 category, 40 product (ảnh /static/img/products/pNN.jpg), 4 coupon, tài khoản admin + customer
  (mật khẩu từ env; thiếu thì sinh ngẫu nhiên và log đúng một lần).
- scripts/deploy.sh (EC2): `git pull --ff-only` -> `docker compose up -d --build` -> chờ healthy -> chạy `smoke`.
  scripts/dev.ps1: run, test, lint, build image, up/down.
- `cmd/smoke`: nhận BASE_URL, gọi lần lượt các endpoint chính và kiểm tra mã/header/shape.
- Security Group: mở 80/443, cổng 22 chỉ IP của chủ lab.

## 11. Phase và tiêu chí nghiệm thu

Phase 0 – Bootstrap: git init (nếu chưa có), go mod, khung thư mục, archtest, config, logger, middleware (request id, access log, recover, header chung),
/api/health, Dockerfile, compose, nginx, deploy.sh, `.gitattributes`. Nghiệm thu: `docker compose up` local và curl /api/health ra 200 qua nginx.
Phase 1 – Lab: toàn bộ mục 5 + genassets + static qua nginx + HitCounter. Test: unit (usecase/lab) + handler test.
Phase 2 – Nghiệp vụ: domain + usecase viết test trước, rồi repository sqlite, handler, seed. Nghiệm thu: luồng
đăng ký -> mua -> pay -> admin giao -> review -> ticket chạy được bằng smoke; các quy tắc ở mục 3 đều có test.
Phase 3 – UI và Tutorial: storefront tối giản, /tutorial + runner + scenarios P0, /docs.
Phase 4 – Hoàn thiện: OpenAPI đầy đủ + contract test, payload-hygiene test, `cmd/smoke`, README.md (chạy local, deploy EC2, 4 hostname),
docs/ARCHITECTURE.md (các lớp, dependency rule, cách thêm một usecase), docs/TEST_PLAN.md (tính năng trong docs CMC -> kịch bản -> endpoint -> kết quả kỳ vọng), DECISIONS.md.
Phase 5 (P2) – HTTPS ở nginx (cert mount qua volume), kịch bản P1, GitHub Actions (vet, lint, test -race, docker build).

Definition of Done:

- `go vet`, golangci-lint, `go test -race ./...` đều pass; archtest, contract test, payload-hygiene test pass.
- Clone sạch -> `cp .env.example .env` -> `docker compose up -d --build` chạy được; smoke pass với BASE_URL=http://localhost.
- /openapi.json hợp lệ OpenAPI 3.0.x.
- Không có lỗ hổng cố ý; không secret trong repo.

## 12. Đọc docs CMC trước khi viết nội dung tutorial (để tên menu và tên trường đúng)

Đã đọc kỹ (có thể dùng ngay): waap/overview, waap/how-it-works, waap/quy-trinh-trien-khai, waap/guide/{app-management, api-protection, rate-limiting,
bot-management, rules-management, whitelist-management, geo-access-control, view-logs}, multi-cdn/{overview, how-it-works, cdn-providers, quy-trinh-trien-khai},
multi-cdn/guides/{cau-hinh-cache, origin-control, page-rule, firewall-rule, opimization, analytics}, multi-cdn/usage-scenarios/{us1, us5}.
Gốc: https://docs.cmctelecom.io/vi/ccsp/ (bổ sung đường dẫn sau tiền tố đó).

CHƯA đọc kỹ, phải đọc trước khi viết kịch bản liên quan: waap/guide/{ip-list, configure-load-balance, quan-ly-ssl, dashboard, process-update-rule},
multi-cdn/guides/{cau-hinh-chien-luoc-dinh-tuyen-toi-cac-provider, cau-hinh-multi-cdn, quan-ly-va-them-site-moi, ssl-tls, dashboard},
multi-cdn/usage-scenarios/{us2, us3}, multi-cdn/faq/_, multi-cdn/troubleshooting/_.
Chỗ docs không nói rõ (cách nối WAAP với CDN, header WAAP/CDN thêm vào, vị trí IP client) thì ghi "cần quan sát thực tế" vào kịch bản
và dùng /api/headers để xác nhận, đừng bịa.
