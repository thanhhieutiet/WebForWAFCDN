import React, { useState, useRef } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import {
  MapPin,
  CreditCard,
  CheckCircle2,
  ArrowRight,
  AlertTriangle,
  Lock,
} from 'lucide-react';
import { useCart } from '@/lib/cartContext';
import { useAuth } from '@/lib/authContext';
import { api, ApiError } from '@/lib/api';
import { formatVND } from '@/lib/formatters';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { toast } from 'sonner';
import { getErrorMessage } from '@/lib/messages';
import type { Order, PaymentMethod } from '@/types';

export function CheckoutPage(): React.JSX.Element {
  const { cart, refreshCart } = useCart();
  const { user, isAuthenticated } = useAuth();
  const navigate = useNavigate();

  // Form State
  const [recipientName, setRecipientName] = useState(user?.name || '');
  const [recipientPhone, setRecipientPhone] = useState(user?.phone || '0901234567');
  const [shippingAddress, setShippingAddress] = useState(user?.address || '123 Đường Cầu Giấy, Hà Nội');
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>('mock_gateway');

  const [submitting, setSubmitting] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  // Generate unique Idempotency-Key once per checkout session / retry attempt
  const idempotencyKeyRef = useRef<string>(
    `order-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`
  );

  if (!isAuthenticated) {
    return (
      <div className="mx-auto max-w-2xl px-4 py-16 text-center space-y-4">
        <Lock className="mx-auto h-12 w-12 text-slate-400" />
        <h1 className="text-2xl font-bold">Yêu cầu đăng nhập</h1>
        <p className="text-sm text-slate-500">
          Vui lòng đăng nhập để hoàn tất quá trình thanh toán và đặt hàng.
        </p>
        <Link to="/login">
          <Button>Đăng nhập ngay</Button>
        </Link>
      </div>
    );
  }

  const items = cart?.items || [];
  if (items.length === 0 && !submitting) {
    return (
      <div className="mx-auto max-w-2xl px-4 py-16 text-center space-y-4">
        <h1 className="text-2xl font-bold">Giỏ hàng rỗng</h1>
        <p className="text-sm text-slate-500">
          Bạn không có sản phẩm nào trong giỏ để thực hiện thanh toán.
        </p>
        <Link to="/products">
          <Button>Quay lại mua sắm</Button>
        </Link>
      </div>
    );
  }

  const handleSubmitOrder = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!recipientName.trim() || !recipientPhone.trim() || !shippingAddress.trim()) {
      setErrorMsg('Vui lòng điền đầy đủ thông tin giao hàng.');
      return;
    }

    try {
      setSubmitting(true);
      setErrorMsg(null);

      const payload = {
        recipient_name: recipientName.trim(),
        recipient_phone: recipientPhone.trim(),
        shipping_address: shippingAddress.trim(),
        payment_method: paymentMethod,
      };

      // Send order request with Idempotency-Key
      const res = await api.post<Order>('/api/orders', payload, {
        idempotencyKey: idempotencyKeyRef.current,
      });

      const createdOrder = res.data;
      toast.success('Đặt hàng thành công!', {
        description: `Mã đơn hàng: ${createdOrder.id}`,
      });

      await refreshCart();
      navigate(`/orders/${createdOrder.id}`);
    } catch (err: any) {
      const msg = err instanceof ApiError ? err.message : getErrorMessage('unknown_error');
      setErrorMsg(msg);
      toast.error('Đặt hàng thất bại', { description: msg });
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8 space-y-8">
      <div className="border-b border-slate-200 dark:border-slate-800 pb-4">
        <h1 className="text-2xl font-bold tracking-tight text-slate-900 dark:text-white">
          Thanh toán Đơn hàng (Checkout)
        </h1>
        <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
          Bảo mật giao dịch với khóa chống trùng lặp Idempotency-Key: <span className="font-mono text-teal-600 dark:text-teal-400">{idempotencyKeyRef.current}</span>
        </p>
      </div>

      <form onSubmit={handleSubmitOrder} className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Left: 2 Columns with Section 1 (Shipping) & Section 2 (Payment) */}
        <div className="lg:col-span-2 space-y-6">
          {/* Section 1: Thông tin Giao hàng */}
          <section className="rounded-card border border-slate-200 bg-white p-6 shadow-sm dark:border-slate-800 dark:bg-slate-900 space-y-4">
            <div className="flex items-center gap-2 text-sm font-bold text-slate-900 dark:text-white border-b border-slate-100 dark:border-slate-800 pb-3">
              <span className="flex h-6 w-6 items-center justify-center rounded-full bg-teal-600 text-xs text-white">
                1
              </span>
              <MapPin className="h-4 w-4 text-teal-600 dark:text-teal-400" />
              <span>Địa chỉ nhận hàng</span>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-slate-700 dark:text-slate-300">
                  Họ và tên người nhận *
                </label>
                <Input
                  required
                  value={recipientName}
                  onChange={(e) => setRecipientName(e.target.value)}
                  placeholder="Nguyễn Văn A"
                  className="text-xs"
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-slate-700 dark:text-slate-300">
                  Số điện thoại liên hệ *
                </label>
                <Input
                  required
                  value={recipientPhone}
                  onChange={(e) => setRecipientPhone(e.target.value)}
                  placeholder="0912345678"
                  className="text-xs"
                />
              </div>

              <div className="sm:col-span-2 space-y-1.5">
                <label className="text-xs font-semibold text-slate-700 dark:text-slate-300">
                  Địa chỉ chi tiết (Số nhà, đường, quận/huyện, tỉnh/thành) *
                </label>
                <Textarea
                  required
                  value={shippingAddress}
                  onChange={(e) => setShippingAddress(e.target.value)}
                  placeholder="123 Phố Duy Tân, Cầu Giấy, Hà Nội"
                  rows={2}
                  className="text-xs"
                />
              </div>
            </div>
          </section>

          {/* Section 2: Phương thức Thanh toán */}
          <section className="rounded-card border border-slate-200 bg-white p-6 shadow-sm dark:border-slate-800 dark:bg-slate-900 space-y-4">
            <div className="flex items-center gap-2 text-sm font-bold text-slate-900 dark:text-white border-b border-slate-100 dark:border-slate-800 pb-3">
              <span className="flex h-6 w-6 items-center justify-center rounded-full bg-teal-600 text-xs text-white">
                2
              </span>
              <CreditCard className="h-4 w-4 text-teal-600 dark:text-teal-400" />
              <span>Phương thức thanh toán</span>
            </div>

            <div className="space-y-3">
              <label
                className={`flex items-start gap-3 rounded-control border p-4 cursor-pointer transition-colors ${
                  paymentMethod === 'mock_gateway'
                    ? 'border-teal-600 bg-teal-50/40 dark:bg-teal-950/30'
                    : 'border-slate-200 hover:bg-slate-50 dark:border-slate-800 dark:hover:bg-slate-800/50'
                }`}
              >
                <input
                  type="radio"
                  name="payment"
                  checked={paymentMethod === 'mock_gateway'}
                  onChange={() => setPaymentMethod('mock_gateway')}
                  className="mt-0.5 text-teal-600 focus:ring-teal-600"
                />
                <div>
                  <div className="text-xs font-bold text-slate-900 dark:text-white">
                    Cổng thanh toán Trực tuyến (Mock Payment Gateway)
                  </div>
                  <div className="text-[11px] text-slate-500 dark:text-slate-400 mt-0.5">
                    Thử nghiệm luồng thanh toán chuyển trạng thái PENDING → PAID tức thì bằng cổng giả lập Origin.
                  </div>
                </div>
              </label>

              <label
                className={`flex items-start gap-3 rounded-control border p-4 cursor-pointer transition-colors ${
                  paymentMethod === 'cod'
                    ? 'border-teal-600 bg-teal-50/40 dark:bg-teal-950/30'
                    : 'border-slate-200 hover:bg-slate-50 dark:border-slate-800 dark:hover:bg-slate-800/50'
                }`}
              >
                <input
                  type="radio"
                  name="payment"
                  checked={paymentMethod === 'cod'}
                  onChange={() => setPaymentMethod('cod')}
                  className="mt-0.5 text-teal-600 focus:ring-teal-600"
                />
                <div>
                  <div className="text-xs font-bold text-slate-900 dark:text-white">
                    Thanh toán khi nhận hàng (COD - Cash on Delivery)
                  </div>
                  <div className="text-[11px] text-slate-500 dark:text-slate-400 mt-0.5">
                    Thanh toán tiền mặt cho shipper khi đơn hàng được vận chuyển tới nơi.
                  </div>
                </div>
              </label>
            </div>
          </section>
        </div>

        {/* Right: Section 3 (Review & Confirm) */}
        <div className="space-y-6">
          <section className="rounded-card border border-slate-200 bg-white p-6 shadow-sm dark:border-slate-800 dark:bg-slate-900 space-y-6">
            <div className="flex items-center gap-2 text-sm font-bold text-slate-900 dark:text-white border-b border-slate-100 dark:border-slate-800 pb-3">
              <span className="flex h-6 w-6 items-center justify-center rounded-full bg-teal-600 text-xs text-white">
                3
              </span>
              <CheckCircle2 className="h-4 w-4 text-teal-600 dark:text-teal-400" />
              <span>Xác nhận & Đặt hàng</span>
            </div>

            {/* Items Mini List */}
            <div className="max-h-56 overflow-y-auto space-y-3 divide-y divide-slate-100 dark:divide-slate-800 text-xs pr-1">
              {items.map((it) => (
                <div key={it.product_id} className="pt-2 first:pt-0 flex items-center justify-between gap-3">
                  <div className="flex items-center gap-2">
                    <img
                      src={it.image_path || '/static/img/products/p01.jpg'}
                      alt=""
                      className="h-10 w-10 rounded object-cover"
                    />
                    <div>
                      <div className="font-semibold text-slate-900 dark:text-white line-clamp-1">
                        {it.product_name}
                      </div>
                      <div className="text-[11px] text-slate-400">
                        x{it.quantity}
                      </div>
                    </div>
                  </div>
                  <div className="font-mono font-bold text-slate-900 dark:text-white shrink-0">
                    {formatVND(it.price * it.quantity)}
                  </div>
                </div>
              ))}
            </div>

            {/* Financial Summary */}
            <div className="space-y-2 pt-4 border-t border-slate-100 dark:border-slate-800 text-xs">
              <div className="flex justify-between text-slate-600 dark:text-slate-400">
                <span>Tạm tính:</span>
                <span className="font-mono font-medium text-slate-900 dark:text-white">
                  {formatVND(cart?.subtotal || 0)}
                </span>
              </div>

              {(cart?.discount || 0) > 0 && (
                <div className="flex justify-between text-teal-600 dark:text-teal-400 font-medium">
                  <span>Giảm giá ({cart?.coupon_code}):</span>
                  <span className="font-mono">-{formatVND(cart?.discount || 0)}</span>
                </div>
              )}

              <div className="flex justify-between text-slate-600 dark:text-slate-400">
                <span>Phí vận chuyển:</span>
                <span className="font-mono font-medium text-slate-900 dark:text-white">
                  {cart?.shipping_fee === 0 ? 'Miễn phí' : formatVND(cart?.shipping_fee || 0)}
                </span>
              </div>

              <div className="pt-3 border-t border-slate-200 dark:border-slate-800 flex justify-between items-baseline">
                <span className="text-sm font-bold text-slate-900 dark:text-white">Tổng thanh toán:</span>
                <span className="text-xl font-extrabold text-teal-600 dark:text-teal-400 font-mono">
                  {formatVND(cart?.total || 0)}
                </span>
              </div>
            </div>

            {errorMsg && (
              <div className="rounded-control bg-rose-50 p-3 text-xs text-rose-700 dark:bg-rose-950/50 dark:text-rose-300 flex items-start gap-2">
                <AlertTriangle className="h-4 w-4 shrink-0 text-rose-500 mt-0.5" />
                <span>{errorMsg}</span>
              </div>
            )}

            {/* Submit Button locked during request */}
            <Button
              type="submit"
              size="lg"
              loading={submitting}
              disabled={submitting}
              className="w-full gap-2"
            >
              <span>Xác nhận Đặt hàng</span>
              <ArrowRight className="h-4 w-4" />
            </Button>
          </section>
        </div>
      </form>
    </div>
  );
}
