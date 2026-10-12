import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import {
  BookOpen,
  ArrowRight,
  Database,
  ShieldCheck,
  ShoppingBag,
  Terminal,
} from 'lucide-react';
import { SCENARIOS } from '@/data/scenariosData';
import { TopologyDiagram } from '@/components/lab/TopologyDiagram';
import type { ScenarioGroup } from '@/types';

export function TutorialPage(): React.JSX.Element {
  const [selectedGroup, setSelectedGroup] = useState<ScenarioGroup | 'all'>('all');

  const filteredScenarios = SCENARIOS.filter((s) => {
    if (selectedGroup === 'all') return true;
    return s.group === selectedGroup;
  });

  const cdnCount = SCENARIOS.filter((s) => s.group === 'cdn').length;
  const waapCount = SCENARIOS.filter((s) => s.group === 'waap').length;
  const shopCount = SCENARIOS.filter((s) => s.group === 'shop').length;

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8 space-y-12">
      {/* Header */}
      <div className="space-y-3">
        <div className="inline-flex items-center gap-2 rounded-control bg-teal-50 px-3 py-1 text-xs font-semibold text-teal-700 ring-1 ring-inset ring-teal-600/20 dark:bg-teal-950/60 dark:text-teal-300">
          <BookOpen className="h-3.5 w-3.5" />
          <span>Tài liệu & Kịch bản Thực hành CMC WAAP & Multi-CDN</span>
        </div>

        <h1 className="text-3xl font-bold tracking-tight text-slate-900 dark:text-white">
          Hướng dẫn Thực hành (Tutorial & Scenarios)
        </h1>
        <p className="text-sm text-slate-600 dark:text-slate-300 max-w-3xl leading-relaxed">
          Tập hợp các bài thực hành thực chiến được thiết kế bám sát giao diện cấu hình thực tế trên
          CMC Cloud Portal. Mọi kịch bản đều có kịch bản từng bước, mẫu kiểm thử, kết quả kỳ vọng
          và bộ công cụ Runner tích hợp ngay trên trình duyệt.
        </p>
      </div>

      {/* 4-Route Topologies Section */}
      <section className="space-y-4">
        <div>
          <h2 className="text-xl font-bold text-slate-900 dark:text-white">
            Kiến trúc 4 Tuyến đường Mạng trong Phòng Lab
          </h2>
          <p className="text-xs text-slate-500 mt-1">
            Hiểu rõ luồng di chuyển của gói tin HTTP giữa Client, Edge CDN/WAAP và Origin
          </p>
        </div>

        <TopologyDiagram />
      </section>

      {/* Scenario List Section */}
      <section className="space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-slate-200 dark:border-slate-800">
          <div>
            <h2 className="text-xl font-bold text-slate-900 dark:text-white">
              Danh sách Kịch bản Thực nghiệm
            </h2>
            <p className="text-xs text-slate-500 mt-1">
              Chọn kịch bản để xem hướng dẫn chi tiết và chạy thử nghiệm
            </p>
          </div>

          {/* Group Filter Buttons */}
          <div className="flex items-center gap-1.5 text-xs">
            {[
              { id: 'all', label: `Tất cả (${SCENARIOS.length})` },
              { id: 'cdn', label: `Multi-CDN (${cdnCount})` },
              { id: 'waap', label: `CMC WAAP (${waapCount})` },
              { id: 'shop', label: `E-Commerce (${shopCount})` },
            ].map((btn) => (
              <button
                key={btn.id}
                type="button"
                onClick={() => setSelectedGroup(btn.id as ScenarioGroup | 'all')}
                className={`rounded-control px-3 py-1.5 font-medium transition-colors ${
                  selectedGroup === btn.id
                    ? 'bg-teal-600 text-white font-semibold'
                    : 'bg-slate-100 text-slate-600 hover:bg-slate-200 dark:bg-slate-800 dark:text-slate-300 dark:hover:bg-slate-700'
                }`}
              >
                {btn.label}
              </button>
            ))}
          </div>
        </div>

        {/* Scenarios Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredScenarios.map((sc) => {
            return (
              <Link
                key={sc.id}
                to={`/tutorial/${sc.id}`}
                className="group flex flex-col justify-between rounded-card border border-slate-200 bg-white p-6 transition-all hover:border-teal-500 hover:shadow-md dark:border-slate-800 dark:bg-slate-900"
              >
                <div className="space-y-3">
                  <div className="flex items-center justify-between gap-2">
                    <span
                      className={`inline-flex items-center gap-1 rounded px-2 py-0.5 text-[11px] font-bold uppercase tracking-wider ${
                        sc.group === 'cdn'
                          ? 'bg-teal-100 text-teal-800 dark:bg-teal-950 dark:text-teal-300'
                          : sc.group === 'waap'
                          ? 'bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300'
                          : 'bg-blue-100 text-blue-800 dark:bg-blue-950 dark:text-blue-300'
                      }`}
                    >
                      {sc.group === 'cdn' ? (
                        <Database className="h-3 w-3" />
                      ) : sc.group === 'waap' ? (
                        <ShieldCheck className="h-3 w-3" />
                      ) : (
                        <ShoppingBag className="h-3 w-3" />
                      )}
                      <span>{sc.group.toUpperCase()}</span>
                    </span>

                    <span className="rounded bg-slate-100 px-2 py-0.5 text-[10px] font-bold text-slate-700 dark:bg-slate-800 dark:text-slate-300">
                      {sc.priority}
                    </span>
                  </div>

                  <h3 className="text-base font-bold text-slate-900 group-hover:text-teal-600 dark:text-white dark:group-hover:text-teal-400 transition-colors">
                    {sc.title}
                  </h3>

                  <p className="text-xs text-slate-600 dark:text-slate-300 line-clamp-3 leading-relaxed">
                    {sc.goal}
                  </p>
                </div>

                <div className="mt-6 pt-4 border-t border-slate-100 dark:border-slate-800 flex items-center justify-between text-xs text-teal-600 dark:text-teal-400 font-semibold">
                  <span className="flex items-center gap-1">
                    <Terminal className="h-3.5 w-3.5" />
                    <span>Mở kịch bản & Runner</span>
                  </span>
                  <ArrowRight className="h-4 w-4 transition-transform group-hover:translate-x-1" />
                </div>
              </Link>
            );
          })}
        </div>
      </section>
    </div>
  );
}
