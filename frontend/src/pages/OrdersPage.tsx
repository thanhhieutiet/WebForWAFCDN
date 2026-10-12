import React from 'react';
import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Package, ArrowRight, AlertCircle } from 'lucide-react';
import { api } from '@/lib/api';
import { useAuth } from '@/lib/authContext';
import { formatVND, formatDateTime } from '@/lib/formatters';
import { OrderStatusBadge } from '@/components/shop/OrderStatusBadge';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import type { Order } from '@/types';

export function OrdersPage(): React.JSX.Element {
  const { isAuthenticated } = useAuth();

  const { data, isLoading, isError, error, refetch } = useQuery({
    queryKey: ['my-orders'],
    queryFn: async () => {
      const res = await api.get<{ items: Order[]; total: number }>('/api/orders?page=1&page_size=20');
      return res.data;
    },
    enabled: isAuthenticated,
  });

  if (!isAuthenticated) {
    return (
      <div className="mx-auto max-w-xl px-4 py-16 text-center space-y-4">
        <Package className="mx-auto h-12 w-12 text-slate-400" />
        <h1 className="text-2xl font-bold">Lịch sử Đơn hàng</h1>
        <p className="text-sm text-slate-500">
          Vui lòng đăng nhập để xem các đơn hàng bạn đã đặt.
        </p>
        <Link to="/login">
          <Button>Đăng nhập ngay</Button>
        </Link>
      </div>
    );
  }

  const orders = data?.items || [];

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8 space-y-6">
      <div className="border-b border-slate-200 dark:border-slate-800 pb-4">
        <h1 className="text-2xl font-bold tracking-tight text-slate-900 dark:text-white">
          Đơn hàng của tôi
        </h1>
        <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
          Theo dõi tiến độ vận chuyển và lịch sử thanh toán các đơn hàng
        </p>
      </div>

      {isLoading && (
        <div className="space-y-4">
          <Skeleton className="h-28 w-full rounded-card" />
          <Skeleton className="h-28 w-full rounded-card" />
        </div>
      )}

      {isError && (
        <div className="rounded-card border border-rose-200 bg-rose-50 p-6 text-center text-xs text-rose-700 dark:border-rose-900/50 dark:bg-rose-950/20">
          <AlertCircle className="mx-auto h-6 w-6 text-rose-500 mb-2" />
          <div className="font-semibold">Không thể tải danh sách đơn hàng</div>
          <p className="mt-1">{(error as any)?.message || 'Vui lòng thử lại sau.'}</p>
          <Button size="sm" variant="outline" onClick={() => refetch()} className="mt-3">
            Thử lại
          </Button>
        </div>
      )}

      {!isLoading && !isError && orders.length === 0 && (
        <div className="rounded-card border border-dashed border-slate-300 p-12 text-center dark:border-slate-800">
          <Package className="mx-auto h-10 w-10 text-slate-400 mb-3" />
          <h2 className="text-base font-bold text-slate-900 dark:text-white">
            Bạn chưa có đơn hàng nào
          </h2>
          <p className="text-xs text-slate-500 max-w-sm mx-auto mt-1">
            Các đơn hàng bạn đặt sẽ xuất hiện tại đây cùng trạng thái xử lý chi tiết.
          </p>
          <Link to="/products" className="mt-4 inline-block">
            <Button size="sm">Bắt đầu mua sắm</Button>
          </Link>
        </div>
      )}

      {!isLoading && !isError && orders.length > 0 && (
        <div className="space-y-4">
          {orders.map((ord) => (
            <div
              key={ord.id}
              className="rounded-card border border-slate-200 bg-white p-6 shadow-sm dark:border-slate-800 dark:bg-slate-900 flex flex-col md:flex-row md:items-center justify-between gap-4"
            >
              <div className="space-y-2">
                <div className="flex flex-wrap items-center gap-3">
                  <span className="font-mono text-xs font-bold text-slate-900 dark:text-white">
                    Đơn hàng #{ord.id}
                  </span>
                  <OrderStatusBadge status={ord.status} />
                  <span className="text-[11px] text-slate-400 font-mono">
                    {formatDateTime(ord.created_at)}
                  </span>
                </div>

                <div className="text-xs text-slate-600 dark:text-slate-300">
                  Người nhận: <strong>{ord.shipping_name || ord.recipient_name}</strong> •{' '}
                  {ord.items?.length || 0} sản phẩm
                </div>
              </div>

              <div className="flex items-center justify-between md:justify-end gap-6 pt-3 md:pt-0 border-t md:border-t-0 border-slate-100 dark:border-slate-800">
                <div className="text-right">
                  <div className="text-[11px] text-slate-400">Tổng tiền</div>
                  <div className="text-base font-bold font-mono text-teal-600 dark:text-teal-400">
                    {formatVND(ord.total)}
                  </div>
                </div>

                <Link to={`/orders/${ord.id}`}>
                  <Button variant="outline" size="sm" className="gap-1 text-xs">
                    <span>Chi tiết</span>
                    <ArrowRight className="h-3.5 w-3.5" />
                  </Button>
                </Link>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
