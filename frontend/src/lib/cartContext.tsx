import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import type { CartView } from '@/types';
import { api, ApiError } from './api';
import { useAuth } from './authContext';
import { toast } from 'sonner';
import { getErrorMessage } from './messages';

interface CartContextType {
  cart: CartView | null;
  itemCount: number;
  isLoading: boolean;
  updateQuantity: (productId: string, quantity: number) => Promise<void>;
  removeItem: (productId: string) => Promise<void>;
  applyCoupon: (code: string) => Promise<void>;
  removeCoupon: () => Promise<void>;
  clearCart: () => Promise<void>;
  refreshCart: () => Promise<void>;
}

const CartContext = createContext<CartContextType | undefined>(undefined);

export function CartProvider({ children }: { children: React.ReactNode }): React.JSX.Element {
  const { isAuthenticated } = useAuth();
  const [cart, setCart] = useState<CartView | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(false);

  const refreshCart = useCallback(async () => {
    if (!isAuthenticated) {
      setCart(null);
      return;
    }
    try {
      setIsLoading(true);
      const res = await api.get<CartView>('/api/cart');
      setCart(res.data);
    } catch {
      // User may not have created cart yet or unauthorized
    } finally {
      setIsLoading(false);
    }
  }, [isAuthenticated]);

  useEffect(() => {
    refreshCart();
  }, [refreshCart]);

  // Optimistic quantity update with rollback
  const updateQuantity = async (productId: string, quantity: number) => {
    if (!cart) return;

    const previousCart = { ...cart, items: [...cart.items] };

    // Optimistic UI state
    const nextItems = cart.items
      .map((item) => (item.product_id === productId ? { ...item, quantity } : item))
      .filter((item) => item.quantity > 0);

    const nextCount = nextItems.reduce((acc, it) => acc + it.quantity, 0);

    setCart((prev) => (prev ? { ...prev, items: nextItems, item_count: nextCount } : null));

    try {
      const res = await api.put<CartView>(`/api/cart/items/${productId}`, { quantity });
      setCart(res.data);
    } catch (err: any) {
      // Rollback
      setCart(previousCart);
      const msg = err instanceof ApiError ? err.message : getErrorMessage('update_failed');
      toast.error('Cập nhật giỏ hàng thất bại', { description: msg });
      throw err;
    }
  };

  const removeItem = async (productId: string) => {
    if (!cart) return;
    const previousCart = { ...cart, items: [...cart.items] };

    setCart((prev) =>
      prev
        ? {
            ...prev,
            items: prev.items.filter((it) => it.product_id !== productId),
            item_count: prev.items.filter((it) => it.product_id !== productId).reduce((a, b) => a + b.quantity, 0),
          }
        : null
    );

    try {
      const res = await api.delete<CartView>(`/api/cart/items/${productId}`);
      setCart(res.data);
      toast.success('Đã xóa sản phẩm khỏi giỏ hàng');
    } catch (err: any) {
      setCart(previousCart);
      const msg = err instanceof ApiError ? err.message : getErrorMessage('remove_failed');
      toast.error('Không thể xóa sản phẩm', { description: msg });
      throw err;
    }
  };

  const applyCoupon = async (code: string) => {
    try {
      const res = await api.post<CartView>('/api/cart/coupon', { code });
      setCart(res.data);
      toast.success(`Đã áp dụng mã giảm giá ${code.toUpperCase()}`);
    } catch (err: any) {
      const msg = err instanceof ApiError ? err.message : getErrorMessage(err?.code || 'coupon_not_found');
      toast.error('Áp dụng mã thất bại', { description: msg });
      throw err;
    }
  };

  const removeCoupon = async () => {
    try {
      const res = await api.delete<CartView>('/api/cart/coupon');
      setCart(res.data);
      toast.info('Đã hủy áp dụng mã giảm giá');
    } catch (err: any) {
      const msg = err instanceof ApiError ? err.message : getErrorMessage('unknown_error');
      toast.error('Thao tác thất bại', { description: msg });
      throw err;
    }
  };

  const clearCart = async () => {
    try {
      await api.delete('/api/cart');
      setCart(null);
    } catch {
      // Ignored
    }
  };

  const itemCount = cart?.item_count ?? 0;

  return (
    <CartContext.Provider
      value={{
        cart,
        itemCount,
        isLoading,
        updateQuantity,
        removeItem,
        applyCoupon,
        removeCoupon,
        clearCart,
        refreshCart,
      }}
    >
      {children}
    </CartContext.Provider>
  );
}

export function useCart(): CartContextType {
  const context = useContext(CartContext);
  if (!context) {
    throw new Error('useCart must be used within a CartProvider');
  }
  return context;
}
