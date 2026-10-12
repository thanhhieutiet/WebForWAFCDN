// Error code to Vietnamese user-friendly message mapping
// Do not expose raw server stack traces or uncensored messages to users

export const ERROR_MESSAGES: Record<string, string> = {
  // Common HTTP & Domain Errors
  internal_server_error: 'Hệ thống đang gặp sự cố tạm thời. Vui lòng thử lại sau giây lát.',
  validation_error: 'Thông tin nhập vào chưa hợp lệ. Vui lòng kiểm tra lại các trường được đánh dấu.',
  unauthorized: 'Phiên đăng nhập đã hết hạn hoặc bạn chưa đăng nhập. Vui lòng đăng nhập lại.',
  forbidden: 'Bạn không có quyền thực hiện hành động này.',
  not_found: 'Tài nguyên yêu cầu không tồn tại hoặc đã bị xóa.',
  conflict: 'Dữ liệu đã tồn tại hoặc xảy ra xung đột phiên bản.',

  // Auth
  invalid_credentials: 'Email hoặc mật khẩu không chính xác. Vui lòng thử lại.',
  email_already_exists: 'Địa chỉ email này đã được sử dụng. Vui lòng chọn email khác hoặc đăng nhập.',
  user_not_found: 'Không tìm thấy thông tin tài khoản.',
  password_too_weak: 'Mật khẩu chưa đủ mạnh. Vui lòng nhập tối thiểu 8 ký tự kèm chữ hoa, số và ký tự đặc biệt.',

  // Cart & Coupon
  coupon_not_found: 'Mã giảm giá không tồn tại. Vui lòng kiểm tra lại.',
  coupon_expired: 'Mã giảm giá đã hết hạn sử dụng.',
  coupon_not_active: 'Mã giảm giá chưa đến thời gian áp dụng.',
  coupon_limit_reached: 'Mã giảm giá đã hết lượt sử dụng.',
  coupon_min_subtotal: 'Giá trị đơn hàng chưa đạt mức tối thiểu để áp dụng mã giảm giá này.',
  invalid_coupon: 'Mã giảm giá không hợp lệ hoặc không áp dụng được cho đơn hàng này.',
  invalid_coupon_type: 'Loại mã giảm giá không được hỗ trợ.',
  empty_cart: 'Giỏ hàng của bạn đang trống.',
  product_out_of_stock: 'Sản phẩm đã hết hàng hoặc không đủ số lượng trong kho.',

  // Order & Payment
  idempotency_conflict: 'Yêu cầu đặt hàng đã được xử lý hoặc trùng lặp với nội dung khác.',
  order_not_found: 'Không tìm thấy đơn hàng yêu cầu.',
  invalid_order_status: 'Trạng thái đơn hàng không hợp lệ cho thao tác này.',
  payment_failed: 'Giao dịch thanh toán chưa thành công. Vui lòng kiểm tra phương thức hoặc thử lại.',
  order_already_paid: 'Đơn hàng này đã được thanh toán.',
  order_cancelled: 'Đơn hàng này đã bị hủy.',

  // Reviews
  no_delivered_order: 'Bạn chỉ có thể đánh giá sản phẩm sau khi đã nhận hàng thành công.',
  review_already_exists: 'Bạn đã đánh giá sản phẩm này rồi.',

  // Support
  ticket_not_found: 'Không tìm thấy yêu cầu hỗ trợ.',
  ticket_closed: 'Yêu cầu hỗ trợ đã đóng, không thể gửi thêm phản hồi.',
  file_too_large: 'Kích thước tệp đính kèm vượt quá giới hạn cho phép (tối đa 10MB).',
  invalid_file_type: 'Định dạng tệp không được hỗ trợ. Chỉ chấp nhận ảnh, PDF hoặc văn bản.',

  // Edge / WAAP
  edge_blocked: 'Yêu cầu bị chặn bởi tường lửa ứng dụng web (WAAP/Edge Rule).',
  rate_limit_exceeded: 'Bạn đã gửi quá nhiều yêu cầu trong thời gian ngắn. Vui lòng chờ giây lát.',
  upstream_timeout: 'Hệ thống máy chủ xử lý phản hồi quá lâu (Timeout). Vui lòng thử lại sau.',
};

/**
 * Returns a clean Vietnamese error message for the provided error code.
 */
export function getErrorMessage(code?: string, fallback = 'Đã có lỗi xảy ra. Vui lòng thử lại.'): string {
  if (!code) return fallback;
  return ERROR_MESSAGES[code] || fallback;
}
