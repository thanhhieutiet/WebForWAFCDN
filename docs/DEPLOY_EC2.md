# Hướng dẫn & Checklist Triển khai LabShop trên AWS EC2

Tài liệu này hướng dẫn chi tiết quy trình triển khai LabShop lên máy chủ EC2 Ubuntu (`ubuntu@ip-172-31-33-24`) sử dụng tên miền `thanhhieutiet.shop` theo **Phương án 1 (Docker Nginx độc lập & Mount Certbot SSL)**.

---

## 1. Tổng quan Kiến trúc Triển khai (Phương án 1)

```
[ Internet / Khách hàng / CMC WAAP & Multi-CDN ]
                      │
           Ports 80 & 443 (HTTP & HTTPS)
                      ▼
┌────────────────────────────────────────────────────────┐
│ Máy chủ AWS EC2 (Ubuntu 22.04 / 24.04)                 │
│                                                        │
│  [Native Host Nginx]: ĐÃ DỪNG (systemctl stop nginx)   │
│  /etc/letsencrypt/: Chứa chứng chỉ Certbot đã ký      │
│         │                                              │
│         │ Mount volume (:ro)                           │
│         ▼                                              │
│  ┌──────────────────────────────────────────────────┐  │
│  │ Docker Nginx Container (Ports 80 & 443)          │  │
│  │ - SSL: /etc/letsencrypt/live/thanhhieutiet.shop  │  │
│  │ - Fallback: Self-signed dummy cert nếu thiếu cert│  │
│  │ - Static assets (/srv/static & /assets)          │  │
│  └───────────────────────┬──────────────────────────┘  │
│                          │ proxy_pass http://api:8080  │
│                          ▼                             │
│  ┌──────────────────────────────────────────────────┐  │
│  │ Docker Go API Origin Container (Port 8080)       │  │
│  │ - Chi router, SQLite WAL, Lab & Business Engine  │  │
│  └──────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────┘
```

---

## 2. Checklist Thao tác trên EC2 (Từng bước)

### Bước 1: Đăng nhập vào máy chủ EC2
```bash
ssh -i <your-key.pem> ubuntu@<EC2-PUBLIC-IP>
```

---

### Bước 2: Dừng và tắt Nginx native trên máy chủ Host
Tránh xung đột cổng 80 và 443 giữa Nginx cài trên host và Nginx trong Docker:
```bash
# Kiểm tra trạng thái nginx host
sudo systemctl status nginx

# Dừng Nginx host
sudo systemctl stop nginx

# Tắt tự khởi động cùng OS để tránh xung đột khi reboot EC2
sudo systemctl disable nginx

# Kiểm tra đảm bảo không còn process nào chiếm cổng 80 và 443
sudo ss -tulpn | grep -E ':80|:443'
# (Nếu không có output nào xuất hiện nghĩa là cổng đã hoàn toàn giải phóng)
```

---

### Bước 3: Di chuyển vào thư mục dự án và Pull code mới nhất
*(Lưu ý: Sau khi bạn push code từ máy local lên GitHub)*
```bash
cd ~/WebForWAF-CDN   # Hoặc đường dẫn thư mục repo trên EC2 của bạn

# Kiểm tra branch
git status

# Kéo code mới nhất từ GitHub
git pull origin master
```

---

### Bước 4: Kiểm tra và cấu hình file `.env`
Đảm bảo file `.env` tồn tại và trỏ đúng đường dẫn SSL trên EC2:
```bash
# Nếu chưa có .env, copy từ .env.example
if [ ! -f .env ]; then cp .env.example .env; fi

# Xem hoặc chỉnh sửa cấu hình
nano .env
```

**Các biến quan trọng cần kiểm tra trong `.env`:**
```ini
APP_ENV=production
HTTP_ADDR=:8080
JWT_SECRET=labshop-super-secret-jwt-key-minimum-32-chars-long!
ADMIN_TOKEN=labshop-admin-secret-token-change-in-production

# Đường dẫn mount certbot từ Host vào Docker container
SSL_CERTS_DIR=/etc/letsencrypt
SSL_CERT_PATH=/etc/letsencrypt/live/thanhhieutiet.shop/fullchain.pem
SSL_KEY_PATH=/etc/letsencrypt/live/thanhhieutiet.shop/privkey.pem

MAX_UPLOAD_MB=10
CATALOG_CACHE_SECONDS=60
ORIGIN_GZIP=off
SEED=true
```

---

### Bước 5: Phân quyền thư mục chứng chỉ Certbot (Nếu cần)
Docker Nginx chạy với non-root hoặc alpine user cần quyền đọc thư mục `/etc/letsencrypt/live/` và `/etc/letsencrypt/archive/`:
```bash
# Kiểm tra sự tồn tại của file cert
sudo ls -la /etc/letsencrypt/live/thanhhieutiet.shop/

# Đảm bảo Docker container có quyền đọc certs:
sudo chmod -R 755 /etc/letsencrypt/archive
sudo chmod -R 755 /etc/letsencrypt/live
```

---

### Bước 6: Build và Khởi chạy Docker Compose
```bash
# Build và chạy ngầm
docker compose up -d --build

# Kiểm tra trạng thái các container
docker compose ps
```
Kết quả mong muốn:
- `api`: State `Up (healthy)`
- `nginx`: State `Up` (Ports `0.0.0.0:80->80/tcp`, `0.0.0.0:443->443/tcp`)

Xem logs Nginx để xác nhận SSL đã nạp thành công:
```bash
docker compose logs nginx
# Mong đợi thấy dòng:
# [LabShop Nginx] Found active SSL certificates at /etc/letsencrypt/live/thanhhieutiet.shop/fullchain.pem
```

---

### Bước 7: Chạy kiểm thử tự động (Smoke Test)
```bash
# Kiểm thử trực tiếp qua curl
curl -i http://localhost/api/health
curl -i -k https://localhost/api/health

# Chạy Smoke Test CLI của dự án (kiểm tra toàn bộ 18 test cases)
docker compose exec api /usr/local/bin/server -version || true
# Hoặc nếu trên EC2 có cài Go:
go run ./cmd/smoke -base-url="http://localhost"
```

---

### Bước 8: Kiểm tra từ Internet / Trình duyệt
Mở trình duyệt hoặc terminal từ máy ngoài:
- `http://thanhhieutiet.shop` -> Tự động phục vụ Storefront LabShop
- `https://thanhhieutiet.shop` -> Kết nối HTTPS an toàn với chứng chỉ Let's Encrypt hợp lệ
- `https://thanhhieutiet.shop/tutorial` -> Giao diện hướng dẫn và điều khiển Lab Multi-CDN & WAAP
- `https://thanhhieutiet.shop/docs` -> Tài liệu API Swagger/OpenAPI

---

## 3. Quy trình Gia hạn Chứng chỉ SSL Let's Encrypt (Certbot Renewal)

Khi chứng chỉ Let's Encrypt sắp hết hạn (sau mỗi 60-90 ngày), bạn có thể gia hạn theo 1 trong 2 cách sau:

### Cách 1: Sử dụng Certbot Webroot qua Docker Nginx (Không gián đoạn dịch vụ)
Nginx container đã cấu hình phục vụ thư mục tĩnh `/.well-known/acme-challenge/`.
```bash
sudo certbot renew --webroot -w $(pwd)/frontend/dist

# Reload Nginx container để nhận chứng chỉ mới mà không cần restart
docker compose exec nginx nginx -s reload
```

### Cách 2: Tạm dừng Docker Nginx ngắn hạn để Certbot standalone gia hạn
```bash
# 1. Tạm dừng Nginx container trong 10 giây
docker compose stop nginx

# 2. Chạy certbot gia hạn standalone
sudo certbot renew --standalone

# 3. Khởi động lại Nginx container
docker compose start nginx
```

---

## 4. Xử lý sự cố thường gặp (Troubleshooting)

| Hiện tượng | Nguyên nhân | Cách khắc phục |
| :--- | :--- | :--- |
| `bind: address already in use` | Nginx host hoặc process khác đang chiếm port 80/443 | Chạy `sudo systemctl stop nginx && sudo fuser -k 80/tcp 443/tcp` rồi chạy lại `docker compose up -d` |
| `permission denied` khi đọc `/etc/letsencrypt/live` | Thư mục archive bị khóa chỉ cho root đọc | Chạy `sudo chmod -R 755 /etc/letsencrypt/archive /etc/letsencrypt/live` |
| Nginx báo `Using self-signed fallback cert` | Đường dẫn `SSL_CERT_PATH` trong `.env` chưa đúng | Kiểm tra lại tên domain trong `/etc/letsencrypt/live/` và cập nhật `.env` |
| API container báo `unhealthy` | Database SQLite lock hoặc port 8080 bị chặn | Chạy `docker compose logs api` để xem log chi tiết |
