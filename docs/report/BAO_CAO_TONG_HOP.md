# BÁO CÁO THỰC THI NHIỆM VỤ ĐO KIỂM HIỆU NĂNG & AN NINH HỆ THỐNG
## (ĐÁNH GIÁ ĐỊNH LƯỢNG CMC WAAP & MULTI-CDN VỚI LOCUST VÀ GRAFANA)

> **Người thực hiện:** @hieuttm  
> **Nhiệm vụ:** Dựng cụm local server test tải, đo kiểm host từ các môi trường mạng đa dạng, đánh giá hiệu năng CMC Multi-CDN & CMC WAAP  
> **Hệ thống mục tiêu:** Ứng dụng LabShop trên AWS EC2 (`t3.micro`, us-east-1 North Virginia, Mỹ)  
> **Thời gian thực hiện:** Tháng 10/2026  

---

## MỤC LỤC BÁO CÁO
1. [PHẦN 1: TỔNG QUAN BÀI TEST & MỤC TIÊU ĐO KIỂM](#phần-1-tổng-quan-bài-test--mục-tiêu-đo-kiểm)
2. [PHẦN 2: MÔ HÌNH TRIỂN KHAI & KIẾN TRÚC KỸ THUẬT](#phần-2-mô-hình-triển-khai--kiến-trúc-kỹ-thuật)
3. [PHẦN 3: KẾT QUẢ ĐO KIỂM & PHÂN TÍCH CHUYÊN SÂU](#phần-3-kết-quả-đo-kiểm--phân-tích-chuyên-sâu)
   * [3.1. Phân Hệ 1 (BM1): So sánh 4 Kịch bản Kiến trúc WAF & Multi-CDN](#31-phân-hệ-1-bm1--so-sánh-4-kịch-bản-kiến-trúc-waf--multi-cdn)
   * [3.2. Phân Hệ 2 (BM2): Đo kiểm Đa Môi trường Mạng (CMC FTTH vs 4G LTE)](#32-phân-hệ-2-bm2--đo-kiểm-đa-môi-trường-mạng-cmc-ftth-vs-4g-lte)
4. [PHẦN 4: KẾT LUẬN & ĐÁNH GIÁ TỔNG THỂ](#phần-4-kết-luận--đánh-giá-tổng-thể)

---

# PHẦN 1: TỔNG QUAN BÀI TEST & MỤC TIÊU ĐO KIỂM

## 1.1. Bối Cảnh & Yêu Cầu Từ Mentor
Trong quá trình triển khai và đánh giá bộ giải pháp an ninh mạng và tăng tốc web của **CMC Telecom** (gồm **CMC WAAP** và **CMC Multi-CDN**) bảo vệ ứng dụng thương mại điện tử thử nghiệm (**LabShop**), mentor đã giao bổ sung nhiệm vụ kỹ thuật:

> *"Thêm 1 task bổ sung học cách dựng local server test tải, test host từ các mạng nhé @hieuttm. Gợi ý có thể tham khảo từ Grafana, stress test có thể dùng Locust."*

Báo cáo này tổng hợp toàn bộ kết quả triển khai hạ tầng đo kiểm, số liệu thực nghiệm đo được trên 4 mô hình kiến trúc và 2 môi trường mạng viễn thông thực tế (Mạng cáp quang văn phòng vs Mạng di động 4G).

## 1.2. Mục Tiêu Trọng Tâm
1. **Làm chủ bộ công cụ đo kiểm tải và giám sát tập trung (Observability Stack):** Tự động hóa việc triển khai cụm công cụ kiểm thử tải phân tán (Locust) kết hợp với hệ thống thu thập số liệu (Prometheus), trực quan hóa hiệu năng thời gian thực (Grafana Dashboard) và gom log truy cập (Loki & Promtail).
2. **Đo lường định lượng giá trị của CMC Multi-CDN & WAAP:**
   * **Tăng tốc & Giảm độ trễ:** Đánh giá hiệu quả của bộ nhớ đệm tại biên (Edge Caching) khi máy chủ gốc (Origin) đặt ở khoảng cách xa (AWS EC2 khu vực us-east-1 North Virginia, Mỹ).
   * **Bảo vệ & Giảm tải máy chủ gốc (Origin Offload):** Đo lường mức độ giảm tải CPU và lưu lượng card mạng trên EC2 khi có CDN gánh tải.
   * **Chi phí độ trễ an ninh (Security Latency Overhead):** Xác định thời gian trễ tăng thêm khi gói tin phải đi qua bộ lọc kiểm tra sâu chữ ký tấn công của CMC WAAP.
3. **Đánh giá trải nghiệm thực tế từ các môi trường mạng (Multi-Network Benchmark):** So sánh chất lượng dịch vụ (QoE) của người dùng cuối khi truy cập qua đường truyền Internet cáp quang doanh nghiệp (CMC FTTH) và đường truyền di động không dây (4G LTE).

## 1.3. Phạm Vi & Các Gói Thử Nghiệm Thực Hiện
Bài kiểm thử được chia thành 2 phân hệ thử nghiệm độc lập:
* **Phân hệ 1 (BM1 - Kiến trúc hạ tầng):** Đo kiểm 4 kịch bản định tuyến mạng cốt lõi với tải định danh cao (**300 CCU, ramp-up 20 users/s, thời gian 5 phút**):
  1. *Case 1 (Direct EC2):* Client $\rightarrow$ EC2 Origin (Đo baseline hiệu năng gốc).
  2. *Case 2 (WAF Only):* Client $\rightarrow$ CMC WAAP $\rightarrow$ EC2 Origin (Đo tải qua bộ lọc bảo mật).
  3. *Case 3 (CDN Only):* Client $\rightarrow$ CMC Multi-CDN $\rightarrow$ EC2 Origin (Đo năng lực Edge Caching).
  4. *Case 4 (Full Stack):* Client $\rightarrow$ CMC Multi-CDN $\rightarrow$ CMC WAAP $\rightarrow$ EC2 Origin (Mô hình Production chuẩn).
* **Phân hệ 2 (BM2 - Đa môi trường mạng):** Đo kiểm so sánh giữa 2 hạ tầng mạng viễn thông với cấu hình chuẩn (**30 CCU, ramp-up 10 users/s, thời gian 120s**):
  1. Mạng Cáp quang Doanh nghiệp (CMC FTTH) $\rightarrow$ CDN vs EC2.
  2. Mạng Di động Cá nhân (4G LTE Hotspot) $\rightarrow$ CDN vs EC2.

---

# PHẦN 2: MÔ HÌNH TRIỂN KHAI & KIẾN TRÚC KỸ THUẬT

## 2.1. Kiến Trúc Hệ Thống Mục Tiêu (System Under Test - SUT)
Hệ thống mục tiêu được thiết kế theo mô hình phòng thủ theo chiều sâu (Defense-in-Depth) với 3 tầng xử lý:

```
[ Người dùng / Client ]
           │
           ▼
┌────────────────────────────────────────────────────────┐
│ TẦNG 1: CMC MULTI-CDN (Anycast Edge PoPs tại Việt Nam) │
│ - Caching RAM/SSD tại biên cho tĩnh & API              │
│ - Tối ưu nén, SSL Termination                          │
└──────────────────────────┬─────────────────────────────┘
                           │ (Cache Miss / Pass-through)
                           ▼
┌────────────────────────────────────────────────────────┐
│ TẦNG 2: CMC CLOUD WAAP (Tường Lửa Web & API)           │
│ - Kiểm tra sâu gói tin (DPI), OWASP Top 10             │
│ - Rate Limiting, Geo-blocking, Bot Protection          │
└──────────────────────────┬─────────────────────────────┘
                           │ (Clean Traffic)
                           ▼
┌────────────────────────────────────────────────────────┐
│ TẦNG 3: MÁY CHỦ GỐC ORIGIN (AWS EC2 us-east-1, Mỹ)     │
│ - Nginx Reverse Proxy (Gzip, Rate limit, JSON log)     │
│ - Go API Core Clean Architecture + SQLite (WAL mode)   │
│ - Static Asset Storage (/srv/static: Ảnh, Banner, PDF) │
└────────────────────────────────────────────────────────┘
```

### Thông số kỹ thuật chi tiết:
1. **Máy chủ gốc (Origin Server):**
   * Hạ tầng: AWS EC2 Instance `t3.micro` (1 vCPU, 1 GB RAM, Ubuntu 22.04 LTS), vị trí địa lý tại **Bắc Virginia (us-east-1), Mỹ**.
   * Web Server: Nginx Reverse Proxy, cấu hình nén `gzip on`, định dạng log chuẩn `json_combined`, khôi phục Client IP thật qua `real_ip_header X-Forwarded-For`.
   * Backend Application: Go RESTful API (Chi Router, kết nối SQLite qua Driver CGO với chế độ WAL - Write-Ahead Logging).
2. **Cấu hình CMC Multi-CDN & WAAP:**
   * Cấu hình Cache Rule chuẩn hóa trên Portal CMC CDN:
     * File tĩnh: Đuôi mở rộng `js|css|jpg|png` $\rightarrow$ Cache **1 Day** trên Edge.
     * API động: Thư mục `/api/` $\rightarrow$ Cache **1 Hour** trên Edge.
     * Mặc định: Thư mục `/` $\rightarrow$ Cache **30 Day**.
   * Domain trỏ CNAME: `app.thanhhieutiet.shop` $\rightarrow$ Endpoint CMC CDN (`uer5m09gjo7c.cmccdn.net`).
   * Origin của CDN trỏ về CMC WAAP (`w1.cmctelecom.io:443`).
   * Origin của WAAP trỏ về IP công khai của EC2 (`52.201.167.9:80/443`).

## 2.2. Kiến Trúc Cụm Đo Kiểm & Giám Sát Cục Bộ (Local Observability Stack)
Cụm công cụ kiểm thử tải được container hóa hoàn toàn bằng Docker Compose, chạy trên máy trạm độc lập để đảm bảo tải phát ra đi qua mạng WAN Internet thực tế:

```
┌──────────────────────────────────────────────────────────────────┐
│ PHÂN HỆ TẠO TẢI & QUAN SÁT (Local Docker Stack)                  │
│                                                                  │
│  ┌─────────────────┐       ┌─────────────────┐                   │
│  │  Locust Engine  │ ────► │ Locust Exporter │                   │
│  │  (Python Users) │       │   (Go Service)  │                   │
│  └────────┬────────┘       └────────┬────────┘                   │
│           │                         │ Pull Metrics               │
│           │ Bắn HTTPS               ▼                            │
│           │ qua WAN        ┌─────────────────┐                   │
│           ▼                │ Prometheus TSDB │                   │
│     [ Internet ]           └────────┬────────┘                   │
│                                     │ PromQL                     │
│                                     ▼                            │
│                            ┌─────────────────┐                   │
│                            │ Grafana Dashbrd │ (Trực quan hóa)   │
│                            └─────────────────┘                   │
│                                                                  │
│  ┌─────────────────┐       ┌─────────────────┐                   │
│  │ Promtail Agent  │ ────► │    Loki TSDB    │ (Log truy cập L7) │
│  └─────────────────┘       └─────────────────┘                   │
└──────────────────────────────────────────────────────────────────┘
```

1. **Locust Load Generator:**
   * Sử dụng cơ chế Event-Driven Coroutine (Gevent) siêu nhẹ, cho phép 1 laptop duy trì hàng trăm kết nối đồng thời với độ chính xác cao.
   * Kịch bản test giả lập hành vi người dùng thực tế: duyệt danh mục (`/api/categories`), xem chi tiết sản phẩm (`/api/products`), kiểm tra thời gian thực (`/api/time`), tải tài liệu hướng dẫn (`/static/docs/sample.pdf`) và kéo ảnh dung lượng lớn (`/static/img/banner-large.jpg`).
2. **Prometheus & Locust Exporter:** Thu thập các chỉ số RPS tức thời, độ trễ phân vị (P50, P95), tỷ lệ lỗi theo chu kỳ lấy mẫu 2 giây/lần.
3. **Grafana Dashboard:** Cung cấp giao diện quan sát thời gian thực với các panel:
   * Tổng quan: Tốc độ bắn tải (RPS), Độ trễ trung bình, Số CCU hoạt động, Tỷ lệ lỗi (Failure %).
   * Đồ thị chuỗi thời gian (Time Series): Biểu đồ RPS vs Failures, Độ trễ phản hồi (Average vs Min Latency), Chi tiết theo từng Endpoint.
4. **Loki & Promtail (SIEM Web Log):** Thu thập toàn bộ Nginx access log dưới dạng JSON cấu trúc, cho phép lọc log theo IP nguồn, mã trạng thái HTTP, URL path và trạng thái bộ đệm CDN.

---

> 📷 **[VỊ TRÍ CHÈN ẢNH 1]:** *Ảnh chụp kiến trúc triển khai hoặc ảnh chụp các container Docker đang chạy (`docker compose ps`)*  
> *(Gợi ý caption: Hình 2.1 - Cụm dịch vụ Docker đo kiểm và giám sát tải hoạt động trên máy cục bộ)*

---

# PHẦN 3: KẾT QUẢ ĐO KIỂM & PHÂN TÍCH CHUYÊN SÂU

## 3.1. Phân Hệ 1 (BM1) — So Sánh 4 Kịch Bản Kiến Trúc WAF & Multi-CDN
* **Cấu hình tải:** 300 Virtual Users (CCU) đồng thời | Tốc độ tăng tải: 20 users/giây | Thời gian chạy: 5 phút (300 giây).
* **Nguồn dữ liệu:** Trích xuất từ file log CSV chuẩn của Locust và hệ thống giám sát AWS CloudWatch / CMC Portal.

### A. Bảng Số Liệu Tổng Hợp 4 Kịch Bản Kiến Trúc

| STT | Phân nhóm chỉ số | Chỉ số đo lường | Đơn vị | Case 1: Direct EC2 | Case 2: WAF Only | Case 3: CDN Only | Case 4: Full Stack (CDN+WAF) | Đánh giá kỹ thuật |
| :---: | :--- | :--- | :---: | :---: | :---: | :---: | :---: | :--- |
| **0** | **Cấu hình** | **Host kiểm thử** | URL/IP | `http://<EC2_IP>` | `https://waap...` | `https://cdn...` | `https://app.thanhhieutiet.shop` | Endpoint đích thực tế |
| **1** | **Thông lượng** | **Total RPS (Trung bình)** | req/s | **316.9** | **351.0** | **626.3** | **515.3** | Tốc độ xử lý ổn định của toàn web |
| **2** | **Thông lượng** | **Total RPS (Đỉnh - Peak)**| req/s | **420** | **392** | **672** | **673** | Ngưỡng chịu tải tối đa ghi nhận |
| **3** | **Độ trễ Tĩnh** | **Static Avg Latency** | ms | **399.3** | **363.4** | **47.9** | **121.1** | Tải tài nguyên ảnh, banner, PDF |
| **4** | **Độ trễ Tĩnh** | **Static P95 Latency** | ms | **352.5** | **500.0** | **117.5** | **84.8** | 95% request tĩnh hoàn tất dưới mức này |
| **5** | **Độ trễ Động** | **Dynamic API Avg Latency**| ms | **630.3** | **379.3** | **51.7** | **195.6** | Gọi API backend / cơ sở dữ liệu |
| **6** | **Độ trễ Động** | **Dynamic API P95 Latency**| ms | **487.3** | **583.0** | **155.0** | **161.2** | Độ trễ đuôi (tail latency) của API |
| **7** | **Độ ổn định** | **Tỷ lệ lỗi (Failure Ratio)**| % | **0.55%** | **0.00%** | **0.00%** | **0.18%** | Tỷ lệ HTTP 5xx hoặc rớt kết nối |
| **8** | **Tải Origin** | **EC2 CPU Peak (%)** | % | **7.44%** | **11.52%** | **0.56%** | **0.86%** | Mức chiếm dụng CPU trên máy chủ gốc |
| **9** | **Lưu lượng** | **Total Data to Client** | MB | **96.4 MB** | **63.4 MB** | **191.5 MB** | **150.5 MB** | Dung lượng nội dung HTTP client nhận |
| **10**| **Bộ đệm Edge**| **Tỷ lệ Cache Offload** | % | **0.0%** | **0.0%** | **99.9%** | **99.8%** | Tỷ lệ request do CDN Edge phục vụ |

---

### B. Vị Trí Chèn Ảnh Minh Chứng Cho BM1

> 📷 **[VỊ TRÍ CHÈN ẢNH 2.1]:** *Ảnh chụp Dashboard Grafana — Case 1: Direct EC2*  
> *(Gợi ý caption: Hình 3.1 - Hiệu năng Case 1 Direct EC2: Tốc độ 317 RPS, CPU EC2 chạm 7.44%, xuất hiện 0.55% lỗi)*

> 📷 **[VỊ TRÍ CHÈN ẢNH 2.2]:** *Ảnh chụp Dashboard Grafana — Case 2: WAF Only*  
> *(Gợi ý caption: Hình 3.2 - Hiệu năng Case 2 qua CMC WAAP: Tốc độ 351 RPS, 0% lỗi, độ trễ ổn định)*

> 📷 **[VỊ TRÍ CHÈN ẢNH 2.3]:** *Ảnh chụp Dashboard Grafana — Case 3: CDN Only (Bản mới)*  
> *(Gợi ý caption: Hình 3.3 - Hiệu năng Case 3 qua CMC Multi-CDN: Tốc độ tăng vọt 626 RPS, độ trễ chạm sàn 48ms)*

> 📷 **[VỊ TRÍ CHÈN ẢNH 2.4]:** *Ảnh chụp Dashboard Grafana — Case 4: Full Stack (CDN + WAF)*  
> *(Gợi ý caption: Hình 3.4 - Hiệu năng Case 4 Production: Đạt 515 RPS, cân bằng tối ưu giữa an ninh WAF và tốc độ CDN)*

> 📷 **[VỊ TRÍ CHÈN ẢNH 2.5]:** *Ảnh chụp AWS CloudWatch Console (CPU & Network Out)*  
> *(Gợi ý caption: Hình 3.5 - Giám sát tài nguyên máy chủ gốc AWS EC2: CPU giảm ngoạn mục từ 7.44% về 0.56% khi có CDN)*

> 📷 **[VỊ TRÍ CHÈN ẢNH 2.6]:** *Ảnh chụp Donut Chart Hit Ratio trên Portal CMC Multi-CDN*  
> *(Gợi ý caption: Hình 3.6 - Tỷ lệ Cache Hit Ratio đạt 99.97% và 100% mã 2xx trên cổng quản trị CMC Telecom)*

---

### C. Phân Tích Kỹ Thuật Chuyên Sâu Cho Phân Hệ BM1
1. **Sức mạnh giải phóng tải của CMC Multi-CDN (Case 3 vs Case 1):**
   * Khi kích hoạt Edge Caching, thông lượng xử lý tăng gần gấp đôi (từ **316.9 req/s lên 626.3 req/s**).
   * Độ trễ tải tài nguyên tĩnh và API giảm từ ~400ms – 630ms xuống mức siêu tốc **48 ms** (nhanh gấp 8 – 13 lần).
   * **Bảo vệ tuyệt đối máy chủ gốc:** Mức tiêu thụ CPU trên EC2 giảm từ **7.44% xuống chỉ còn 0.56%** (giảm hơn 13 lần). Hơn **99.9%** request được giải quyết ngay trên bộ nhớ RAM/SSD của các máy chủ biên CDN tại Việt Nam, máy chủ gốc tại Mỹ gần như ở trạng thái nghỉ ngơi hoàn toàn.
2. **Chi phí độ trễ an ninh của WAF (Case 3 vs Case 4):**
   * Case 3 (chỉ có CDN) đạt độ trễ **48 ms**, trong khi Case 4 (Full Stack) đạt độ trễ **121 ms (Static)** và **196 ms (Dynamic)**.
   * Khoảng chênh lệch ~70ms – 140ms này phản ánh đúng thực tế kỹ thuật: Khi lưu lượng đi qua **CMC WAAP**, hệ thống phải thực hiện phân tích sâu gói tin (DPI), đối soát danh sách chữ ký tấn công (OWASP Top 10, SQLi, XSS) và kiểm tra hạn ngạch Rate Limiting. Đây là khoản chi phí độ trễ an ninh (Security Overhead) hoàn toàn xứng đáng để đổi lấy sự bảo vệ an toàn cho hệ thống.
3. **Giải mã chỉ số 'Total Data Transferred to Client' (Dòng 9):**
   * Con số **191.5 MB** ở Case 3 là cao nhất toàn bảng không phải vì máy chủ EC2 bị nghẽn mạng, mà vì hệ thống phục vụ quá nhanh nên người dùng tải được nhiều trang nhất (**187,921 requests**). Lượng dữ liệu này hoàn toàn do máy chủ biên CDN chi trả, không hề gây áp lực lên băng thông Origin.

---

## 3.2. Phân Hệ 2 (BM2) — Đo Kiểm Đa Môi Trường Mạng (CMC FTTH vs 4G LTE)
* **Cấu hình tải:** 30 Virtual Users (CCU) | Tốc độ tăng tải: 10 users/giây | Thời gian chạy: 120 giây (2 phút).
* **Mục tiêu:** Đo lường trải nghiệm thực tế của người dùng từ mạng cáp quang doanh nghiệp và mạng di động không dây khi truy cập vào CDN nội địa so với đi thẳng sang Origin bên Mỹ.

### A. Bảng Số Liệu So Sánh Đa Mạng (Bản V2 Chuẩn Xác)

| Chỉ số đo lường | Đơn vị | CMC FTTH $\rightarrow$ CDN | 4G LTE $\rightarrow$ CDN | CMC FTTH $\rightarrow$ EC2 (Origin) | 4G LTE $\rightarrow$ EC2 (Origin) | Ý nghĩa kỹ thuật |
| :--- | :---: | :---: | :---: | :---: | :---: | :--- |
| **Tổng requests hoàn thành** | req | **7,620** | **6,584** | **5,032** | **2,801** | Số request hoàn tất trong 120s |
| **Thông lượng trung bình (RPS)**| req/s | **63.6** | **54.9** | **41.9** | **23.4** | Tốc độ đáp ứng của hệ thống |
| **Độ trễ trung bình (Overall Avg)**| ms | **49.6 ms** | **108.6 ms** | **265.6 ms** | **794.8 ms** | Trải nghiệm người dùng tổng thể |
| **Độ trễ P95 (Overall P95)** | ms | **110.0 ms** | **320.0 ms** | **280.0 ms** | **1,900.0 ms** | 95% request hoàn thành dưới mức này |
| **Độ trễ tối thiểu (Min Latency)**| ms | **25.5 ms** | **19.2 ms** | **245.3 ms** | **294.4 ms** | **Độ trễ vật lý sàn (RTT)** |
| **Độ trễ tĩnh (Static Avg)** | ms | **49.4 ms** | **129.4 ms** | **265.2 ms** | **802.7 ms** | Tải tài nguyên tĩnh |
| **Độ trễ động (Dynamic API Avg)**| ms | **49.9 ms** | **69.6 ms** | **266.3 ms** | **784.3 ms** | Tải API JSON nghiệp vụ |
| **Tỷ lệ lỗi (Failure Ratio)** | % | **0.00%** | **0.00%** | **0.00%** | **0.00%** | Độ ổn định (không lỗi HTTP) |
| **Tổng dữ liệu tải về (Client Data)**| MB | **7.84 MB** | **6.93 MB** | **5.38 MB** | **3.21 MB** | Băng thông client nhận về |

---

### B. Vị Trí Chèn Ảnh Minh Chứng Cho BM2

> 📷 **[VỊ TRÍ CHÈN ẢNH 3.1]:** *Ảnh chụp Dashboard Grafana — Kịch bản CMC FTTH $\rightarrow$ CDN vs EC2*  
> *(Gợi ý caption: Hình 3.7 - So sánh độ trễ mạng cáp quang văn phòng CMC khi qua CDN (~49ms) so với đi thẳng EC2 (~265ms))*

> 📷 **[VỊ TRÍ CHÈN ẢNH 3.2]:** *Ảnh chụp Dashboard Grafana — Kịch bản 4G LTE Hotspot $\rightarrow$ CDN vs EC2*  
> *(Gợi ý caption: Hình 3.8 - So sánh độ trễ mạng 4G di động: CDN giúp giảm độ trễ từ gần 800ms xuống chỉ còn 108ms)*

---

### C. Phân Tích Kỹ Thuật Chuyên Sâu Cho Phân Hệ BM2
1. **Min Latency phản ánh hoàn hảo quy luật khoảng cách địa lý:**
   * Khi qua CDN: Min Latency đạt **19.2ms – 25.5ms** nhờ các PoP Anycast của CMC đặt ngay tại Hà Nội / TP.HCM.
   * Khi đi thẳng EC2 Mỹ: Min Latency là **245.3 ms (FTTH)** và **294.4 ms (4G LTE)**. Con số này khớp chính xác với độ trễ truyền dẫn ánh sáng qua hệ thống cáp quang biển xuyên Thái Bình Dương (AAG/APG). Độ trễ của 4G cao hơn FTTH khoảng **49 ms**, đúng bằng thời gian trễ xử lý sóng vô tuyến (Air Interface Delay) của trạm BTS di động.
2. **Giá trị "cứu mạng" của CDN đối với người dùng di động (4G LTE):**
   * Khi **không có CDN**, người dùng 4G phải chịu độ trễ cực cao (**794.8 ms**, P95 chạm **1,900 ms**), thông lượng tụt xuống chỉ còn **23.4 RPS** vì mỗi kết nối TCP phải chờ tín hiệu bay nửa vòng trái đất.
   * Khi **có CMC Multi-CDN**, độ trễ của người dùng di động giảm mạnh xuống **108.6 ms (nhanh gấp 7.3 lần!)**, thông lượng tăng vọt lên **54.9 RPS (gấp 2.4 lần)**.

---

# PHẦN 4: KẾT LUẬN & ĐÁNH GIÁ TỔNG THỂ

## 4.1. Trả Lời Trực Diện Yêu Cầu Của Mentor

| Câu hỏi của Mentor | Kết quả thực tế đạt được |
| :--- | :--- |
| **Đã làm những gì?** | 1. Triển khai trọn vẹn cụm công cụ tạo tải và giám sát thời gian thực trên môi trường máy cục bộ (Docker Stack: Locust Master-Worker, Prometheus, Grafana Telemetry Dashboard và Loki Web SIEM).<br>2. Thực hiện đo kiểm định lượng trên **4 kịch bản kiến trúc** (Direct EC2, WAF Only, CDN Only, Full Stack) ở mức tải 300 CCU.<br>3. Đo kiểm thực tế **trên 2 môi trường mạng viễn thông** (Cáp quang văn phòng CMC FTTH vs Mạng di động 4G LTE Hotspot). |
| **Mục đích để làm gì?** | 1. Xác định năng lực giảm tải (Offload) và tăng tốc độ truy cập của CMC Multi-CDN khi máy chủ ứng dụng đặt xa tại Mỹ (AWS EC2 us-east-1).<br>2. Đo lường chi phí độ trễ an ninh (Security Latency Overhead) khi bật tầng bảo vệ CMC WAAP.<br>3. Đánh giá chất lượng dịch vụ (QoE) của người dùng cuối trên các loại hạ tầng mạng Internet khác nhau. |
| **Đạt được kết quả gì?** | 1. **Hiệu năng vượt bậc nhờ CDN:** Thông lượng tăng từ **317 req/s lên 626 req/s**, độ trễ giảm từ ~400ms–630ms xuống **48 ms** (nhanh gấp 8–13 lần).<br>2. **Bảo vệ máy chủ gốc:** Tỷ lệ Cache Offload đạt tới **99.9%**, giúp giải phóng CPU EC2 từ **7.44% xuống chỉ còn 0.56%**.<br>3. **Đo lường an ninh WAAP:** Xác định mức overhead độ trễ khi qua WAF là khoảng **~70ms – 140ms**, đổi lại khả năng phòng chống toàn diện các cuộc tấn công L7 và quét lỗ hổng.<br>4. **Trải nghiệm mạng di động:** CDN giúp người dùng 4G giảm độ trễ từ **794.8 ms xuống 108.6 ms** (nhanh gấp **7.3 lần**), tăng gấp đôi thông lượng. |

## 4.2. Nhận Định Về Bài Toán Stress Test & Giới Hạn Hạ Tầng
* **Điểm gãy của máy chủ gốc (Origin Breaking Point):** Ở kịch bản truy cập trực tiếp (Direct EC2), với 300 CCU, máy chủ EC2 `t3.micro` đã bắt đầu xuất hiện lỗi **0.55%** do nghẽn hàng đợi kết nối. Nếu không có lớp đệm CDN, máy chủ gốc chắc chắn sẽ sập (HTTP 502/504) khi lưu lượng tiếp tục tăng cao.
* **Sức mạnh lá chắn của CMC Multi-CDN:** Khi đặt sau Multi-CDN, dưới cùng mức tải 300 CCU, CPU của EC2 chỉ tiêu thụ **0.56%**. Điều này chứng minh kiến trúc Edge Caching đã **dời hoàn toàn điểm gãy chịu tải từ máy chủ gốc sang hạ tầng phân tán toàn cầu của CMC Telecom**. Ngưỡng chịu tải của hệ thống lúc này phụ thuộc vào quy mô băng thông của các Edge PoP chứ không còn bị giới hạn bởi cấu hình phần cứng của server nội bộ.

## 4.3. Bài Học Kinh Nghiệm & Đề Xuất Vận Hành
1. **Tối ưu hóa Cache Rule:** Phải luôn cấu hình quy tắc lưu đệm tường minh cho cả đuôi file tĩnh (`js|css|jpg|png`) và đường dẫn API đọc dữ liệu (`/api/`). Việc để cấu hình mặc định sẽ khiến CDN fallback về header Origin (`max-age=60`), dẫn tới hiện tượng trễ hàng chục giây do Cold Cache khi có đột biến tải.
2. **Quản trị DNS TTL khi chuyển dịch hệ thống:** Do các nhà mạng di động thường lưu cache DNS lâu hơn mạng cáp quang, khi lập kế hoạch chuyển đổi (Cut-over) domain sang CDN/WAF, cần hạ thấp giá trị TTL (xuống 60s – 300s) trước ít nhất 24 giờ để đảm bảo 100% người dùng được điều hướng chuẩn xác.
3. **Ứng dụng cụm Grafana & Locust vào CI/CD:** Bộ công cụ local test tải đã dựng hoàn toàn có thể tái sử dụng để làm Performance Gate trong quy trình release phần mềm định kỳ, giúp phát hiện sớm các điểm nghẽn hiệu năng trước khi đưa mã nguồn lên Production.
