import React from 'react';
import { Server, Database, AlertCircle, ShieldAlert, Clock, Cpu } from 'lucide-react';
import type { LabBadgeType } from '@/types';

interface ResponseBadgesProps {
  badges: LabBadgeType[];
  originInstance?: string;
  cacheStatus?: string;
  className?: string;
}

export function ResponseBadges({ badges, originInstance, cacheStatus, className = '' }: ResponseBadgesProps): React.JSX.Element {
  return (
    <div className={`flex flex-wrap items-center gap-1.5 ${className}`}>
      {badges.includes('origin') && (
        <span
          className="inline-flex items-center gap-1 rounded-control bg-blue-50 px-2 py-0.5 text-xs font-semibold text-blue-700 ring-1 ring-inset ring-blue-700/20 dark:bg-blue-950/60 dark:text-blue-300 dark:ring-blue-500/30"
          title={`Phản hồi từ Origin Server: ${originInstance || 'LabShop Origin'}`}
        >
          <Server className="h-3 w-3" />
          <span>Origin{originInstance ? ` (${originInstance})` : ''}</span>
        </span>
      )}

      {badges.includes('hit') && (
        <span
          className="inline-flex items-center gap-1 rounded-control bg-teal-50 px-2 py-0.5 text-xs font-semibold text-teal-700 ring-1 ring-inset ring-teal-600/20 dark:bg-teal-950/60 dark:text-teal-300 dark:ring-teal-500/30"
          title="Nội dung được trả trực tiếp từ bộ đệm của Edge CDN"
        >
          <Database className="h-3 w-3" />
          <span>Cache HIT{cacheStatus ? ` (${cacheStatus})` : ''}</span>
        </span>
      )}

      {badges.includes('miss') && (
        <span
          className="inline-flex items-center gap-1 rounded-control bg-amber-50 px-2 py-0.5 text-xs font-semibold text-amber-700 ring-1 ring-inset ring-amber-600/20 dark:bg-amber-950/60 dark:text-amber-300 dark:ring-amber-500/30"
          title="Edge CDN chưa có cache, phải chuyển tiếp về Origin lấy dữ liệu"
        >
          <AlertCircle className="h-3 w-3" />
          <span>Cache MISS{cacheStatus ? ` (${cacheStatus})` : ''}</span>
        </span>
      )}

      {badges.includes('stale') && (
        <span
          className="inline-flex items-center gap-1 rounded-control bg-orange-50 px-2 py-0.5 text-xs font-semibold text-orange-700 ring-1 ring-inset ring-orange-600/20 dark:bg-orange-950/60 dark:text-orange-300 dark:ring-orange-500/30"
          title="Cache đã hết hạn hoặc đang phục vụ nội dung cũ (Stale)"
        >
          <Clock className="h-3 w-3" />
          <span>Stale</span>
        </span>
      )}

      {badges.includes('blocked') && (
        <span
          className="inline-flex items-center gap-1 rounded-control bg-rose-50 px-2 py-0.5 text-xs font-semibold text-rose-700 ring-1 ring-inset ring-rose-600/20 dark:bg-rose-950/60 dark:text-rose-300 dark:ring-rose-500/30"
          title="Nghi bị chặn ở edge/WAAP (Không có header Origin)"
        >
          <ShieldAlert className="h-3 w-3" />
          <span>Edge Blocked</span>
        </span>
      )}

      {badges.includes('compressed') && (
        <span
          className="inline-flex items-center gap-1 rounded-control bg-purple-50 px-2 py-0.5 text-xs font-semibold text-purple-700 ring-1 ring-inset ring-purple-600/20 dark:bg-purple-950/60 dark:text-purple-300 dark:ring-purple-500/30"
          title="Dữ liệu được nén (Gzip/Brotli)"
        >
          <Cpu className="h-3 w-3" />
          <span>Compressed</span>
        </span>
      )}
    </div>
  );
}
