import React from 'react';
import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import {
  ShieldCheck,
  Zap,
  ArrowRight,
  Database,
  Server,
  Terminal,
  ShoppingBag,
  Sparkles,
} from 'lucide-react';
import { api } from '@/lib/api';
import type { Product, Category, ProductListResponse } from '@/types';
import { ProductCard } from '@/components/shop/ProductCard';
import { Skeleton } from '@/components/ui/skeleton';
import { Button } from '@/components/ui/button';

export function HomePage(): React.JSX.Element {
  // Query popular products
  const { data: productsData, isLoading: loadingProducts } = useQuery({
    queryKey: ['home-popular-products'],
    queryFn: async () => {
      const res = await api.get<ProductListResponse>('/api/products?page=1&page_size=8');
      return res.data;
    },
  });

  // Query categories
  const { data: categoriesData } = useQuery({
    queryKey: ['categories'],
    queryFn: async () => {
      const res = await api.get<Category[]>('/api/categories');
      return res.data;
    },
  });

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8 space-y-12">
      {/* Hero Section */}
      <section className="relative overflow-hidden rounded-card border border-slate-200 bg-white p-8 sm:p-12 shadow-sm dark:border-slate-800 dark:bg-slate-900">
        <div className="absolute top-0 right-0 -mr-16 -mt-16 h-64 w-64 rounded-full bg-teal-500/10 blur-3xl pointer-events-none" />

        <div className="relative z-10 max-w-2xl space-y-4">
          <div className="inline-flex items-center gap-2 rounded-control bg-teal-50 px-3 py-1 text-xs font-semibold text-teal-700 ring-1 ring-inset ring-teal-600/20 dark:bg-teal-950/60 dark:text-teal-300">
            <Sparkles className="h-3.5 w-3.5" />
            <span>Phòng Lab Bảo Mật & Tăng Tốc Toàn Diện</span>
          </div>

          <h1 className="text-3xl font-bold tracking-tight text-slate-900 sm:text-4xl lg:text-5xl dark:text-white leading-tight">
            Mua sắm thông minh, <br />
            <span className="text-teal-600 dark:text-teal-400">Thực chiến WAF & CDN</span>
          </h1>

          <p className="text-base text-slate-600 dark:text-slate-300 leading-relaxed">
            LabShop là hệ thống kép kết hợp cửa hàng thương mại điện tử thực tế và bảng điều khiển
            phân tích HTTP trực quan. Giúp kỹ sư bảo mật và DevOps kiểm chứng hành vi của CMC WAAP
            và Multi-CDN trên từng yêu cầu.
          </p>

          <div className="flex flex-wrap items-center gap-3 pt-2">
            <Link to="/products">
              <Button size="lg" className="gap-2">
                <ShoppingBag className="h-4 w-4" />
                <span>Khám phá Cửa hàng</span>
                <ArrowRight className="h-4 w-4" />
              </Button>
            </Link>

            <Link to="/tutorial">
              <Button variant="outline" size="lg" className="gap-2">
                <Terminal className="h-4 w-4 text-teal-600 dark:text-teal-400" />
                <span>Kịch bản Thực hành (Tutorial)</span>
              </Button>
            </Link>
          </div>
        </div>

        {/* Feature Highlights Grid */}
        <div className="mt-8 pt-8 border-t border-slate-100 dark:border-slate-800 grid grid-cols-2 sm:grid-cols-4 gap-4 text-xs font-mono">
          <div className="flex items-center gap-2 text-slate-600 dark:text-slate-300">
            <Server className="h-4 w-4 text-blue-500 shrink-0" />
            <span>Origin Go + SQLite</span>
          </div>
          <div className="flex items-center gap-2 text-slate-600 dark:text-slate-300">
            <Database className="h-4 w-4 text-teal-500 shrink-0" />
            <span>Multi-CDN Edge Caching</span>
          </div>
          <div className="flex items-center gap-2 text-slate-600 dark:text-slate-300">
            <ShieldCheck className="h-4 w-4 text-rose-500 shrink-0" />
            <span>CMC WAAP / Core Rules</span>
          </div>
          <div className="flex items-center gap-2 text-slate-600 dark:text-slate-300">
            <Zap className="h-4 w-4 text-amber-500 shrink-0" />
            <span>Lab Console (Ctrl+Shift+L)</span>
          </div>
        </div>
      </section>

      {/* Category List */}
      {categoriesData && categoriesData.length > 0 && (
        <section className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-bold tracking-tight text-slate-900 dark:text-white">
              Danh mục Thiết bị
            </h2>
            <Link
              to="/products"
              className="text-xs font-semibold text-teal-600 hover:underline dark:text-teal-400"
            >
              Xem tất cả
            </Link>
          </div>

          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-3">
            {categoriesData.map((cat) => (
              <Link
                key={cat.id}
                to={`/products?category=${encodeURIComponent(cat.id)}`}
                className="flex flex-col items-center justify-center p-4 rounded-card border border-slate-200 bg-white hover:border-teal-500 hover:shadow-sm transition-all text-center dark:border-slate-800 dark:bg-slate-900 group"
              >
                <span className="font-semibold text-sm text-slate-800 dark:text-slate-200 group-hover:text-teal-600 dark:group-hover:text-teal-400">
                  {cat.name}
                </span>
                <span className="text-[11px] text-slate-400 mt-1 line-clamp-1">
                  {cat.description || 'Khám phá sản phẩm'}
                </span>
              </Link>
            ))}
          </div>
        </section>
      )}

      {/* Popular Products Grid */}
      <section className="space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-xl font-bold tracking-tight text-slate-900 dark:text-white">
              Sản phẩm Nổi bật
            </h2>
            <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
              Dữ liệu được truy vấn qua API Origin và có thể quan sát cache trên Multi-CDN
            </p>
          </div>
          <Link
            to="/products"
            className="text-xs font-semibold text-teal-600 hover:underline dark:text-teal-400"
          >
            Xem danh mục đầy đủ ({productsData?.total || 40} sản phẩm)
          </Link>
        </div>

        {loadingProducts ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
            {Array.from({ length: 8 }).map((_, i) => (
              <div key={i} className="space-y-3 rounded-card border p-4">
                <Skeleton className="aspect-square w-full rounded-control" />
                <Skeleton className="h-4 w-3/4" />
                <Skeleton className="h-4 w-1/2" />
              </div>
            ))}
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
            {productsData?.items?.map((product: Product) => (
              <ProductCard key={product.id} product={product} />
            ))}
          </div>
        )}
      </section>

      {/* Lab Observability Banner */}
      <section className="rounded-card border border-teal-200 bg-teal-50/50 p-6 sm:p-8 dark:border-teal-900/50 dark:bg-teal-950/20">
        <div className="flex flex-col md:flex-row items-center justify-between gap-6">
          <div className="space-y-2 text-center md:text-left">
            <h3 className="text-lg font-bold text-slate-900 dark:text-white">
              Bắt đầu bài thực hành CMC WAAP & Multi-CDN
            </h3>
            <p className="text-sm text-slate-600 dark:text-slate-300 max-w-xl">
              Thực hành 18 kịch bản chuẩn (P0) từ Onboard site, thiết lập Cache Rules, xử lý
              rò rỉ dữ liệu cá nhân cho đến kiểm tra bộ luật WAF và Rate Limit.
            </p>
          </div>
          <Link to="/tutorial" className="shrink-0">
            <Button className="gap-2">
              <Terminal className="h-4 w-4" />
              <span>Xem các Kịch bản Lab</span>
            </Button>
          </Link>
        </div>
      </section>
    </div>
  );
}
