import React, { useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import {
  Search,
  SlidersHorizontal,
  X,
  RotateCcw,
  AlertCircle,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react';
import { api } from '@/lib/api';
import type { Category, ProductListResponse } from '@/types';
import { ProductCard } from '@/components/shop/ProductCard';
import { Skeleton } from '@/components/ui/skeleton';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

export function ProductsPage(): React.JSX.Element {
  const [searchParams, setSearchParams] = useSearchParams();
  const [mobileFilterOpen, setMobileFilterOpen] = useState(false);

  // URL query state
  const q = searchParams.get('q') || '';
  const category = searchParams.get('category') || '';
  const sort = searchParams.get('sort') || 'newest';
  const minPrice = searchParams.get('min_price') || '';
  const maxPrice = searchParams.get('max_price') || '';
  const page = parseInt(searchParams.get('page') || '1', 10);
  const pageSize = 12;

  // Local state for search & filter inputs before applying
  const [searchInput, setSearchInput] = useState(q);
  const [minPriceInput, setMinPriceInput] = useState(minPrice);
  const [maxPriceInput, setMaxPriceInput] = useState(maxPrice);

  // Query categories
  const { data: categories } = useQuery({
    queryKey: ['categories'],
    queryFn: async () => {
      const res = await api.get<{ categories: Category[] }>('/api/categories');
      return res.data.categories || [];
    },
  });

  // Query products with active filters
  const { data, isLoading, isError, error, refetch } = useQuery({
    queryKey: ['products', q, category, sort, minPrice, maxPrice, page],
    queryFn: async () => {
      const params = new URLSearchParams();
      if (q) params.set('q', q);
      if (category) params.set('category', category);
      if (sort) params.set('sort', sort);
      if (minPrice) params.set('min_price', minPrice);
      if (maxPrice) params.set('max_price', maxPrice);
      params.set('page', String(page));
      params.set('page_size', String(pageSize));

      const res = await api.get<ProductListResponse>(`/api/products?${params.toString()}`);
      return res.data;
    },
  });

  const updateParam = (key: string, val: string) => {
    const next = new URLSearchParams(searchParams);
    if (val) {
      next.set(key, val);
    } else {
      next.delete(key);
    }
    // reset to page 1 on filter change
    if (key !== 'page') {
      next.set('page', '1');
    }
    setSearchParams(next);
  };

  const handleApplyFilter = (e: React.FormEvent) => {
    e.preventDefault();
    const next = new URLSearchParams(searchParams);
    if (searchInput.trim()) next.set('q', searchInput.trim());
    else next.delete('q');

    if (minPriceInput) next.set('min_price', minPriceInput);
    else next.delete('min_price');

    if (maxPriceInput) next.set('max_price', maxPriceInput);
    else next.delete('max_price');

    next.set('page', '1');
    setSearchParams(next);
    setMobileFilterOpen(false);
  };

  const handleClearFilters = () => {
    setSearchInput('');
    setMinPriceInput('');
    setMaxPriceInput('');
    setSearchParams(new URLSearchParams());
    setMobileFilterOpen(false);
  };

  const hasActiveFilters = !!(q || category || minPrice || maxPrice || (sort && sort !== 'newest'));
  const totalPages = data ? Math.ceil(data.total / pageSize) : 1;

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      {/* Header & Filter Controls Bar */}
      <div className="mb-6 flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-6 border-b border-slate-200 dark:border-slate-800">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-slate-900 dark:text-white">
            Danh mục Sản phẩm
          </h1>
          <p className="mt-1 text-xs text-slate-500 dark:text-slate-400">
            {data ? `Hiển thị ${data.items.length} trên tổng số ${data.total} sản phẩm` : 'Đang tải danh mục...'}
          </p>
        </div>

        <div className="flex items-center gap-3">
          {/* Mobile Filter Toggle */}
          <Button
            variant="outline"
            size="sm"
            onClick={() => setMobileFilterOpen(true)}
            className="md:hidden gap-1.5"
          >
            <SlidersHorizontal className="h-4 w-4" />
            <span>Bộ lọc</span>
            {hasActiveFilters && <span className="h-2 w-2 rounded-full bg-teal-600" />}
          </Button>

          {/* Sort Selector */}
          <div className="flex items-center gap-2 text-xs">
            <span className="text-slate-500 hidden sm:inline">Sắp xếp:</span>
            <select
              value={sort}
              onChange={(e) => updateParam('sort', e.target.value)}
              className="h-9 rounded-control border border-slate-300 bg-white px-2.5 py-1 text-xs text-slate-800 shadow-sm focus:outline-none focus:ring-2 focus:ring-teal-600 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-200"
            >
              <option value="newest">Mới nhất</option>
              <option value="price_asc">Giá: Thấp đến Cao</option>
              <option value="price_desc">Giá: Cao đến Thấp</option>
              <option value="rating">Đánh giá cao nhất</option>
            </select>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
        {/* Left Filter Sidebar (Desktop) */}
        <aside className="hidden md:block space-y-6">
          <form onSubmit={handleApplyFilter} className="space-y-6">
            {/* Search Input */}
            <div className="space-y-1.5">
              <label className="text-xs font-semibold text-slate-700 dark:text-slate-300">
                Tìm kiếm từ khóa
              </label>
              <div className="relative">
                <Input
                  type="search"
                  placeholder="Tên hoặc mô tả..."
                  value={searchInput}
                  onChange={(e) => setSearchInput(e.target.value)}
                  className="pl-8 text-xs h-9"
                />
                <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-slate-400 pointer-events-none" />
              </div>
            </div>

            {/* Categories */}
            <div className="space-y-2">
              <div className="text-xs font-semibold text-slate-700 dark:text-slate-300">
                Danh mục
              </div>
              <div className="space-y-1 text-xs">
                <button
                  type="button"
                  onClick={() => updateParam('category', '')}
                  className={`block w-full text-left px-2.5 py-1.5 rounded-control transition-colors ${
                    !category
                      ? 'bg-teal-50 text-teal-700 font-semibold dark:bg-teal-950/60 dark:text-teal-300'
                      : 'text-slate-600 hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-800'
                  }`}
                >
                  Tất cả danh mục
                </button>
                {categories?.map((cat) => (
                  <button
                    key={cat.id}
                    type="button"
                    onClick={() => updateParam('category', cat.id)}
                    className={`block w-full text-left px-2.5 py-1.5 rounded-control transition-colors ${
                      category === cat.id
                        ? 'bg-teal-50 text-teal-700 font-semibold dark:bg-teal-950/60 dark:text-teal-300'
                        : 'text-slate-600 hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-800'
                    }`}
                  >
                    {cat.name}
                  </button>
                ))}
              </div>
            </div>

            {/* Price Range Filter */}
            <div className="space-y-2">
              <div className="text-xs font-semibold text-slate-700 dark:text-slate-300">
                Khoảng giá (VND)
              </div>
              <div className="grid grid-cols-2 gap-2">
                <Input
                  type="number"
                  placeholder="Từ (VNĐ)"
                  value={minPriceInput}
                  onChange={(e) => setMinPriceInput(e.target.value)}
                  className="text-xs h-9"
                />
                <Input
                  type="number"
                  placeholder="Đến (VNĐ)"
                  value={maxPriceInput}
                  onChange={(e) => setMaxPriceInput(e.target.value)}
                  className="text-xs h-9"
                />
              </div>
            </div>

            <div className="flex gap-2 pt-2">
              <Button type="submit" size="sm" className="flex-1">
                Áp dụng
              </Button>
              {hasActiveFilters && (
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={handleClearFilters}
                  title="Xóa bộ lọc"
                >
                  <RotateCcw className="h-3.5 w-3.5" />
                </Button>
              )}
            </div>
          </form>
        </aside>

        {/* Product Grid Area (3 Columns) */}
        <div className="md:col-span-3 space-y-8">
          {/* Active Filter Badges */}
          {hasActiveFilters && (
            <div className="flex flex-wrap items-center gap-2 text-xs">
              <span className="text-slate-400">Đang lọc theo:</span>
              {q && (
                <span className="inline-flex items-center gap-1 rounded bg-slate-100 px-2 py-0.5 text-slate-800 dark:bg-slate-800 dark:text-slate-200">
                  Từ khóa: "{q}"
                  <button type="button" onClick={() => updateParam('q', '')}>
                    <X className="h-3 w-3 text-slate-400 hover:text-slate-600" />
                  </button>
                </span>
              )}
              {category && (
                <span className="inline-flex items-center gap-1 rounded bg-slate-100 px-2 py-0.5 text-slate-800 dark:bg-slate-800 dark:text-slate-200">
                  Danh mục: {categories?.find((c) => c.id === category)?.name || category}
                  <button type="button" onClick={() => updateParam('category', '')}>
                    <X className="h-3 w-3 text-slate-400 hover:text-slate-600" />
                  </button>
                </span>
              )}
              {minPrice && (
                <span className="inline-flex items-center gap-1 rounded bg-slate-100 px-2 py-0.5 text-slate-800 dark:bg-slate-800 dark:text-slate-200">
                  Từ: {Number(minPrice).toLocaleString()} ₫
                  <button type="button" onClick={() => updateParam('min_price', '')}>
                    <X className="h-3 w-3 text-slate-400 hover:text-slate-600" />
                  </button>
                </span>
              )}
              {maxPrice && (
                <span className="inline-flex items-center gap-1 rounded bg-slate-100 px-2 py-0.5 text-slate-800 dark:bg-slate-800 dark:text-slate-200">
                  Đến: {Number(maxPrice).toLocaleString()} ₫
                  <button type="button" onClick={() => updateParam('max_price', '')}>
                    <X className="h-3 w-3 text-slate-400 hover:text-slate-600" />
                  </button>
                </span>
              )}
              <button
                type="button"
                onClick={handleClearFilters}
                className="text-xs text-teal-600 hover:underline dark:text-teal-400 font-medium ml-1"
              >
                Xóa tất cả
              </button>
            </div>
          )}

          {/* Loading State */}
          {isLoading && (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
              {Array.from({ length: 6 }).map((_, i) => (
                <div key={i} className="space-y-3 rounded-card border p-4">
                  <Skeleton className="aspect-square w-full rounded-control" />
                  <Skeleton className="h-4 w-3/4" />
                  <Skeleton className="h-4 w-1/2" />
                </div>
              ))}
            </div>
          )}

          {/* Error State */}
          {isError && (
            <div className="rounded-card border border-rose-200 bg-rose-50 p-8 text-center dark:border-rose-900/50 dark:bg-rose-950/20">
              <AlertCircle className="mx-auto h-8 w-8 text-rose-500 mb-2" />
              <h3 className="text-sm font-semibold text-rose-800 dark:text-rose-300">
                Không thể tải danh sách sản phẩm
              </h3>
              <p className="mt-1 text-xs text-rose-600 dark:text-rose-400">
                {(error as any)?.message || 'Lỗi khi kết nối với máy chủ API Origin'}
              </p>
              <div className="mt-4">
                <Button size="sm" variant="outline" onClick={() => refetch()}>
                  Thử lại
                </Button>
              </div>
            </div>
          )}

          {/* Empty State */}
          {!isLoading && !isError && data?.items?.length === 0 && (
            <div className="rounded-card border border-dashed border-slate-300 bg-white p-12 text-center dark:border-slate-800 dark:bg-slate-900">
              <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-slate-100 dark:bg-slate-800 text-slate-400">
                <Search className="h-6 w-6" />
              </div>
              <h3 className="mt-4 text-base font-semibold text-slate-900 dark:text-white">
                Không tìm thấy sản phẩm phù hợp
              </h3>
              <p className="mt-1 text-xs text-slate-500 dark:text-slate-400 max-w-sm mx-auto">
                Không có sản phẩm nào khớp với tiêu chí tìm kiếm hoặc bộ lọc hiện tại của bạn.
              </p>
              <div className="mt-6">
                <Button size="sm" onClick={handleClearFilters}>
                  Xóa bộ lọc tìm kiếm
                </Button>
              </div>
            </div>
          )}

          {/* Product Grid */}
          {!isLoading && !isError && data && data.items.length > 0 && (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
              {data.items.map((prod) => (
                <ProductCard key={prod.id} product={prod} />
              ))}
            </div>
          )}

          {/* Pagination Controls */}
          {!isLoading && !isError && totalPages > 1 && (
            <div className="flex items-center justify-between border-t border-slate-200 dark:border-slate-800 pt-6">
              <Button
                variant="outline"
                size="sm"
                disabled={page <= 1}
                onClick={() => updateParam('page', String(page - 1))}
                className="gap-1"
              >
                <ChevronLeft className="h-4 w-4" />
                <span>Trang trước</span>
              </Button>

              <div className="text-xs font-mono text-slate-600 dark:text-slate-400">
                Trang <strong className="text-slate-900 dark:text-white">{page}</strong> / {totalPages}
              </div>

              <Button
                variant="outline"
                size="sm"
                disabled={page >= totalPages}
                onClick={() => updateParam('page', String(page + 1))}
                className="gap-1"
              >
                <span>Trang sau</span>
                <ChevronRight className="h-4 w-4" />
              </Button>
            </div>
          )}
        </div>
      </div>

      {/* Mobile Filter Sheet Modal */}
      {mobileFilterOpen && (
        <div className="fixed inset-0 z-50 flex md:hidden" role="dialog" aria-modal="true">
          <div
            className="fixed inset-0 bg-slate-900/60 backdrop-blur-sm"
            onClick={() => setMobileFilterOpen(false)}
          />
          <div className="relative ml-auto h-full w-full max-w-xs bg-white p-6 shadow-xl dark:bg-slate-900 overflow-y-auto">
            <div className="flex items-center justify-between pb-4 border-b border-slate-200 dark:border-slate-800 mb-4">
              <h2 className="text-base font-bold text-slate-900 dark:text-white">Bộ lọc sản phẩm</h2>
              <button
                type="button"
                onClick={() => setMobileFilterOpen(false)}
                className="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            <form onSubmit={handleApplyFilter} className="space-y-6">
              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-slate-700 dark:text-slate-300">
                  Từ khóa
                </label>
                <Input
                  type="search"
                  placeholder="Tên sản phẩm..."
                  value={searchInput}
                  onChange={(e) => setSearchInput(e.target.value)}
                  className="text-xs"
                />
              </div>

              <div className="space-y-2">
                <div className="text-xs font-semibold text-slate-700 dark:text-slate-300">
                  Danh mục
                </div>
                <div className="space-y-1 text-xs">
                  <button
                    type="button"
                    onClick={() => updateParam('category', '')}
                    className={`block w-full text-left px-2.5 py-1.5 rounded-control ${
                      !category ? 'bg-teal-50 text-teal-700 font-semibold' : 'text-slate-600'
                    }`}
                  >
                    Tất cả danh mục
                  </button>
                  {categories?.map((cat) => (
                    <button
                      key={cat.id}
                      type="button"
                      onClick={() => updateParam('category', cat.id)}
                      className={`block w-full text-left px-2.5 py-1.5 rounded-control ${
                        category === cat.id ? 'bg-teal-50 text-teal-700 font-semibold' : 'text-slate-600'
                      }`}
                    >
                      {cat.name}
                    </button>
                  ))}
                </div>
              </div>

              <div className="space-y-2">
                <div className="text-xs font-semibold text-slate-700 dark:text-slate-300">
                  Khoảng giá (VND)
                </div>
                <div className="grid grid-cols-2 gap-2">
                  <Input
                    type="number"
                    placeholder="Từ"
                    value={minPriceInput}
                    onChange={(e) => setMinPriceInput(e.target.value)}
                    className="text-xs"
                  />
                  <Input
                    type="number"
                    placeholder="Đến"
                    value={maxPriceInput}
                    onChange={(e) => setMaxPriceInput(e.target.value)}
                    className="text-xs"
                  />
                </div>
              </div>

              <div className="flex gap-2 pt-4 border-t border-slate-200 dark:border-slate-800">
                <Button type="submit" className="flex-1">
                  Áp dụng
                </Button>
                <Button type="button" variant="outline" onClick={handleClearFilters}>
                  Xóa
                </Button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
