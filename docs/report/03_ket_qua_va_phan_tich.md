# PHẦN 3: KẾT QUẢ ĐO KIỂM & PHÂN TÍCH CHUYÊN SÂU

---

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
