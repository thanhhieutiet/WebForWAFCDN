import React from 'react';
import {
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  CartesianGrid,
} from 'recharts';
import { formatVND } from '@/lib/formatters';

interface DayRevenueItem {
  date: string;
  revenue: number;
  revenue_vnd?: string;
  order_count: number;
}

interface AdminRevenueChartProps {
  data: DayRevenueItem[];
}

export default function AdminRevenueChart({ data }: AdminRevenueChartProps): React.JSX.Element {
  if (!data || data.length === 0) {
    return (
      <div className="flex h-64 items-center justify-center text-xs text-slate-400">
        Chưa có số liệu doanh thu theo ngày
      </div>
    );
  }

  // Format data for Recharts
  const chartData = data.map((d) => ({
    date: d.date.slice(5), // MM-DD
    revenue: Number(d.revenue) || 0,
    orders: d.order_count || 0,
  }));

  return (
    <div className="h-72 w-full pt-4">
      <ResponsiveContainer width="100%" height="100%">
        <BarChart data={chartData} margin={{ top: 10, right: 10, left: 10, bottom: 20 }}>
          <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#334155" opacity={0.2} />
          <XAxis
            dataKey="date"
            stroke="#64748b"
            fontSize={11}
            tickLine={false}
            axisLine={{ stroke: '#cbd5e1' }}
          />
          <YAxis
            stroke="#64748b"
            fontSize={11}
            tickLine={false}
            axisLine={false}
            tickFormatter={(val) => `${(val / 1000000).toFixed(0)}M`}
          />
          <Tooltip
            formatter={(value: any) => [formatVND(Number(value) || 0), 'Doanh thu']}
            labelFormatter={(label) => `Ngày ${label}`}
            contentStyle={{
              backgroundColor: '#0f172a',
              borderColor: '#1e293b',
              borderRadius: '8px',
              fontSize: '12px',
              color: '#f8fafc',
            }}
          />
          <Bar dataKey="revenue" fill="#0d9488" radius={[4, 4, 0, 0]} />
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}
