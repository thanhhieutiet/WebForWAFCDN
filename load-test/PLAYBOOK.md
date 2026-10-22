# 📖 CẨM NANG HƯỚNG DẪN TEST TẢI (LOAD TESTING PLAYBOOK)
## Đánh giá Hiệu năng & Khả năng Phòng thủ CMC Multi-CDN & WAAP

Tài liệu này được biên soạn để bạn **tự tay thực hiện toàn bộ quy trình kiểm thử tải**, quan sát biểu đồ thời gian thực trên **Locust & Grafana**, và **tự chụp ảnh minh chứng (POC)** để đưa vào báo cáo gửi Mentor.

---

## 📑 MỤC LỤC
1. [Khởi Động Trạm Bắn Tải (Local Docker Stack)](#1-khởi-động-trạm-bắn-tải-local-docker-stack)
2. [Các Cổng Dịch Vụ & Giao Diện Quản Trị](#2-các-cổng-dịch-vụ--giao-diện-quản-trị)
3. [Thiết Lập 4 Kịch Bản Đo Kiểm (Topology Benchmark)](#3-thiết-lập-4-kịch-bản-đo-kiểm-topology-benchmark)
4. [Quy Trình Chạy 3 Bài Test Thực Chiến & Chụp POC](#4-quy-trình-chạy-3-bài-test-thực-chiến--chụp-poc)
   * [Bài Test 1: Sức Mạnh Caching của Multi-CDN (Static Cache Stress)](#bài-test-1-sức-mạnh-caching-của-multi-cdn-static-cache-stress)
   * [Bài Test 2: Đo Ngưỡng Chịu Tải của Máy Chủ Origin (Dynamic API Stress)](#bài-test-2-đo-ngưỡng-chịu-tải-của-máy-chủ-origin-dynamic-api-stress)
   * [Bài Test 3: Kích Hoạt Luật Phòng Thủ WAF (Rate Limiting & Anti-Brute Force)](#bài-test-3-kích-hoạt-luật-phòng-thủ-waf-rate-limiting--anti-brute-force)
5. [Mẫu Bảng Tổng Hợp Số Liệu Nộp Mentor](#5-mẫu-bảng-tổng-hợp-số-liệu-nộp-mentor)

---

## 1. Khởi Động Trạm Bắn Tải (Local Docker Stack)

Trên máy tính cá nhân của bạn, mở terminal tại thư mục dự án và chuyển vào `load-test/`:

```powershell
cd c:\CMC\WAF\WebForWAF&CDN\load-test

# Khởi động toàn bộ 4 dịch vụ (Locust, Exporter, Prometheus, Grafana)
docker compose up -d

# Kiểm tra trạng thái các container đang chạy
docker compose ps
```

Kết quả hiển thị `Up` cho cả 4 container:
* `labshop-locust` (Port 8089)
* `labshop-locust-exporter` (Port 9191)
* `labshop-prometheus` (Port 9090)
* `labshop-grafana` (Port 3000)

---

## 2. Các Cổng Dịch Vụ & Giao Diện Quản Trị

| Giao diện | Địa chỉ URL | Tài khoản / Cách truy cập | Nhiệm vụ chính |
| :--- | :--- | :--- | :--- |
| **Locust Web UI** | [http://localhost:8089](http://localhost:8089) | Không cần mật khẩu | Cấu hình số CCU, Host, bấm **Start/Stop** |
| **Grafana Dashboard** | [http://localhost:3000](http://localhost:3000) | `admin` / `admin` | Xem biểu đồ RPS, Latency p95/p99 tự động |
| **Prometheus Targets** | [http://localhost:9090/targets](http://localhost:9090/targets) | Không cần mật khẩu | Kiểm tra trạng thái kết nối Locust Exporter |

> [!TIP]
> **Dashboard Grafana đã được cấu hình tự động (Zero-Config)!**  
> Khi bạn đăng nhập vào Grafana `http://localhost:3000`, chỉ cần vào menu **Dashboards** ➔ Mở thư mục **Load Testing** ➔ Chọn dashboard **`CMC LabShop - Locust Load Testing & WAF/CDN Benchmark`**. Đồ thị sẽ tự động cập nhật mỗi 5 giây.

---

## 3. Thiết Lập 4 Kịch Bản Đo Kiểm (Topology Benchmark)

Khi chạy test trong giao diện Locust ([http://localhost:8089](http://localhost:8089)), bạn chỉ cần nhập URL của target vào ô **Host**:

| Kịch bản | Host URL cần nhập vào Locust | Ý nghĩa đo lường |
| :--- | :--- | :--- |
| **Kịch bản 0 (Baseline)** | `http://<IP_EC2>` | Đo năng lực gốc của máy chủ Origin (không có lá chắn nào). |
| **Kịch bản 1** | `https://waap.thanhhieutiet.shop` *(nếu có)* | Đo độ trễ xử lý gói tin của CMC WAF (WAF Overhead). |
| **Kịch bản 2** | `https://cdn.thanhhieutiet.shop` *(nếu có)* | Đo hiệu năng Edge Cache của CMC Multi-CDN. |
| **Kịch bản 3 (Full)** | `https://app.thanhhieutiet.shop` | Đo mô hình hoàn chỉnh hiện tại: `Client ➔ CDN ➔ WAF ➔ EC2`. |

---

## 4. Quy Trình Chạy 3 Bài Test Thực Chiến & Chụp POC

---

### Bài Test 1: Sức Mạnh Caching của Multi-CDN (Static Cache Stress)
> **Mục tiêu:** Chứng minh CDN "hấp thụ" được tải tệp tĩnh, giảm độ trễ từ hàng trăm mili-giây xuống chỉ còn vài mili-giây, máy chủ EC2 Origin không bị ảnh hưởng.

1. **Thao tác trên Locust Web UI ([http://localhost:8089](http://localhost:8089)):**
   * **Number of users:** `200`
   * **Ramp up (users started/second):** `20`
   * **Host:** `https://app.thanhhieutiet.shop`
   * **Advanced Options (User class):** Chọn duy nhất `StaticAssetUser` (hoặc để trống để chạy tất cả).
   * Bấm **Start swarming**.
2. **Quan sát & Ghi nhận:**
   * Cho chạy liên tục trong **2 - 3 phút**.
   * Quan sát trên Locust: Tốc độ RPS đạt cao (500 - 1,500 RPS).
   * Độ trễ (Average Response Time): Rất thấp (**10ms - 25ms**) do dữ liệu được trả về ngay từ node CDN gần nhất.
   * Mở terminal EC2 kiểm tra CPU: `htop` hoặc `top` ➔ CPU của con EC2 vẫn cực kỳ nhàn nhã (~2-5%).
3. **📸 Góc Chụp Ảnh POC cho Mentor:**
   * **POC 1.1:** Chụp màn hình tab **Statistics** và **Charts** của Locust với thông số Requests/s cao và Failure = 0%.
   * **POC 1.2:** Chụp Dashboard **Grafana** cho thấy đường cong Latency nằm sát đáy (xanh lá).

---

### Bài Test 2: Đo Ngưỡng Chịu Tải của Máy Chủ Origin (Dynamic API Stress)
> **Mục tiêu:** Kiểm tra khả năng xử lý của backend Go API và SQLite khi có tải gọi liên tục vào các endpoint động không lưu cache (`/api/time`, `/api/products`).

1. **Thao tác trên Locust Web UI:**
   * Bấm **Edit** hoặc **New test**.
   * **Number of users:** Bắt đầu từ `50`, tăng dần lên `100`, rồi `200`.
   * **Ramp up:** `10`
   * **Host:** `https://app.thanhhieutiet.shop`
   * **Advanced Options (User class):** Chọn `DynamicApiUser`.
   * Bấm **Start swarming**.
2. **Quan sát & Ghi nhận:**
   * Quan sát độ trễ phản hồi: Độ trễ sẽ dao động từ **50ms - 150ms** (vì mỗi request đều phải chuyển tiếp về EC2 và truy vấn SQLite).
   * Mở terminal EC2 xem log Nginx/Go:
     ```bash
     sudo docker logs -f --tail 20 webforwafcdn-api-1
     ```
   * Bạn sẽ thấy các bản ghi log truy vấn liên tục đổ về origin.
3. **📸 Góc Chụp Ảnh POC cho Mentor:**
   * **POC 2.1:** Chụp biểu đồ RPS và Response Time trên Grafana thể hiện rõ sự khác biệt giữa bài Test 1 (Static) và bài Test 2 (Dynamic).

---

### Bài Test 3: Kích Hoạt Luật Phòng Thủ WAF (Rate Limiting & Anti-Brute Force)
> **Mục tiêu:** Chứng minh khi có kẻ tấn công hoặc bot cố tình spam request dồn dập vào endpoint nhạy cảm (`/api/auth/login`), hệ thống WAF/Origin sẽ tự động kích hoạt mã lỗi `429 Too Many Requests` hoặc `403 Forbidden` để bảo vệ tài khoản.

1. **Thao tác trên Locust Web UI:**
   * **Number of users:** `50`
   * **Ramp up:** `25` (tăng tốc đột ngột)
   * **Host:** `https://app.thanhhieutiet.shop`
   * **Advanced Options (User class):** Chọn `WafRateLimitProbeUser`.
   * Bấm **Start swarming**.
2. **Quan sát & Ghi nhận:**
   * Locust sẽ gửi hàng chục lượt login thử mỗi giây.
   * Ngay sau khi vượt quá 5 lần thất bại trong 5 phút từ IP của bạn (theo luật `LoginLimiter`), server sẽ lập tức trả về **HTTP 429** (*Too Many Requests*).
   * Nếu có luật WAF chặn bot / injection, WAF sẽ trả về **HTTP 403**.
   * Trên Locust: Các request nhận 429/403 được kịch bản định nghĩa là *phòng thủ thành công*.
3. **📸 Góc Chụp Ảnh POC cho Mentor:**
   * **POC 3.1:** Chụp bảng phân bố mã trạng thái HTTP trên Locust (thấy xuất hiện các phản hồi 429/403).
   * **POC 3.2:** Mở portal **CMC WAAP > Logs > View Logs** chụp lại các bản ghi WAF cảnh báo / chặn request.

---

## 5. Mẫu Bảng Tổng Hợp Số Liệu Nộp Mentor

Sau khi hoàn thành các bài test, bạn điền các số liệu đo được vào bảng sau để đưa vào báo cáo nghiệm thu:

| Kịch bản Test | Target Host | CCU | Max RPS | Avg Latency (ms) | P95 Latency (ms) | Tỉ lệ lỗi (%) | Đánh giá |
| :--- | :--- | :---: | :---: | :---: | :---: | :---: | :--- |
| **1. CDN Static Stress** | `app.thanhhieutiet.shop` | 200 | *Ví dụ: 1200* | *18ms* | *32ms* | 0% | CDN cache mượt, EC2 Origin nhàn rỗi |
| **2. Origin Dynamic API** | `app.thanhhieutiet.shop` | 100 | *Ví dụ: 350* | *85ms* | *140ms* | 0% | Origin xử lý tốt, không bị sập DB |
| **3. WAF Rate Limit** | `app.thanhhieutiet.shop` | 50 | *Ví dụ: 150* | *45ms* | *70ms* | 0% | Phạt đúng mã 429 khi spam đăng nhập |

---

## 6. Dừng & Dọn Dẹp Cụm Test Tải Khi Hoàn Thành

Khi làm xong bài test và chụp xong ảnh POC, bạn chỉ cần gõ lệnh:

```powershell
cd c:\CMC\WAF\WebForWAF&CDN\load-test
docker compose down
```

Hệ thống sẽ tắt sạch các container để giải phóng tài nguyên CPU/RAM cho máy tính của bạn. Dữ liệu cấu hình và kịch bản vẫn được lưu nguyên vẹn trong thư mục `load-test/`.
