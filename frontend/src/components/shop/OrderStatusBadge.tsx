import React from 'react';
import type { OrderStatus } from '@/types';
import { Badge } from '@/components/ui/badge';
import { Clock, CheckCircle2, Truck, PackageCheck, XCircle } from 'lucide-react';

interface OrderStatusBadgeProps {
  status: OrderStatus;
  className?: string;
}

export function OrderStatusBadge({ status, className = '' }: OrderStatusBadgeProps): React.JSX.Element {
  switch (status) {
    case 'PENDING':
      return (
        <span
          className={`inline-flex items-center gap-1 rounded-control bg-amber-50 px-2.5 py-0.5 text-xs font-semibold text-amber-700 ring-1 ring-inset ring-amber-600/20 dark:bg-amber-950/60 dark:text-amber-300 ${className}`}
        >
          <Clock className="h-3 w-3" />
          <span>Chờ thanh toán (PENDING)</span>
        </span>
      );
    case 'PAID':
      return (
        <span
          className={`inline-flex items-center gap-1 rounded-control bg-blue-50 px-2.5 py-0.5 text-xs font-semibold text-blue-700 ring-1 ring-inset ring-blue-600/20 dark:bg-blue-950/60 dark:text-blue-300 ${className}`}
        >
          <CheckCircle2 className="h-3 w-3" />
          <span>Đã thanh toán (PAID)</span>
        </span>
      );
    case 'SHIPPED':
      return (
        <span
          className={`inline-flex items-center gap-1 rounded-control bg-purple-50 px-2.5 py-0.5 text-xs font-semibold text-purple-700 ring-1 ring-inset ring-purple-600/20 dark:bg-purple-950/60 dark:text-purple-300 ${className}`}
        >
          <Truck className="h-3 w-3" />
          <span>Đang giao hàng (SHIPPED)</span>
        </span>
      );
    case 'DELIVERED':
      return (
        <span
          className={`inline-flex items-center gap-1 rounded-control bg-teal-50 px-2.5 py-0.5 text-xs font-semibold text-teal-700 ring-1 ring-inset ring-teal-600/20 dark:bg-teal-950/60 dark:text-teal-300 ${className}`}
        >
          <PackageCheck className="h-3 w-3" />
          <span>Giao thành công (DELIVERED)</span>
        </span>
      );
    case 'CANCELLED':
      return (
        <span
          className={`inline-flex items-center gap-1 rounded-control bg-rose-50 px-2.5 py-0.5 text-xs font-semibold text-rose-700 ring-1 ring-inset ring-rose-600/20 dark:bg-rose-950/60 dark:text-rose-300 ${className}`}
        >
          <XCircle className="h-3 w-3" />
          <span>Đã hủy (CANCELLED)</span>
        </span>
      );
    default:
      return (
        <Badge variant="secondary" className={className}>
          {status}
        </Badge>
      );
  }
}
