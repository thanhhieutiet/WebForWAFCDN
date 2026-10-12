import React from 'react';
import { ArrowRight } from 'lucide-react';

export function TopologyDiagram(): React.JSX.Element {
  const routes = [
    {
      id: 'direct',
      name: 'Tuyến 1: Direct Origin',
      desc: 'Truy cập trực tiếp máy chủ Go trên EC2 (Bỏ qua WAF & CDN để đo baseline hiệu năng và kiểm chứng logic).',
      flow: ['Trình duyệt', 'Origin EC2 (Go :8080)'],
      color: 'border-blue-300 dark:border-blue-900 bg-blue-50/50 dark:bg-blue-950/20 text-blue-700 dark:text-blue-300',
    },
    {
      id: 'waap',
      name: 'Tuyến 2: WAAP Only',
      desc: 'Đi qua CMC WAAP kiểm tra chữ ký SQLi, XSS, Path Traversal, Bot và Rate Limit trước khi forward về Origin.',
      flow: ['Trình duyệt', 'CMC WAAP Edge', 'Origin EC2'],
      color: 'border-rose-300 dark:border-rose-900 bg-rose-50/50 dark:bg-rose-950/20 text-rose-700 dark:text-rose-300',
    },
    {
      id: 'cdn',
      name: 'Tuyến 3: Multi-CDN Only',
      desc: 'Đi qua CMC Multi-CDN để kiểm chứng Cache HIT/MISS, Cache Rule, Purge, Gzip/Brotli và tối ưu tĩnh.',
      flow: ['Trình duyệt', 'CMC Multi-CDN PoP', 'Origin EC2'],
      color: 'border-teal-300 dark:border-teal-900 bg-teal-50/50 dark:bg-teal-950/20 text-teal-700 dark:text-teal-300',
    },
    {
      id: 'full',
      name: 'Tuyến 4: Full Topology (CDN + WAAP)',
      desc: 'Mô hình đầy đủ chuẩn doanh nghiệp: Client → CDN PoP (Cache & DDoS L3/L4) → WAAP (L7 Inspection) → Origin EC2.',
      flow: ['Trình duyệt', 'Multi-CDN', 'CMC WAAP', 'Origin EC2'],
      color: 'border-purple-300 dark:border-purple-900 bg-purple-50/50 dark:bg-purple-950/20 text-purple-700 dark:text-purple-300',
    },
  ];

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {routes.map((r) => (
          <div
            key={r.id}
            className={`rounded-card border p-4 transition-all shadow-sm ${r.color}`}
          >
            <div className="font-bold text-sm text-slate-900 dark:text-white mb-1">
              {r.name}
            </div>
            <p className="text-xs text-slate-600 dark:text-slate-300 mb-3 leading-relaxed">
              {r.desc}
            </p>

            {/* Visual route flow */}
            <div className="flex items-center gap-1.5 flex-wrap text-xs font-mono font-medium">
              {r.flow.map((node, i) => (
                <React.Fragment key={i}>
                  <span className="rounded bg-white/80 dark:bg-slate-900/80 px-2 py-0.5 border border-slate-200 dark:border-slate-800 shadow-xs text-slate-800 dark:text-slate-200">
                    {node}
                  </span>
                  {i < r.flow.length - 1 && (
                    <ArrowRight className="h-3 w-3 text-slate-400 shrink-0" />
                  )}
                </React.Fragment>
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
