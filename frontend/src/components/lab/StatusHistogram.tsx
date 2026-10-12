import React from 'react';

interface StatusHistogramProps {
  stats: Record<number, number>;
  total: number;
}

export function StatusHistogram({ stats, total }: StatusHistogramProps): React.JSX.Element {
  if (total === 0) return <div className="text-xs text-slate-500">Chưa có dữ liệu thống kê</div>;

  const entries = Object.entries(stats).sort(([a], [b]) => Number(a) - Number(b));

  const getColorClass = (code: number) => {
    if (code >= 200 && code < 300) return 'bg-teal-500 text-teal-700 dark:text-teal-300';
    if (code >= 300 && code < 400) return 'bg-blue-500 text-blue-700 dark:text-blue-300';
    if (code === 403) return 'bg-rose-500 text-rose-700 dark:text-rose-300';
    if (code === 429) return 'bg-orange-500 text-orange-700 dark:text-orange-300';
    if (code >= 400 && code < 500) return 'bg-amber-500 text-amber-700 dark:text-amber-300';
    return 'bg-purple-500 text-purple-700 dark:text-purple-300';
  };

  return (
    <div className="space-y-2">
      <div className="text-xs font-semibold text-slate-700 dark:text-slate-300">
        Phân bố mã phản hồi HTTP ({total} requests)
      </div>
      <div className="flex h-4 w-full overflow-hidden rounded-control bg-slate-100 dark:bg-slate-800">
        {entries.map(([codeStr, count]) => {
          const code = Number(codeStr);
          const percent = (count / total) * 100;
          return (
            <div
              key={code}
              style={{ width: `${percent}%` }}
              className={`h-full ${getColorClass(code).split(' ')[0]} transition-all`}
              title={`HTTP ${code}: ${count} (${percent.toFixed(1)}%)`}
            />
          );
        })}
      </div>
      <div className="flex flex-wrap gap-3 text-xs">
        {entries.map(([codeStr, count]) => {
          const code = Number(codeStr);
          const percent = ((count / total) * 100).toFixed(1);
          return (
            <div key={code} className="flex items-center gap-1.5 font-mono">
              <span className={`h-2.5 w-2.5 rounded-full ${getColorClass(code).split(' ')[0]}`} />
              <span className="font-semibold">{code}:</span>
              <span className="text-slate-600 dark:text-slate-400">
                {count} ({percent}%)
              </span>
            </div>
          );
        })}
      </div>
    </div>
  );
}
