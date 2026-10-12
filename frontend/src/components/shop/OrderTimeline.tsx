import React from 'react';
import type { OrderStatus } from '@/types';
import { Check, X, Clock, Truck, PackageCheck, CreditCard } from 'lucide-react';

interface OrderTimelineProps {
  status: OrderStatus;
  createdAt?: string;
  updatedAt?: string;
}

export function OrderTimeline({ status }: OrderTimelineProps): React.JSX.Element {
  if (status === 'CANCELLED') {
    return (
      <div className="flex items-center gap-3 p-4 rounded-control bg-rose-50 text-rose-800 dark:bg-rose-950/40 dark:text-rose-200 border border-rose-200 dark:border-rose-900">
        <div className="flex h-8 w-8 items-center justify-center rounded-full bg-rose-600 text-white shrink-0">
          <X className="h-4 w-4" />
        </div>
        <div className="text-xs">
          <div className="font-bold">Đơn hàng đã bị hủy (CANCELLED)</div>
          <div className="text-rose-600 dark:text-rose-300">Đơn hàng không tiếp tục xử lý vận chuyển.</div>
        </div>
      </div>
    );
  }

  const steps = [
    { key: 'PENDING', label: 'Đặt hàng', icon: Clock },
    { key: 'PAID', label: 'Thanh toán', icon: CreditCard },
    { key: 'SHIPPED', label: 'Vận chuyển', icon: Truck },
    { key: 'DELIVERED', label: 'Hoàn tất', icon: PackageCheck },
  ];

  const getStepState = (stepKey: string) => {
    const sequence = ['PENDING', 'PAID', 'SHIPPED', 'DELIVERED'];
    const currentIndex = sequence.indexOf(status);
    const stepIndex = sequence.indexOf(stepKey);

    if (stepIndex < currentIndex) return 'completed';
    if (stepIndex === currentIndex) return 'current';
    return 'upcoming';
  };

  return (
    <div className="p-4 rounded-control border border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-900">
      <div className="flex items-center justify-between relative">
        {/* Progress bar line behind */}
        <div className="absolute left-6 right-6 top-4 h-0.5 bg-slate-200 dark:bg-slate-800 -z-0" />

        {steps.map((step) => {
          const state = getStepState(step.key);
          const Icon = step.icon;

          return (
            <div key={step.key} className="flex flex-col items-center z-10 text-center flex-1">
              <div
                className={`flex h-8 w-8 items-center justify-center rounded-full border-2 transition-colors ${
                  state === 'completed'
                    ? 'border-teal-600 bg-teal-600 text-white'
                    : state === 'current'
                    ? 'border-teal-600 bg-white text-teal-600 dark:bg-slate-900'
                    : 'border-slate-300 bg-white text-slate-400 dark:border-slate-700 dark:bg-slate-900'
                }`}
              >
                {state === 'completed' ? <Check className="h-4 w-4" /> : <Icon className="h-4 w-4" />}
              </div>
              <span
                className={`mt-2 text-[11px] font-semibold ${
                  state === 'completed' || state === 'current'
                    ? 'text-slate-900 dark:text-white'
                    : 'text-slate-400 dark:text-slate-500'
                }`}
              >
                {step.label}
              </span>
            </div>
          );
        })}
      </div>
    </div>
  );
}
