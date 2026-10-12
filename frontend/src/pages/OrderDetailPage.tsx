import React, { useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  ArrowLeft,
  CreditCard,
  Ban,
  MapPin,
  AlertTriangle,
  Receipt,
} from 'lucide-react';
import { api, ApiError } from '@/lib/api';
import { formatVND, formatDateTime } from '@/lib/formatters';
import { OrderStatusBadge } from '@/components/shop/OrderStatusBadge';
import { OrderTimeline } from '@/components/shop/OrderTimeline';
import { Button } from '@/components/ui/button';
import { Dialog } from '@/components/ui/dialog';
import { Skeleton } from '@/components/ui/skeleton';
import { toast } from 'sonner';
import { getErrorMessage } from '@/lib/messages';
import type { Order } from '@/types';

export function OrderDetailPage(): React.JSX.Element {
  const { id } = useParams<{ id: string }>();
  const queryClient = useQueryClient();

  const [cancelModalOpen, setCancelModalOpen] = useState(false);

  // Fetch Order
  const { data: order, isLoading, isError, error } = useQuery({
    queryKey: ['order-detail', id],
    queryFn: async () => {
      const res = await api.get<Order>(`/api/orders/${id}`);
      return res.data;
    },
    enabled: !!id,
  });

  // Pay mutation
  const payMutation = useMutation({
    mutationFn: async () => {
      const res = await api.post<Order>(`/api/orders/${id}/pay`);
      return res.data;
    },
    onSuccess: (updated) => {
      toast.success('Thanh toán đơn hàng thành công!', {
        description: 'Trạng thái chuyển sang PAID.',
      });
      queryClient.setQueryData(['order-detail', id], updated);
      queryClient.invalidateQueries({ queryKey: ['my-orders'] });
    },
    onError: (err: any) => {
      const msg = err instanceof ApiError ? err.message : getErrorMessage('payment_failed');
      toast.error('Thanh toán thất bại', { description: msg });
    },
  });

  // Cancel mutation
  const cancelMutation = useMutation({
    mutationFn: async () => {
      const res = await api.post<Order>(`/api/orders/${id}/cancel`);
      return res.data;
    },
    onSuccess: (updated) => {
      toast.info('Đơn hàng đã được hủy.');
      setCancelModalOpen(false);
      queryClient.setQueryData(['order-detail', id], updated);
      queryClient.invalidateQueries({ queryKey: ['my-orders'] });
    },
    onError: (err: any) => {
      const msg = err instanceof ApiError ? err.message : getErrorMessage('unknown_error');
      toast.error('Không thể hủy đơn hàng', { description: msg });
    },
  });

  if (isLoading) {
    return (
      <div className="mx-auto max-w-4xl px-4 py-8 space-y-6">
        <Skeleton className="h-6 w-32" />
        <Skeleton className="h-28 w-full rounded-card" />
        <Skeleton className="h-48 w-full rounded-card" />
      </div>
    );
  }

  if (isError || !order) {
    return (
      <div className="mx-auto max-w-xl px-4 py-16 text-center space-y-4">
        <AlertTriangle className="mx-auto h-12 w-12 text-rose-500" />
        <h2 className="text-xl font-bold">Không tìm thấy thông tin đơn hàng</h2>
        <p className="text-sm text-slate-500">{(error as any)?.message || 'Đơn hàng không tồn tại.'}</p>
        <Link to="/orders">
          <Button variant="outline">Quay lại danh sách đơn</Button>
        </Link>
      </div>
    );
  }

  const isPending = order.status === 'PENDING';
  const canCancel = order.status === 'PENDING' || order.status === 'PAID';

  return (
    <div className="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8 space-y-8">
      {/* Back button */}
      <Link
        to="/orders"
        className="inline-flex items-center gap-1.5 text-xs font-semibold text-slate-600 hover:text-teal-600 dark:text-slate-400 dark:hover:text-teal-400"
      >
        <ArrowLeft className="h-4 w-4" />
        <span>Quay lại Đơn hàng của tôi</span>
      </Link>

      {/* Top Header Card */}
      <div className="rounded-card border border-slate-200 bg-white p-6 shadow-sm dark:border-slate-800 dark:bg-slate-900 space-y-4">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-slate-100 dark:border-slate-800">
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-xl font-bold text-slate-900 dark:text-white font-mono">
                Đơn hàng #{order.id}
              </h1>
              <OrderStatusBadge status={order.status} />
            </div>
            <div className="text-xs text-slate-500 mt-1">
              Thời gian tạo: <span className="font-mono">{formatDateTime(order.created_at)}</span>
            </div>
          </div>

          {/* Quick Pay or Cancel Buttons */}
          <div className="flex items-center gap-2">
            {isPending && order.payment_method === 'mock_gateway' && (
              <Button
                size="sm"
                loading={payMutation.isPending}
                onClick={() => payMutation.mutate()}
                className="gap-1.5"
              >
                <CreditCard className="h-4 w-4" />
                <span>Thanh toán ngay</span>
              </Button>
            )}

            {canCancel && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => setCancelModalOpen(true)}
                className="text-rose-600 hover:bg-rose-50 dark:text-rose-400 dark:hover:bg-rose-950/50"
              >
                <Ban className="h-4 w-4 mr-1" />
                <span>Hủy đơn</span>
              </Button>
            )}
          </div>
        </div>

        {/* Timeline Status */}
        <div className="pt-2">
          <OrderTimeline status={order.status} />
        </div>
      </div>

      {/* Shipping & Payment Details */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="rounded-card border border-slate-200 bg-white p-6 shadow-sm dark:border-slate-800 dark:bg-slate-900 space-y-3">
          <div className="flex items-center gap-2 text-xs font-bold text-slate-900 dark:text-white uppercase tracking-wider">
            <MapPin className="h-4 w-4 text-teal-600 dark:text-teal-400" />
            <span>Địa chỉ nhận hàng</span>
          </div>
          <div className="text-xs space-y-1 text-slate-600 dark:text-slate-300">
            <div className="font-bold text-slate-900 dark:text-white">
              {order.shipping_name || order.recipient_name}
            </div>
            <div>SĐT: {order.shipping_phone || order.recipient_phone}</div>
            <div>{order.shipping_address}</div>
          </div>
        </div>

        <div className="rounded-card border border-slate-200 bg-white p-6 shadow-sm dark:border-slate-800 dark:bg-slate-900 space-y-3">
          <div className="flex items-center gap-2 text-xs font-bold text-slate-900 dark:text-white uppercase tracking-wider">
            <Receipt className="h-4 w-4 text-teal-600 dark:text-teal-400" />
            <span>Phương thức & Trạng thái</span>
          </div>
          <div className="text-xs space-y-1 text-slate-600 dark:text-slate-300">
            <div>
              Hình thức:{' '}
              <strong>
                {order.payment_method === 'mock_gateway' ? 'Mock Payment Gateway' : 'COD (Tiền mặt)'}
              </strong>
            </div>
            <div>
              Trạng thái thanh toán:{' '}
              <strong className={order.status === 'PAID' || order.status === 'DELIVERED' ? 'text-teal-600' : 'text-amber-600'}>
                {order.status === 'PENDING' ? 'Chưa thanh toán' : 'Đã thanh toán'}
              </strong>
            </div>
            {order.idempotency_key && (
              <div className="text-[11px] font-mono text-slate-400 truncate">
                Idempotency-Key: {order.idempotency_key}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Items Summary Table */}
      <div className="rounded-card border border-slate-200 bg-white p-6 shadow-sm dark:border-slate-800 dark:bg-slate-900 space-y-4">
        <h2 className="text-sm font-bold text-slate-900 dark:text-white">
          Chi tiết sản phẩm ({order.items?.length || 0})
        </h2>

        <div className="divide-y divide-slate-100 dark:divide-slate-800 text-xs">
          {order.items?.map((it, idx) => (
            <div key={idx} className="py-3 flex items-center justify-between gap-4">
              <div className="flex items-center gap-3">
                <img
                  src={it.image_path || '/static/img/products/p01.jpg'}
                  alt=""
                  className="h-12 w-12 rounded object-cover"
                />
                <div>
                  <div className="font-semibold text-slate-900 dark:text-white">
                    {it.product_name}
                  </div>
                  <div className="text-slate-400">
                    {formatVND(it.price ?? it.product_price ?? 0)} x {it.quantity}
                  </div>
                </div>
              </div>

              <div className="font-mono font-bold text-slate-900 dark:text-white">
                {formatVND((it.price ?? it.product_price ?? 0) * it.quantity)}
              </div>
            </div>
          ))}
        </div>

        {/* Pricing Subtotal, Discount, Shipping, Total */}
        <div className="pt-4 border-t border-slate-200 dark:border-slate-800 space-y-2 text-xs">
          <div className="flex justify-between text-slate-500">
            <span>Tạm tính:</span>
            <span className="font-mono text-slate-800 dark:text-slate-200">
              {formatVND(order.subtotal)}
            </span>
          </div>
          {order.discount > 0 && (
            <div className="flex justify-between text-teal-600 font-medium">
              <span>Giảm giá ({order.coupon_code || 'Voucher'}):</span>
              <span className="font-mono">-{formatVND(order.discount)}</span>
            </div>
          )}
          <div className="flex justify-between text-slate-500">
            <span>Phí vận chuyển:</span>
            <span className="font-mono text-slate-800 dark:text-slate-200">
              {order.shipping_fee === 0 ? 'Miễn phí' : formatVND(order.shipping_fee)}
            </span>
          </div>
          <div className="pt-2 border-t border-slate-100 dark:border-slate-800 flex justify-between items-baseline text-sm font-bold text-slate-900 dark:text-white">
            <span>Tổng thanh toán:</span>
            <span className="text-lg font-extrabold font-mono text-teal-600 dark:text-teal-400">
              {formatVND(order.total)}
            </span>
          </div>
        </div>
      </div>

      {/* Cancel Order Confirmation Modal */}
      <Dialog
        open={cancelModalOpen}
        onOpenChange={setCancelModalOpen}
        title="Xác nhận hủy đơn hàng"
        description={`Bạn có chắc chắn muốn hủy đơn hàng #${order.id}? Thao tác này không thể hoàn tác.`}
      >
        <div className="flex justify-end gap-2 pt-4">
          <Button variant="outline" onClick={() => setCancelModalOpen(false)}>
            Quay lại
          </Button>
          <Button
            variant="destructive"
            loading={cancelMutation.isPending}
            onClick={() => cancelMutation.mutate()}
          >
            Xác nhận Hủy đơn
          </Button>
        </div>
      </Dialog>
    </div>
  );
}
