import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import App from './App';

describe('App', () => {
  it('renders the LabShop header and brand badge correctly', () => {
    render(<App />);
    expect(screen.getByText('LabShop')).toBeInTheDocument();
    expect(screen.getByText('WAAP & CDN Lab')).toBeInTheDocument();
    expect(
      screen.getByText('Học WAAP & CDN qua trải nghiệm thực chiến')
    ).toBeInTheDocument();
  });

  it('renders the semantic lab status indicators', () => {
    render(<App />);
    expect(screen.getByText(/Origin \(Blue\)/i)).toBeInTheDocument();
    expect(screen.getByText(/HIT \(Teal\)/i)).toBeInTheDocument();
    expect(screen.getByText(/MISS \(Amber\)/i)).toBeInTheDocument();
    expect(screen.getByText(/Blocked \(Rose\)/i)).toBeInTheDocument();
    expect(screen.getByText(/Stale \(Orange\)/i)).toBeInTheDocument();
  });

  it('renders storefront placeholders for modules', () => {
    render(<App />);
    expect(
      screen.getByText('Khu Vực Cửa Hàng (Storefront Placeholder)')
    ).toBeInTheDocument();
    expect(screen.getByText('Module Catalog & Sản phẩm')).toBeInTheDocument();
    expect(screen.getByText('Module Giỏ hàng & Thanh toán')).toBeInTheDocument();
    expect(screen.getByText('Module Lab & Kịch bản WAF/CDN')).toBeInTheDocument();
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
});
