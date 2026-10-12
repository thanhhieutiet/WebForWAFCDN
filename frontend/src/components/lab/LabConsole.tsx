import React, { useState, useEffect, useMemo } from 'react';
import {
  Terminal,
  ChevronUp,
  ChevronDown,
  Trash2,
  Download,
  Copy,
  Check,
  X,
  Filter,
  ShieldAlert,
} from 'lucide-react';
import type { LabRequestLog } from '@/types';
import {
  getRequestHistory,
  clearRequestHistory,
  subscribeRequestHistory,
  exportHistoryJSON,
} from '@/lib/api';
import { ResponseBadges } from './ResponseBadges';
import { HeadersTable } from './HeadersTable';
import { toast } from 'sonner';

const CONSOLE_OPEN_KEY = 'labshop_console_open';

type FilterType = 'all' | 'errors' | 'hit' | 'miss' | 'blocked' | 'origin';

export function LabConsole(): React.JSX.Element {
  const [isOpen, setIsOpen] = useState<boolean>(() => {
    try {
      const saved = localStorage.getItem(CONSOLE_OPEN_KEY);
      return saved === 'true';
    } catch {
      return false;
    }
  });

  const [logs, setLogs] = useState<LabRequestLog[]>(() => getRequestHistory());
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [filter, setFilter] = useState<FilterType>('all');
  const [copiedCurlId, setCopiedCurlId] = useState<string | null>(null);

  // Subscribe to history changes from Single Fetch Wrapper
  useEffect(() => {
    const unsubscribe = subscribeRequestHistory((newLogs) => {
      setLogs(newLogs);
    });
    return unsubscribe;
  }, []);

  // Listen for Ctrl+Shift+L shortcut
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.ctrlKey && e.shiftKey && (e.key === 'L' || e.key === 'l')) {
        e.preventDefault();
        setIsOpen((prev) => {
          const next = !prev;
          try {
            localStorage.setItem(CONSOLE_OPEN_KEY, String(next));
          } catch {
            // ignore
          }
          return next;
        });
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  const toggleOpen = () => {
    setIsOpen((prev) => {
      const next = !prev;
      try {
        localStorage.setItem(CONSOLE_OPEN_KEY, String(next));
      } catch {
        // ignore
      }
      return next;
    });
  };

  const handleExportJSON = () => {
    try {
      const dataStr = exportHistoryJSON();
      const blob = new Blob([dataStr], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `labshop-requests-${Date.now()}.json`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
      toast.success('Đã xuất lịch sử request dạng JSON');
    } catch {
      toast.error('Không thể xuất tệp JSON');
    }
  };

  const handleCopyCurl = async (log: LabRequestLog) => {
    try {
      let cmd = `curl -i -X ${log.method} "${log.fullUrl}"`;
      if (log.requestHeaders) {
        Object.entries(log.requestHeaders).forEach(([k, v]) => {
          cmd += ` \\\n  -H "${k}: ${v}"`;
        });
      }
      if (log.requestBodyPreview) {
        cmd += ` \\\n  -d '${log.requestBodyPreview}'`;
      }
      await navigator.clipboard.writeText(cmd);
      setCopiedCurlId(log.id);
      toast.success('Đã sao chép cURL vào clipboard');
      setTimeout(() => setCopiedCurlId(null), 2000);
    } catch {
      toast.error('Lỗi khi sao chép lệnh cURL');
    }
  };

  const filteredLogs = useMemo(() => {
    return logs.filter((log) => {
      if (filter === 'errors') return log.status >= 400 || log.status === 0;
      if (filter === 'hit') return log.badges.includes('hit');
      if (filter === 'miss') return log.badges.includes('miss');
      if (filter === 'blocked') return log.badges.includes('blocked');
      if (filter === 'origin') return log.badges.includes('origin');
      return true;
    });
  }, [logs, filter]);

  const selectedLog = useMemo(() => {
    return logs.find((l) => l.id === selectedId) || null;
  }, [logs, selectedId]);

  return (
    <div className="fixed bottom-0 left-0 right-0 z-50 transition-all duration-200">
      {/* Floating Toggle Bar (Always visible or drawer handle) */}
      <div className="flex justify-between items-center bg-slate-900 border-t border-slate-700 px-4 py-2 text-slate-100 shadow-2xl">
        <div className="flex items-center gap-3">
          <button
            type="button"
            onClick={toggleOpen}
            className="flex items-center gap-2 text-xs font-semibold hover:text-teal-400 transition-colors focus:outline-none focus:ring-1 focus:ring-teal-400 rounded px-1.5 py-0.5"
            aria-label="Toggle Lab Console"
          >
            <Terminal className="h-4 w-4 text-teal-400" />
            <span>Lab Console</span>
            <span className="rounded-full bg-slate-800 px-2 py-0.5 text-[10px] font-mono text-teal-300 border border-slate-700">
              {logs.length}/50
            </span>
          </button>

          <kbd className="hidden sm:inline-block rounded border border-slate-700 bg-slate-800 px-1.5 py-0.5 text-[10px] font-mono text-slate-400">
            Ctrl+Shift+L
          </kbd>

          {/* Quick status summary badge */}
          {logs.length > 0 && (
            <div className="hidden md:flex items-center gap-2 text-[11px] text-slate-400 font-mono">
              <span className="text-teal-400">
                HIT: {logs.filter((l) => l.badges.includes('hit')).length}
              </span>
              <span>•</span>
              <span className="text-amber-400">
                MISS: {logs.filter((l) => l.badges.includes('miss')).length}
              </span>
              <span>•</span>
              <span className="text-rose-400">
                Blocked: {logs.filter((l) => l.badges.includes('blocked')).length}
              </span>
            </div>
          )}
        </div>

        <div className="flex items-center gap-2">
          {isOpen && (
            <>
              {/* Filter selector */}
              <div className="hidden lg:flex items-center gap-1 text-xs">
                <Filter className="h-3 w-3 text-slate-400 mr-1" />
                {(
                  [
                    { id: 'all', label: 'Tất cả' },
                    { id: 'errors', label: 'Lỗi (≥400)' },
                    { id: 'hit', label: 'HIT' },
                    { id: 'miss', label: 'MISS' },
                    { id: 'blocked', label: 'Edge Block' },
                    { id: 'origin', label: 'Origin' },
                  ] as const
                ).map((f) => (
                  <button
                    key={f.id}
                    type="button"
                    onClick={() => setFilter(f.id)}
                    className={`rounded px-2 py-0.5 text-[11px] transition-colors ${
                      filter === f.id
                        ? 'bg-teal-600 text-white font-medium'
                        : 'text-slate-400 hover:bg-slate-800 hover:text-slate-200'
                    }`}
                  >
                    {f.label}
                  </button>
                ))}
              </div>

              {/* Actions */}
              <button
                type="button"
                onClick={handleExportJSON}
                className="inline-flex items-center gap-1 rounded bg-slate-800 px-2 py-1 text-[11px] text-slate-300 hover:bg-slate-700 hover:text-white transition-colors border border-slate-700"
                title="Xuất JSON lịch sử request"
              >
                <Download className="h-3 w-3" />
                <span className="hidden sm:inline">Xuất JSON</span>
              </button>

              <button
                type="button"
                onClick={clearRequestHistory}
                className="inline-flex items-center gap-1 rounded bg-slate-800 px-2 py-1 text-[11px] text-rose-300 hover:bg-rose-950/50 hover:text-rose-200 transition-colors border border-slate-700"
                title="Xóa lịch sử 50 requests"
              >
                <Trash2 className="h-3 w-3" />
                <span className="hidden sm:inline">Xóa</span>
              </button>
            </>
          )}

          <button
            type="button"
            onClick={toggleOpen}
            className="p-1 rounded text-slate-400 hover:text-white hover:bg-slate-800 transition-colors"
            aria-label={isOpen ? 'Thu gọn console' : 'Mở rộng console'}
          >
            {isOpen ? <ChevronDown className="h-4 w-4" /> : <ChevronUp className="h-4 w-4" />}
          </button>
        </div>
      </div>

      {/* Expanded Console Body */}
      {isOpen && (
        <div className="h-80 md:h-96 bg-slate-950 text-slate-200 border-t border-slate-800 flex flex-col md:flex-row overflow-hidden font-sans">
          {/* Request List (Left column) */}
          <div className="w-full md:w-1/2 lg:w-2/5 border-b md:border-b-0 md:border-r border-slate-800 overflow-y-auto flex flex-col">
            {filteredLogs.length === 0 ? (
              <div className="flex-1 flex flex-col items-center justify-center p-6 text-slate-500 text-xs italic text-center">
                <Terminal className="h-8 w-8 mb-2 opacity-30" />
                Chưa có request HTTP nào được ghi nhận.
                <br />
                Thực hiện tương tác trên web hoặc chạy kịch bản Lab để quan sát.
              </div>
            ) : (
              <div className="divide-y divide-slate-800/80">
                {filteredLogs.map((log) => {
                  const isSelected = log.id === selectedId;
                  const isErr = log.status >= 400 || log.status === 0;

                  return (
                    <div
                      key={log.id}
                      onClick={() => setSelectedId(log.id)}
                      className={`p-2.5 text-xs cursor-pointer transition-colors ${
                        isSelected
                          ? 'bg-slate-800/90 border-l-2 border-teal-500'
                          : 'hover:bg-slate-900/80'
                      }`}
                    >
                      <div className="flex items-center justify-between gap-2 mb-1">
                        <div className="flex items-center gap-1.5 font-mono">
                          <span
                            className={`font-bold px-1.5 py-0.2 rounded text-[10px] ${
                              log.method === 'GET'
                                ? 'bg-blue-950 text-blue-300'
                                : log.method === 'POST'
                                ? 'bg-teal-950 text-teal-300'
                                : log.method === 'DELETE'
                                ? 'bg-rose-950 text-rose-300'
                                : 'bg-purple-950 text-purple-300'
                            }`}
                          >
                            {log.method}
                          </span>
                          <span
                            className={`font-semibold ${
                              isErr
                                ? 'text-rose-400'
                                : log.status === 304
                                ? 'text-blue-400'
                                : 'text-teal-400'
                            }`}
                          >
                            {log.status || 'ERR'}
                          </span>
                          <span className="text-slate-400 text-[11px] truncate max-w-[140px] lg:max-w-[180px]">
                            {log.path}
                          </span>
                        </div>
                        <div className="flex items-center gap-1.5 text-[11px] font-mono text-slate-400 shrink-0">
                          <span>{log.latencyMs}ms</span>
                        </div>
                      </div>

                      {/* Badges line */}
                      <div className="flex items-center justify-between gap-1 text-[11px]">
                        <ResponseBadges
                          badges={log.badges}
                          originInstance={log.originInstance}
                          cacheStatus={log.cacheStatus}
                        />
                        {log.requestId && (
                          <span className="text-[10px] font-mono text-slate-500 truncate max-w-[90px]" title={log.requestId}>
                            id:{log.requestId.slice(0, 8)}
                          </span>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>

          {/* Inspector Panel (Right column) */}
          <div className="w-full md:w-1/2 lg:w-3/5 p-4 overflow-y-auto bg-slate-900/60">
            {selectedLog ? (
              <div className="space-y-4">
                <div className="flex items-start justify-between gap-2 pb-2 border-b border-slate-800">
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-mono font-bold text-sm text-teal-400">
                        {selectedLog.method}
                      </span>
                      <span className="font-mono text-xs text-slate-200 select-all">
                        {selectedLog.fullUrl}
                      </span>
                    </div>
                    <div className="flex items-center gap-3 mt-1 text-xs text-slate-400">
                      <span>Status: <strong className={selectedLog.status >= 400 ? 'text-rose-400' : 'text-teal-400'}>{selectedLog.status}</strong></span>
                      <span>Độ trễ: <strong>{selectedLog.latencyMs} ms</strong></span>
                      <span>Thời điểm: {new Date(selectedLog.timestamp).toLocaleTimeString()}</span>
                    </div>
                  </div>

                  <div className="flex items-center gap-2">
                    <button
                      type="button"
                      onClick={() => handleCopyCurl(selectedLog)}
                      className="inline-flex items-center gap-1 rounded bg-slate-800 px-2.5 py-1 text-xs text-slate-200 hover:bg-slate-700 transition-colors border border-slate-700"
                    >
                      {copiedCurlId === selectedLog.id ? (
                        <Check className="h-3 w-3 text-teal-400" />
                      ) : (
                        <Copy className="h-3 w-3" />
                      )}
                      <span>Copy cURL</span>
                    </button>
                    <button
                      type="button"
                      onClick={() => setSelectedId(null)}
                      className="text-slate-400 hover:text-white p-1"
                      aria-label="Đóng chi tiết"
                    >
                      <X className="h-4 w-4" />
                    </button>
                  </div>
                </div>

                {/* Observation Speculation Note */}
                {selectedLog.speculatedEdgeBlocked && (
                  <div className="rounded-control border border-rose-900/60 bg-rose-950/40 p-3 text-xs text-rose-200 space-y-1">
                    <div className="flex items-center gap-1.5 font-semibold text-rose-400">
                      <ShieldAlert className="h-4 w-4 text-rose-400" />
                      Nghi bị chặn ở Edge / WAAP (Suy đoán)
                    </div>
                    <p className="text-rose-300">
                      {selectedLog.speculationEvidence ||
                        'Request không có header X-Origin-Instance và trả về lỗi hoặc định dạng không phải JSON.'}
                    </p>
                  </div>
                )}

                {/* Headers Tables */}
                <div className="grid grid-cols-1 gap-3">
                  <HeadersTable headers={selectedLog.responseHeaders} title="Response Headers" />
                  <HeadersTable headers={selectedLog.requestHeaders} title="Request Headers" />
                </div>

                {/* Body Previews */}
                {selectedLog.requestBodyPreview && (
                  <div className="space-y-1">
                    <div className="text-xs font-semibold text-slate-300">Request Body Preview</div>
                    <pre className="max-h-32 overflow-x-auto rounded-control bg-slate-950 p-2 text-xs font-mono text-slate-300 border border-slate-800">
                      {selectedLog.requestBodyPreview}
                    </pre>
                  </div>
                )}

                {selectedLog.responseBodyPreview && (
                  <div className="space-y-1">
                    <div className="text-xs font-semibold text-slate-300">Response Body Preview</div>
                    <pre className="max-h-40 overflow-x-auto rounded-control bg-slate-950 p-2 text-xs font-mono text-teal-300 border border-slate-800 select-all">
                      {selectedLog.responseBodyPreview}
                    </pre>
                  </div>
                )}
              </div>
            ) : (
              <div className="h-full flex flex-col items-center justify-center text-slate-500 text-xs italic">
                Bấm vào một request bên trái để kiểm tra chi tiết headers và payload.
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
