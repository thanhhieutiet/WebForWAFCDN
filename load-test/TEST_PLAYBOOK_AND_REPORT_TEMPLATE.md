# 📊 SỔ TAY KIỂM THỬ TẢI & BẢNG MẪU BÁO CÁO KẾT QUẢ ĐO KIỂM
## (Load & Security Testing Playbook and Benchmark Report Template)

> **Dự án:** CMC Telecom LabShop E-Commerce System  
> **Kiến trúc hệ thống:** Client ➔ CMC Multi-CDN ➔ CMC Cloud WAAP ➔ AWS EC2 Origin (Nginx + Go Clean Architecture + SQLite WAL)  
> **Cụm phát tải:** Distributed Locust (Master + 4 Workers) ➔ Locust Exporter ➔ Prometheus TSDB ➔ Grafana Dashboard  
> **Mục tiêu:** Đo lường hiệu năng 4 mô hình kiến trúc, định tuyến liên mạng (Multi-ISP) và kiểm chứng năng lực phòng vệ WAF.

---

## 📑 MỤC LỤC
1. [Tổng Quan 4 Kịch Bản Kiến Trúc (Architecture Cases)](#1-tổng-quan-4-kịch-bản-kiến-trúc-architecture-cases)
2. [Cải Tiến Kỹ Thuật Đã Áp Dụng (Anti-Bottleneck & Isolation)](#2-cải-tiến-kỹ-thuật-đã-áp-dụng-anti-bottleneck--isolation)
3. [Quy Trình Khởi Động & Điều Khiển Cụm Bắn Tải](#3-quy-trình-khởi-động--điều-khiển-cụm-bắn-tải)
4. [BẢNG MẪU 1: So Sánh Hiệu Năng 4 Kịch Bản Kiến Trúc](#4-bảng-mẫu-1-so-sánh-hiệu-năng-4-kịch-bản-kiến-trúc)
5. [BẢNG MẪU 2: So Sánh Định Tuyến Đa Mạng (Multi-ISP: Công Ty vs Nhà vs 4G)](#5-bảng-mẫu-2-so-sánh-định-tuyến-đa-mạng-multi-isp-công-ty-vs-nhà-vs-4g)
6. [BẢNG MẪU 3: Kiểm Chứng Năng Lực Phòng Vệ WAF & Rate Limiting](#6-bảng-mẫu-3-kiểm-chứng-năng-lực-phòng-vệ-waf--rate-limiting)
7. [Danh Mục Ảnh Chụp Minh Chứng (POC Checklist) Cho Báo Cáo Mentor](#7-danh-mục-ảnh-chụp-minh-chứng-poc-checklist-cho-báo-cáo-mentor)
8. [Mẫu Nhận Xét & Kết Luận Kỹ Thuật Gửi Mentor](#8-mẫu-nhận-xét--kết-luận-kỹ-thuật-gửi-mentor)

---

## 1. Tổng Quan 4 Kịch Bản Kiến Trúc (Architecture Cases)

Để có số liệu đối chiếu chuẩn xác, chúng ta thực hiện đo kiểm trên 4 cấu hình topology:

```
[Case 1: Direct]   Client ───────────────────────────────────────────────► EC2 Origin (Port 80/443)
[Case 2: WAF Only] Client ──────────────────► CMC WAAP ─────────────────► EC2 Origin
[Case 3: CDN Only] Client ────► CMC CDN ─────────────────────────────────► EC2 Origin
[Case 4: Full]     Client ────► CMC CDN ────► CMC WAAP ──────────────────► EC2 Origin
```

| Kịch bản | Host URL đo kiểm | Mục tiêu phân tích |
| :--- | :--- | :--- |
| **Case 1 (Direct)** | `http://52.201.167.9` *(hoặc direct domain)* | Baseline gốc của EC2 Origin (Nginx + Go + SQLite). Đo ngưỡng chịu tải trần của server. |
| **Case 2 (WAF Only)** | `https://waap.thanhhieutiet.shop` *(nếu có CNAME)* | Đo độ trễ cộng thêm của tầng WAF (WAF Inspection Overhead) khi phải phân tích gói tin. |
| **Case 3 (CDN Only)** | `https://uer5m09gjo7c.cmccdn.net` *(Edge CNAME)* | Đo năng lực hấp thụ tải tĩnh và tỷ lệ Cache Offload Ratio của CDN. |
| **Case 4 (Full Stack)** | `https://app.thanhhieutiet.shop` | Đánh giá kiến trúc sản xuất thực tế: Tận dụng cả sức mạnh tăng tốc của CDN và lớp chắn an ninh WAF. |

---

## 2. Cải Tiến Kỹ Thuật Đã Áp Dụng (Anti-Bottleneck & Isolation)

Trước khi chạy, hãy đảm bảo hệ thống đã kích hoạt 3 cải tiến quan trọng sau:

1. **Khử Nghẽn Băng Thông Tải Xuống (Range Header Protection):**
   * Trong `StaticAssetUser`, cờ `USE_RANGE_HEADER=true` sẽ tự động gửi kèm header `Range: bytes=0-1024` khi tải các file nặng (`banner-large.jpg`, `sample.pdf`, ảnh sản phẩm).
   * **Hiệu quả:** CDN và Origin vẫn phải thực hiện toàn bộ quy trình kiểm tra Cache HIT/MISS, đọc đĩa I/O và trả về mã `206 Partial Content` (hoặc `200`), nhưng đường truyền máy local không bị quá tải (tránh tình trạng nghẽn cổ chai mạng làm vọt latency lên 80s).
2. **Cô Lập Kịch Bản (Blast Radius Elimination):**
   * Khi `TEST_MODE=performance`: Kịch bản `WafRateLimitProbeUser` bị tắt hoàn toàn (`weight=0`). Nhờ đó, IP của máy test không bị WAF kích hoạt Rate Limit (chặn 429/403 hàng loạt), giữ cho biểu đồ RPS và Latency luôn mượt mà.
   * Khi `TEST_MODE=security`: Chỉ chạy riêng `WafRateLimitProbeUser` để đo đạc tỷ lệ chặn của WAF.
3. **Phân Tán Luồng Tạo Tải (Locust Master-Worker):**
   * Cụm được triển khai với 1 Master và **4 Workers** độc lập, vượt qua giới hạn GIL (Global Interpreter Lock) của Python đơn nhân, dễ dàng tạo tải từ 500 đến 5.000 CCU.

---

## 3. Quy Trình Khởi Động & Điều Khiển Cụm Bắn Tải

### Cách 1: Sử dụng Script điều khiển nhanh (Khuyên dùng)

* **Trên Linux / macOS / Git Bash:**
  ```bash
  cd load-test
  chmod +x run_test_suite.sh
  ./run_test_suite.sh          # Menu tương tác
  # Hoặc chạy lệnh trực tiếp:
  ./run_test_suite.sh perf https://app.thanhhieutiet.shop 4
  ```

* **Trên Windows PowerShell:**
  ```powershell
  cd c:\CMC\WAF\WebForWAF&CDN\load-test
  .\run_test_suite.ps1         # Menu tương tác
  # Hoặc chạy lệnh trực tiếp:
  .\run_test_suite.ps1 perf https://app.thanhhieutiet.shop 4
  ```

* **Sử dụng Makefile:**
  ```bash
  make perf                    # Khởi động chế độ Performance (4 Workers)
  make security                # Khởi động chế độ WAF Security
  make network                 # Khởi động chế độ Multi-ISP
  make scale WORKERS=8         # Mở rộng lên 8 Workers nếu cần tải cực lớn
  make down                    # Tắt toàn bộ dịch vụ
  ```

### Các Cổng Truy Cập Dashboard:
* **Locust Web UI:** [http://localhost:8089](http://localhost:8089)
* **Grafana Dashboard:** [http://localhost:3000](http://localhost:3000) (`admin` / `admin`)
* **Prometheus Targets:** [http://localhost:9090/targets](http://localhost:9090/targets)
* **Locust Exporter Metrics:** [http://localhost:9646/metrics](http://localhost:9646/metrics)

---

## 4. BẢNG MẪU 1: So Sánh Hiệu Năng 4 Kịch Bản Kiến Trúc

> **Thông số test chuẩn:**  
> * **CCU:** `300` người dùng đồng thời  
> * **Ramp-up:** `20` users/giây  
> * **Thời gian chạy:** `5 phút` mỗi kịch bản  
> * **Chế độ:** `TEST_MODE=performance` (Range Header = true)

### 📋 Bảng Ghi Nhận Số Liệu:

| Chỉ số đo lường | Case 1: Direct EC2 | Case 2: WAF Only | Case 3: CDN Only | Case 4: Full Stack (CDN+WAF) |
| :--- | :--- | :--- | :--- | :--- |
| **Host kiểm thử** | `http://<EC2_IP>` | `https://waap...` | `https://cdn...` | `https://app.thanhhieutiet.shop` |
| **Total RPS (Trung bình)** | *(ví dụ: 85 req/s)* | *(ví dụ: 78 req/s)* | *(ví dụ: 450 req/s)* | *(ví dụ: 420 req/s)* |
| **Total RPS (Đỉnh - Peak)** | | | | |
| **Static Avg Latency** | *(ví dụ: 120 ms)* | *(ví dụ: 145 ms)* | *(ví dụ: 18 ms)* | *(ví dụ: 22 ms)* |
| **Static P95 Latency** | *(ví dụ: 250 ms)* | *(ví dụ: 290 ms)* | *(ví dụ: 35 ms)* | *(ví dụ: 40 ms)* |
| **Dynamic API Avg Latency**| *(ví dụ: 65 ms)* | *(ví dụ: 85 ms)* | *(ví dụ: 70 ms)* | *(ví dụ: 88 ms)* |
| **Dynamic API P95 Latency**| *(ví dụ: 140 ms)* | *(ví dụ: 180 ms)* | *(ví dụ: 145 ms)* | *(ví dụ: 190 ms)* |
| **Failure Ratio (%)** | *(mục tiêu: 0.0%)* | *(mục tiêu: 0.0%)* | *(mục tiêu: 0.0%)* | *(mục tiêu: 0.0%)* |
| **EC2 CPU Peak (%)** | *(ví dụ: 75%)* | *(ví dụ: 70%)* | *(ví dụ: 12%)* | *(ví dụ: 15%)* |
| **EC2 Network Out (MB)** | *(ví dụ: 850 MB)* | *(ví dụ: 820 MB)* | *(ví dụ: 45 MB)* | *(ví dụ: 50 MB)* |
| **Tỷ lệ Cache Offload (%)**| `0.0%` (Không có CDN)| `0.0%` | *(ví dụ: 96.5%)* | *(ví dụ: 96.2%)* |

### 🔍 Hướng Dẫn Phân Tích & Đối Soát:
1. **WAF Latency Overhead:** So sánh **Case 2 vs Case 1** và **Case 4 vs Case 3**. Phần chênh lệch (thường khoảng 10–25ms) chính là thời gian bộ máy WAF kiểm tra chữ ký gói tin (Deep Packet Inspection).
2. **CDN Cache Offload Effect:** So sánh **Case 4 vs Case 1**. Nhận thấy CPU EC2 giảm mạnh từ ~70% xuống < 15% và Network Out giảm tới hơn 90%, chứng minh toàn bộ tải tĩnh đã được chặn đứng và phục vụ ngay tại Edge.
3. **Định luật Little:** Kiểm chứng mối tương quan giữa CCU, RPS và Latency: $N = X \times R$.

---

## 5. BẢNG MẪU 2: So Sánh Định Tuyến Đa Mạng (Multi-ISP: Công Ty vs Nhà vs 4G)

> **Thông số test chuẩn:**  
> * **CCU:** `30` người dùng đồng thời (tải nhẹ để tránh bão hòa đường truyền uplink cá nhân)  
> * **Ramp-up:** `10` users/giây  
> * **Thời gian chạy:** `2 phút`  
> * **Mục đích:** Đánh giá chất lượng hạ tầng mạng và năng lực Anycast/PoP Edge của CMC CDN khi người dùng truy cập từ nhiều nhà mạng khác nhau.

### 📋 Bảng Ghi Nhận Số Liệu:

| Môi trường mạng & Nhà mạng (ISP) | Kịch bản so sánh | Min Latency (ms) | Avg Latency (ms) | P95 Latency (ms) | Max Latency (ms) | Mức cải thiện độ trễ (%) |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Mạng Công ty / CMC Telecom**<br>*(Đường truyền Leased Line / Fiber Doanh nghiệp)* | **Case 1 (Direct EC2)** | | | | | *Baseline* |
| | **Case 4 (Full Stack CDN)** | | | | | *(ví dụ: Giảm 65%)* |
| **Cáp quang gia đình (Home FTTH)**<br>*(Viettel / VNPT / FPT)* | **Case 1 (Direct EC2)** | | | | | *Baseline* |
| | **Case 4 (Full Stack CDN)** | | | | | *(ví dụ: Giảm 72%)* |
| **Mạng Di Động (4G / 5G Hotspot)**<br>*(Viettel / Vinaphone / Mobifone)* | **Case 1 (Direct EC2)** | | | | | *Baseline* |
| | **Case 4 (Full Stack CDN)** | | | | | *(ví dụ: Giảm 55%)* |

### 🔍 Hướng Dẫn Phân Tích:
* **Edge Proximity:** Khi truy cập từ mạng 4G hoặc cáp quang gia đình, đường truyền tới máy chủ AWS EC2 (Singapore / US) thường có độ trễ vật lý cao (60–120ms).
* Khi đi qua CMC Multi-CDN, request chỉ cần đến Node PoP gần nhất trong nước (Hà Nội / TP.HCM), độ trễ giảm xuống chỉ còn **15–30ms**.

---

## 6. BẢNG MẪU 3: Kiểm Chứng Năng Lực Phòng Vệ WAF & Rate Limiting

> **Thông số test chuẩn:**  
> * **Kịch bản:** Chạy riêng `TEST_MODE=security` (hoặc cờ `--tags waf`)  
> * **CCU:** `100` người dùng đồng thời  
> * **Ramp-up:** `20` users/giây  
> * **Thời gian chạy:** `2 phút`  
> * **Host kiểm thử:** `https://app.thanhhieutiet.shop`

### 📋 Bảng Ghi Nhận Số Liệu:

| Vector tấn công kiểm thử | Tổng request đã phát | Mã HTTP mong đợi | Mã HTTP thực tế | Tỷ lệ chặn thành công (%) | Mã Event ID ghi nhận trên CMC Portal | Tác động CPU Origin EC2 |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **SQL Injection (SQLi Probe)**<br>`GET /api/products?q=' OR '1'='1' --` | *(ví dụ: 1,200)* | `403 Forbidden` | `403` *(hoặc 400)* | **100.0%** | *(Điền Block ID từ Portal CMC WAF)* | **0% ảnh hưởng** (Bị chặn ngay tại biên) |
| **Brute-Force Login Burst**<br>`POST /api/auth/login [Spam Sai Mật Khẩu]` | *(ví dụ: 3,500)* | `429 Too Many Requests` | `429` *(sau ngưỡng rate limit)* | **~85% - 95%** *(các request đầu trả 401, sau đó 429)* | *(Điền Rate-Limit Rule Name)* | **Bảo vệ an toàn** (Hàm Argon2id không bị cạn kiệt CPU) |

---

## 7. Danh Mục Ảnh Chụp Minh Chứng (POC Checklist) Cho Báo Cáo Mentor

Để báo cáo có sức thuyết phục tuyệt đối, bạn hãy chụp lại 6 bức ảnh minh chứng thực tế:

- [ ] **POC 1 (Grafana Overview):** Chụp Dashboard Grafana ở Case 4 (RPS đỉnh đạt > 400, Avg Latency thấp, Failure = 0%).
- [ ] **POC 2 (Locust Statistics Table):** Chụp bảng chi tiết của Locust cho thấy các endpoint `/static/*` trả về mã `200/206` với độ trễ < 25ms.
- [ ] **POC 3 (Origin EC2 CPU Monitor):** Chụp màn hình terminal EC2 chạy lệnh `htop` trong lúc bắn tải Case 4, minh chứng CPU server vẫn duy trì ở mức thấp (< 20%).
- [ ] **POC 4 (So Sánh Latency Biểu Đồ):** Chụp đồ thị Latency giữa Case 1 (đường cao) và Case 4 (đường nằm sát đáy).
- [ ] **POC 5 (CMC WAF Portal Blocks):** Chụp trang quản trị CMC Cloud WAAP phần **Security Logs / Blocked Events** hiển thị các request bị chặn bởi luật SQLi (mã 403).
- [ ] **POC 6 (CMC WAF Rate Limiting):** Chụp biểu đồ Rate Limiting trên portal CMC WAF ghi nhận việc bóp nghẽn các IP thực hiện Brute-force Login.

---

## 8. Mẫu Nhận Xét & Kết Luận Kỹ Thuật Gửi Mentor

*(Bạn có thể sao chép đoạn mẫu này và điền số liệu thực tế để đưa vào email hoặc slide báo cáo)*

```markdown
Kính gửi Mentor và Bộ phận Đánh giá Kỹ thuật,

Sau quá trình triển khai đo kiểm tải phân tán (Distributed Load Testing) bằng bộ công cụ Locust Master-Worker (4 Workers) kết hợp Prometheus và Grafana, em xin đúc kết các kết quả kỹ thuật thu được trên hệ thống CMC Telecom LabShop như sau:

1. Năng Lực Giảm Tải Của CMC Multi-CDN (Offload Ratio > 96%):
   - Khi chạy ở mức 300 CCU, toàn bộ lưu lượng tải tệp tĩnh và hình ảnh được phục vụ trực tiếp từ Edge Cache với độ trễ trung bình đạt [Điền số: ví dụ 22ms], nhanh gấp [Điền số: ví dụ 5.5 lần] so với truy cập trực tiếp Origin EC2.
   - Nhờ cơ chế Cache Edge, mức tiêu thụ CPU của máy chủ EC2 chỉ dao động quanh [Điền số: ví dụ 15%], và băng thông mạng ra (Network Out) giảm hơn 90%, chứng minh hiệu quả giảm tải vượt trội.

2. Đánh Giá Độ Trễ Tầng WAF (WAF Inspection Overhead):
   - Lớp phòng vệ CMC WAAP bổ sung độ trễ trung bình khoảng [Điền số: ví dụ 15 - 20ms] cho quá trình phân tích gói tin sâu (DPI). Đây là mức đánh đổi hoàn toàn tối ưu và chấp nhận được đối với một hệ thống E-Commerce tiêu chuẩn.

3. Năng Lực Phòng Thủ An Ninh WAF & Rate Limiting:
   - Cơ chế WAF hoạt động chính xác 100% đối với các mẫu thử SQL Injection (trả về HTTP 403 kèm Block ID trên Portal).
   - Cơ chế Rate Limiting kích hoạt kịp thời khi xuất hiện đột biến tấn công vét cạn đăng nhập (Brute-force), phản hồi mã HTTP 429 và bảo vệ an toàn cho hàm băm mật khẩu Argon2id trên máy chủ gốc.

Em xin đính kèm bảng số liệu chi tiết và các ảnh chụp minh chứng (POC) trong tài liệu này.
```
