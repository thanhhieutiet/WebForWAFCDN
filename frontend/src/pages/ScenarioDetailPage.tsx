import React, { useState, useRef } from 'react';
import { useParams, Link } from 'react-router-dom';
import {
  ArrowLeft,
  Terminal,
  Play,
  RotateCw,
  Zap,
  Square,
  ExternalLink,
  ShieldAlert,
  CheckCircle2,
  AlertTriangle,
  Info,
  CheckSquare,
  Square as SquareOutline,
  Layers,
} from 'lucide-react';
import { SCENARIOS } from '@/data/scenariosData';
import { api } from '@/lib/api';
import { ResponseBadges } from '@/components/lab/ResponseBadges';
import { HeadersTable } from '@/components/lab/HeadersTable';
import { CurlBlock } from '@/components/lab/CurlBlock';
import { StatusHistogram } from '@/components/lab/StatusHistogram';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Dialog } from '@/components/ui/dialog';
import { toast } from 'sonner';
import type { LabRequestLog } from '@/types';

interface RunResultItem {
  index: number;
  status: number;
  latencyMs: number;
  cacheStatus?: string;
  age?: string;
  requestId?: string;
  timestamp: string;
}

export function ScenarioDetailPage(): React.JSX.Element {
  const { id } = useParams<{ id: string }>();
  const scenario = SCENARIOS.find((s) => s.id === id);

  // Stepper state
  const [currentStep, setCurrentStep] = useState<number>(0);
  const steps = [
    { title: 'Chuẩn bị', desc: 'Kiểm tra môi trường và endpoint ban đầu' },
    { title: 'Cấu hình portal', desc: 'Thao tác cấu hình quy tắc trên CMC Cloud Portal' },
    { title: 'Chạy thử', desc: 'Sử dụng Runner Panel để gửi các mẫu request' },
    { title: 'Quan sát', desc: 'Theo dõi header, trạng thái Cache/WAF trên Lab Console' },
    { title: 'Dọn dẹp', desc: 'Xóa hoặc khôi phục quy tắc sau bài lab' },
  ];

  // Checklist state stored in localStorage
  const [checklist, setChecklist] = useState<Record<string, boolean>>(() => {
    try {
      const saved = localStorage.getItem(`labshop_checklist_${id}`);
      return saved ? JSON.parse(saved) : {};
    } catch {
      return {};
    }
  });

  const toggleChecklistItem = (key: string) => {
    setChecklist((prev) => {
      const next = { ...prev, [key]: !prev[key] };
      try {
        localStorage.setItem(`labshop_checklist_${id}`, JSON.stringify(next));
      } catch {
        // ignore
      }
      return next;
    });
  };

  // Runner Panel States
  const [selectedReqIndex, setSelectedReqIndex] = useState(0);
  const [runningOnce, setRunningOnce] = useState(false);
  const [lastSingleLog, setLastSingleLog] = useState<LabRequestLog | null>(null);

  // Run N States
  const [nTimes, setNTimes] = useState(3);
  const [runningN, setRunningN] = useState(false);
  const [nResults, setNResults] = useState<RunResultItem[]>([]);

  // Burst States
  const [burstCount, setBurstCount] = useState(30);
  const [burstConcurrency, setBurstConcurrency] = useState(10);
  const [burstConfirmOpen, setBurstConfirmOpen] = useState(false);
  const [runningBurst, setRunningBurst] = useState(false);
  const [burstStats, setBurstStats] = useState<Record<number, number>>({});
  const [burstCompleted, setBurstCompleted] = useState(0);
  const abortControllerRef = useRef<AbortController | null>(null);

  // Probe Confirmation Dialog
  const [probeConfirmedSession, setProbeConfirmedSession] = useState(false);
  const [probeConfirmModalOpen, setProbeConfirmModalOpen] = useState(false);
  const [pendingRunAction, setPendingRunAction] = useState<(() => void) | null>(null);

  if (!scenario) {
    return (
      <div className="mx-auto max-w-xl px-4 py-16 text-center space-y-4">
        <AlertTriangle className="mx-auto h-12 w-12 text-rose-500" />
        <h1 className="text-xl font-bold">Không tìm thấy kịch bản #{id}</h1>
        <Link to="/tutorial">
          <Button variant="outline">Quay lại danh sách kịch bản</Button>
        </Link>
      </div>
    );
  }

  const activeReq = scenario.requests[selectedReqIndex] || scenario.requests[0];

  // Check probe safety before running
  const executeWithSafetyCheck = (action: () => void) => {
    if (activeReq?.isProbe && !probeConfirmedSession) {
      setPendingRunAction(() => action);
      setProbeConfirmModalOpen(true);
    } else {
      action();
    }
  };

  // 1. Run Once
  const handleRunOnce = () => {
    executeWithSafetyCheck(async () => {
      if (!activeReq || !activeReq.runnable) return;
      try {
        setRunningOnce(true);
        const res = await api.get(activeReq.path, {
          method: activeReq.method,
          body: activeReq.body,
        });
        setLastSingleLog(res.log);
        toast.success(`Đã gửi request: HTTP ${res.log.status}`);
      } catch (err: any) {
        if (err.log) {
          setLastSingleLog(err.log);
        }
        toast.error(`Request hoàn tất: HTTP ${err.status || 0}`);
      } finally {
        setRunningOnce(false);
      }
    });
  };

  // 2. Run N Times sequentially
  const handleRunN = () => {
    executeWithSafetyCheck(async () => {
      if (!activeReq || !activeReq.runnable) return;
      setRunningN(true);
      setNResults([]);
      const results: RunResultItem[] = [];

      for (let i = 1; i <= Math.min(20, Math.max(1, nTimes)); i++) {
        try {
          const res = await api.get(activeReq.path, {
            method: activeReq.method,
            body: activeReq.body,
          });
          results.push({
            index: i,
            status: res.log.status,
            latencyMs: res.log.latencyMs,
            cacheStatus: res.log.cacheStatus,
            age: res.log.age,
            requestId: res.log.requestId,
            timestamp: res.log.timestamp,
          });
        } catch (err: any) {
          results.push({
            index: i,
            status: err.status || 0,
            latencyMs: err.log?.latencyMs || 0,
            cacheStatus: err.log?.cacheStatus,
            age: err.log?.age,
            requestId: err.log?.requestId,
            timestamp: new Date().toISOString(),
          });
        }
        setNResults([...results]);
        // slight tick
        await new Promise((r) => setTimeout(r, 100));
      }
      setRunningN(false);
      toast.success(`Hoàn tất chạy ${nTimes} lần!`);
    });
  };

  // 3. Burst Test with concurrency
  const handleStartBurst = async () => {
    if (!activeReq) return;
    setBurstConfirmOpen(false);
    setRunningBurst(true);
    setBurstCompleted(0);
    setBurstStats({});

    const controller = new AbortController();
    abortControllerRef.current = controller;

    const total = Math.min(200, Math.max(5, burstCount));
    const limit = Math.min(20, Math.max(1, burstConcurrency));

    const statsAcc: Record<number, number> = {};
    let executed = 0;

    const runWorker = async () => {
      while (executed < total && !controller.signal.aborted) {
        executed++;
        try {
          const res = await api.get(activeReq.path, {
            method: activeReq.method,
            body: activeReq.body,
            signal: controller.signal,
          });
          const code = res.log.status;
          statsAcc[code] = (statsAcc[code] || 0) + 1;
        } catch (err: any) {
          const code = err.status || 0;
          statsAcc[code] = (statsAcc[code] || 0) + 1;
        }
        setBurstStats({ ...statsAcc });
        setBurstCompleted((prev) => prev + 1);
      }
    };

    const workers = Array.from({ length: limit }, () => runWorker());
    await Promise.all(workers);

    setRunningBurst(false);
    toast.success('Hoàn thành Burst Test!');
  };

  const handleStopBurst = () => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      setRunningBurst(false);
      toast.info('Đã dừng Burst Test.');
    }
  };

  // Identify duplicate request IDs in Run N results for cache HIT coloring
  const reqIdCounts = nResults.reduce((acc, r) => {
    if (r.requestId) acc[r.requestId] = (acc[r.requestId] || 0) + 1;
    return acc;
  }, {} as Record<string, number>);

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8 space-y-8">
      {/* Top Breadcrumb & Title */}
      <div className="space-y-3">
        <Link
          to="/tutorial"
          className="inline-flex items-center gap-1.5 text-xs font-semibold text-slate-600 hover:text-teal-600 dark:text-slate-400 dark:hover:text-teal-400"
        >
          <ArrowLeft className="h-4 w-4" />
          <span>Quay lại Danh sách Kịch bản</span>
        </Link>

        <div className="flex flex-wrap items-center justify-between gap-4">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <span className="rounded bg-teal-100 text-teal-800 dark:bg-teal-950 dark:text-teal-300 px-2 py-0.5 text-xs font-bold uppercase">
                {scenario.group.toUpperCase()}
              </span>
              <span className="rounded bg-slate-100 dark:bg-slate-800 px-2 py-0.5 text-xs font-mono font-bold text-slate-700 dark:text-slate-300">
                {scenario.priority}
              </span>
            </div>
            <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-slate-900 dark:text-white">
              {scenario.title}
            </h1>
          </div>

          {scenario.docsUrl && (
            <a
              href={scenario.docsUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1.5 rounded-control border border-slate-300 bg-white px-3 py-1.5 text-xs font-medium text-slate-700 shadow-sm hover:bg-slate-50 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-200 dark:hover:bg-slate-700 transition-colors"
            >
              <span>Xem tài liệu CMC Portal</span>
              <ExternalLink className="h-3.5 w-3.5 text-teal-600" />
            </a>
          )}
        </div>

        <p className="text-sm text-slate-600 dark:text-slate-300 leading-relaxed max-w-3xl">
          <strong>Mục tiêu:</strong> {scenario.goal}
        </p>
      </div>

      {/* Stepper Navigation */}
      <section className="rounded-card border border-slate-200 bg-white p-4 shadow-sm dark:border-slate-800 dark:bg-slate-900">
        <div className="grid grid-cols-2 sm:grid-cols-5 gap-2 text-xs">
          {steps.map((st, idx) => {
            const isCurrent = idx === currentStep;
            const isCompleted = idx < currentStep;

            return (
              <button
                key={idx}
                type="button"
                onClick={() => setCurrentStep(idx)}
                className={`p-2.5 rounded-control text-left transition-colors flex flex-col justify-between ${
                  isCurrent
                    ? 'bg-teal-50 border border-teal-300 text-teal-900 dark:bg-teal-950/60 dark:border-teal-800 dark:text-teal-200'
                    : isCompleted
                    ? 'bg-slate-50 text-slate-700 dark:bg-slate-800/50 dark:text-slate-300'
                    : 'text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800/30'
                }`}
              >
                <div className="flex items-center gap-1.5 font-bold mb-1">
                  <span className="flex h-5 w-5 items-center justify-center rounded-full bg-teal-600 text-[10px] text-white">
                    {idx + 1}
                  </span>
                  <span>{st.title}</span>
                </div>
                <div className="text-[10px] text-slate-500 line-clamp-1">{st.desc}</div>
              </button>
            );
          })}
        </div>
      </section>

      {/* Two Column Layout: Left Content (Portal Steps & Expect) | Right Runner Panel */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
        {/* Left Column (7 Cols) */}
        <div className="lg:col-span-7 space-y-6">
          {/* Portal Steps with Checklist */}
          <section className="rounded-card border border-slate-200 bg-white p-6 shadow-sm dark:border-slate-800 dark:bg-slate-900 space-y-4">
            <div className="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-3">
              <h2 className="text-base font-bold text-slate-900 dark:text-white flex items-center gap-2">
                <Layers className="h-4 w-4 text-teal-600" />
                <span>Các bước thực hiện trên CMC Portal</span>
              </h2>
              <span className="text-xs text-slate-400">Đánh dấu để lưu tiến độ</span>
            </div>

            <div className="space-y-3">
              {scenario.portalSteps.map((step, idx) => {
                const itemKey = `step_${idx}`;
                const isChecked = !!checklist[itemKey];

                return (
                  <div
                    key={idx}
                    onClick={() => toggleChecklistItem(itemKey)}
                    className={`flex items-start gap-3 p-3 rounded-control border cursor-pointer transition-colors ${
                      isChecked
                        ? 'bg-teal-50/50 border-teal-200 text-teal-900 dark:bg-teal-950/30 dark:border-teal-900 dark:text-teal-200'
                        : 'border-slate-200 hover:bg-slate-50 dark:border-slate-800 dark:hover:bg-slate-800/40 text-slate-800 dark:text-slate-200'
                    }`}
                  >
                    <button type="button" className="mt-0.5 text-teal-600">
                      {isChecked ? (
                        <CheckSquare className="h-4 w-4" />
                      ) : (
                        <SquareOutline className="h-4 w-4 text-slate-400" />
                      )}
                    </button>
                    <div className="text-xs leading-relaxed flex-1 select-none">
                      <span className="font-bold mr-1.5">Bước {idx + 1}:</span>
                      <span className={isChecked ? 'line-through text-slate-500' : ''}>
                        {step}
                      </span>
                    </div>
                  </div>
                );
              })}
            </div>
          </section>

          {/* Expect Table */}
          <section className="rounded-card border border-slate-200 bg-white p-6 shadow-sm dark:border-slate-800 dark:bg-slate-900 space-y-4">
            <h2 className="text-base font-bold text-slate-900 dark:text-white flex items-center gap-2 border-b border-slate-100 dark:border-slate-800 pb-3">
              <CheckCircle2 className="h-4 w-4 text-teal-600" />
              <span>Bảng Kết quả Kỳ vọng (Expect Table)</span>
            </h2>

            <div className="rounded-control border border-slate-200 overflow-hidden dark:border-slate-800 text-xs">
              <table className="w-full text-left">
                <thead className="bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 font-semibold">
                  <tr>
                    <th className="p-3 w-2/5">Điều kiện / Cấu hình</th>
                    <th className="p-3">Kết quả quan sát kỳ vọng</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
                  {scenario.expect.map((exp, idx) => (
                    <tr key={idx} className="hover:bg-slate-50/50 dark:hover:bg-slate-900/50">
                      <td className="p-3 font-semibold text-slate-900 dark:text-white">
                        {exp.condition}
                      </td>
                      <td className="p-3 text-slate-600 dark:text-slate-300 leading-relaxed">
                        {exp.expectedResult}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {scenario.notes && (
              <div className="rounded-control bg-slate-50 dark:bg-slate-800/50 p-3 text-xs text-slate-600 dark:text-slate-300 flex items-start gap-2">
                <Info className="h-4 w-4 shrink-0 text-teal-600 mt-0.5" />
                <span>
                  <strong>Ghi chú:</strong> {scenario.notes}
                </span>
              </div>
            )}
          </section>
        </div>

        {/* Right Column: Runner Panel (5 Cols) */}
        <div className="lg:col-span-5 space-y-6">
          <div className="rounded-card border border-slate-200 bg-white p-6 shadow-sm dark:border-slate-800 dark:bg-slate-900 space-y-6 sticky top-20">
            <div className="border-b border-slate-100 dark:border-slate-800 pb-3 flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Terminal className="h-4 w-4 text-teal-600" />
                <h2 className="text-base font-bold text-slate-900 dark:text-white">Runner Panel</h2>
              </div>
              <span className="text-[11px] font-mono text-slate-400">
                {activeReq?.method} {activeReq?.path}
              </span>
            </div>

            {/* Request selector if multiple requests exist */}
            {scenario.requests.length > 1 && (
              <div className="space-y-1">
                <label className="text-xs font-semibold text-slate-600">Chọn mẫu Request:</label>
                <div className="flex flex-wrap gap-1.5">
                  {scenario.requests.map((r, i) => (
                    <button
                      key={i}
                      type="button"
                      onClick={() => setSelectedReqIndex(i)}
                      className={`px-2 py-1 rounded text-xs font-mono font-medium ${
                        selectedReqIndex === i
                          ? 'bg-teal-600 text-white'
                          : 'bg-slate-100 text-slate-700 hover:bg-slate-200 dark:bg-slate-800 dark:text-slate-300'
                      }`}
                    >
                      #{i + 1} {r.method} {r.path}
                    </button>
                  ))}
                </div>
              </div>
            )}

            {/* cURL Display */}
            {activeReq?.curl && <CurlBlock curl={activeReq.curl} label="Lệnh cURL mẫu" />}

            {/* If Non-Runnable: display reason only */}
            {!activeReq?.runnable ? (
              <div className="rounded-control bg-amber-50 p-4 text-xs text-amber-900 dark:bg-amber-950/60 dark:text-amber-200 space-y-2 border border-amber-200 dark:border-amber-900">
                <div className="flex items-center gap-1.5 font-bold">
                  <AlertTriangle className="h-4 w-4 text-amber-600" />
                  Kịch bản chỉ chạy qua cURL bên ngoài (runnable = false)
                </div>
                <p className="leading-relaxed">
                  {activeReq?.nonRunnableReason ||
                    'Trình duyệt hạn chế can thiệp các header đặc quyền (như User-Agent, Host, Cookie). Hãy sao chép lệnh cURL ở trên và chạy tại Terminal của bạn.'}
                </p>
              </div>
            ) : (
              /* Runnable Action Controls */
              <div className="space-y-4">
                {/* Mode 1: Run Once */}
                <div className="flex items-center gap-3">
                  <Button
                    size="sm"
                    loading={runningOnce}
                    onClick={handleRunOnce}
                    className="flex-1 gap-1.5"
                  >
                    <Play className="h-3.5 w-3.5" />
                    <span>Run once (1 lần)</span>
                  </Button>
                </div>

                {/* Mode 2: Run N */}
                <div className="p-3 rounded-control border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/50 space-y-3">
                  <div className="flex items-center justify-between text-xs">
                    <span className="font-semibold text-slate-700 dark:text-slate-300">
                      Run N lần (1 - 20):
                    </span>
                    <div className="flex items-center gap-2">
                      <Input
                        type="number"
                        min={1}
                        max={20}
                        value={nTimes}
                        onChange={(e) => setNTimes(Number(e.target.value))}
                        className="w-16 h-8 text-xs text-center"
                      />
                      <Button
                        size="sm"
                        variant="outline"
                        loading={runningN}
                        onClick={handleRunN}
                        className="h-8 gap-1"
                      >
                        <RotateCw className="h-3 w-3" />
                        <span>Chạy N</span>
                      </Button>
                    </div>
                  </div>

                  {/* Run N Results Table */}
                  {nResults.length > 0 && (
                    <div className="max-h-48 overflow-y-auto rounded border border-slate-200 dark:border-slate-800 text-[11px] font-mono">
                      <table className="w-full text-left">
                        <thead className="bg-slate-100 dark:bg-slate-800 text-slate-600">
                          <tr>
                            <th className="p-1.5">#</th>
                            <th className="p-1.5">Status</th>
                            <th className="p-1.5">ms</th>
                            <th className="p-1.5">Cache</th>
                            <th className="p-1.5">Request ID</th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
                          {nResults.map((row) => {
                            const isRepeatedReqId = (reqIdCounts[row.requestId || ''] || 0) > 1;

                            return (
                              <tr
                                key={row.index}
                                className={
                                  isRepeatedReqId
                                    ? 'bg-teal-50 dark:bg-teal-950/60 font-semibold text-teal-900 dark:text-teal-200'
                                    : ''
                                }
                              >
                                <td className="p-1.5">{row.index}</td>
                                <td className="p-1.5 font-bold">{row.status}</td>
                                <td className="p-1.5">{row.latencyMs}ms</td>
                                <td className="p-1.5">{row.cacheStatus || '-'}</td>
                                <td className="p-1.5 truncate max-w-[80px]" title={row.requestId}>
                                  {row.requestId ? row.requestId.slice(0, 8) : '-'}
                                </td>
                              </tr>
                            );
                          })}
                        </tbody>
                      </table>
                    </div>
                  )}
                </div>

                {/* Mode 3: Burst Test */}
                <div className="p-3 rounded-control border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/50 space-y-3">
                  <div className="flex items-center justify-between text-xs">
                    <span className="font-semibold text-slate-700 dark:text-slate-300">
                      Burst Test (Rate Limit):
                    </span>

                    {runningBurst ? (
                      <Button
                        size="sm"
                        variant="destructive"
                        onClick={handleStopBurst}
                        className="h-8 gap-1 text-xs"
                      >
                        <Square className="h-3 w-3 fill-current" />
                        <span>Dừng ({burstCompleted}/{burstCount})</span>
                      </Button>
                    ) : (
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => executeWithSafetyCheck(() => setBurstConfirmOpen(true))}
                        className="h-8 gap-1 text-xs text-amber-600 border-amber-300 hover:bg-amber-50"
                      >
                        <Zap className="h-3 w-3" />
                        <span>Cấu hình Burst</span>
                      </Button>
                    )}
                  </div>

                  {/* Status Histogram for Burst results */}
                  {Object.keys(burstStats).length > 0 && (
                    <StatusHistogram stats={burstStats} total={burstCompleted} />
                  )}
                </div>

                {/* Single Run Result Preview */}
                {lastSingleLog && (
                  <div className="p-3 rounded-control border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 space-y-2 text-xs">
                    <div className="flex items-center justify-between">
                      <div className="font-bold flex items-center gap-1.5">
                        <span className={lastSingleLog.status >= 400 ? 'text-rose-600' : 'text-teal-600'}>
                          HTTP {lastSingleLog.status}
                        </span>
                        <span>•</span>
                        <span className="text-slate-500 font-mono">{lastSingleLog.latencyMs} ms</span>
                      </div>
                      <ResponseBadges badges={lastSingleLog.badges} />
                    </div>

                    <HeadersTable headers={lastSingleLog.responseHeaders} title="Response Headers" />

                    {lastSingleLog.responseBodyPreview && (
                      <div>
                        <div className="font-semibold text-slate-500 text-[11px] mb-1">Response Body:</div>
                        <pre className="max-h-28 overflow-x-auto rounded bg-slate-900 text-teal-300 p-2 font-mono text-[11px]">
                          {lastSingleLog.responseBodyPreview}
                        </pre>
                      </div>
                    )}
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Burst Test Confirmation Dialog */}
      <Dialog
        open={burstConfirmOpen}
        onOpenChange={setBurstConfirmOpen}
        title="Xác nhận chạy Burst Test"
        description="Gửi lượng lớn request đồng thời để kiểm tra ngưỡng Rate Limiting của WAF hoặc năng lực chịu tải."
      >
        <div className="space-y-4 text-xs">
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="font-semibold block mb-1">Tổng số requests (tối đa 200)</label>
              <Input
                type="number"
                min={5}
                max={200}
                value={burstCount}
                onChange={(e) => setBurstCount(Number(e.target.value))}
              />
            </div>
            <div>
              <label className="font-semibold block mb-1">Concurrency (tối đa 20)</label>
              <Input
                type="number"
                min={1}
                max={20}
                value={burstConcurrency}
                onChange={(e) => setBurstConcurrency(Number(e.target.value))}
              />
            </div>
          </div>

          <div className="p-3 rounded bg-amber-50 text-amber-800 dark:bg-amber-950/50 dark:text-amber-200">
            Lưu ý: Burst test gửi liên tục các request có thể kích hoạt Rule Rate Limit và chặn IP của bạn trong thời gian cấu hình.
          </div>

          <div className="flex justify-end gap-2 pt-2 border-t">
            <Button variant="outline" size="sm" onClick={() => setBurstConfirmOpen(false)}>
              Hủy
            </Button>
            <Button size="sm" onClick={handleStartBurst}>
              Bắt đầu Burst ({burstCount} reqs)
            </Button>
          </div>
        </div>
      </Dialog>

      {/* WAF Probe Safety Modal */}
      <Dialog
        open={probeConfirmModalOpen}
        onOpenChange={setProbeConfirmModalOpen}
        title="Xác nhận an toàn thử nghiệm WAF Probe"
        description="Bạn đang chuẩn bị gửi chuỗi thử nghiệm mang chữ ký tấn công (SQLi / XSS / Path Traversal)."
      >
        <div className="space-y-4 text-xs">
          <div className="p-3 rounded-control bg-rose-50 text-rose-800 dark:bg-rose-950/50 dark:text-rose-200 flex items-start gap-2">
            <ShieldAlert className="h-5 w-5 shrink-0 text-rose-600 mt-0.5" />
            <div className="leading-relaxed">
              <strong>Cảnh báo an toàn:</strong> Chỉ thực hiện gửi các mẫu thử nghiệm này tới chính máy chủ phòng lab mà bạn có quyền kiểm soát. Không dùng công cụ này quét các hệ thống bên ngoài.
            </div>
          </div>

          <div className="flex justify-end gap-2 pt-2 border-t">
            <Button variant="outline" size="sm" onClick={() => setProbeConfirmModalOpen(false)}>
              Hủy bỏ
            </Button>
            <Button
              size="sm"
              onClick={() => {
                setProbeConfirmedSession(true);
                setProbeConfirmModalOpen(false);
                if (pendingRunAction) pendingRunAction();
              }}
            >
              Tôi xác nhận & Tiếp tục
            </Button>
          </div>
        </div>
      </Dialog>
    </div>
  );
}
