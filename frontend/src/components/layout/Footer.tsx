import React from 'react';
import { Link } from 'react-router-dom';
import { ShieldCheck, ExternalLink, GitBranch } from 'lucide-react';

export function Footer(): React.JSX.Element {
  return (
    <footer className="border-t border-slate-200 bg-white pb-16 pt-8 text-slate-600 dark:border-slate-800 dark:bg-slate-900 dark:text-slate-400">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
          {/* Col 1: Brand & Overview */}
          <div className="space-y-3 md:col-span-1">
            <div className="flex items-center gap-2">
              <div className="flex h-8 w-8 items-center justify-center rounded-control bg-teal-600 text-white">
                <ShieldCheck className="h-5 w-5" />
              </div>
              <span className="font-bold text-base text-slate-900 dark:text-white">LabShop</span>
            </div>
            <p className="text-xs leading-relaxed">
              Môi trường thực hành CMC WAAP & Multi-CDN qua tương tác ứng dụng thương mại điện tử thực chiến.
            </p>
            <div className="flex items-center gap-2 text-xs font-mono text-slate-500">
              <GitBranch className="h-3.5 w-3.5" />
              <span>v1.0.0 (Phase 3 Vite + React)</span>
            </div>
          </div>

          {/* Col 2: Navigation Links */}
          <div className="space-y-2 text-xs">
            <div className="font-semibold text-slate-900 dark:text-white">Điều hướng</div>
            <ul className="space-y-1.5">
              <li>
                <Link to="/products" className="hover:text-teal-600 dark:hover:text-teal-400">
                  Danh mục sản phẩm
                </Link>
              </li>
              <li>
                <Link to="/tutorial" className="hover:text-teal-600 dark:hover:text-teal-400">
                  Kịch bản thực hành (Tutorial)
                </Link>
              </li>
              <li>
                <Link to="/docs" className="hover:text-teal-600 dark:hover:text-teal-400">
                  Tài liệu API (OpenAPI 3.0)
                </Link>
              </li>
              <li>
                <Link to="/cart" className="hover:text-teal-600 dark:hover:text-teal-400">
                  Giỏ hàng & Thanh toán
                </Link>
              </li>
            </ul>
          </div>

          {/* Col 3: 4-Route Topologies */}
          <div className="space-y-2 text-xs font-mono">
            <div className="font-semibold text-slate-900 dark:text-white font-sans">4 Tuyến đường Lab</div>
            <ul className="space-y-1 text-[11px] text-slate-500 dark:text-slate-400">
              <li>• Direct: Client → Origin EC2</li>
              <li>• WAAP: Client → CMC WAAP → Origin</li>
              <li>• CDN: Client → Multi-CDN → Origin</li>
              <li>• Full: Client → CDN → WAAP → Origin</li>
            </ul>
          </div>

          {/* Col 4: CMC Cloud Documentation */}
          <div className="space-y-2 text-xs">
            <div className="font-semibold text-slate-900 dark:text-white">Tài liệu tham khảo</div>
            <ul className="space-y-1.5">
              <li>
                <a
                  href="https://cloud.cmctelecom.vn/docs/cdn"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-1 hover:text-teal-600 dark:hover:text-teal-400"
                >
                  <span>CMC CDN Docs</span>
                  <ExternalLink className="h-3 w-3" />
                </a>
              </li>
              <li>
                <a
                  href="https://cloud.cmctelecom.vn/docs/waap"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-1 hover:text-teal-600 dark:hover:text-teal-400"
                >
                  <span>CMC WAAP Docs</span>
                  <ExternalLink className="h-3 w-3" />
                </a>
              </li>
              <li>
                <span className="text-slate-400">
                  Phím tắt: <kbd className="rounded border border-slate-300 dark:border-slate-700 px-1 py-0.5 font-mono text-[10px]">Ctrl+Shift+L</kbd> mở Lab Console
                </span>
              </li>
            </ul>
          </div>
        </div>

        <div className="mt-8 pt-4 border-t border-slate-200 dark:border-slate-800 text-center text-xs text-slate-500">
          © 2026 LabShop Project. Dành cho mục đích đào tạo và thực nghiệm an toàn mạng.
        </div>
      </div>
    </footer>
  );
}
