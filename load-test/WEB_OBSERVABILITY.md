# 🌐 Hướng Dẫn Giám Sát Lưu Lượng Web & An Ninh Với Grafana + Loki (Web Observability & SIEM Insights)

> **Mục tiêu**: Cung cấp góc nhìn toàn diện từ máy chủ gốc (**Origin Server / Nginx**), gom log thời gian thực về **Loki**, giải mã IP thật qua lớp **CMC Multi-CDN** và **CMC WAAP**, trực quan hóa trên **Grafana**.

---

## 1. Sơ Đồ Kiến Trúc Luồng Log & Dữ Liệu

```
                                  ┌─────────────────────────────┐
                                  │   Client / Browser / Locust │
                                  └──────────────┬──────────────┘
                                                 │
                                                 ▼
                                  ┌─────────────────────────────┐
                                  │ CMC Multi-CDN / WAAP Layer  │
                                  │  (Adds X-Forwarded-For)     │
                                  └──────────────┬──────────────┘
                                                 │
                                                 ▼
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                              MÁY CHỦ GỐC ORIGIN (AWS EC2 / DOCKER)                      │
│                                                                                         │
│   ┌────────────────────────┐         ┌─────────────────────────┐                        │
│   │   Nginx Reverse Proxy  │ ──────> │  Go Clean Architecture  │                        │
│   │   (Real-IP Restoration)│ <────── │  Core API (Port 8080)   │                        │
│   └───────────┬────────────┘         └─────────────────────────┘                        │
│               │                                                                         │
│               ▼ Ghi structured JSON log                                                 │
│       /var/log/nginx/access.log                                                         │
│               │                                                                         │
│               ▼ File scrape & Parse JSON                                                │
│   ┌────────────────────────┐                                                            │
│   │  Promtail Agent (9080) │                                                            │
│   └───────────┬────────────┘                                                            │
└───────────────┼─────────────────────────────────────────────────────────────────────────┘
                │ HTTP Push LogQL (Port 3100)
                ▼
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                     PHÂN HỆ GIÁM SÁT & QUAN SÁT TẬP TRUNG (OBSERVABILITY STACK)          │
│                                                                                         │
│   ┌────────────────────────┐               ┌────────────────────────────────────────┐   │
│   │    Grafana Loki TSDB   │ ────────────> │           Grafana (Port 3000)          │   │
│   │      (Port 3100)       │  Loki Source  │  - Dashboard: Web Traffic Observability│   │
│   └────────────────────────┘               │  - Dashboard: Locust Load Testing      │   │
│                                            └────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. Các Thành Phần Kỹ Thuật Đã Cấu Hình

### 2.1. Nginx Structured JSON Log & Phục Hồi Real IP (`nginx/nginx.conf.template`)
* **Khôi phục IP Client thật qua CDN & WAF**:
  ```nginx
  set_real_ip_from 0.0.0.0/0;
  real_ip_header X-Forwarded-For;
  real_ip_recursive on;
  ```
  * `$remote_addr`: Tự động trở thành **IP thật của người dùng cuối** (thay vì IP của Edge CDN).
  * `$realip_remote_addr`: Giữ nguyên **IP kết nối trực tiếp** (IP của POP CDN hoặc WAF Gateway).
* **Định dạng Log JSON giàu thông tin**:
  * `time_local`: Timestamp chuẩn ISO8601.
  * `remote_addr`: Client IP thật.
  * `realip_remote_addr`: Edge IP (WAF/CDN).
  * `host`: Hostname định tuyến (`direct.*`, `waap.*`, `cdn.*`, `full.*`).
  * `method`, `uri`, `status`, `body_bytes_sent`.
  * `request_time`: Tổng thời gian xử lý của Nginx.
  * `upstream_response_time`: Thời gian phản hồi riêng của Go API core.
  * `cdn_cache`: Header trạng thái cache (`HIT`, `MISS`, `BYPASS`).
  * `ua`, `referer`, `ssl_protocol`.
* **Ghi log kép (Dual Logging)**:
  * `/dev/stdout`: Cho phép dùng `docker compose logs -f nginx`.
  * `/var/log/nginx/access.log`: Được mount ra thư mục `./logs/nginx` trên máy chủ để Promtail đọc ổn định với hiệu năng cao.

### 2.2. Promtail Agent (`load-test/promtail/promtail-config.yaml`)
* Đọc trực tiếp `/var/log/nginx/access.log`.
* Tự động parse JSON và tách các label phục vụ truy vấn tốc độ cao: `status`, `method`, `host`, `cdn_cache`.
* Đồng bộ mốc thời gian timestamp chính xác theo thời điểm phát sinh request.

### 2.3. Grafana Loki (`load-test/loki/loki-config.yaml`)
* Chạy engine TSDB tối ưu hóa lưu trữ log dạng chuỗi thời gian.
* Tự động đánh chỉ mục label và lưu trữ dữ liệu nén trong volume `loki_data`.

---

## 3. Hướng Dẫn Vận Hành

### Kịch Bản A: Chạy Cùng Lúc Tại Local / Trên Cùng 1 Máy Chủ EC2
Nếu bạn chạy cả Web Origin và cụm Load-test / Observability trên cùng một máy:

```bash
# 1. Khởi động Web Origin & Nginx (tại thư mục gốc)
docker compose up -d --build

# 2. Khởi động cụm Observability & Load Testing (tại thư mục load-test)
cd load-test
docker compose up -d

# 3. Kiểm tra các dịch vụ đang chạy
docker compose ps
```

Các cổng dịch vụ sẵn sàng:
* **Grafana**: `http://localhost:3000` (User: `admin` / Password: `admin`)
* **Loki**: `http://localhost:3100`
* **Locust Web UI**: `http://localhost:8089`
* **Prometheus**: `http://localhost:9090`

---

### Kịch Bản B: Web Origin Trên EC2, Grafana/Loki Trên Máy Trạm (Hoặc Server Giám Sát Riêng)
Nếu LabShop Origin đang chạy trên AWS EC2 (`direct.thanhhieutiet.shop`):

1. **Trên máy giám sát (Local / Monitoring Server)**:
   * Chạy cụm `load-test`:
     ```bash
     cd load-test
     docker compose up -d
     ```
   * Mở port **3100** trên Security Group của máy giám sát (hoặc qua VPN/SSH Tunnel) để nhận log từ EC2.

2. **Trên máy chủ AWS EC2 (Origin)**:
   * Chạy script triển khai Promtail Agent với IP của máy giám sát:
     ```bash
     cd /opt/labshop # Hoặc thư mục chứa mã nguồn trên EC2
     chmod +x scripts/deploy-promtail-agent.sh
     ./scripts/deploy-promtail-agent.sh http://<IP_MÁY_GIÁM_SÁT>:3100
     ```
   * Promtail trên EC2 sẽ ngay lập tức tail file `/var/log/nginx/access.log` và đẩy liên tục về Loki.

---

## 4. Hướng Dẫn Khai Thác Dashboard Trên Grafana

Truy cập `http://localhost:3000`, vào menu **Dashboards** ➔ Thư mục **LabShop Observability** ➔ Chọn **`🌐 LabShop — Web Traffic & Security Observability`**.

Dashboard được thiết kế theo 5 tầng quan sát chuẩn SOC/SRE:

### 1. Tầng Chỉ Số Tổng Quan (KPIs):
* **Tổng Requests**: Tổng số request trong khung giờ đã chọn.
* **Tốc Độ Lưu Lượng (RPS)**: Tốc độ request thời gian thực.
* **Tỷ Lệ Lỗi (Error Rate %)**: Tỷ lệ phần trăm các request lỗi (`4xx` và `5xx`).
* **Độ Trễ P95**: Thời gian xử lý của 95% request nhanh nhất.
* **Sự Kiện Chặn An Ninh**: Số lượng request bị Nginx chặn với mã `403 Forbidden`.

### 2. Tầng Diễn Biến Lưu Lượng & Cache:
* **Biểu đồ RPS theo HTTP Status**: Phân biệt rõ các luồng `200` (xanh lá), `304` (xanh nhạt), `403` (tím), `429` (cam), `500/502` (đỏ).
* **Biểu đồ CDN Cache (HIT / MISS / BYPASS)**: Cho biết tỷ lệ request được CDN xử lý tại Edge so với request lọt về Origin.

### 3. Tầng Phân Tích Độ Trễ & Định Tuyến:
* **So sánh độ trễ Nginx Gateway vs Go Backend Core**: Biết được thời gian trễ nằm ở tầng proxy mạng hay tầng xử lý logic DB của ứng dụng Go.
* **Phân phối lưu lượng qua 4 Hostname**: So sánh trực tiếp tải vào `direct.*`, `waap.*`, `cdn.*`, `full.*`.

### 4. Tầng Thực Thể (Entities Breakdown):
* **Top 10 Client IP**: Nhận diện IP thực của người dùng hoặc máy tấn công (kể cả khi đi qua CDN).
* **Top 10 Endpoints**: Xem URL nào đang chịu tải nhiều nhất (ví dụ `/api/products`, `/api/orders`).
* **Top User-Agents**: Nhận diện lưu lượng đến từ Chrome/Safari, từ Locust (`locust`), hay từ công cụ dò quét (`sqlmap`, `curl`).

### 5. Tầng An Ninh & SIEM Insights (Threat Detection):
* **Cảnh báo Brute-force**: Theo dõi lưu lượng request bất thường vào endpoint đăng nhập `/api/auth/login`.
* **Dòng Log Sự Kiện 403**: Tự động lọc các hành vi truy cập trái phép bị chặn (ví dụ cố tình đọc file `.env`, `.git`, các chuỗi SQLi hoặc công cụ scanner bị rule Nginx chặn đứng).
* **Live Structured Log Stream**: Tìm kiếm từ khóa tức thì (gõ IP, path, status) trong toàn bộ luồng log thô.

---

## 5. Thử Nghiệm Kiểm Chứng (Validation)

Để thấy ngay dữ liệu trên Grafana Dashboard:

1. **Sinh lưu lượng bình thường**:
   * Chạy Smoke Test:
     ```bash
     go run ./cmd/smoke -base-url=http://localhost
     ```
2. **Sinh lưu lượng tải cao**:
   * Mở Locust Web UI (`http://localhost:8089`), đặt 50 users, ramp-up 5 users/s, Host `http://localhost` (hoặc `https://direct.thanhhieutiet.shop`) và bấm **Start swarming**.
   * Quan sát Grafana: Biểu đồ RPS, Latency và Log stream sẽ nhảy số liên tục theo thời gian thực!
3. **Mô phỏng hành vi tấn công (Security Probe Test)**:
   * Thử gọi request truy cập file nhạy cảm:
     ```bash
     curl -i http://localhost/.env
     curl -i http://localhost/.git/config
     ```
   * Thử giả mạo User-Agent của công cụ scanner:
     ```bash
     curl -i -A "sqlmap/1.8" http://localhost/api/products
     ```
   * **Kết quả**: Request lập tức nhận về `HTTP 403 Forbidden` và xuất hiện trực tiếp trong bảng **"Nhật Ký Các Cuộc Dò Quét & Tấn Công Bị Chặn"** trên Grafana Dashboard!
