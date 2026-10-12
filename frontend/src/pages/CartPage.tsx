import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import {
  Trash2,
  ArrowRight,
  ShoppingBag,
  Tag,
  X,
  AlertCircle,
} from 'lucide-react';
import { useCart } from '@/lib/cartContext';
import { useAuth } from '@/lib/authContext';
import { formatVND } from '@/lib/formatters';
import { QuantityStepper } from '@/components/shop/QuantityStepper';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

export function CartPage(): React.JSX.Element {
  const { cart, isLoading, updateQuantity, removeItem, applyCoupon, removeCoupon } = useCart();
  const { isAuthenticated } = useAuth();
  const navigate = useNavigate();

  const [couponInput, setCouponInput] = useState('');
  const [applyingCoupon, setApplyingCoupon] = useState(false);
  const [couponError, setCouponError] = useState<string | null>(null);

  const handleApplyCoupon = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!couponInput.trim()) return;

    try {
      setApplyingCoupon(true);
      setCouponError(null);
      await applyCoupon(couponInput.trim().toUpperCase());
      setCouponInput('');
    } catch (err: any) {
      setCouponError(err?.message || 'Mã giảm giá không hợp lệ');
    } finally {
      setApplyingCoupon(false);
    }
  };

  const handleRemoveCoupon = async () => {
    try {
      setCouponError(null);
      await removeCoupon();
    } catch {
      // toast in cartContext
    }
  };

  if (!isAuthenticated) {
    return (
      <div className="mx-auto max-w-7xl px-4 py-16 sm:px-6 lg:px-8 text-center space-y-4">
        <ShoppingBag className="mx-auto h-12 w-12 text-slate-400" />
        <h1 className="text-2xl font-bold tracking-tight text-slate-900 dark:text-white">
          Giỏ hàng của bạn
        </h1>
        <p className="text-sm text-slate-500 max-w-md mx-auto">
          Vui lòng đăng nhập để lưu trữ và quản lý các sản phẩm trong giỏ hàng của bạn.
        </p>
        <div className="pt-2">
          <Link to="/login">
            <Button className="gap-2">
              <span>Đăng nhập ngay</span>
              <ArrowRight className="h-4 w-4" />
            </Button>
          </Link>
        </div>
      </div>
    );
  }

  const items = cart?.items || [];
  const isEmpty = items.length === 0;

  if (isEmpty && !isLoading) {
    return (
      <div className="mx-auto max-w-7xl px-4 py-16 sm:px-6 lg:px-8 text-center space-y-4">
        <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-slate-100 text-slate-400 dark:bg-slate-800">
          <ShoppingBag className="h-8 w-8" />
        </div>
        <h1 className="text-2xl font-bold tracking-tight text-slate-900 dark:text-white">
          Giỏ hàng đang trống
        </h1>
        <p className="text-sm text-slate-500 max-w-md mx-auto">
          Bạn chưa chọn sản phẩm nào. Hãy khám phá danh mục để tìm kiếm các thiết bị phù hợp nhé!
        </p>
        <div className="pt-2">
          <Link to="/products">
            <Button className="gap-2">
              <span>Duyệt sản phẩm ngay</span>
              <ArrowRight className="h-4 w-4" />
            </Button>
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      <div className="mb-6 pb-4 border-b border-slate-200 dark:border-slate-800">
        <h1 className="text-2xl font-bold tracking-tight text-slate-900 dark:text-white">
          Giỏ hàng ({cart?.item_count || 0} sản phẩm)
        </h1>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 items-start">
        {/* Left: Cart Items Table/List (2 Cols) */}
        <div className="lg:col-span-2 space-y-4">
          <div className="rounded-card border border-slate-200 bg-white overflow-hidden shadow-sm dark:border-slate-800 dark:bg-slate-900">
            <div className="divide-y divide-slate-200 dark:divide-slate-800">
              {items.map((item) => {
                const itemTotal = item.price * item.quantity;

                return (
                  <div
                    key={item.product_id}
                    className="flex flex-col sm:flex-row items-start sm:items-center justify-between p-4 gap-4"
                  >
                    <div className="flex items-center gap-4 flex-1">
                      <Link
                        to={`/products/${item.slug}`}
                        className="h-16 w-16 shrink-0 overflow-hidden rounded-control border border-slate-100 bg-slate-50 dark:border-slate-800 dark:bg-slate-800"
                      >
                        <img
                          src={item.image_path || '/static/img/products/p01.jpg'}
                          alt={item.product_name}
                          className="h-full w-full object-cover object-center"
                        />
                      </Link>

                      <div className="space-y-1">
                        <Link
                          to={`/products/${item.slug}`}
                          className="text-sm font-semibold text-slate-900 hover:text-teal-600 dark:text-white dark:hover:text-teal-400 line-clamp-1"
                        >
                          {item.product_name}
                        </Link>
                        <div className="text-xs font-mono text-slate-500">
                          Đơn giá: {formatVND(item.price)}
                        </div>
                      </div>
                    </div>

                    <div className="flex items-center justify-between w-full sm:w-auto gap-4 sm:gap-6 pt-2 sm:pt-0 border-t sm:border-t-0 border-slate-100 dark:border-slate-800">
                      {/* Stepper with optimistic update */}
                      <QuantityStepper
                        value={item.quantity}
                        min={1}
                        max={10}
                        onChange={(qty) => updateQuantity(item.product_id, qty)}
                      />

                      <div className="text-right min-w-[100px]">
                        <div className="text-sm font-bold text-slate-900 dark:text-white">
                          {formatVND(itemTotal)}
                        </div>
                      </div>

                      <button
                        type="button"
                        onClick={() => removeItem(item.product_id)}
                        className="p-1.5 text-slate-400 hover:text-rose-600 transition-colors rounded hover:bg-slate-50 dark:hover:bg-slate-800"
                        title="Xóa sản phẩm"
                        aria-label={`Xóa ${item.product_name}`}
                      >
                        <Trash2 className="h-4 w-4" />
                      </button>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          {/* Quick coupon suggestion badges for lab testing */}
          <div className="p-4 rounded-card border border-slate-200 bg-slate-50/50 dark:border-slate-800 dark:bg-slate-900/50 text-xs space-y-2">
            <span className="font-semibold text-slate-700 dark:text-slate-300">
              Gợi ý mã giảm giá để thử nghiệm Lab:
            </span>
            <div className="flex flex-wrap gap-2">
              <button
                type="button"
                onClick={() => setCouponInput('WELCOME10')}
                className="rounded border border-teal-300 bg-teal-50 px-2 py-0.5 text-teal-800 font-mono text-[11px] hover:bg-teal-100 dark:border-teal-800 dark:bg-teal-950/60 dark:text-teal-300"
              >
                WELCOME10 (10% tối đa 100k)
              </button>
              <button
                type="button"
                onClick={() => setCouponInput('FREESHIP')}
                className="rounded border border-teal-300 bg-teal-50 px-2 py-0.5 text-teal-800 font-mono text-[11px] hover:bg-teal-100 dark:border-teal-800 dark:bg-teal-950/60 dark:text-teal-300"
              >
                FREESHIP (Miễn phí ship)
              </button>
              <button
                type="button"
                onClick={() => setCouponInput('EXPIRED5')}
                className="rounded border border-rose-300 bg-rose-50 px-2 py-0.5 text-rose-800 font-mono text-[11px] hover:bg-rose-100 dark:border-rose-800 dark:bg-rose-950/60 dark:text-rose-300"
              >
                EXPIRED5 (Test mã hết hạn)
              </button>
            </div>
          </div>
        </div>

        {/* Right: Sticky Summary Box (1 Col) */}
        <div className="sticky top-24 space-y-6">
          <div className="rounded-card border border-slate-200 bg-white p-6 shadow-sm dark:border-slate-800 dark:bg-slate-900 space-y-6">
            <h2 className="text-base font-bold text-slate-900 dark:text-white">
              Tóm tắt đơn hàng
            </h2>

            {/* Coupon form */}
            <div className="space-y-2 pt-2 border-t border-slate-100 dark:border-slate-800">
              {cart?.coupon_code ? (
                <div className="flex items-center justify-between rounded-control border border-teal-200 bg-teal-50 p-2.5 text-xs text-teal-900 dark:border-teal-800 dark:bg-teal-950/60 dark:text-teal-200">
                  <div className="flex items-center gap-1.5 font-semibold">
                    <Tag className="h-4 w-4 text-teal-600" />
                    <span>Mã áp dụng: {cart.coupon_code}</span>
                  </div>
                  <button
                    type="button"
                    onClick={handleRemoveCoupon}
                    className="p-1 text-teal-700 hover:text-rose-600 dark:text-teal-300"
                    title="Hủy mã giảm giá"
                  >
                    <X className="h-3.5 w-3.5" />
                  </button>
                </div>
              ) : (
                <form onSubmit={handleApplyCoupon} className="space-y-1.5">
                  <label className="text-xs font-semibold text-slate-700 dark:text-slate-300">
                    Mã giảm giá (Coupon)
                  </label>
                  <div className="flex gap-2">
                    <Input
                      type="text"
                      placeholder="Nhập mã ưu đãi..."
                      value={couponInput}
                      onChange={(e) => setCouponInput(e.target.value.toUpperCase())}
                      className="text-xs uppercase font-mono h-9"
                    />
                    <Button
                      type="submit"
                      size="sm"
                      loading={applyingCoupon}
                      disabled={!couponInput.trim()}
                    >
                      Áp dụng
                    </Button>
                  </div>
                </form>
              )}

              {couponError && (
                <div className="rounded-control bg-rose-50 p-2.5 text-xs text-rose-700 dark:bg-rose-950/50 dark:text-rose-300 flex items-start gap-1.5">
                  <AlertCircle className="h-4 w-4 shrink-0 text-rose-500 mt-0.5" />
                  <span>{couponError}</span>
                </div>
              )}
            </div>

            {/* Price Calculations */}
            <div className="space-y-2.5 pt-4 border-t border-slate-100 dark:border-slate-800 text-xs">
              <div className="flex justify-between text-slate-600 dark:text-slate-400">
                <span>Tạm tính:</span>
                <span className="font-mono font-medium text-slate-900 dark:text-white">
                  {formatVND(cart?.subtotal || 0)}
                </span>
              </div>

              {(cart?.discount || 0) > 0 && (
                <div className="flex justify-between text-teal-600 dark:text-teal-400 font-medium">
                  <span>Giảm giá voucher:</span>
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
                <span className="text-sm font-bold text-slate-900 dark:text-white">Tổng cộng:</span>
                <span className="text-xl font-extrabold text-teal-600 dark:text-teal-400 font-mono">
                  {formatVND(cart?.total || 0)}
                </span>
              </div>
            </div>

            {/* Checkout Button */}
            <Button
              size="lg"
              className="w-full gap-2"
              onClick={() => navigate('/checkout')}
            >
              <span>Tiến hành Thanh toán</span>
              <ArrowRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
