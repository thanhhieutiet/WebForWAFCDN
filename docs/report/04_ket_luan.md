# PHẦN 4: KẾT LUẬN & ĐÁNH GIÁ TỔNG THỂ

---

## 4.1. Trả Lời Trực Diện Yêu Cầu Của Mentor

| Câu hỏi của Mentor | Kết quả thực tế đạt được |
| :--- | :--- |
| **Đã làm những gì?** | 1. Triển khai trọn vẹn cụm công cụ tạo tải và giám sát thời gian thực trên môi trường máy cục bộ (Docker Stack: Locust Master-Worker, Prometheus, Grafana Telemetry Dashboard và Loki Web SIEM).<br>2. Thực hiện đo kiểm định lượng trên **4 kịch bản kiến trúc** (Direct EC2, WAF Only, CDN Only, Full Stack) ở mức tải 300 CCU.<br>3. Đo kiểm thực tế **trên 2 môi trường mạng viễn thông** (Cáp quang văn phòng CMC FTTH vs Mạng di động 4G LTE Hotspot). |
| **Mục đích để làm gì?** | 1. Xác định năng lực giảm tải (Offload) và tăng tốc độ truy cập của CMC Multi-CDN khi máy chủ ứng dụng đặt xa tại Mỹ (AWS EC2 us-east-1).<br>2. Đo lường chi phí độ trễ an ninh (Security Latency Overhead) khi bật tầng bảo vệ CMC WAAP.<br>3. Đánh giá chất lượng dịch vụ (QoE) của người dùng cuối trên các loại hạ tầng mạng Internet khác nhau. |
| **Đạt được kết quả gì?** | 1. **Hiệu năng vượt bậc nhờ CDN:** Thông lượng tăng từ **317 req/s lên 626 req/s**, độ trễ giảm từ ~400ms–630ms xuống **48 ms** (nhanh gấp 8–13 lần).<br>2. **Bảo vệ máy chủ gốc:** Tỷ lệ Cache Offload đạt tới **99.9%**, giúp giải phóng CPU EC2 từ **7.44% xuống chỉ còn 0.56%**.<br>3. **Đo lường an ninh WAAP:** Xác định mức overhead độ trễ khi qua WAF là khoảng **~70ms – 140ms**, đổi lại khả năng phòng chống toàn diện các cuộc tấn công L7 và quét lỗ hổng.<br>4. **Trải nghiệm mạng di động:** CDN giúp người dùng 4G giảm độ trễ từ **794.8 ms xuống 108.6 ms** (nhanh gấp **7.3 lần**), tăng gấp đôi thông lượng. |

---

## 4.2. Nhận Định Về Bài Toán Stress Test & Giới Hạn Hạ Tầng
* **Điểm gãy của máy chủ gốc (Origin Breaking Point):** Ở kịch bản truy cập trực tiếp (Direct EC2), với 300 CCU, máy chủ EC2 `t3.micro` đã bắt đầu xuất hiện lỗi **0.55%** do nghẽn hàng đợi kết nối. Nếu không có lớp đệm CDN, máy chủ gốc chắc chắn sẽ sập (HTTP 502/504) khi lưu lượng tiếp tục tăng cao.
* **Sức mạnh lá chắn của CMC Multi-CDN:** Khi đặt sau Multi-CDN, dưới cùng mức tải 300 CCU, CPU của EC2 chỉ tiêu thụ **0.56%**. Điều này chứng minh kiến trúc Edge Caching đã **dời hoàn toàn điểm gãy chịu tải từ máy chủ gốc sang hạ tầng phân tán toàn cầu của CMC Telecom**. Ngưỡng chịu tải của hệ thống lúc này phụ thuộc vào quy mô băng thông của các Edge PoP chứ không còn bị giới hạn bởi cấu hình phần cứng của server nội bộ.

---

## 4.3. Bài Học Kinh Nghiệm & Đề Xuất Vận Hành
1. **Tối ưu hóa Cache Rule:** Phải luôn cấu hình quy tắc lưu đệm tường minh cho cả đuôi file tĩnh (`js|css|jpg|png`) và đường dẫn API đọc dữ liệu (`/api/`). Việc để cấu hình mặc định sẽ khiến CDN fallback về header Origin (`max-age=60`), dẫn tới hiện tượng trễ hàng chục giây do Cold Cache khi có đột biến tải.
2. **Quản trị DNS TTL khi chuyển dịch hệ thống:** Do các nhà mạng di động thường lưu cache DNS lâu hơn mạng cáp quang, khi lập kế hoạch chuyển đổi (Cut-over) domain sang CDN/WAF, cần hạ thấp giá trị TTL (xuống 60s – 300s) trước ít nhất 24 giờ để đảm bảo 100% người dùng được điều hướng chuẩn xác.
3. **Ứng dụng cụm Grafana & Locust vào CI/CD:** Bộ công cụ local test tải đã dựng hoàn toàn có thể tái sử dụng để làm Performance Gate trong quy trình release phần mềm định kỳ, giúp phát hiện sớm các điểm nghẽn hiệu năng trước khi đưa mã nguồn lên Production.
