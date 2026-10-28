# PHẦN 2: MÔ HÌNH TRIỂN KHAI & KIẾN TRÚC KỸ THUẬT

---

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

---

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
> 
> *(Gợi ý caption: Hình 2.1 - Cụm dịch vụ Docker đo kiểm và giám sát tải hoạt động trên máy cục bộ)*
