import React from 'react';

interface HeadersTableProps {
  headers: Record<string, string>;
  title?: string;
}

const HIGHLIGHT_HEADERS = new Set([
  'x-origin-instance',
  'x-origin-time',
  'x-request-id',
  'x-cache-status',
  'age',
  'server',
  'via',
  'content-encoding',
  'cache-control',
  'cf-ray',
  'x-waf-status',
]);

export function HeadersTable({ headers, title }: HeadersTableProps): React.JSX.Element {
  const entries = Object.entries(headers).sort(([a], [b]) => a.localeCompare(b));

  if (entries.length === 0) {
    return (
      <div className="py-2 text-xs italic text-slate-400 dark:text-slate-500">
        Không có header nào
      </div>
    );
  }

  return (
    <div className="space-y-1.5">
      {title && <div className="text-xs font-semibold text-slate-700 dark:text-slate-300">{title}</div>}
      <div className="max-h-60 overflow-y-auto rounded-control border border-slate-200 bg-slate-50 text-xs font-mono dark:border-slate-800 dark:bg-slate-950">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="border-b border-slate-200 bg-slate-100 text-slate-600 dark:border-slate-800 dark:bg-slate-900 dark:text-slate-400">
              <th className="py-1 px-3 w-1/3">Header Name</th>
              <th className="py-1 px-3">Value</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-200 dark:divide-slate-800">
            {entries.map(([key, val]) => {
              const isHighlight = HIGHLIGHT_HEADERS.has(key.toLowerCase());
              return (
                <tr
                  key={key}
                  className={`hover:bg-slate-100/50 dark:hover:bg-slate-900/50 ${
                    isHighlight ? 'bg-teal-50/50 font-semibold text-teal-900 dark:bg-teal-950/30 dark:text-teal-200' : 'text-slate-700 dark:text-slate-300'
                  }`}
                >
                  <td className="py-1 px-3 break-all text-slate-500 dark:text-slate-400 select-all">
                    {key}
                  </td>
                  <td className="py-1 px-3 break-all select-all">{val}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
