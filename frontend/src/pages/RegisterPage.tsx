import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import {
  ShieldCheck,
  Eye,
  EyeOff,
  Lock,
  Mail,
  User,
  Phone,
  MapPin,
  AlertCircle,
  ArrowRight,
  Shield,
  LockKeyhole,
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
import { getErrorMessage } from '@/lib/messages';
import type { User as UserType } from '@/types';

const registerSchema = z.object({
  name: z.string().min(2, 'Họ và tên phải có ít nhất 2 ký tự'),
  email: z.string().email('Địa chỉ email không hợp lệ'),
  password: z
    .string()
    .min(8, 'Mật khẩu phải có ít nhất 8 ký tự')
    .regex(/[A-Z]/, 'Cần ít nhất 1 chữ hoa')
    .regex(/[0-9]/, 'Cần ít nhất 1 chữ số'),
  phone: z.string().min(9, 'Số điện thoại không hợp lệ'),
  address: z.string().min(5, 'Địa chỉ giao hàng quá ngắn'),
});

type RegisterFormData = z.infer<typeof registerSchema>;

export function RegisterPage(): React.JSX.Element {
  const { login } = useAuth();
  const { resolvedTheme, toggleTheme } = useTheme();
  const navigate = useNavigate();

  const [showPassword, setShowPassword] = useState(false);
  const [generalError, setGeneralError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      name: '',
      email: '',
      password: '',
      phone: '',
      address: '',
    },
  });

  const onSubmit = async (data: RegisterFormData) => {
    try {
      setGeneralError(null);
      const res = await api.post<{ token: string; user: UserType }>('/api/auth/register', data, {
        skipAuth: true,
      });

      login(res.data.token, res.data.user);
      toast.success('Đăng ký tài khoản thành công!');
      navigate('/');
    } catch (err: any) {
      if (err instanceof ApiError) {
        if (err.details && err.details.length > 0) {
          err.details.forEach((issue) => {
            const field = issue.field as keyof RegisterFormData;
            setError(field, { message: issue.issue });
          });
        }
        setGeneralError(err.message);
      } else {
        setGeneralError(getErrorMessage('validation_error'));
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

        {/* Center Content */}
        <div className="relative z-10 my-auto py-10 space-y-8">
          <div className="space-y-3">
            <h1 className="text-3xl xl:text-4xl font-extrabold tracking-tight text-white leading-tight">
              Tạo Tài Khoản & <br />
              <span className="bg-gradient-to-r from-teal-400 via-teal-200 to-sky-400 bg-clip-text text-transparent">
                Tham Gia Môi Trường Lab
              </span>
            </h1>
            <p className="text-sm text-slate-400 max-w-lg leading-relaxed">
              Khởi tạo hồ sơ người dùng để trải nghiệm đầy đủ các tính năng thương mại điện tử,
              thử nghiệm giỏ hàng, đặt hàng Idempotent và thực hành kịch bản an ninh mạng.
            </p>
          </div>

          <div className="grid grid-cols-1 gap-4 max-w-lg">
            <div className="flex items-start gap-3.5 rounded-xl bg-slate-800/60 p-3.5 ring-1 ring-slate-700/60 backdrop-blur-sm">
              <div className="rounded-lg bg-teal-500/20 p-2 text-teal-400">
                <Shield className="h-4 w-4" />
              </div>
              <div className="text-xs">
                <div className="font-semibold text-slate-200">Bảo mật mật khẩu Argon2id</div>
                <div className="text-slate-400 mt-0.5">Mật khẩu được băm an toàn theo tiêu chuẩn mã hóa chống dò quét GPU hiện đại.</div>
              </div>
            </div>

            <div className="flex items-start gap-3.5 rounded-xl bg-slate-800/60 p-3.5 ring-1 ring-slate-700/60 backdrop-blur-sm">
              <div className="rounded-lg bg-sky-500/20 p-2 text-sky-400">
                <LockKeyhole className="h-4 w-4" />
              </div>
              <div className="text-xs">
                <div className="font-semibold text-slate-200">Xác thực Stateless JWT</div>
                <div className="text-slate-400 mt-0.5">Phiên làm việc bảo mật cao với chuẩn JSON Web Token mã hóa khóa ký 256-bit.</div>
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
          RIGHT PANEL: Pure Register Form
      ========================================================================= */}
      <div className="w-full lg:w-1/2 flex flex-col justify-center items-center p-6 sm:p-12 relative overflow-y-auto">
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
        <div className="w-full max-w-md space-y-6 my-auto py-8">
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
              Khởi tạo tài khoản thực hành phòng Lab
            </p>
          </div>

          {/* Form Header */}
          <div className="space-y-1.5 text-center lg:text-left">
            <h2 className="text-2xl font-bold tracking-tight text-slate-900 dark:text-white">
              Đăng ký tài khoản
            </h2>
            <p className="text-xs text-slate-500 dark:text-slate-400">
              Nhập thông tin cá nhân để tạo tài khoản người dùng LabShop
            </p>
          </div>

          {/* Register Card */}
          <div className="rounded-2xl border border-slate-200 bg-white p-7 shadow-xl shadow-slate-200/50 dark:border-slate-800/80 dark:bg-slate-900/90 dark:shadow-none space-y-5 backdrop-blur-sm">
            {generalError && (
              <div className="rounded-xl bg-rose-50 p-3.5 text-xs text-rose-700 dark:bg-rose-950/60 dark:text-rose-300 flex items-start gap-2.5 border border-rose-200/60 dark:border-rose-900/40">
                <AlertCircle className="h-4 w-4 shrink-0 text-rose-500 mt-0.5" />
                <span className="leading-relaxed">{generalError}</span>
              </div>
            )}

            <form onSubmit={handleSubmit(onSubmit)} className="space-y-3.5">
              <div className="space-y-1">
                <label className="text-xs font-semibold text-slate-700 dark:text-slate-300">
                  Họ và tên <span className="text-rose-500">*</span>
                </label>
                <div className="relative">
                  <Input
                    placeholder="Nguyễn Văn A"
                    {...register('name')}
                    error={!!errors.name}
                    className="pl-9 text-xs h-10 rounded-xl"
                  />
                  <User className="absolute left-3 top-3 h-4 w-4 text-slate-400 pointer-events-none" />
                </div>
                {errors.name && <p className="text-[11px] text-rose-500">{errors.name.message}</p>}
              </div>

              <div className="space-y-1">
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
                {errors.email && <p className="text-[11px] text-rose-500">{errors.email.message}</p>}
              </div>

              <div className="space-y-1">
                <label className="text-xs font-semibold text-slate-700 dark:text-slate-300">
                  Mật khẩu (≥ 8 ký tự, có chữ hoa và số) <span className="text-rose-500">*</span>
                </label>
                <div className="relative">
                  <Input
                    type={showPassword ? 'text' : 'password'}
                    placeholder="••••••••"
                    autoComplete="new-password"
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
                {errors.password && <p className="text-[11px] text-rose-500">{errors.password.message}</p>}
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div className="space-y-1">
                  <label className="text-xs font-semibold text-slate-700 dark:text-slate-300">
                    Số điện thoại <span className="text-rose-500">*</span>
                  </label>
                  <div className="relative">
                    <Input
                      placeholder="0912345678"
                      {...register('phone')}
                      error={!!errors.phone}
                      className="pl-9 text-xs h-10 rounded-xl"
                    />
                    <Phone className="absolute left-3 top-3 h-4 w-4 text-slate-400 pointer-events-none" />
                  </div>
                  {errors.phone && <p className="text-[11px] text-rose-500">{errors.phone.message}</p>}
                </div>

                <div className="space-y-1">
                  <label className="text-xs font-semibold text-slate-700 dark:text-slate-300">
                    Địa chỉ giao hàng <span className="text-rose-500">*</span>
                  </label>
                  <div className="relative">
                    <Input
                      placeholder="Hà Nội, Việt Nam"
                      {...register('address')}
                      error={!!errors.address}
                      className="pl-9 text-xs h-10 rounded-xl"
                    />
                    <MapPin className="absolute left-3 top-3 h-4 w-4 text-slate-400 pointer-events-none" />
                  </div>
                  {errors.address && <p className="text-[11px] text-rose-500">{errors.address.message}</p>}
                </div>
              </div>

              <Button
                type="submit"
                size="lg"
                loading={isSubmitting}
                className="w-full gap-2 rounded-xl h-10 bg-teal-600 hover:bg-teal-700 text-white font-medium shadow-md shadow-teal-600/20 transition-all mt-3"
              >
                <span>Tạo tài khoản LabShop</span>
                <ArrowRight className="h-4 w-4" />
              </Button>
            </form>

            <div className="text-center text-xs text-slate-500 dark:text-slate-400 pt-3 border-t border-slate-100 dark:border-slate-800/80">
              Đã có tài khoản?{' '}
              <Link to="/login" className="font-semibold text-teal-600 hover:text-teal-700 dark:text-teal-400 hover:underline">
                Đăng nhập ngay
              </Link>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
