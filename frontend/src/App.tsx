import React, { useState } from 'react';
import {
  ShieldCheck,
  Terminal,
  ShoppingBag,
  Zap,
  Globe,
  Radio,
  Server,
  Database,
  AlertTriangle,
  Flame,
  ChevronUp,
  ChevronDown,
} from 'lucide-react';
import { Toaster, toast } from 'sonner';

export default function App(): React.JSX.Element {
  const [consoleOpen, setConsoleOpen] = useState(false);

  const triggerToast = () => {
    toast.success('LabShop hệ thống sẵn sàng', {
      description: 'Phase 0 scaffold thành công với Vite + React + Tailwind CSS',
    });
  };

  return (
    <div className="flex min-h-screen flex-col bg-slate-50 text-slate-900 dark:bg-slate-950 dark:text-slate-100 font-sans">
      <Toaster position="top-right" richColors />

      {/* Header */}
      <header className="sticky top-0 z-40 border-b border-slate-200 bg-white/95 backdrop-blur dark:border-slate-800 dark:bg-slate-900/95">
        <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-control bg-teal-600 text-white shadow-sm">
              <ShieldCheck className="h-6 w-6" aria-hidden="true" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="text-xl font-bold tracking-tight text-slate-900 dark:text-white">
                  LabShop
                </span>
                <span className="inline-flex items-center gap-1 rounded-control bg-teal-50 px-2.5 py-0.5 text-xs font-semibold text-teal-700 ring-1 ring-inset ring-teal-600/20 dark:bg-teal-950/60 dark:text-teal-300 dark:ring-teal-500/30">
                  <span className="h-1.5 w-1.5 rounded-full bg-teal-600 dark:bg-teal-400" />
                  WAAP & CDN Lab
                </span>
              </div>
              <p className="text-xs text-slate-500 dark:text-slate-400">
                Học WAAP & CDN qua trải nghiệm thực chiến
              </p>
            </div>
          </div>

          <div className="flex items-center gap-3">
            <button
              type="button"
              onClick={triggerToast}
              className="inline-flex items-center gap-1.5 rounded-control bg-teal-600 px-3.5 py-2 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-teal-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-teal-600 focus-visible:ring-offset-2"
            >
              <Zap className="h-4 w-4" />
              <span>Kiểm tra trạng thái</span>
            </button>

            <button
              type="button"
              onClick={() => setConsoleOpen((prev) => !prev)}
              aria-label="Toggle Lab Console"
              className="inline-flex items-center gap-1.5 rounded-control border border-slate-300 bg-white px-3 py-2 text-sm font-medium text-slate-700 shadow-sm transition-colors hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-teal-600 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-200 dark:hover:bg-slate-700"
            >
              <Terminal className="h-4 w-4 text-teal-600 dark:text-teal-400" />
              <span className="hidden sm:inline">Lab Console</span>
              <kbd className="ml-1 hidden rounded border border-slate-200 bg-slate-100 px-1.5 py-0.5 font-mono text-[10px] text-slate-600 sm:inline-block dark:border-slate-600 dark:bg-slate-700 dark:text-slate-300">
                Ctrl+Shift+L
              </kbd>
            </button>
          </div>
        </div>
      </header>

      {/* Main Content / Storefront Placeholder */}
      <main className="flex-1">
        <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
          {/* Hero Banner */}
          <section className="mb-8 rounded-card border border-slate-200 bg-white p-6 shadow-sm dark:border-slate-800 dark:bg-slate-900">
            <div className="flex flex-col items-start justify-between gap-4 md:flex-row md:items-center">
              <div>
                <h1 className="text-2xl font-bold tracking-tight text-slate-900 sm:text-3xl dark:text-white">
                  Môi trường phòng lab An ninh & Tăng tốc Ứng dụng
                </h1>
                <p className="mt-2 text-sm text-slate-600 dark:text-slate-300">
                  Hai phân khu song song: <strong className="text-teal-600 dark:text-teal-400">Shop</strong> (thương mại điện tử đầy đủ) và{' '}
                  <strong className="text-teal-600 dark:text-teal-400">Lab Console & Scenario Runner</strong> (quan sát HTTP, WAF, CDN cache).
                </p>
              </div>
              <div className="flex flex-wrap gap-2 text-xs font-mono">
                <span className="inline-flex items-center gap-1.5 rounded-control bg-blue-50 px-2.5 py-1 text-blue-700 dark:bg-blue-950/50 dark:text-blue-300">
                  <Server className="h-3.5 w-3.5" /> Origin (Blue)
                </span>
                <span className="inline-flex items-center gap-1.5 rounded-control bg-teal-50 px-2.5 py-1 text-teal-700 dark:bg-teal-950/50 dark:text-teal-300">
                  <Database className="h-3.5 w-3.5" /> HIT (Teal)
                </span>
                <span className="inline-flex items-center gap-1.5 rounded-control bg-amber-50 px-2.5 py-1 text-amber-700 dark:bg-amber-950/50 dark:text-amber-300">
                  <Radio className="h-3.5 w-3.5" /> MISS (Amber)
                </span>
                <span className="inline-flex items-center gap-1.5 rounded-control bg-rose-50 px-2.5 py-1 text-rose-700 dark:bg-rose-950/50 dark:text-rose-300">
                  <AlertTriangle className="h-3.5 w-3.5" /> Blocked (Rose)
                </span>
                <span className="inline-flex items-center gap-1.5 rounded-control bg-orange-50 px-2.5 py-1 text-orange-700 dark:bg-orange-950/50 dark:text-orange-300">
                  <Flame className="h-3.5 w-3.5" /> Stale (Orange)
                </span>
              </div>
            </div>
          </section>

          {/* Storefront Placeholder Grid */}
          <section className="mb-8">
            <div className="mb-4 flex items-center justify-between">
              <div className="flex items-center gap-2">
                <ShoppingBag className="h-5 w-5 text-teal-600" />
                <h2 className="text-lg font-semibold text-slate-900 dark:text-white">
                  Khu Vực Cửa Hàng (Storefront Placeholder)
                </h2>
              </div>
              <span className="text-xs text-slate-500">Phase 0 Skeleton</span>
            </div>

            <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
              {[
                {
                  id: 1,
                  title: 'Module Catalog & Sản phẩm',
                  desc: 'Danh mục linh kiện thiết bị mạng, lọc theo URL search params, ảnh tối ưu, phân trang.',
                  status: 'Ready for Phase 1',
                },
                {
                  id: 2,
                  title: 'Module Giỏ hàng & Thanh toán',
                  desc: 'Giỏ hàng optimistic updates, Idempotency-Key checkout, timeline đơn hàng 5 trạng thái.',
                  status: 'Ready for Phase 2',
                },
                {
                  id: 3,
                  title: 'Module Lab & Kịch bản WAF/CDN',
                  desc: '20+ kịch bản thực hành: SQLi, XSS, Rate limit, Cache rules, Stale-while-revalidate.',
                  status: 'Ready for Phase 3',
                },
              ].map((card) => (
                <div
                  key={card.id}
                  className="flex flex-col justify-between rounded-card border border-slate-200 bg-white p-5 shadow-sm transition-all hover:border-teal-500/50 hover:shadow-md dark:border-slate-800 dark:bg-slate-900"
                >
                  <div>
                    <span className="inline-block rounded-control bg-slate-100 px-2 py-0.5 text-xs font-medium text-slate-700 dark:bg-slate-800 dark:text-slate-300">
                      Phase 0 Scaffolding
                    </span>
                    <h3 className="mt-3 text-base font-semibold text-slate-900 dark:text-white">
                      {card.title}
                    </h3>
                    <p className="mt-2 text-sm text-slate-600 dark:text-slate-400">
                      {card.desc}
                    </p>
                  </div>
                  <div className="mt-5 border-t border-slate-100 pt-3 dark:border-slate-800">
                    <span className="text-xs font-mono font-medium text-teal-600 dark:text-teal-400">
                      {card.status}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </section>
        </div>
      </main>

      {/* Lab Console Placeholder Drawer */}
      <footer
        className={`sticky bottom-0 z-30 border-t border-slate-200 bg-white shadow-lg transition-all duration-200 dark:border-slate-800 dark:bg-slate-900 ${
          consoleOpen ? 'h-64' : 'h-12'
        }`}
      >
        <div className="mx-auto flex h-12 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
          <div className="flex items-center gap-3">
            <button
              type="button"
              onClick={() => setConsoleOpen((prev) => !prev)}
              className="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-slate-700 hover:text-teal-600 dark:text-slate-300 dark:hover:text-teal-400"
            >
              <Terminal className="h-4 w-4 text-teal-600" />
              <span>Lab Console</span>
              {consoleOpen ? (
                <ChevronDown className="h-4 w-4" />
              ) : (
                <ChevronUp className="h-4 w-4" />
              )}
            </button>
            <span className="rounded bg-slate-100 px-1.5 py-0.5 font-mono text-[11px] text-slate-600 dark:bg-slate-800 dark:text-slate-400">
              0 requests recorded
            </span>
          </div>

          <div className="flex items-center gap-4 text-xs font-mono text-slate-500">
            <span className="flex items-center gap-1">
              <Globe className="h-3.5 w-3.5 text-teal-600" /> Direct Mode (Localhost)
            </span>
            <span className="hidden sm:inline">Ctrl+Shift+L to toggle</span>
          </div>
        </div>

        {consoleOpen && (
          <div className="h-[calc(100%-3rem)] border-t border-slate-100 bg-slate-950 p-4 font-mono text-xs text-slate-300 dark:border-slate-800 overflow-y-auto">
            <div className="flex items-center justify-between pb-2 text-slate-400 border-b border-slate-800">
              <span>Method | Path | Status | Origin/Edge | Cache | Latency</span>
              <span>History (Max 50)</span>
            </div>
            <div className="py-4 text-center text-slate-500">
              Chưa có request HTTP nào được ghi nhận. Các request từ fetch wrapper sẽ hiển thị tại đây kèm chi tiết headers, cURL, và phán đoán WAF/CDN.
            </div>
          </div>
        )}
      </footer>
    </div>
  );
}
