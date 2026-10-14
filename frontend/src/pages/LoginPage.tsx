import React, { useState } from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { ShieldCheck, Eye, EyeOff, Lock, Mail, AlertCircle, ArrowRight } from 'lucide-react';
import { api, ApiError } from '@/lib/api';
import { useAuth } from '@/lib/authContext';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { toast } from 'sonner';
import { getErrorMessage } from '@/lib/messages';
import type { User } from '@/types';

const loginSchema = z.object({
  email: z.string().email('Địa chỉ email không đúng định dạng'),
  password: z.string().min(6, 'Mật khẩu phải có ít nhất 6 ký tự'),
});

type LoginFormData = z.infer<typeof loginSchema>;

export function LoginPage(): React.JSX.Element {
  const { login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  const [showPassword, setShowPassword] = useState(false);
  const [generalError, setGeneralError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: '',
      password: '',
    },
  });

  const onSubmit = async (data: LoginFormData) => {
    try {
      setGeneralError(null);
      const res = await api.post<{ token: string; user: User }>('/api/auth/login', data, {
        skipAuth: true,
      });

      login(res.data.token, res.data.user);
      toast.success(`Đăng nhập thành công! Chào mừng ${res.data.user.name}.`);

      const returnUrl = (location.state as any)?.from?.pathname || '/';
      navigate(returnUrl, { replace: true });
    } catch (err: any) {
      if (err instanceof ApiError) {
        if (err.details && err.details.length > 0) {
          err.details.forEach((issue) => {
            if (issue.field === 'email' || issue.field === 'password') {
              setError(issue.field as keyof LoginFormData, { message: issue.issue });
            }
          });
        }
        setGeneralError(err.message);
      } else {
        setGeneralError(getErrorMessage('invalid_credentials'));
      }
    }
  };

  return (
    <div className="mx-auto flex min-h-[calc(100vh-14rem)] max-w-md flex-col justify-center px-4 py-12 sm:px-6 lg:px-8">
      <div className="sm:mx-auto sm:w-full sm:max-w-md text-center space-y-2">
        <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-control bg-teal-600 text-white shadow-sm">
          <ShieldCheck className="h-7 w-7" />
        </div>
        <h1 className="text-2xl font-bold tracking-tight text-slate-900 dark:text-white">
          Đăng nhập tài khoản
        </h1>
        <p className="text-xs text-slate-500 dark:text-slate-400">
          Truy cập hệ thống LabShop để trải nghiệm mua sắm và lab an ninh
        </p>
      </div>

      <div className="mt-8 sm:mx-auto sm:w-full sm:max-w-md">
        <div className="rounded-card border border-slate-200 bg-white p-6 shadow-sm dark:border-slate-800 dark:bg-slate-900 sm:p-8 space-y-6">
          {generalError && (
            <div className="rounded-control bg-rose-50 p-3 text-xs text-rose-700 dark:bg-rose-950/50 dark:text-rose-300 flex items-start gap-2">
              <AlertCircle className="h-4 w-4 shrink-0 text-rose-500 mt-0.5" />
              <span>{generalError}</span>
            </div>
          )}

          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            <div className="space-y-1.5">
              <label className="text-xs font-semibold text-slate-700 dark:text-slate-300">
                Email *
              </label>
              <div className="relative">
                <Input
                  type="email"
                  placeholder="name@example.com"
                  {...register('email')}
                  error={!!errors.email}
                  className="pl-9 text-xs"
                />
                <Mail className="absolute left-3 top-3 h-4 w-4 text-slate-400 pointer-events-none" />
              </div>
              {errors.email && (
                <p className="text-[11px] text-rose-500">{errors.email.message}</p>
              )}
            </div>

            <div className="space-y-1.5">
              <div className="flex items-center justify-between">
                <label className="text-xs font-semibold text-slate-700 dark:text-slate-300">
                  Mật khẩu *
                </label>
              </div>
              <div className="relative">
                <Input
                  type={showPassword ? 'text' : 'password'}
                  placeholder="••••••••"
                  {...register('password')}
                  error={!!errors.password}
                  className="pl-9 pr-9 text-xs"
                />
                <Lock className="absolute left-3 top-3 h-4 w-4 text-slate-400 pointer-events-none" />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-3 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200"
                  aria-label={showPassword ? 'Ẩn mật khẩu' : 'Hiện mật khẩu'}
                >
                  {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </button>
              </div>
              {errors.password && (
                <p className="text-[11px] text-rose-500">{errors.password.message}</p>
              )}
            </div>

            <Button type="submit" size="lg" loading={isSubmitting} className="w-full gap-2">
              <span>Đăng nhập</span>
              <ArrowRight className="h-4 w-4" />
            </Button>
          </form>

          {/* Security Notice */}
          <div className="pt-4 border-t border-slate-100 dark:border-slate-800 space-y-1.5 text-center">
            <div className="flex items-center justify-center gap-1.5 text-xs text-teal-700 dark:text-teal-400 font-medium">
              <ShieldCheck className="h-4 w-4" />
              <span>Hệ thống phòng Lab được bảo vệ</span>
            </div>
            <p className="text-[11px] text-slate-500 dark:text-slate-400">
              Vui lòng đăng nhập để truy cập trang web. Hệ thống tự động khóa IP nếu phát hiện thử mật khẩu bất thường.
            </p>
          </div>

          <div className="text-center text-xs text-slate-500">
            Chưa có tài khoản?{' '}
            <Link to="/register" className="font-semibold text-teal-600 hover:underline dark:text-teal-400">
              Đăng ký ngay
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
