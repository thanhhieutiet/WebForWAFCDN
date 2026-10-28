# PHẦN 1: TỔNG QUAN BÀI TEST & MỤC TIÊU ĐO KIỂM

---

## 1.1. Bối Cảnh & Yêu Cầu Từ Mentor
Trong quá trình triển khai và đánh giá bộ giải pháp an ninh mạng và tăng tốc web của **CMC Telecom** (gồm **CMC WAAP** và **CMC Multi-CDN**) bảo vệ ứng dụng thương mại điện tử thử nghiệm (**LabShop**), mentor đã giao bổ sung nhiệm vụ kỹ thuật:

> *"Thêm 1 task bổ sung học cách dựng local server test tải, test host từ các mạng nhé @hieuttm. Gợi ý có thể tham khảo từ Grafana, stress test có thể dùng Locust."*

Báo cáo này tổng hợp toàn bộ kết quả triển khai hạ tầng đo kiểm, số liệu thực nghiệm đo được trên 4 mô hình kiến trúc và 2 môi trường mạng viễn thông thực tế (Mạng cáp quang văn phòng vs Mạng di động 4G).

---

## 1.2. Mục Tiêu Trọng Tâm
1. **Làm chủ bộ công cụ đo kiểm tải và giám sát tập trung (Observability Stack):** Tự động hóa việc triển khai cụm công cụ kiểm thử tải phân tán (Locust) kết hợp với hệ thống thu thập số liệu (Prometheus), trực quan hóa hiệu năng thời gian thực (Grafana Dashboard) và gom log truy cập (Loki & Promtail).
2. **Đo lường định lượng giá trị của CMC Multi-CDN & WAAP:**
   * **Tăng tốc & Giảm độ trễ:** Đánh giá hiệu quả của bộ nhớ đệm tại biên (Edge Caching) khi máy chủ gốc (Origin) đặt ở khoảng cách xa (AWS EC2 khu vực us-east-1 North Virginia, Mỹ).
   * **Bảo vệ & Giảm tải máy chủ gốc (Origin Offload):** Đo lường mức độ giảm tải CPU và lưu lượng card mạng trên EC2 khi có CDN gánh tải.
   * **Chi phí độ trễ an ninh (Security Latency Overhead):** Xác định thời gian trễ tăng thêm khi gói tin phải đi qua bộ lọc kiểm tra sâu chữ ký tấn công của CMC WAAP.
3. **Đánh giá trải nghiệm thực tế từ các môi trường mạng (Multi-Network Benchmark):** So sánh chất lượng dịch vụ (QoE) của người dùng cuối khi truy cập qua đường truyền Internet cáp quang doanh nghiệp (CMC FTTH) và đường truyền di động không dây (4G LTE).

---

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
