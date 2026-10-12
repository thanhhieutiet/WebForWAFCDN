import React from 'react';
import { Minus, Plus } from 'lucide-react';

interface QuantityStepperProps {
  value: number;
  min?: number;
  max?: number;
  onChange: (val: number) => void;
  disabled?: boolean;
}

export function QuantityStepper({
  value,
  min = 1,
  max = 10,
  onChange,
  disabled = false,
}: QuantityStepperProps): React.JSX.Element {
  const handleDecrement = () => {
    if (value > min) {
      onChange(value - 1);
    }
  };

  const handleIncrement = () => {
    if (value < max) {
      onChange(value + 1);
    }
  };

  return (
    <div className="inline-flex items-center rounded-control border border-slate-300 bg-white dark:border-slate-700 dark:bg-slate-900 shadow-sm">
      <button
        type="button"
        onClick={handleDecrement}
        disabled={disabled || value <= min}
        aria-label="Giảm số lượng"
        className="flex h-8 w-8 items-center justify-center text-slate-500 hover:text-slate-800 disabled:opacity-40 disabled:hover:text-slate-500 dark:text-slate-400 dark:hover:text-slate-200 transition-colors"
      >
        <Minus className="h-3.5 w-3.5" />
      </button>

      <span
        className="w-10 text-center font-mono text-sm font-semibold text-slate-900 dark:text-white select-none"
        aria-live="polite"
      >
        {value}
      </span>

      <button
        type="button"
        onClick={handleIncrement}
        disabled={disabled || value >= max}
        aria-label="Tăng số lượng"
        className="flex h-8 w-8 items-center justify-center text-slate-500 hover:text-slate-800 disabled:opacity-40 disabled:hover:text-slate-500 dark:text-slate-400 dark:hover:text-slate-200 transition-colors"
      >
        <Plus className="h-3.5 w-3.5" />
      </button>
    </div>
  );
}
