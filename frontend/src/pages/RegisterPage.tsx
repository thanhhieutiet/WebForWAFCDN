import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { ShieldCheck, Eye, EyeOff, Lock, Mail, User, Phone, MapPin, AlertCircle, ArrowRight } from 'lucide-react';
import { api, ApiError } from '@/lib/api';
import { useAuth } from '@/lib/authContext';
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
    <div className="mx-auto flex min-h-[calc(100vh-14rem)] max-w-lg flex-col justify-center px-4 py-12 sm:px-6 lg:px-8">
      <div className="sm:mx-auto sm:w-full text-center space-y-2">
        <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-control bg-teal-600 text-white shadow-sm">
          <ShieldCheck className="h-7 w-7" />
        </div>
        <h1 className="text-2xl font-bold tracking-tight text-slate-900 dark:text-white">
          Tạo tài khoản LabShop
        </h1>
        <p className="text-xs text-slate-500 dark:text-slate-400">
          Đăng ký để lưu giỏ hàng, đặt hàng và thực hành các kịch bản an ninh mạng
        </p>
      </div>

      <div className="mt-8 sm:mx-auto sm:w-full">
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
                Họ và tên *
              </label>
              <div className="relative">
                <Input
                  placeholder="Nguyễn Văn A"
                  {...register('name')}
                  error={!!errors.name}
                  className="pl-9 text-xs"
                />
                <User className="absolute left-3 top-3 h-4 w-4 text-slate-400 pointer-events-none" />
              </div>
              {errors.name && <p className="text-[11px] text-rose-500">{errors.name.message}</p>}
            </div>

            <div className="space-y-1.5">
              <label className="text-xs font-semibold text-slate-700 dark:text-slate-300">
                Địa chỉ Email *
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
              {errors.email && <p className="text-[11px] text-rose-500">{errors.email.message}</p>}
            </div>

            <div className="space-y-1.5">
              <label className="text-xs font-semibold text-slate-700 dark:text-slate-300">
                Mật khẩu (≥ 8 ký tự, có chữ hoa và số) *
              </label>
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
              {errors.password && <p className="text-[11px] text-rose-500">{errors.password.message}</p>}
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-slate-700 dark:text-slate-300">
                  Số điện thoại *
                </label>
                <div className="relative">
                  <Input
                    placeholder="0912345678"
                    {...register('phone')}
                    error={!!errors.phone}
                    className="pl-9 text-xs"
                  />
                  <Phone className="absolute left-3 top-3 h-4 w-4 text-slate-400 pointer-events-none" />
                </div>
                {errors.phone && <p className="text-[11px] text-rose-500">{errors.phone.message}</p>}
              </div>

              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-slate-700 dark:text-slate-300">
                  Địa chỉ giao hàng *
                </label>
                <div className="relative">
                  <Input
                    placeholder="Hà Nội, Việt Nam"
                    {...register('address')}
                    error={!!errors.address}
                    className="pl-9 text-xs"
                  />
                  <MapPin className="absolute left-3 top-3 h-4 w-4 text-slate-400 pointer-events-none" />
                </div>
                {errors.address && <p className="text-[11px] text-rose-500">{errors.address.message}</p>}
              </div>
            </div>

            <Button type="submit" size="lg" loading={isSubmitting} className="w-full gap-2 mt-2">
              <span>Tạo tài khoản</span>
              <ArrowRight className="h-4 w-4" />
            </Button>
          </form>

          <div className="text-center text-xs text-slate-500 pt-2 border-t border-slate-100 dark:border-slate-800">
            Đã có tài khoản?{' '}
            <Link to="/login" className="font-semibold text-teal-600 hover:underline dark:text-teal-400">
              Đăng nhập ngay
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
