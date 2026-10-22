import React, { useState } from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import {
  ShieldCheck,
  Eye,
  EyeOff,
  Lock,
  Mail,
  AlertCircle,
  ArrowRight,
  Shield,
  Zap,
  LockKeyhole,
  Activity,
  Sun,
  Moon,
  Server,
} from 'lucide-react';
import { api, ApiError } from '@/lib/api';
import { useAuth } from '@/lib/authContext';
import { useTheme } from '@/lib/themeContext';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { toast } from 'sonner';
import type { User } from '@/types';

const loginSchema = z.object({
  email: z.string().email('Địa chỉ email không đúng định dạng'),
  password: z.string().min(6, 'Mật khẩu phải có ít nhất 6 ký tự'),
});

type LoginFormData = z.infer<typeof loginSchema>;

export function LoginPage(): React.JSX.Element {
  const { login } = useAuth();
  const { resolvedTheme, toggleTheme } = useTheme();
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
        if (err.status === 401 || err.code === 'unauthorized' || err.code === 'invalid_credentials') {
          setGeneralError('Email hoặc mật khẩu không chính xác. Vui lòng kiểm tra lại.');
        } else if (err.status === 429 || err.code === 'too_many_attempts') {
          setGeneralError(err.message || 'IP đã bị khóa tạm thời do nhập sai quá 5 lần. Vui lòng thử lại sau.');
        } else {
          setGeneralError(err.message);
        }
      } else {
        setGeneralError('Email hoặc mật khẩu không chính xác. Vui lòng kiểm tra lại.');
      }
    }
  };

  return (
    <div className="flex min-h-screen w-full bg-slate-50 dark:bg-slate-950 text-slate-900 dark:text-slate-100 font-sans selection:bg-teal-500 selection:text-white">
      {/* =========================================================================
          LEFT PANEL: High-Tech Security Branding (Desktop Only)
      ========================================================================= */}
      <div className="hidden lg:flex lg:w-1/2 flex-col justify-between p-12 xl:p-16 bg-slate-900 text-white relative overflow-hidden border-r border-slate-800">
        {/* Glow Effects */}
        <div className="absolute -top-32 -left-32 w-96 h-96 rounded-full bg-teal-500/15 blur-3xl pointer-events-none" />
        <div className="absolute -bottom-32 -right-32 w-96 h-96 rounded-full bg-blue-500/15 blur-3xl pointer-events-none" />
        <div className="absolute inset-0 bg-[radial-gradient(#1e293b_1px,transparent_1px)] [background-size:24px_24px] opacity-40 pointer-events-none" />

        {/* Top Header & Logo */}
        <div className="relative z-10">
          <div className="flex items-center gap-3">
            <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-teal-500/20 text-teal-400 ring-1 ring-teal-500/40 shadow-inner">
              <ShieldCheck className="h-6 w-6 text-teal-400" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="text-xl font-bold tracking-tight text-white">LabShop</span>
                <span className="rounded-full bg-teal-500/20 px-2 py-0.5 text-[10px] font-semibold text-teal-300 ring-1 ring-teal-500/30">
                  WAAP & CDN
                </span>
              </div>
              <p className="text-xs text-slate-400">CMC Telecom Cloud Security Hands-on Lab</p>
            </div>
          </div>
        </div>

        {/* Center Content: Architecture & Security Highlights */}
        <div className="relative z-10 my-auto py-10 space-y-8">
          <div className="space-y-3">
            <h1 className="text-3xl xl:text-4xl font-extrabold tracking-tight text-white leading-tight">
              Cổng Xác Thực & <br />
              <span className="bg-gradient-to-r from-teal-400 via-teal-200 to-sky-400 bg-clip-text text-transparent">
                Phòng Thủ Tấn Công Đa Tầng
              </span>
            </h1>
            <p className="text-sm text-slate-400 max-w-lg leading-relaxed">
              Môi trường đào tạo thực chiến về giải pháp bảo vệ ứng dụng Web & API (WAAP),
              mạng lưới phân phối Multi-CDN, phòng chống rò rỉ dữ liệu và chống Brute Force.
            </p>
          </div>

          {/* Security Features List */}
          <div className="grid grid-cols-1 gap-4 max-w-lg">
            <div className="flex items-start gap-3.5 rounded-xl bg-slate-800/60 p-3.5 ring-1 ring-slate-700/60 backdrop-blur-sm">
              <div className="rounded-lg bg-teal-500/20 p-2 text-teal-400">
                <Shield className="h-4 w-4" />
              </div>
              <div className="text-xs">
                <div className="font-semibold text-slate-200">Bảo vệ toàn diện CMC WAAP</div>
                <div className="text-slate-400 mt-0.5">Phát hiện và ngăn chặn SQLi, XSS, Path Traversal, Bot crawling độc hại.</div>
              </div>
            </div>

            <div className="flex items-start gap-3.5 rounded-xl bg-slate-800/60 p-3.5 ring-1 ring-slate-700/60 backdrop-blur-sm">
              <div className="rounded-lg bg-sky-500/20 p-2 text-sky-400">
                <Zap className="h-4 w-4" />
              </div>
              <div className="text-xs">
                <div className="font-semibold text-slate-200">Tăng tốc Edge Multi-CDN</div>
                <div className="text-slate-400 mt-0.5">Bộ nhớ đệm thông minh, nén nội dung động và tối ưu hóa phân phối toàn cầu.</div>
              </div>
            </div>

            <div className="flex items-start gap-3.5 rounded-xl bg-slate-800/60 p-3.5 ring-1 ring-slate-700/60 backdrop-blur-sm">
              <div className="rounded-lg bg-amber-500/20 p-2 text-amber-400">
                <LockKeyhole className="h-4 w-4" />
              </div>
              <div className="text-xs">
                <div className="font-semibold text-slate-200">Xác thực Zero-Trust Gateway</div>
                <div className="text-slate-400 mt-0.5">Cô lập hoàn toàn tài nguyên và API nội bộ, chỉ cấp quyền sau khi đăng nhập.</div>
              </div>
            </div>

            <div className="flex items-start gap-3.5 rounded-xl bg-slate-800/60 p-3.5 ring-1 ring-slate-700/60 backdrop-blur-sm">
              <div className="rounded-lg bg-rose-500/20 p-2 text-rose-400">
                <Activity className="h-4 w-4" />
              </div>
              <div className="text-xs">
                <div className="font-semibold text-slate-200">Chống Brute-Force & Dò quét</div>
                <div className="text-slate-400 mt-0.5">Rate limit thông minh, tự động cô lập IP có hành vi thử mật khẩu bất thường.</div>
              </div>
            </div>
          </div>
        </div>

        {/* Bottom Topology Indicator */}
        <div className="relative z-10 pt-6 border-t border-slate-800/80 flex items-center justify-between text-xs text-slate-500">
          <div className="flex items-center gap-2 font-mono text-[11px] text-slate-400">
            <Server className="h-3.5 w-3.5 text-teal-400" />
            <span>Client</span>
            <span>➔</span>
            <span className="text-teal-300">CMC CDN</span>
            <span>➔</span>
            <span className="text-sky-300">CMC WAF</span>
            <span>➔</span>
            <span className="text-slate-200">EC2 Origin</span>
          </div>
          <div>v2.0 Enterprise</div>
        </div>
      </div>

      {/* =========================================================================
          RIGHT PANEL: Pure Login Form
      ========================================================================= */}
      <div className="w-full lg:w-1/2 flex flex-col justify-center items-center p-6 sm:p-12 relative">
        {/* Top Controls: Dark/Light Mode Toggle */}
        <div className="absolute top-6 right-6 z-20">
          <button
            type="button"
            onClick={toggleTheme}
            className="flex h-9 w-9 items-center justify-center rounded-xl border border-slate-200 bg-white text-slate-600 shadow-sm transition hover:bg-slate-100 hover:text-slate-900 dark:border-slate-800 dark:bg-slate-900 dark:text-slate-400 dark:hover:bg-slate-800 dark:hover:text-slate-100"
            aria-label="Chuyển đổi giao diện Sáng / Tối"
            title="Chuyển đổi giao diện"
          >
            {resolvedTheme === 'dark' ? <Sun className="h-4 w-4 text-amber-400" /> : <Moon className="h-4 w-4 text-slate-600" />}
          </button>
        </div>

        {/* Form Container */}
        <div className="w-full max-w-md space-y-8">
          {/* Mobile-only Branding */}
          <div className="lg:hidden text-center space-y-2">
            <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-teal-600 text-white shadow-md">
              <ShieldCheck className="h-7 w-7" />
            </div>
            <div className="flex items-center justify-center gap-2">
              <h2 className="text-xl font-bold tracking-tight text-slate-900 dark:text-white">LabShop</h2>
              <span className="rounded-full bg-teal-100 px-2 py-0.5 text-[10px] font-semibold text-teal-800 dark:bg-teal-950 dark:text-teal-300">
                WAAP & CDN
              </span>
            </div>
            <p className="text-xs text-slate-500 dark:text-slate-400">
              Cổng xác thực an ninh hệ thống phòng Lab
            </p>
          </div>

          {/* Form Header */}
          <div className="space-y-1.5 text-center lg:text-left">
            <div className="hidden lg:inline-flex items-center gap-2 rounded-full bg-teal-50 dark:bg-teal-950/60 px-3 py-1 text-xs font-medium text-teal-700 dark:text-teal-300 ring-1 ring-teal-600/20 mb-2">
              <Lock className="h-3 w-3" />
              <span>Yêu cầu đăng nhập để truy cập</span>
            </div>
            <h2 className="text-2xl font-bold tracking-tight text-slate-900 dark:text-white">
              Đăng nhập tài khoản
            </h2>
            <p className="text-xs text-slate-500 dark:text-slate-400">
              Vui lòng nhập thông tin xác thực được cấp để tiếp tục
            </p>
          </div>

          {/* Login Card */}
          <div className="rounded-2xl border border-slate-200 bg-white p-7 shadow-xl shadow-slate-200/50 dark:border-slate-800/80 dark:bg-slate-900/90 dark:shadow-none space-y-6 backdrop-blur-sm">
            {generalError && (
              <div className="rounded-xl bg-rose-50 p-3.5 text-xs text-rose-700 dark:bg-rose-950/60 dark:text-rose-300 flex items-start gap-2.5 border border-rose-200/60 dark:border-rose-900/40">
                <AlertCircle className="h-4 w-4 shrink-0 text-rose-500 mt-0.5" />
                <span className="leading-relaxed">{generalError}</span>
              </div>
            )}

            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-slate-700 dark:text-slate-300">
                  Địa chỉ Email <span className="text-rose-500">*</span>
                </label>
                <div className="relative">
                  <Input
                    type="email"
                    placeholder="name@example.com"
                    autoComplete="email"
                    {...register('email')}
                    error={!!errors.email}
                    className="pl-9 text-xs h-10 rounded-xl"
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
                    Mật khẩu <span className="text-rose-500">*</span>
                  </label>
                </div>
                <div className="relative">
                  <Input
                    type={showPassword ? 'text' : 'password'}
                    placeholder="••••••••"
                    autoComplete="current-password"
                    {...register('password')}
                    error={!!errors.password}
                    className="pl-9 pr-9 text-xs h-10 rounded-xl"
                  />
                  <Lock className="absolute left-3 top-3 h-4 w-4 text-slate-400 pointer-events-none" />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-3 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 transition"
                    aria-label={showPassword ? 'Ẩn mật khẩu' : 'Hiện mật khẩu'}
                  >
                    {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                  </button>
                </div>
                {errors.password && (
                  <p className="text-[11px] text-rose-500">{errors.password.message}</p>
                )}
              </div>

              <Button
                type="submit"
                size="lg"
                loading={isSubmitting}
                className="w-full gap-2 rounded-xl h-10 bg-teal-600 hover:bg-teal-700 text-white font-medium shadow-md shadow-teal-600/20 transition-all mt-2"
              >
                <span>Đăng nhập hệ thống</span>
                <ArrowRight className="h-4 w-4" />
              </Button>
            </form>

            {/* Security Notice */}
            <div className="pt-4 border-t border-slate-100 dark:border-slate-800/80 space-y-1.5 text-center">
              <div className="flex items-center justify-center gap-1.5 text-xs text-teal-700 dark:text-teal-400 font-medium">
                <ShieldCheck className="h-4 w-4" />
                <span>Hệ thống phòng Lab được bảo vệ</span>
              </div>
              <p className="text-[11px] text-slate-500 dark:text-slate-400 leading-relaxed">
                Hệ thống tự động khóa tạm thời 15 phút nếu phát hiện thử sai mật khẩu quá 5 lần liên tiếp.
              </p>
            </div>

            <div className="text-center text-xs text-slate-500 dark:text-slate-400">
              Chưa có tài khoản?{' '}
              <Link to="/register" className="font-semibold text-teal-600 hover:text-teal-700 dark:text-teal-400 hover:underline">
                Đăng ký tài khoản
              </Link>
            </div>
          </div>

          <div className="text-center text-[11px] text-slate-400 dark:text-slate-600">
            CMC Telecom Multi-CDN & WAAP Security Lab • All Rights Reserved
          </div>
        </div>
      </div>
    </div>
  );
}
