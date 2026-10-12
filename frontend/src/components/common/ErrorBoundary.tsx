import { Component, ErrorInfo, ReactNode } from 'react';
import { AlertTriangle, RotateCcw, Home } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface Props {
  children: ReactNode;
  fallbackTitle?: string;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
    error: null,
  };

  public static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo): void {
    console.error('ErrorBoundary caught an error:', error, errorInfo);
  }

  private handleReset = (): void => {
    this.setState({ hasError: false, error: null });
    window.location.reload();
  };

  private handleGoHome = (): void => {
    this.setState({ hasError: false, error: null });
    window.location.href = '/';
  };

  public render(): ReactNode {
    if (this.state.hasError) {
      return (
        <div className="min-h-[400px] flex items-center justify-center p-6">
          <div className="w-full max-w-md rounded-card border border-rose-200 bg-white p-6 shadow-sm dark:border-rose-900/50 dark:bg-slate-900 text-center space-y-4">
            <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-rose-50 text-rose-600 dark:bg-rose-950/50 dark:text-rose-400">
              <AlertTriangle className="h-6 w-6" />
            </div>

            <div>
              <h2 className="text-lg font-bold text-slate-900 dark:text-white">
                {this.props.fallbackTitle || 'Đã xảy ra lỗi khi hiển thị trang'}
              </h2>
              <p className="mt-1 text-xs text-slate-500 dark:text-slate-400 break-words">
                {this.state.error?.message || 'Lỗi không xác định. Vui lòng thử tải lại trang.'}
              </p>
            </div>

            <div className="flex items-center justify-center gap-3 pt-2">
              <Button size="sm" variant="outline" onClick={this.handleReset} className="gap-1.5">
                <RotateCcw className="h-4 w-4" />
                <span>Tải lại trang</span>
              </Button>
              <Button size="sm" onClick={this.handleGoHome} className="gap-1.5">
                <Home className="h-4 w-4" />
                <span>Về trang chủ</span>
              </Button>
            </div>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
