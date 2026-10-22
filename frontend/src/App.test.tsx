import { describe, it, expect, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import App from './App';
import { formatVND } from './lib/formatters';
import { getErrorMessage } from './lib/messages';
import { getRequestHistory, clearRequestHistory } from './lib/api';

describe('App & Core Components', () => {
  beforeEach(() => {
    clearRequestHistory();
    localStorage.clear();
  });

  it('renders the secure split-screen login page when unauthenticated', () => {
    render(<App />);

    // Branding and Security Highlights
    expect(screen.getAllByText('LabShop')[0]).toBeInTheDocument();
    expect(screen.getAllByText('WAAP & CDN')[0]).toBeInTheDocument();
    expect(screen.getAllByText(/Cổng Xác Thực/i)[0]).toBeInTheDocument();
    expect(screen.getByText('Bảo vệ toàn diện CMC WAAP')).toBeInTheDocument();

    // Login Form
    expect(screen.getByText('Đăng nhập tài khoản')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('name@example.com')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('••••••••')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Đăng nhập hệ thống/i })).toBeInTheDocument();

    // Verify Navbar and Lab Console are HIDDEN when unauthenticated (Zero-Trust Security requirement)
    expect(screen.queryByText('Cửa hàng')).not.toBeInTheDocument();
    expect(screen.queryByText('Hướng dẫn Lab')).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /Toggle Lab Console/i })).not.toBeInTheDocument();
  });

  it('formats Vietnamese VND currency accurately', () => {
    const formatted = formatVND(1250000);
    expect(formatted).toMatch(/1\.250\.000/);
    expect(formatted).toMatch(/₫|VND/);
  });

  it('maps domain error codes to Vietnamese messages in messages.ts', () => {
    expect(getErrorMessage('coupon_expired')).toBe('Mã giảm giá đã hết hạn sử dụng.');
    expect(getErrorMessage('no_delivered_order')).toBe(
      'Bạn chỉ có thể đánh giá sản phẩm sau khi đã nhận hàng thành công.'
    );
    expect(getErrorMessage('edge_blocked')).toBe(
      'Yêu cầu bị chặn bởi tường lửa ứng dụng web (WAAP/Edge Rule).'
    );
  });

  it('manages request history in Single Fetch Wrapper', () => {
    expect(getRequestHistory().length).toBe(0);
  });
});
