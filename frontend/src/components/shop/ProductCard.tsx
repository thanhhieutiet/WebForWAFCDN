import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { ShoppingBag, Check } from 'lucide-react';
import type { Product } from '@/types';
import { formatVND } from '@/lib/formatters';
import { RatingStars } from './RatingStars';
import { useCart } from '@/lib/cartContext';
import { useAuth } from '@/lib/authContext';
import { toast } from 'sonner';

interface ProductCardProps {
  product: Product;
}

export function ProductCard({ product }: ProductCardProps): React.JSX.Element {
  const { updateQuantity, cart } = useCart();
  const { isAuthenticated } = useAuth();
  const navigate = useNavigate();
  const [adding, setAdding] = useState(false);
  const [added, setAdded] = useState(false);

  // Fallback image handling
  const [imgSrc, setImgSrc] = useState(product.image_path || '/static/img/products/p01.jpg');

  const isOutOfStock = product.stock <= 0;
  const isLowStock = product.stock > 0 && product.stock <= 5;

  const handleQuickAdd = async (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();

    if (!isAuthenticated) {
      toast.info('Vui lòng đăng nhập để thêm sản phẩm vào giỏ hàng');
      navigate('/login');
      return;
    }

    if (isOutOfStock) {
      toast.error('Sản phẩm tạm thời hết hàng');
      return;
    }

    try {
      setAdding(true);
      const currentItem = cart?.items.find((i) => i.product_id === product.id);
      const nextQty = (currentItem?.quantity || 0) + 1;
      await updateQuantity(product.id, nextQty);
      setAdded(true);
      toast.success(`Đã thêm "${product.name}" vào giỏ`);
      setTimeout(() => setAdded(false), 2000);
    } catch {
      // Toast already handled by cartContext
    } finally {
      setAdding(false);
    }
  };

  return (
    <div className="group relative flex flex-col overflow-hidden rounded-card border border-slate-200 bg-white transition-all hover:shadow-md dark:border-slate-800 dark:bg-slate-900">
      {/* Product Image */}
      <Link to={`/products/${product.slug}`} className="aspect-square relative overflow-hidden bg-slate-100 dark:bg-slate-800 block">
        <img
          src={imgSrc}
          alt={product.name}
          loading="lazy"
          onError={() => setImgSrc('/static/img/products/p01.jpg')}
          className="h-full w-full object-cover object-center transition-transform duration-300 group-hover:scale-105"
        />

        {/* Stock badge */}
        <div className="absolute top-2 left-2 flex flex-col gap-1">
          {isOutOfStock ? (
            <span className="rounded bg-rose-600 px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider text-white shadow">
              Hết hàng
            </span>
          ) : isLowStock ? (
            <span className="rounded bg-amber-500 px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider text-white shadow">
              Chỉ còn {product.stock}
            </span>
          ) : null}
        </div>
      </Link>

      {/* Product Info */}
      <div className="flex flex-1 flex-col p-4 justify-between">
        <div>
          <div className="text-[11px] font-medium text-teal-600 dark:text-teal-400 mb-1">
            {product.category_name || 'Thiết bị'}
          </div>

          <Link
            to={`/products/${product.slug}`}
            className="line-clamp-2 text-sm font-semibold text-slate-900 hover:text-teal-600 dark:text-white dark:hover:text-teal-400 transition-colors"
            title={product.name}
          >
            {product.name}
          </Link>

          <p className="mt-1 line-clamp-2 text-xs text-slate-500 dark:text-slate-400">
            {product.description}
          </p>
        </div>

        <div className="mt-4 pt-3 border-t border-slate-100 dark:border-slate-800 flex items-center justify-between">
          <div>
            <div className="text-sm font-bold text-slate-900 dark:text-white">
              {formatVND(product.price)}
            </div>
            <RatingStars rating={product.rating_avg} count={product.rating_count} size="sm" />
          </div>

          <button
            type="button"
            onClick={handleQuickAdd}
            disabled={isOutOfStock || adding}
            className={`flex h-9 w-9 items-center justify-center rounded-control transition-colors shadow-sm focus:outline-none focus:ring-2 focus:ring-teal-600 disabled:opacity-40 disabled:cursor-not-allowed ${
              added
                ? 'bg-teal-600 text-white'
                : 'bg-slate-100 text-slate-800 hover:bg-teal-600 hover:text-white dark:bg-slate-800 dark:text-slate-200 dark:hover:bg-teal-600 dark:hover:text-white'
            }`}
            aria-label={`Thêm ${product.name} vào giỏ hàng`}
            title="Thêm nhanh vào giỏ"
          >
            {added ? <Check className="h-4 w-4" /> : <ShoppingBag className="h-4 w-4" />}
          </button>
        </div>
      </div>
    </div>
  );
}
