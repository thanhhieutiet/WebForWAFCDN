import React, { useState } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  ShoppingBag,
  ArrowLeft,
  Check,
  AlertTriangle,
  Star,
  MessageSquare,
  ShieldAlert,
  Clock,
  Send,
} from 'lucide-react';
import { api, ApiError } from '@/lib/api';
import type { Product, Review } from '@/types';
import { formatVND, formatDateTime } from '@/lib/formatters';
import { RatingStars } from '@/components/shop/RatingStars';
import { QuantityStepper } from '@/components/shop/QuantityStepper';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Skeleton } from '@/components/ui/skeleton';
import { useCart } from '@/lib/cartContext';
import { useAuth } from '@/lib/authContext';
import { toast } from 'sonner';
import { getErrorMessage } from '@/lib/messages';

export function ProductDetailPage(): React.JSX.Element {
  const { slug } = useParams<{ slug: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { updateQuantity, cart } = useCart();
  const { isAuthenticated } = useAuth();

  const [quantity, setQuantity] = useState<number>(1);
  const [adding, setAdding] = useState(false);
  const [added, setAdded] = useState(false);

  // Review form state
  const [reviewRating, setReviewRating] = useState<number>(5);
  const [reviewComment, setReviewComment] = useState<string>('');
  const [reviewError, setReviewError] = useState<string | null>(null);

  // Fallback image
  const [imgSrc, setImgSrc] = useState<string>('/static/img/products/p01.jpg');

  // Query product detail
  const { data: product, isLoading, isError, error } = useQuery({
    queryKey: ['product', slug],
    queryFn: async () => {
      const res = await api.get<Product>(`/api/products/${slug}`);
      setImgSrc(res.data.image_path || '/static/img/products/p01.jpg');
      return res.data;
    },
    enabled: !!slug,
  });

  // Query reviews
  const { data: reviewsData, isLoading: loadingReviews } = useQuery({
    queryKey: ['product-reviews', slug],
    queryFn: async () => {
      const res = await api.get<{ items: Review[]; total: number }>(
        `/api/products/${slug}/reviews`
      );
      return res.data;
    },
    enabled: !!slug,
  });

  // Submit review mutation
  const reviewMutation = useMutation({
    mutationFn: async () => {
      setReviewError(null);
      const res = await api.post(`/api/products/${slug}/reviews`, {
        rating: reviewRating,
        comment: reviewComment.trim(),
      });
      return res.data;
    },
    onSuccess: () => {
      toast.success('Đã gửi đánh giá sản phẩm thành công!');
      setReviewComment('');
      queryClient.invalidateQueries({ queryKey: ['product-reviews', slug] });
      queryClient.invalidateQueries({ queryKey: ['product', slug] });
    },
    onError: (err: any) => {
      const msg = err instanceof ApiError ? err.message : getErrorMessage('unknown_error');
      setReviewError(msg);
      toast.error('Gửi đánh giá không thành công', { description: msg });
    },
  });

  const isOutOfStock = (product?.stock ?? 0) <= 0;
  const isLowStock = (product?.stock ?? 0) > 0 && (product?.stock ?? 0) <= 5;

  const handleAddToCart = async () => {
    if (!isAuthenticated) {
      toast.info('Vui lòng đăng nhập để thêm sản phẩm vào giỏ hàng');
      navigate('/login');
      return;
    }
    if (!product || isOutOfStock) return;

    try {
      setAdding(true);
      const currentItem = cart?.items.find((i) => i.product_id === product.id);
      const currentQty = currentItem?.quantity || 0;
      const nextQty = Math.min(10, currentQty + quantity);

      await updateQuantity(product.id, nextQty);
      setAdded(true);
      toast.success(`Đã thêm ${quantity} x "${product.name}" vào giỏ hàng`);
      setTimeout(() => setAdded(false), 2000);
    } catch {
      // Handled in cart context
    } finally {
      setAdding(false);
    }
  };

  const reviews = reviewsData?.items || [];

  // Calculate star rating distribution from actual reviews or estimate
  const ratingDistribution = [5, 4, 3, 2, 1].map((stars) => {
    const count = reviews.filter((r) => r.rating === stars).length;
    const percent = reviews.length > 0 ? (count / reviews.length) * 100 : 0;
    return { stars, count, percent };
  });

  if (isLoading) {
    return (
      <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8 space-y-8">
        <Skeleton className="h-6 w-32" />
        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
          <Skeleton className="aspect-square w-full rounded-card" />
          <div className="space-y-4">
            <Skeleton className="h-8 w-3/4" />
            <Skeleton className="h-6 w-1/3" />
            <Skeleton className="h-20 w-full" />
            <Skeleton className="h-10 w-40" />
          </div>
        </div>
      </div>
    );
  }

  if (isError || !product) {
    return (
      <div className="mx-auto max-w-7xl px-4 py-16 sm:px-6 lg:px-8 text-center space-y-4">
        <AlertTriangle className="mx-auto h-12 w-12 text-rose-500" />
        <h2 className="text-xl font-bold text-slate-900 dark:text-white">
          Không tìm thấy thông tin sản phẩm
        </h2>
        <p className="text-sm text-slate-500 dark:text-slate-400">
          {(error as any)?.message || 'Sản phẩm có thể đã ngừng kinh doanh hoặc đường dẫn sai.'}
        </p>
        <Link to="/products">
          <Button variant="outline" className="gap-2">
            <ArrowLeft className="h-4 w-4" />
            <span>Quay lại Cửa hàng</span>
          </Button>
        </Link>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8 space-y-12">
      {/* Back button */}
      <Link
        to="/products"
        className="inline-flex items-center gap-1.5 text-xs font-semibold text-slate-600 hover:text-teal-600 dark:text-slate-400 dark:hover:text-teal-400 transition-colors"
      >
        <ArrowLeft className="h-4 w-4" />
        <span>Quay lại danh mục sản phẩm</span>
      </Link>

      {/* Main Product Info (Image + Details) */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-8 lg:gap-12">
        {/* Left: Product Image */}
        <div className="aspect-square overflow-hidden rounded-card border border-slate-200 bg-white p-4 shadow-sm dark:border-slate-800 dark:bg-slate-900 flex items-center justify-center">
          <img
            src={imgSrc}
            alt={product.name}
            onError={() => setImgSrc('/static/img/products/p01.jpg')}
            className="h-full w-full object-contain object-center rounded-control"
          />
        </div>

        {/* Right: Buy Box & Specs */}
        <div className="flex flex-col justify-between space-y-6">
          <div className="space-y-4">
            <div className="text-xs font-semibold uppercase tracking-wider text-teal-600 dark:text-teal-400">
              {product.category_name || 'Thiết bị công nghệ'}
            </div>

            <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-slate-900 dark:text-white">
              {product.name}
            </h1>

            {/* Rating summary */}
            <div className="flex items-center gap-3">
              <RatingStars rating={product.rating_avg} count={product.rating_count} size="md" showNumber />
              <span className="text-slate-300 dark:text-slate-700">|</span>
              <span className="text-xs text-slate-500 dark:text-slate-400">
                Mã SP: <span className="font-mono">{product.id}</span>
              </span>
            </div>

            {/* Price */}
            <div className="text-3xl font-bold tracking-tight text-slate-900 dark:text-white">
              {formatVND(product.price)}
            </div>

            {/* Stock indicator badge */}
            <div>
              {isOutOfStock ? (
                <span className="inline-flex items-center gap-1.5 rounded-control bg-rose-50 px-2.5 py-1 text-xs font-semibold text-rose-700 ring-1 ring-inset ring-rose-600/20 dark:bg-rose-950/60 dark:text-rose-300">
                  <AlertTriangle className="h-3.5 w-3.5" />
                  Hết hàng tạm thời
                </span>
              ) : isLowStock ? (
                <span className="inline-flex items-center gap-1.5 rounded-control bg-amber-50 px-2.5 py-1 text-xs font-semibold text-amber-700 ring-1 ring-inset ring-amber-600/20 dark:bg-amber-950/60 dark:text-amber-300">
                  <Clock className="h-3.5 w-3.5" />
                  Sắp hết hàng (Chỉ còn {product.stock} sản phẩm)
                </span>
              ) : (
                <span className="inline-flex items-center gap-1.5 rounded-control bg-teal-50 px-2.5 py-1 text-xs font-semibold text-teal-700 ring-1 ring-inset ring-teal-600/20 dark:bg-teal-950/60 dark:text-teal-300">
                  <Check className="h-3.5 w-3.5" />
                  Còn hàng trong kho ({product.stock} sản phẩm)
                </span>
              )}
            </div>

            <p className="text-sm leading-relaxed text-slate-600 dark:text-slate-300 pt-2 border-t border-slate-100 dark:border-slate-800">
              {product.description}
            </p>
          </div>

          {/* Action Box */}
          <div className="space-y-4 pt-6 border-t border-slate-200 dark:border-slate-800">
            <div className="flex items-center gap-4">
              <label className="text-xs font-semibold text-slate-700 dark:text-slate-300">
                Số lượng:
              </label>
              <QuantityStepper
                value={quantity}
                min={1}
                max={Math.min(10, Math.max(1, product.stock))}
                onChange={setQuantity}
                disabled={isOutOfStock}
              />
            </div>

            <div className="flex gap-3">
              <Button
                size="lg"
                disabled={isOutOfStock || adding}
                onClick={handleAddToCart}
                className="flex-1 gap-2"
              >
                {added ? <Check className="h-5 w-5" /> : <ShoppingBag className="h-5 w-5" />}
                <span>{added ? 'Đã thêm vào giỏ' : 'Thêm vào giỏ hàng'}</span>
              </Button>
            </div>
          </div>
        </div>
      </div>

      {/* Reviews & Ratings Section */}
      <section className="space-y-8 pt-8 border-t border-slate-200 dark:border-slate-800">
        <div>
          <h2 className="text-xl font-bold tracking-tight text-slate-900 dark:text-white">
            Đánh giá & Nhận xét
          </h2>
          <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
            Phản hồi thực tế từ những khách hàng đã mua và nhận sản phẩm thành công
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          {/* Star Distribution Summary */}
          <div className="space-y-4 rounded-card border border-slate-200 bg-white p-6 dark:border-slate-800 dark:bg-slate-900">
            <div className="text-center space-y-1">
              <div className="text-4xl font-extrabold text-slate-900 dark:text-white">
                {product.rating_avg.toFixed(1)}
              </div>
              <RatingStars rating={product.rating_avg} size="md" />
              <div className="text-xs text-slate-500">
                Dựa trên {product.rating_count} lượt đánh giá
              </div>
            </div>

            <div className="space-y-2 pt-4 border-t border-slate-100 dark:border-slate-800 text-xs">
              {ratingDistribution.map((row) => (
                <div key={row.stars} className="flex items-center gap-2">
                  <span className="w-12 text-slate-600 dark:text-slate-400 flex items-center gap-1 font-mono">
                    {row.stars} <Star className="h-3 w-3 fill-amber-400 text-amber-400 inline" />
                  </span>
                  <div className="flex-1 h-2 rounded-full bg-slate-100 dark:bg-slate-800 overflow-hidden">
                    <div
                      className="h-full bg-amber-400 rounded-full"
                      style={{ width: `${row.percent}%` }}
                    />
                  </div>
                  <span className="w-8 text-right font-mono text-slate-400">{row.count}</span>
                </div>
              ))}
            </div>
          </div>

          {/* Reviews List & Submission Form */}
          <div className="md:col-span-2 space-y-6">
            {/* Submit Review Form */}
            <div className="rounded-card border border-slate-200 bg-white p-6 dark:border-slate-800 dark:bg-slate-900 space-y-4">
              <h3 className="text-sm font-semibold text-slate-900 dark:text-white flex items-center gap-2">
                <MessageSquare className="h-4 w-4 text-teal-600 dark:text-teal-400" />
                Viết đánh giá của bạn
              </h3>

              {!isAuthenticated ? (
                <div className="rounded-control bg-slate-50 p-4 text-xs text-slate-600 dark:bg-slate-800/50 dark:text-slate-300">
                  Vui lòng{' '}
                  <Link to="/login" className="font-semibold text-teal-600 underline dark:text-teal-400">
                    đăng nhập
                  </Link>{' '}
                  để có thể gửi đánh giá cho sản phẩm này.
                </div>
              ) : (
                <form
                  onSubmit={(e) => {
                    e.preventDefault();
                    reviewMutation.mutate();
                  }}
                  className="space-y-4"
                >
                  <div className="space-y-1.5">
                    <label className="text-xs font-semibold text-slate-700 dark:text-slate-300">
                      Chọn mức đánh giá sao:
                    </label>
                    <div className="flex items-center gap-2">
                      {[1, 2, 3, 4, 5].map((val) => (
                        <button
                          key={val}
                          type="button"
                          onClick={() => setReviewRating(val)}
                          className="p-1 hover:scale-110 transition-transform"
                          aria-label={`${val} sao`}
                        >
                          <Star
                            className={`h-6 w-6 ${
                              val <= reviewRating
                                ? 'fill-amber-400 text-amber-400'
                                : 'text-slate-300 dark:text-slate-700'
                            }`}
                          />
                        </button>
                      ))}
                      <span className="text-xs font-mono font-semibold text-slate-700 dark:text-slate-300 ml-2">
                        {reviewRating} sao
                      </span>
                    </div>
                  </div>

                  <div className="space-y-1.5">
                    <label className="text-xs font-semibold text-slate-700 dark:text-slate-300">
                      Nhận xét chi tiết:
                    </label>
                    <Textarea
                      required
                      placeholder="Chia sẻ trải nghiệm của bạn về chất lượng sản phẩm..."
                      value={reviewComment}
                      onChange={(e) => setReviewComment(e.target.value)}
                      error={!!reviewError}
                      className="text-xs"
                      rows={3}
                    />
                  </div>

                  {reviewError && (
                    <div className="rounded-control bg-rose-50 p-3 text-xs text-rose-700 dark:bg-rose-950/50 dark:text-rose-300 flex items-start gap-2">
                      <ShieldAlert className="h-4 w-4 shrink-0 text-rose-500 mt-0.5" />
                      <span>{reviewError}</span>
                    </div>
                  )}

                  <div className="flex items-center justify-between pt-1">
                    <span className="text-[11px] text-slate-400">
                      * Yêu cầu: Bạn phải có đơn hàng đã giao thành công (DELIVERED) chứa sản phẩm này.
                    </span>
                    <Button
                      type="submit"
                      size="sm"
                      loading={reviewMutation.isPending}
                      disabled={!reviewComment.trim()}
                      className="gap-1.5"
                    >
                      <Send className="h-3.5 w-3.5" />
                      <span>Gửi đánh giá</span>
                    </Button>
                  </div>
                </form>
              )}
            </div>

            {/* Existing Reviews List */}
            <div className="space-y-3">
              <h4 className="text-xs font-semibold text-slate-700 dark:text-slate-300">
                Các nhận xét ({reviews.length})
              </h4>

              {loadingReviews ? (
                <div className="space-y-3">
                  <Skeleton className="h-20 w-full rounded-control" />
                  <Skeleton className="h-20 w-full rounded-control" />
                </div>
              ) : reviews.length === 0 ? (
                <div className="rounded-control border border-dashed border-slate-300 bg-slate-50 p-6 text-center text-xs text-slate-500 dark:border-slate-800 dark:bg-slate-900">
                  Chưa có đánh giá nào cho sản phẩm này. Hãy là người đầu tiên trải nghiệm!
                </div>
              ) : (
                <div className="space-y-3">
                  {reviews.map((rev) => (
                    <div
                      key={rev.id}
                      className="rounded-card border border-slate-200 bg-white p-4 dark:border-slate-800 dark:bg-slate-900 space-y-2"
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <span className="text-xs font-bold text-slate-900 dark:text-white">
                            {rev.user_name || 'Khách hàng'}
                          </span>
                          <RatingStars rating={rev.rating} size="sm" />
                        </div>
                        <span className="text-[11px] text-slate-400 font-mono">
                          {formatDateTime(rev.created_at)}
                        </span>
                      </div>
                      <p className="text-xs text-slate-700 dark:text-slate-300 leading-relaxed">
                        {rev.comment}
                      </p>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      </section>
    </div>
  );
}
