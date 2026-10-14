import React from 'react';
import { Navigate, useLocation, Outlet } from 'react-router-dom';
import { useAuth } from '@/lib/authContext';
import { Loader2 } from 'lucide-react';

interface AuthGuardProps {
  children?: React.ReactNode;
}

/**
 * AuthGuard protects routes from unauthenticated access.
 * Unauthenticated users are redirected to /login with the target path saved in state.
 */
export function AuthGuard({ children }: AuthGuardProps): React.JSX.Element {
  const { isAuthenticated, isLoading } = useAuth();
  const location = useLocation();

  if (isLoading) {
    return (
      <div className="flex min-h-[50vh] flex-col items-center justify-center space-y-3">
        <Loader2 className="h-8 w-8 animate-spin text-teal-600" />
        <p className="text-xs font-medium text-slate-500">Đang xác thực phiên làm việc...</p>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  return children ? <>{children}</> : <Outlet />;
}
