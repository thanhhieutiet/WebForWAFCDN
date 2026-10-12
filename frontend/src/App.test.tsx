import { describe, it, expect, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import App from './App';
import { formatVND } from './lib/formatters';
import { getErrorMessage } from './lib/messages';
import { getRequestHistory, clearRequestHistory } from './lib/api';

describe('App & Core Components', () => {
  beforeEach(() => {
    clearRequestHistory();
  });

  it('renders the LabShop header and navigation links correctly', () => {
    render(<App />);
    expect(screen.getAllByText('LabShop')[0]).toBeInTheDocument();
    expect(screen.getAllByText('WAAP & CDN')[0]).toBeInTheDocument();
    expect(screen.getByText('Cửa hàng')).toBeInTheDocument();
    expect(screen.getByText('Hướng dẫn Lab')).toBeInTheDocument();
    expect(screen.getByText('Tài liệu API')).toBeInTheDocument();
  });

  it('renders Environment Bar with host information', () => {
    render(<App />);
    expect(screen.getByText('Host:')).toBeInTheDocument();
    expect(screen.getByText(/Cấu hình Hostname/i)).toBeInTheDocument();
  });

  it('toggles the Lab Console drawer on click', () => {
    render(<App />);
    const consoleButton = screen.getByRole('button', {
      name: /Toggle Lab Console/i,
    });
    expect(consoleButton).toBeInTheDocument();

    // Open console drawer
    fireEvent.click(consoleButton);
    expect(
      screen.getByText(/Chưa có request HTTP nào được ghi nhận/i)
    ).toBeInTheDocument();
  });

  it('formats Vietnamese VND currency accurately', () => {
    const formatted = formatVND(1250000);
    // Non-breaking space or standard space with currency symbol
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
