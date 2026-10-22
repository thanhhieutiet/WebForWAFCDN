import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Toaster } from 'sonner';

import { ThemeProvider } from '@/lib/themeContext';
import { AuthProvider } from '@/lib/authContext';
import { CartProvider } from '@/lib/cartContext';
import { AppLayout } from '@/components/layout/AppLayout';

import { HomePage } from '@/pages/HomePage';
import { ProductsPage } from '@/pages/ProductsPage';
import { ProductDetailPage } from '@/pages/ProductDetailPage';
import { CartPage } from '@/pages/CartPage';
import { CheckoutPage } from '@/pages/CheckoutPage';
import { OrdersPage } from '@/pages/OrdersPage';
import { OrderDetailPage } from '@/pages/OrderDetailPage';
import { LoginPage } from '@/pages/LoginPage';
import { RegisterPage } from '@/pages/RegisterPage';
import { SupportPage } from '@/pages/SupportPage';
import { AdminPage } from '@/pages/AdminPage';
import { TutorialPage } from '@/pages/TutorialPage';
import { ScenarioDetailPage } from '@/pages/ScenarioDetailPage';
import { DocsPage } from '@/pages/DocsPage';

import { ErrorBoundary } from '@/components/common/ErrorBoundary';
import { AuthGuard } from '@/components/common/AuthGuard';

// Initialize React Query Client
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
      staleTime: 30000,
    },
  },
});

export default function App(): React.JSX.Element {
  return (
    <ErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <ThemeProvider>
          <AuthProvider>
            <CartProvider>
              <Toaster position="top-right" richColors />
              <BrowserRouter>
                <Routes>
                  {/* Public Authentication Routes - Pure standalone layouts without Navbar/Footer/LabConsole */}
                  <Route path="/login" element={<LoginPage />} />
                  <Route path="/register" element={<RegisterPage />} />

                  {/* Protected Routes - Authentication Required with Full App Layout */}
                  <Route element={<AuthGuard />}>
                    <Route element={<AppLayout />}>
                      <Route path="/" element={<HomePage />} />
                      <Route path="/products" element={<ProductsPage />} />
                      <Route path="/products/:slug" element={<ProductDetailPage />} />
                      <Route path="/cart" element={<CartPage />} />
                      <Route path="/checkout" element={<CheckoutPage />} />
                      <Route path="/orders" element={<OrdersPage />} />
                      <Route path="/orders/:id" element={<OrderDetailPage />} />
                      <Route path="/support" element={<SupportPage />} />
                      <Route path="/admin" element={<AdminPage />} />
                      <Route path="/tutorial" element={<TutorialPage />} />
                      <Route path="/tutorial/:id" element={<ScenarioDetailPage />} />
                      <Route path="/docs" element={<DocsPage />} />
                    </Route>
                  </Route>

                  <Route path="*" element={<Navigate to="/" replace />} />
                </Routes>
              </BrowserRouter>
            </CartProvider>
          </AuthProvider>
        </ThemeProvider>
      </QueryClientProvider>
    </ErrorBoundary>
  );
}
