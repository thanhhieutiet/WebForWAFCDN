import React from 'react';
import { Star } from 'lucide-react';

interface RatingStarsProps {
  rating: number; // e.g. 4.5
  max?: number;
  count?: number;
  size?: 'sm' | 'md' | 'lg';
  showNumber?: boolean;
}

export function RatingStars({
  rating,
  max = 5,
  count,
  size = 'sm',
  showNumber = false,
}: RatingStarsProps): React.JSX.Element {
  const starSizes = {
    sm: 'h-3.5 w-3.5',
    md: 'h-4 w-4',
    lg: 'h-5 w-5',
  };

  const starClass = starSizes[size];

  return (
    <div className="inline-flex items-center gap-1" aria-label={`Đánh giá ${rating} trên ${max} sao`}>
      <div className="flex items-center text-amber-400">
        {Array.from({ length: max }).map((_, idx) => {
          const filled = idx + 1 <= Math.round(rating);
          return (
            <Star
              key={idx}
              className={`${starClass} ${
                filled ? 'fill-amber-400 text-amber-400' : 'text-slate-300 dark:text-slate-600'
              }`}
              aria-hidden="true"
            />
          );
        })}
      </div>

      {showNumber && (
        <span className="text-xs font-semibold text-slate-700 dark:text-slate-300 ml-0.5">
          {rating.toFixed(1)}
        </span>
      )}

      {count !== undefined && (
        <span className="text-xs text-slate-400 dark:text-slate-500">({count})</span>
      )}
    </div>
  );
}
