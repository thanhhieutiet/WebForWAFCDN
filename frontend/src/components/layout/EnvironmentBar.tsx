import React, { useState, useEffect } from 'react';
import { Globe, Settings, AlertTriangle, ShieldCheck, Check } from 'lucide-react';
import { getHostConfig, saveHostConfig, detectCurrentRouting } from '@/lib/environment';
import { Dialog } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { toast } from 'sonner';
import type { RoutingHostConfig } from '@/types';

export function EnvironmentBar(): React.JSX.Element {
  const [config, setConfig] = useState<RoutingHostConfig>(() => getHostConfig());
  const [routingInfo, setRoutingInfo] = useState(() => detectCurrentRouting(config));
  const [modalOpen, setModalOpen] = useState(false);
  const [formConfig, setFormConfig] = useState<RoutingHostConfig>(config);

  const refreshState = () => {
    const fresh = getHostConfig();
    setConfig(fresh);
    setRoutingInfo(detectCurrentRouting(fresh));
  };

  useEffect(() => {
    window.addEventListener('labshop_host_config_changed', refreshState);
    return () => window.removeEventListener('labshop_host_config_changed', refreshState);
  }, []);

  const handleOpenModal = () => {
    setFormConfig(config);
    setModalOpen(true);
  };

  const handleSave = (e: React.FormEvent) => {
    e.preventDefault();
    saveHostConfig(formConfig);
    setConfig(formConfig);
    setRoutingInfo(detectCurrentRouting(formConfig));
    setModalOpen(false);
    toast.success('Đã lưu cấu hình hostname môi trường Lab');
  };

  return (
    <>
      <div className="border-b border-slate-200 bg-slate-100/90 text-xs text-slate-700 dark:border-slate-800 dark:bg-slate-900/90 dark:text-slate-300">
        <div className="mx-auto flex h-9 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
          <div className="flex items-center gap-3 overflow-hidden">
            <div className="flex items-center gap-1.5 font-mono text-[11px] truncate">
              <Globe className="h-3.5 w-3.5 text-teal-600 dark:text-teal-400 shrink-0" />
              <span className="text-slate-500 dark:text-slate-400 hidden sm:inline">Host:</span>
              <strong className="text-slate-900 dark:text-white truncate">{routingInfo.host}</strong>
            </div>

            <div className="h-3 w-[1px] bg-slate-300 dark:bg-slate-700 hidden sm:block" />

            {/* Routing badge */}
            <div className="flex items-center gap-1.5 font-mono text-[11px]">
              <span className="text-slate-500 dark:text-slate-400 hidden md:inline">Đường đi:</span>
              <span
                className={`inline-flex items-center gap-1 rounded px-2 py-0.5 font-semibold ${
                  routingInfo.routing === 'direct'
                    ? 'bg-blue-100 text-blue-800 dark:bg-blue-950 dark:text-blue-300'
                    : routingInfo.routing === 'waap'
                    ? 'bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300'
                    : routingInfo.routing === 'cdn'
                    ? 'bg-teal-100 text-teal-800 dark:bg-teal-950 dark:text-teal-300'
                    : routingInfo.routing === 'full'
                    ? 'bg-purple-100 text-purple-800 dark:bg-purple-950 dark:text-purple-300'
                    : 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300'
                }`}
              >
                {routingInfo.isUnknown ? (
                  <AlertTriangle className="h-3 w-3 text-amber-600 dark:text-amber-400" />
                ) : (
                  <ShieldCheck className="h-3 w-3" />
                )}
                <span>{routingInfo.label}</span>
              </span>
            </div>

            {routingInfo.isUnknown && (
              <span className="hidden lg:inline text-amber-600 dark:text-amber-400 text-[11px] italic">
                (Hostname lạ - bạn có thể cấu hình trong cài đặt)
              </span>
            )}
          </div>

          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={handleOpenModal}
              className="inline-flex items-center gap-1 rounded px-2 py-1 text-[11px] font-medium text-slate-600 hover:bg-slate-200 dark:text-slate-300 dark:hover:bg-slate-800 transition-colors"
              title="Cấu hình 4 hostname đường đi của bạn"
            >
              <Settings className="h-3 w-3" />
              <span>Cấu hình Hostname</span>
            </button>
          </div>
        </div>
      </div>

      {/* Hostname Configuration Modal */}
      <Dialog
        open={modalOpen}
        onOpenChange={setModalOpen}
        title="Cấu hình Hostname Môi trường Lab"
        description="Nhập 4 hostname tương ứng với 4 tuyến đường đi thực tế của bạn để LabShop tự động nhận diện."
      >
        <form onSubmit={handleSave} className="space-y-4">
          <div className="space-y-3">
            <div>
              <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                1. Direct Origin (Client -&gt; Origin EC2)
              </label>
              <Input
                value={formConfig.direct}
                onChange={(e) => setFormConfig({ ...formConfig, direct: e.target.value })}
                placeholder="vd: localhost:8080 hoặc origin.labshop.local"
                required
              />
            </div>

            <div>
              <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                2. WAAP Only (Client -&gt; WAAP -&gt; Origin)
              </label>
              <Input
                value={formConfig.waap}
                onChange={(e) => setFormConfig({ ...formConfig, waap: e.target.value })}
                placeholder="vd: waap.labshop.local"
                required
              />
            </div>

            <div>
              <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                3. Multi-CDN Only (Client -&gt; CDN -&gt; Origin)
              </label>
              <Input
                value={formConfig.cdn}
                onChange={(e) => setFormConfig({ ...formConfig, cdn: e.target.value })}
                placeholder="vd: cdn.labshop.local"
                required
              />
            </div>

            <div>
              <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                4. Full Topology (Client -&gt; CDN -&gt; WAAP -&gt; Origin)
              </label>
              <Input
                value={formConfig.full}
                onChange={(e) => setFormConfig({ ...formConfig, full: e.target.value })}
                placeholder="vd: shop.labshop.local"
                required
              />
            </div>
          </div>

          <div className="flex justify-end gap-2 pt-2 border-t border-slate-200 dark:border-slate-800">
            <Button type="button" variant="outline" onClick={() => setModalOpen(false)}>
              Hủy
            </Button>
            <Button type="submit">
              <Check className="h-4 w-4 mr-1" />
              Lưu cấu hình
            </Button>
          </div>
        </form>
      </Dialog>
    </>
  );
}
