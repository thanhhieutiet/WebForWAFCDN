import React, { useState } from 'react';
import { Copy, Check } from 'lucide-react';
import { toast } from 'sonner';

interface CurlBlockProps {
  curl: string;
  label?: string;
}

export function CurlBlock({ curl, label }: CurlBlockProps): React.JSX.Element {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(curl);
      setCopied(true);
      toast.success('Đã sao chép lệnh cURL vào clipboard');
      setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error('Không thể sao chép lệnh');
    }
  };

  return (
    <div className="relative rounded-control border border-slate-200 bg-slate-900 p-3 text-slate-100 font-mono text-xs dark:border-slate-800">
      <div className="flex items-center justify-between pb-1.5 border-b border-slate-800 mb-2">
        <span className="text-[11px] font-sans font-medium text-slate-400">
          {label || 'cURL Command'}
        </span>
        <button
          type="button"
          onClick={handleCopy}
          className="inline-flex items-center gap-1 rounded bg-slate-800 px-2 py-0.5 text-[11px] font-sans text-slate-300 hover:bg-slate-700 hover:text-white transition-colors"
          title="Sao chép cURL"
        >
          {copied ? <Check className="h-3 w-3 text-teal-400" /> : <Copy className="h-3 w-3" />}
          <span>{copied ? 'Đã chép' : 'Sao chép'}</span>
        </button>
      </div>
      <pre className="overflow-x-auto whitespace-pre-wrap break-all text-teal-300 select-all leading-relaxed">
        {curl}
      </pre>
    </div>
  );
}
