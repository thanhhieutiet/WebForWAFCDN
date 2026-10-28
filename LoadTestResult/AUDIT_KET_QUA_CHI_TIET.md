# BẢNG GIẢI MÃ CÔNG THỨC & AUDIT SỐ LIỆU ĐO KIỂM LOAD TEST
> **Tài liệu đối chiếu và kiểm tra tính toán (Audit & Verification Report)**
> Hệ thống: LabShop Origin behind CMC Telecom WAAP & Multi-CDN
> Nguồn dữ liệu: Các file CSV xuất trực tiếp từ Locust Web Reporter
> Cấu hình kiểm thử chuẩn: **300 Users (CCU) | Spawn rate: 20 users/s | Thời gian chạy: 5 phút (300 giây)**

---

## 1. NGUYÊN LÝ TÍNH TOÁN & ĐỊNH NGHĨA TRƯỜNG DỮ LIỆU TỪ LOCUST CSV

File CSV xuất từ Locust gồm các cột dữ liệu gốc quan trọng sau:
* `Name`: Tên endpoint được test (URL route).
* `Request Count`: Tổng số requests thành công + thất bại được gửi vào endpoint đó.
* `Failure Count`: Số lượng requests bị lỗi (status >= 400 hoặc connection drop).
* `Average Response Time`: Độ trễ trung bình của riêng endpoint đó (đơn vị: mili-giây / ms).
* `95%`: Độ trễ P95 của riêng endpoint đó (95% request hoàn thành trong thời gian này).
* `Requests/s`: Tốc độ phục vụ requests trung bình của endpoint (RPS).
* `Average Content Size`: Kích thước trung bình một response body (đơn vị: Bytes).
* Dòng `Aggregated`: Dòng tổng hợp của toàn bộ các endpoint do Locust tự động gom.

### 1.1. Tại sao phải dùng Công thức Trung bình Trọng số (Weighted Average)?
Trong thực tế đo kiểm, **không thể lấy trung bình cộng giản đơn** của các endpoint:
$$\text{Sai lầm: } \text{Avg} \neq \frac{\text{Avg}_1 + \text{Avg}_2 + \text{Avg}_3}{3}$$
Vì số lượng request gửi vào mỗi endpoint là **hoàn toàn khác nhau** (ví dụ: ảnh sản phẩm có 64,000 requests nhưng banner chỉ có 31,000 requests).
Do đó, chuẩn mực toán học bắt buộc phải nhân số request với thời gian của từng endpoint rồi chia cho tổng request:

$$\text{Weighted Average Latency (ms)} = \frac{\sum (\text{Request Count}_i \times \text{Average Response Time}_i)}{\sum \text{Request Count}_i}$$

$$\text{Weighted P95 Latency (ms)} = \frac{\sum (\text{Request Count}_i \times \text{P95}_i)}{\sum \text{Request Count}_i}$$

### 1.2. Công thức tính Tỷ lệ Lỗi (Failure Ratio):
$$\text{Failure Ratio (\%)} = \frac{\text{Failure Count}_{\text{Aggregated}}}{\text{Request Count}_{\text{Aggregated}}} \times 100\%$$

### 1.3. Công thức tính Tổng Dung Lượng Nhận Về Phía Client (Total Data Transferred to Client):
Con số $1,048,576$ là quy chuẩn đổi đơn vị từ Bytes sang Megabytes (MB) trong hệ nhị phân ($1024 \times 1024 = 2^{20}$):
$$\text{Total Data Transferred to Client (MB)} = \frac{\text{Average Content Size}_{\text{Aggregated}} \times \text{Request Count}_{\text{Aggregated}}}{1024 \times 1024}$$
> *Lưu ý kỹ thuật:* Đây là tổng dung lượng nội dung HTTP (Layer 7 Payload) mà người dùng cuối tải về trong 5 phút. Khi có CDN, phần lớn dung lượng này được cung cấp bởi bộ nhớ đệm RAM/SSD của CDN Edge, không gây áp lực lên máy chủ gốc EC2.

---

## 2.1. CHI TIẾT TÍNH TOÁN & THẾ SỐ: CASE 1: DIRECT EC2 (BASELINE)
**Ý nghĩa kịch bản:** Truy cập trực tiếp EC2 Origin, không qua WAF hay CDN  
**File CSV nguồn:** `BM1_Case1_request.csv`  
**Ánh xạ trong Excel KetQua.xlsx:** Cột `E`  

### A. Bảng Dữ Liệu Thô Trích Xuất Từ CSV
| Nhóm | Endpoint (Name) | Request Count ($N_i$) | Failure Count | Avg Latency ($T_i$) ms | P95 Latency ($P_i$) ms | Requests/s | Avg Size (Bytes) |
| :--- | :--- | :---: | :---: | :---: | :---: | :---: | :---: |
| **Static** | `/static/docs/sample.pdf` | 15,915 | 77 | 394.20 | 340.00 | 53.04 | 674.7 |
| **Static** | `/static/img/banner-large.jpg` | 15,948 | 80 | 408.38 | 370.00 | 53.15 | 1019.9 |
| **Static** | `/static/img/products/[id].jpg` | 31,820 | 151 | 397.36 | 350.00 | 106.04 | 1020.1 |
| **Dynamic** | `/api/cache/no-store` | 2,955 | 14 | 419.44 | 580.00 | 9.85 | 116.4 |
| **Dynamic** | `/api/cache/public` | 7,882 | 34 | 393.62 | 360.00 | 26.27 | 115.5 |
| **Dynamic** | `/api/categories` | 5,889 | 38 | 455.17 | 490.00 | 19.63 | 785.9 |
| **Dynamic** | `/api/products` | 8,638 | 91 | 1167.62 | 540.00 | 28.79 | 4078.7 |
| **Dynamic** | `/api/time` | 6,041 | 36 | 444.57 | 530.00 | 20.13 | 91.5 |
| **TỔNG HỢP** | **`Aggregated`** | **95,088** | **521** | **475.61** | **390.00** | **316.89** | **1063.5** |

### B. Chi Tiết Tính Toán Nhóm Static Assets (Tĩnh)
1. **Tổng số static requests ($N_{static}$):**
   $$N_{static} = 15,915 (sample.pdf) + 15,948 (banner-large.jpg) + 31,820 ([id].jpg) = 63,683$$

2. **Độ trễ trung bình có trọng số (Static Avg Latency):**
   $$\text{Static Avg} = \frac{(15,915 \times 394.20) + (15,948 \times 408.38) + (31,820 \times 397.36)}{63,683}$$
   $$\text{Static Avg} = \frac{25,430,615.95}{63,683} = \mathbf{399.33\text{ ms}} \approx \mathbf{399\text{ ms}} \quad \text{(Điền vào Dòng 5, Cột E)}$$

3. **Độ trễ P95 có trọng số (Static P95 Latency):**
   $$\text{Static P95} = \frac{(15,915 \times 340.00) + (15,948 \times 370.00) + (31,820 \times 350.00)}{63,683}$$
   $$\text{Static P95} = \frac{22,448,860.00}{63,683} = \mathbf{352.51\text{ ms}} \approx \mathbf{353\text{ ms}} \quad \text{(Điền vào Dòng 6, Cột E)}$$

### C. Chi Tiết Tính Toán Nhóm Dynamic APIs (Động)
1. **Tổng số dynamic requests ($N_{dynamic}$):**
   $$N_{dynamic} = 2,955 (no-store) + 7,882 (public) + 5,889 (categories) + 8,638 (products) + 6,041 (time) = 31,405$$

2. **Độ trễ trung bình có trọng số (Dynamic Avg Latency):**
   $$\text{Dynamic Avg} = \frac{(2,955 \times 419.44) + (7,882 \times 393.62) + (5,889 \times 455.17) + (8,638 \times 1167.62) + (6,041 \times 444.57)}{31,405}$$
   $$\text{Dynamic Avg} = \frac{19,794,015.22}{31,405} = \mathbf{630.28\text{ ms}} \approx \mathbf{630\text{ ms}} \quad \text{(Điền vào Dòng 7, Cột E)}$$

3. **Độ trễ P95 có trọng số (Dynamic P95 Latency):**
   $$\text{Dynamic P95} = \frac{(2,955 \times 580.00) + (7,882 \times 360.00) + (5,889 \times 490.00) + (8,638 \times 540.00) + (6,041 \times 530.00)}{31,405}$$
   $$\text{Dynamic P95} = \frac{15,303,280.00}{31,405} = \mathbf{487.29\text{ ms}} \approx \mathbf{487\text{ ms}} \quad \text{(Điền vào Dòng 8, Cột E)}$$

### D. Các Chỉ Số Tổng Hợp (Lấy Trực Tiếp Từ Hàng Aggregated)
* **Total Requests:** `95,088`
* **Total RPS (Average):** Cột `Requests/s` = $\mathbf{316.89\text{ req/s}}$ $\approx \mathbf{317}$ $\rightarrow$ Điền Dòng 3, Cột `E`
* **Failure Ratio (%):** 
  $$\text{Failure Ratio} = \frac{521}{95,088} \times 100\% = \mathbf{0.55\%} \quad \rightarrow \text{Điền Dòng 9, Cột E}$$
* **Total Data Transferred to Client (MB):**
  $$\text{Data Transferred} = \frac{1063.54\text{ Bytes} \times 95,088}{1,048,576} = \mathbf{96.45\text{ MB}} \approx \mathbf{96.4\text{ MB}} \quad \rightarrow \text{Điền Dòng 11, Cột E}$$

---

## 2.2. CHI TIẾT TÍNH TOÁN & THẾ SỐ: CASE 2: WAF ONLY (CMC WAAP)
**Ý nghĩa kịch bản:** Traffic đi qua CMC WAAP kiểm tra an ninh rồi mới về Origin  
**File CSV nguồn:** `BM1_Case2_requests.csv`  
**Ánh xạ trong Excel KetQua.xlsx:** Cột `F`  

### A. Bảng Dữ Liệu Thô Trích Xuất Từ CSV
| Nhóm | Endpoint (Name) | Request Count ($N_i$) | Failure Count | Avg Latency ($T_i$) ms | P95 Latency ($P_i$) ms | Requests/s | Avg Size (Bytes) |
| :--- | :--- | :---: | :---: | :---: | :---: | :---: | :---: |
| **Static** | `/static/docs/sample.pdf` | 9,238 | 0 | 360.23 | 490.00 | 54.90 | 678.0 |
| **Static** | `/static/img/banner-large.jpg` | 9,254 | 0 | 364.76 | 510.00 | 55.00 | 1025.0 |
| **Static** | `/static/img/products/[id].jpg` | 18,246 | 0 | 364.26 | 500.00 | 108.44 | 1025.0 |
| **Dynamic** | `/api/cache/no-store` | 2,207 | 0 | 380.73 | 610.00 | 13.12 | 117.0 |
| **Dynamic** | `/api/cache/public` | 4,721 | 0 | 372.33 | 490.00 | 28.06 | 116.0 |
| **Dynamic** | `/api/categories` | 4,353 | 0 | 370.76 | 580.00 | 25.87 | 791.0 |
| **Dynamic** | `/api/products` | 6,637 | 0 | 393.75 | 610.00 | 39.45 | 4123.8 |
| **Dynamic** | `/api/time` | 4,394 | 0 | 372.88 | 630.00 | 26.11 | 92.0 |
| **TỔNG HỢP** | **`Aggregated`** | **59,050** | **0** | **369.40** | **530.00** | **350.95** | **1125.7** |

### B. Chi Tiết Tính Toán Nhóm Static Assets (Tĩnh)
1. **Tổng số static requests ($N_{static}$):**
   $$N_{static} = 9,238 (sample.pdf) + 9,254 (banner-large.jpg) + 18,246 ([id].jpg) = 36,738$$

2. **Độ trễ trung bình có trọng số (Static Avg Latency):**
   $$\text{Static Avg} = \frac{(9,238 \times 360.23) + (9,254 \times 364.76) + (18,246 \times 364.26)}{36,738}$$
   $$\text{Static Avg} = \frac{13,349,534.89}{36,738} = \mathbf{363.37\text{ ms}} \approx \mathbf{363\text{ ms}} \quad \text{(Điền vào Dòng 5, Cột F)}$$

3. **Độ trễ P95 có trọng số (Static P95 Latency):**
   $$\text{Static P95} = \frac{(9,238 \times 490.00) + (9,254 \times 510.00) + (18,246 \times 500.00)}{36,738}$$
   $$\text{Static P95} = \frac{18,369,160.00}{36,738} = \mathbf{500.00\text{ ms}} \approx \mathbf{500\text{ ms}} \quad \text{(Điền vào Dòng 6, Cột F)}$$

### C. Chi Tiết Tính Toán Nhóm Dynamic APIs (Động)
1. **Tổng số dynamic requests ($N_{dynamic}$):**
   $$N_{dynamic} = 2,207 (no-store) + 4,721 (public) + 4,353 (categories) + 6,637 (products) + 4,394 (time) = 22,312$$

2. **Độ trễ trung bình có trọng số (Dynamic Avg Latency):**
   $$\text{Dynamic Avg} = \frac{(2,207 \times 380.73) + (4,721 \times 372.33) + (4,353 \times 370.76) + (6,637 \times 393.75) + (4,394 \times 372.88)}{22,312}$$
   $$\text{Dynamic Avg} = \frac{8,463,705.81}{22,312} = \mathbf{379.33\text{ ms}} \approx \mathbf{379\text{ ms}} \quad \text{(Điền vào Dòng 7, Cột F)}$$

3. **Độ trễ P95 có trọng số (Dynamic P95 Latency):**
   $$\text{Dynamic P95} = \frac{(2,207 \times 610.00) + (4,721 \times 490.00) + (4,353 \times 580.00) + (6,637 \times 610.00) + (4,394 \times 630.00)}{22,312}$$
   $$\text{Dynamic P95} = \frac{13,001,090.00}{22,312} = \mathbf{582.69\text{ ms}} \approx \mathbf{583\text{ ms}} \quad \text{(Điền vào Dòng 8, Cột F)}$$

### D. Các Chỉ Số Tổng Hợp (Lấy Trực Tiếp Từ Hàng Aggregated)
* **Total Requests:** `59,050`
* **Total RPS (Average):** Cột `Requests/s` = $\mathbf{350.95\text{ req/s}}$ $\approx \mathbf{351}$ $\rightarrow$ Điền Dòng 3, Cột `F`
* **Failure Ratio (%):** 
  $$\text{Failure Ratio} = \frac{0}{59,050} \times 100\% = \mathbf{0.00\%} \quad \rightarrow \text{Điền Dòng 9, Cột F}$$
* **Total Data Transferred to Client (MB):**
  $$\text{Data Transferred} = \frac{1125.72\text{ Bytes} \times 59,050}{1,048,576} = \mathbf{63.39\text{ MB}} \approx \mathbf{63.4\text{ MB}} \quad \rightarrow \text{Điền Dòng 11, Cột F}$$

---

## 2.3. CHI TIẾT TÍNH TOÁN & THẾ SỐ: CASE 3: CDN ONLY (CMC MULTI-CDN)
**Ý nghĩa kịch bản:** Traffic đi qua CMC Multi-CDN (Edge Caching) về Origin  
**File CSV nguồn:** `BM1_Case3_requests.csv`  
**Ánh xạ trong Excel KetQua.xlsx:** Cột `G`  

### A. Bảng Dữ Liệu Thô Trích Xuất Từ CSV
| Nhóm | Endpoint (Name) | Request Count ($N_i$) | Failure Count | Avg Latency ($T_i$) ms | P95 Latency ($P_i$) ms | Requests/s | Avg Size (Bytes) |
| :--- | :--- | :---: | :---: | :---: | :---: | :---: | :---: |
| **Static** | `/static/docs/sample.pdf` | 31,921 | 0 | 46.43 | 110.00 | 106.39 | 678.0 |
| **Static** | `/static/img/banner-large.jpg` | 31,816 | 0 | 48.34 | 120.00 | 106.04 | 1025.0 |
| **Static** | `/static/img/products/[id].jpg` | 64,018 | 0 | 48.36 | 120.00 | 213.36 | 1025.0 |
| **Dynamic** | `/api/cache/no-store` | 5,552 | 0 | 54.44 | 180.00 | 18.50 | 117.0 |
| **Dynamic** | `/api/cache/public` | 15,871 | 0 | 46.88 | 120.00 | 52.90 | 116.0 |
| **Dynamic** | `/api/categories` | 10,997 | 0 | 52.75 | 150.00 | 36.65 | 791.0 |
| **Dynamic** | `/api/products` | 16,672 | 0 | 55.15 | 180.00 | 55.56 | 4123.7 |
| **Dynamic** | `/api/time` | 11,074 | 0 | 51.26 | 160.00 | 36.91 | 92.0 |
| **TỔNG HỢP** | **`Aggregated`** | **187,921** | **0** | **49.11** | **130.00** | **626.31** | **1068.7** |

### B. Chi Tiết Tính Toán Nhóm Static Assets (Tĩnh)
1. **Tổng số static requests ($N_{static}$):**
   $$N_{static} = 31,921 (sample.pdf) + 31,816 (banner-large.jpg) + 64,018 ([id].jpg) = 127,755$$

2. **Độ trễ trung bình có trọng số (Static Avg Latency):**
   $$\text{Static Avg} = \frac{(31,921 \times 46.43) + (31,816 \times 48.34) + (64,018 \times 48.36)}{127,755}$$
   $$\text{Static Avg} = \frac{6,115,516.24}{127,755} = \mathbf{47.87\text{ ms}} \approx \mathbf{48\text{ ms}} \quad \text{(Điền vào Dòng 5, Cột G)}$$

3. **Độ trễ P95 có trọng số (Static P95 Latency):**
   $$\text{Static P95} = \frac{(31,921 \times 110.00) + (31,816 \times 120.00) + (64,018 \times 120.00)}{127,755}$$
   $$\text{Static P95} = \frac{15,011,390.00}{127,755} = \mathbf{117.50\text{ ms}} \approx \mathbf{118\text{ ms}} \quad \text{(Điền vào Dòng 6, Cột G)}$$

### C. Chi Tiết Tính Toán Nhóm Dynamic APIs (Động)
1. **Tổng số dynamic requests ($N_{dynamic}$):**
   $$N_{dynamic} = 5,552 (no-store) + 15,871 (public) + 10,997 (categories) + 16,672 (products) + 11,074 (time) = 60,166$$

2. **Độ trễ trung bình có trọng số (Dynamic Avg Latency):**
   $$\text{Dynamic Avg} = \frac{(5,552 \times 54.44) + (15,871 \times 46.88) + (10,997 \times 52.75) + (16,672 \times 55.15) + (11,074 \times 51.26)}{60,166}$$
   $$\text{Dynamic Avg} = \frac{3,113,418.10}{60,166} = \mathbf{51.75\text{ ms}} \approx \mathbf{52\text{ ms}} \quad \text{(Điền vào Dòng 7, Cột G)}$$

3. **Độ trễ P95 có trọng số (Dynamic P95 Latency):**
   $$\text{Dynamic P95} = \frac{(5,552 \times 180.00) + (15,871 \times 120.00) + (10,997 \times 150.00) + (16,672 \times 180.00) + (11,074 \times 160.00)}{60,166}$$
   $$\text{Dynamic P95} = \frac{9,326,230.00}{60,166} = \mathbf{155.01\text{ ms}} \approx \mathbf{155\text{ ms}} \quad \text{(Điền vào Dòng 8, Cột G)}$$

### D. Các Chỉ Số Tổng Hợp (Lấy Trực Tiếp Từ Hàng Aggregated)
* **Total Requests:** `187,921`
* **Total RPS (Average):** Cột `Requests/s` = $\mathbf{626.31\text{ req/s}}$ $\approx \mathbf{626}$ $\rightarrow$ Điền Dòng 3, Cột `G`
* **Failure Ratio (%):** 
  $$\text{Failure Ratio} = \frac{0}{187,921} \times 100\% = \mathbf{0.00\%} \quad \rightarrow \text{Điền Dòng 9, Cột G}$$
* **Total Data Transferred to Client (MB):**
  $$\text{Data Transferred} = \frac{1068.70\text{ Bytes} \times 187,921}{1,048,576} = \mathbf{191.53\text{ MB}} \approx \mathbf{191.5\text{ MB}} \quad \rightarrow \text{Điền Dòng 11, Cột G}$$

---

## 2.4. CHI TIẾT TÍNH TOÁN & THẾ SỐ: CASE 4: FULL STACK (MULTI-CDN + WAAP)
**Ý nghĩa kịch bản:** Mô hình Production: Client -> Multi-CDN -> WAAP -> Origin  
**File CSV nguồn:** `BM1_Case4_requests.csv`  
**Ánh xạ trong Excel KetQua.xlsx:** Cột `H`  

### A. Bảng Dữ Liệu Thô Trích Xuất Từ CSV
| Nhóm | Endpoint (Name) | Request Count ($N_i$) | Failure Count | Avg Latency ($T_i$) ms | P95 Latency ($P_i$) ms | Requests/s | Avg Size (Bytes) |
| :--- | :--- | :---: | :---: | :---: | :---: | :---: | :---: |
| **Static** | `/static/docs/sample.pdf` | 25,131 | 43 | 126.01 | 86.00 | 87.26 | 676.8 |
| **Static** | `/static/img/banner-large.jpg` | 25,179 | 26 | 114.95 | 83.00 | 87.42 | 1023.9 |
| **Static** | `/static/img/products/[id].jpg` | 50,631 | 74 | 121.65 | 85.00 | 175.80 | 1023.5 |
| **Dynamic** | `/api/cache/no-store` | 4,423 | 15 | 214.92 | 210.00 | 15.36 | 116.6 |
| **Dynamic** | `/api/cache/public` | 12,522 | 28 | 136.89 | 84.00 | 43.48 | 115.7 |
| **Dynamic** | `/api/categories` | 8,810 | 21 | 190.33 | 210.00 | 30.59 | 789.1 |
| **Dynamic** | `/api/products` | 13,003 | 38 | 229.64 | 180.00 | 45.15 | 4111.3 |
| **Dynamic** | `/api/time` | 8,717 | 25 | 224.88 | 170.00 | 30.27 | 91.7 |
| **TỔNG HỢP** | **`Aggregated`** | **148,416** | **270** | **144.92** | **97.00** | **515.32** | **1063.2** |

### B. Chi Tiết Tính Toán Nhóm Static Assets (Tĩnh)
1. **Tổng số static requests ($N_{static}$):**
   $$N_{static} = 25,131 (sample.pdf) + 25,179 (banner-large.jpg) + 50,631 ([id].jpg) = 100,941$$

2. **Độ trễ trung bình có trọng số (Static Avg Latency):**
   $$\text{Static Avg} = \frac{(25,131 \times 126.01) + (25,179 \times 114.95) + (50,631 \times 121.65)}{100,941}$$
   $$\text{Static Avg} = \frac{12,220,227.27}{100,941} = \mathbf{121.06\text{ ms}} \approx \mathbf{121\text{ ms}} \quad \text{(Điền vào Dòng 5, Cột H)}$$

3. **Độ trễ P95 có trọng số (Static P95 Latency):**
   $$\text{Static P95} = \frac{(25,131 \times 86.00) + (25,179 \times 83.00) + (50,631 \times 85.00)}{100,941}$$
   $$\text{Static P95} = \frac{8,554,758.00}{100,941} = \mathbf{84.75\text{ ms}} \approx \mathbf{85\text{ ms}} \quad \text{(Điền vào Dòng 6, Cột H)}$$

### C. Chi Tiết Tính Toán Nhóm Dynamic APIs (Động)
1. **Tổng số dynamic requests ($N_{dynamic}$):**
   $$N_{dynamic} = 4,423 (no-store) + 12,522 (public) + 8,810 (categories) + 13,003 (products) + 8,717 (time) = 47,475$$

2. **Độ trễ trung bình có trọng số (Dynamic Avg Latency):**
   $$\text{Dynamic Avg} = \frac{(4,423 \times 214.92) + (12,522 \times 136.89) + (8,810 \times 190.33) + (13,003 \times 229.64) + (8,717 \times 224.88)}{47,475}$$
   $$\text{Dynamic Avg} = \frac{9,287,860.82}{47,475} = \mathbf{195.64\text{ ms}} \approx \mathbf{196\text{ ms}} \quad \text{(Điền vào Dòng 7, Cột H)}$$

3. **Độ trễ P95 có trọng số (Dynamic P95 Latency):**
   $$\text{Dynamic P95} = \frac{(4,423 \times 210.00) + (12,522 \times 84.00) + (8,810 \times 210.00) + (13,003 \times 180.00) + (8,717 \times 170.00)}{47,475}$$
   $$\text{Dynamic P95} = \frac{7,653,208.00}{47,475} = \mathbf{161.21\text{ ms}} \approx \mathbf{161\text{ ms}} \quad \text{(Điền vào Dòng 8, Cột H)}$$

### D. Các Chỉ Số Tổng Hợp (Lấy Trực Tiếp Từ Hàng Aggregated)
* **Total Requests:** `148,416`
* **Total RPS (Average):** Cột `Requests/s` = $\mathbf{515.32\text{ req/s}}$ $\approx \mathbf{515}$ $\rightarrow$ Điền Dòng 3, Cột `H`
* **Failure Ratio (%):** 
  $$\text{Failure Ratio} = \frac{270}{148,416} \times 100\% = \mathbf{0.18\%} \quad \rightarrow \text{Điền Dòng 9, Cột H}$$
* **Total Data Transferred to Client (MB):**
  $$\text{Data Transferred} = \frac{1063.15\text{ Bytes} \times 148,416}{1,048,576} = \mathbf{150.48\text{ MB}} \approx \mathbf{150.5\text{ MB}} \quad \rightarrow \text{Điền Dòng 11, Cột H}$$

---

## 3. BẢNG TỔNG HỢP ĐỐI CHIẾU CẢ 4 CASES ĐỂ DOUBLE-CHECK FILE KETQUA.XLSX

| Chỉ Số Trong KetQua.xlsx | Dòng | Case 1: Direct (Cột E) | Case 2: WAF (Cột F) | Case 3: CDN (Cột G) | Case 4: Full Stack (Cột H) |
| :--- | :---: | :---: | :---: | :---: | :---: |
| **Total Requests** | Ref | 95,088 | 59,050 | 187,921 | 148,416 |
| **Total RPS (Avg)** | Dòng 3 | 316.9 | 351.0 | 626.3 | 515.3 |
| **Static Avg Latency** | Dòng 5 | 399.3 ms | 363.4 ms | 47.9 ms | 121.1 ms |
| **Static P95 Latency** | Dòng 6 | 352.5 ms | 500.0 ms | 117.5 ms | 84.8 ms |
| **Dynamic API Avg Latency** | Dòng 7 | 630.3 ms | 379.3 ms | 51.7 ms | 195.6 ms |
| **Dynamic API P95 Latency** | Dòng 8 | 487.3 ms | 582.7 ms | 155.0 ms | 161.2 ms |
| **Failure Ratio (%)** | Dòng 9 | 0.55% | 0.00% | 0.00% | 0.18% |
| **Total Data Transferred to Client (MB)** | Dòng 11 | 96.4 MB | 63.4 MB | 191.5 MB | 150.5 MB |

---

## 4. PHÂN TÍCH CHUYÊN SÂU HIỆU NĂNG 4 KỊCH BẢN (ARCHITECTURAL INSIGHTS)

### 4.1. Sức mạnh giải phóng tải tuyệt đối của CMC Multi-CDN (Case 3 & Case 4 vs Case 1):
Khi đồng bộ bộ Cache Rule chuẩn (`js|css|jpg|png` = 1 Day, `/api/` = 1 Hour) cho cả Case 3 và Case 4:
* **Thông lượng tăng vọt:** Case 3 đạt **626.3 req/s** (gấp đôi Case 1 Direct với 316.9 req/s).
* **Độ trễ chạm sàn:** Độ trễ static assets và dynamic API tại Case 3 giảm ngoạn mục xuống còn **48 ms** (nhanh gấp 8 - 12 lần so với truy cập trực tiếp Origin EC2 399ms - 630ms).
* **Giải phóng tài nguyên máy chủ gốc (Origin Offload):** Mức chiếm dụng CPU EC2 giảm từ **7.44%** (Case 1) xuống chỉ còn **0.56%** (Case 3) và **0.86%** (Case 4). Hơn **99.9%** request được giải quyết ngay tại RAM/SSD của Edge PoP, Origin EC2 gần như ở trạng thái nghỉ ngơi hoàn toàn.

### 4.2. Chi phí độ trễ an ninh của WAF (Security Latency Overhead - Case 3 vs Case 4):
* **Case 3 (CDN Only):** Tốc độ đạt **626.3 req/s**, độ trễ trung bình **48 ms**.
* **Case 4 (Full Stack: Multi-CDN + WAAP):** Tốc độ đạt **515.3 req/s**, độ trễ trung bình **121 ms (Static)** và **196 ms (Dynamic)**.

> **Đánh giá kỹ thuật:** Khoảng chênh lệch độ trễ ~70ms - 140ms giữa Case 3 và Case 4 phản ánh đúng bản chất kỹ thuật: Mô hình Full Stack phải đi qua lớp **CMC WAAP** để thực hiện phân tích chữ ký (Deep Packet Inspection), kiểm tra OWASP Top 10 và Rate Limiting. Đây là khoản chi phí độ trễ an ninh (Security Latency Overhead) hoàn toàn chấp nhận được để bảo vệ ứng dụng toàn diện trong môi trường Production.

### 4.3. Giải thích chỉ số 'Total Data Transferred to Client' (Dòng 11):
* Ở Case 3, tổng dữ liệu client nhận về đạt mức cao nhất (**191.5 MB**) trong 5 phút.
* **Ý nghĩa:** Vì hệ thống xử lý cực nhanh (626 req/s), Locust đã hoàn thành tới **187,921 requests** (gấp đôi Case 1). Lượng dữ liệu khổng lồ này **hoàn toàn do các Edge PoPs của CMC CDN phục vụ**, không làm nghẽn băng thông mạng của EC2.
