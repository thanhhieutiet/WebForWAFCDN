import React, { useState, Suspense, lazy } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import {
  LayoutDashboard,
  ShoppingBag,
  Package,
  Headphones,
  Users,
  ChevronLeft,
  ChevronRight,
  TrendingUp,
  Plus,
  Edit2,
  Trash2,
  Truck,
  PackageCheck,
  ShieldAlert,
} from 'lucide-react';
import { api } from '@/lib/api';
import { useAuth } from '@/lib/authContext';
import { formatVND, formatDateTime } from '@/lib/formatters';
import { OrderStatusBadge } from '@/components/shop/OrderStatusBadge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Dialog } from '@/components/ui/dialog';
import { Skeleton } from '@/components/ui/skeleton';
import { toast } from 'sonner';
import type { Product, Order, OrderStatus } from '@/types';

// Lazy-load Recharts component
const AdminRevenueChart = lazy(() => import('@/components/admin/AdminRevenueChart'));

type AdminTab = 'dashboard' | 'orders' | 'tickets' | 'products' | 'users';

export function AdminPage(): React.JSX.Element {
  const { user, isAdmin, isAuthenticated } = useAuth();
  const queryClient = useQueryClient();

  const [activeTab, setActiveTab] = useState<AdminTab>('dashboard');
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);

  // Products CRUD State
  const [productModalOpen, setProductModalOpen] = useState(false);
  const [editingProduct, setEditingProduct] = useState<Product | null>(null);
  const [prodForm, setProdForm] = useState({
    name: '',
    slug: '',
    category_id: 'cat-smartphones',
    price: 1000000,
    stock: 20,
    description: '',
    image_path: '/static/img/products/p01.jpg',
  });

  // Ticket reply state
  const [replyTicketId, setReplyTicketId] = useState<string | null>(null);
  const [replyContent, setReplyContent] = useState('');
  const [closeOnReply, setCloseOnReply] = useState(false);

  // Queries
  const { data: stats } = useQuery({
    queryKey: ['admin-stats'],
    queryFn: async () => {
      const res = await api.get<any>('/api/admin/stats');
      return res.data;
    },
    enabled: isAdmin,
  });

  const { data: ordersData } = useQuery({
    queryKey: ['admin-orders'],
    queryFn: async () => {
      const res = await api.get<{ items: Order[]; total: number }>('/api/admin/orders?page_size=50');
      return res.data;
    },
    enabled: isAdmin && activeTab === 'orders',
  });

  const { data: ticketsData } = useQuery({
    queryKey: ['admin-tickets'],
    queryFn: async () => {
      const res = await api.get<any>('/api/admin/tickets?page_size=50');
      return res.data;
    },
    enabled: isAdmin && activeTab === 'tickets',
  });

  const { data: productsData } = useQuery({
    queryKey: ['admin-products'],
    queryFn: async () => {
      const res = await api.get<{ items: Product[]; total: number }>('/api/admin/products?page_size=100');
      return res.data;
    },
    enabled: isAdmin && (activeTab === 'products' || activeTab === 'dashboard'),
  });

  const { data: usersData } = useQuery({
    queryKey: ['admin-users'],
    queryFn: async () => {
      const res = await api.get<any>('/api/admin/users?page_size=50');
      return res.data;
    },
    enabled: isAdmin && activeTab === 'users',
  });

  // Order status transition mutation
  const updateOrderStatusMutation = useMutation({
    mutationFn: async ({ orderId, status }: { orderId: string; status: OrderStatus }) => {
      const res = await api.patch(`/api/admin/orders/${orderId}/status`, { status });
      return res.data;
    },
    onSuccess: (_, variables) => {
      toast.success(`Đã cập nhật trạng thái đơn sang ${variables.status}`);
      queryClient.invalidateQueries({ queryKey: ['admin-orders'] });
      queryClient.invalidateQueries({ queryKey: ['admin-stats'] });
    },
    onError: (err: any) => {
      toast.error('Lỗi cập nhật trạng thái đơn', { description: err?.message });
    },
  });

  // Ticket reply mutation
  const replyTicketMutation = useMutation({
    mutationFn: async () => {
      if (!replyTicketId) return;
      const res = await api.post(`/api/admin/tickets/${replyTicketId}/reply`, {
        message: replyContent.trim(),
        close: closeOnReply,
      });
      return res.data;
    },
    onSuccess: () => {
      toast.success('Đã gửi phản hồi ticket');
      setReplyTicketId(null);
      setReplyContent('');
      queryClient.invalidateQueries({ queryKey: ['admin-tickets'] });
    },
    onError: (err: any) => {
      toast.error('Lỗi khi trả lời ticket', { description: err?.message });
    },
  });

  // Product Create/Update mutation
  const saveProductMutation = useMutation({
    mutationFn: async () => {
      if (editingProduct) {
        const res = await api.put(`/api/admin/products/${editingProduct.id}`, prodForm);
        return res.data;
      } else {
        const res = await api.post('/api/admin/products', prodForm);
        return res.data;
      }
    },
    onSuccess: () => {
      toast.success(editingProduct ? 'Đã cập nhật sản phẩm' : 'Đã tạo sản phẩm mới');
      setProductModalOpen(false);
      setEditingProduct(null);
      queryClient.invalidateQueries({ queryKey: ['admin-products'] });
    },
    onError: (err: any) => {
      toast.error('Lỗi lưu sản phẩm', { description: err?.message });
    },
  });

  // Product Delete mutation
  const deleteProductMutation = useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/api/admin/products/${id}`);
    },
    onSuccess: () => {
      toast.success('Đã lưu trữ / xóa sản phẩm');
      queryClient.invalidateQueries({ queryKey: ['admin-products'] });
    },
    onError: (err: any) => {
      toast.error('Lỗi khi xóa sản phẩm', { description: err?.message });
    },
  });

  if (!isAuthenticated || !isAdmin) {
    return (
      <div className="mx-auto max-w-md px-4 py-20 text-center space-y-4">
        <ShieldAlert className="mx-auto h-12 w-12 text-rose-500" />
        <h1 className="text-2xl font-bold text-slate-900 dark:text-white">Truy cập bị từ chối (403)</h1>
        <p className="text-sm text-slate-500">
          Trang Quản trị chỉ dành riêng cho tài khoản có quyền Quản trị viên (Admin).
        </p>
        <Link to="/login">
          <Button>Đăng nhập bằng tài khoản Quản trị</Button>
        </Link>
      </div>
    );
  }

  const handleOpenEditProduct = (p: Product) => {
    setEditingProduct(p);
    setProdForm({
      name: p.name,
      slug: p.slug,
      category_id: p.category_id,
      price: p.price,
      stock: p.stock,
      description: p.description,
      image_path: p.image_path,
    });
    setProductModalOpen(true);
  };

  const handleOpenCreateProduct = () => {
    setEditingProduct(null);
    setProdForm({
      name: '',
      slug: '',
      category_id: 'cat-smartphones',
      price: 15000000,
      stock: 20,
      description: '',
      image_path: '/static/img/products/p01.jpg',
    });
    setProductModalOpen(true);
  };

  return (
    <div className="flex min-h-[calc(100vh-8rem)] bg-slate-50 dark:bg-slate-950">
      {/* Collapsible Sidebar */}
      <aside
        className={`border-r border-slate-200 bg-white transition-all duration-200 dark:border-slate-800 dark:bg-slate-900 ${
          sidebarCollapsed ? 'w-16' : 'w-60'
        } shrink-0 flex flex-col justify-between`}
      >
        <div className="p-3 space-y-4">
          <div className="flex items-center justify-between px-2 py-1">
            {!sidebarCollapsed && (
              <span className="text-xs font-bold uppercase tracking-wider text-slate-400">
                Quản trị viên
              </span>
            )}
            <button
              type="button"
              onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
              className="p-1 rounded text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800"
              aria-label="Thu gọn thanh bên"
            >
              {sidebarCollapsed ? <ChevronRight className="h-4 w-4" /> : <ChevronLeft className="h-4 w-4" />}
            </button>
          </div>

          <nav className="space-y-1">
            {[
              { id: 'dashboard', label: 'Tổng quan (KPI)', icon: LayoutDashboard },
              { id: 'orders', label: 'Đơn hàng', icon: Package },
              { id: 'products', label: 'Sản phẩm (CRUD)', icon: ShoppingBag },
              { id: 'tickets', label: 'Hỗ trợ (Tickets)', icon: Headphones },
              { id: 'users', label: 'Người dùng', icon: Users },
            ].map((item) => {
              const Icon = item.icon;
              const isActive = activeTab === item.id;
              return (
                <button
                  key={item.id}
                  type="button"
                  onClick={() => setActiveTab(item.id as AdminTab)}
                  className={`flex w-full items-center gap-3 rounded-control px-3 py-2 text-xs font-semibold transition-colors ${
                    isActive
                      ? 'bg-teal-50 text-teal-700 dark:bg-teal-950/60 dark:text-teal-300'
                      : 'text-slate-600 hover:bg-slate-100 hover:text-slate-900 dark:text-slate-400 dark:hover:bg-slate-800 dark:hover:text-slate-200'
                  }`}
                  title={item.label}
                >
                  <Icon className="h-4 w-4 shrink-0" />
                  {!sidebarCollapsed && <span>{item.label}</span>}
                </button>
              );
            })}
          </nav>
        </div>

        <div className="p-3 border-t border-slate-100 dark:border-slate-800">
          {!sidebarCollapsed && (
            <div className="text-[11px] text-slate-400 font-mono">
              Admin: {user?.email}
            </div>
          )}
        </div>
      </aside>

      {/* Main Content Area */}
      <main className="flex-1 p-6 lg:p-8 overflow-y-auto max-w-7xl">
        {/* Tab 1: Dashboard / KPIs */}
        {activeTab === 'dashboard' && (
          <div className="space-y-8">
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-slate-900 dark:text-white">
                Bảng điều khiển Quản trị (KPI)
              </h1>
              <p className="text-xs text-slate-500 mt-1">
                Theo dõi tổng quan doanh thu, số lượng đơn hàng và người dùng hệ thống
              </p>
            </div>

            {/* KPI Cards */}
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-6">
              <div className="rounded-card border border-slate-200 bg-white p-6 shadow-sm dark:border-slate-800 dark:bg-slate-900">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-semibold text-slate-500">Tổng doanh thu</span>
                  <div className="flex h-8 w-8 items-center justify-center rounded-full bg-teal-50 text-teal-600 dark:bg-teal-950 dark:text-teal-400">
                    <TrendingUp className="h-4 w-4" />
                  </div>
                </div>
                <div className="mt-2 text-2xl font-bold font-mono text-slate-900 dark:text-white">
                  {formatVND(stats?.total_revenue || 0)}
                </div>
                <div className="mt-1 text-[11px] text-teal-600 font-medium">Doanh thu tích lũy</div>
              </div>

              <div className="rounded-card border border-slate-200 bg-white p-6 shadow-sm dark:border-slate-800 dark:bg-slate-900">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-semibold text-slate-500">Tổng số đơn hàng</span>
                  <div className="flex h-8 w-8 items-center justify-center rounded-full bg-blue-50 text-blue-600 dark:bg-blue-950 dark:text-blue-400">
                    <Package className="h-4 w-4" />
                  </div>
                </div>
                <div className="mt-2 text-2xl font-bold font-mono text-slate-900 dark:text-white">
                  {stats?.total_orders || 0}
                </div>
                <div className="mt-1 text-[11px] text-blue-600 font-medium">
                  {stats?.orders_by_status?.PAID || 0} đơn đã thanh toán
                </div>
              </div>

              <div className="rounded-card border border-slate-200 bg-white p-6 shadow-sm dark:border-slate-800 dark:bg-slate-900">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-semibold text-slate-500">Tổng người dùng</span>
                  <div className="flex h-8 w-8 items-center justify-center rounded-full bg-purple-50 text-purple-600 dark:bg-purple-950 dark:text-purple-400">
                    <Users className="h-4 w-4" />
                  </div>
                </div>
                <div className="mt-2 text-2xl font-bold font-mono text-slate-900 dark:text-white">
                  {stats?.total_users || 0}
                </div>
                <div className="mt-1 text-[11px] text-purple-600 font-medium">Tài khoản khách & admin</div>
              </div>
            </div>

            {/* Lazy Loaded Revenue Chart */}
            <div className="rounded-card border border-slate-200 bg-white p-6 shadow-sm dark:border-slate-800 dark:bg-slate-900 space-y-2">
              <h2 className="text-sm font-bold text-slate-900 dark:text-white">
                Biểu đồ Doanh thu theo Ngày (Lazy-loaded Recharts)
              </h2>
              <Suspense fallback={<Skeleton className="h-72 w-full rounded" />}>
                <AdminRevenueChart data={stats?.revenue_by_day || []} />
              </Suspense>
            </div>
          </div>
        )}

        {/* Tab 2: Orders Management */}
        {activeTab === 'orders' && (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <div>
                <h1 className="text-2xl font-bold tracking-tight text-slate-900 dark:text-white">
                  Quản lý Đơn hàng
                </h1>
                <p className="text-xs text-slate-500 mt-1">
                  Cập nhật trạng thái hợp lệ của đơn hàng (PAID → SHIPPED → DELIVERED)
                </p>
              </div>
            </div>

            <div className="rounded-card border border-slate-200 bg-white overflow-hidden shadow-sm dark:border-slate-800 dark:bg-slate-900">
              <div className="overflow-x-auto">
                <table className="w-full text-left text-xs">
                  <thead className="bg-slate-50 border-b border-slate-200 dark:bg-slate-800/50 dark:border-slate-800 text-slate-600 dark:text-slate-400 font-semibold">
                    <tr>
                      <th className="p-3">Mã đơn</th>
                      <th className="p-3">Người nhận</th>
                      <th className="p-3">Tổng tiền</th>
                      <th className="p-3">Trạng thái</th>
                      <th className="p-3">Thời gian</th>
                      <th className="p-3 text-right">Chuyển trạng thái</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
                    {ordersData?.items?.map((ord) => (
                      <tr key={ord.id} className="hover:bg-slate-50/50 dark:hover:bg-slate-800/50">
                        <td className="p-3 font-mono font-bold">{ord.id}</td>
                        <td className="p-3">
                          <div className="font-semibold">{ord.shipping_name || ord.recipient_name}</div>
                          <div className="text-[11px] text-slate-400">{ord.shipping_phone || ord.recipient_phone}</div>
                        </td>
                        <td className="p-3 font-mono font-bold text-teal-600 dark:text-teal-400">
                          {formatVND(ord.total)}
                        </td>
                        <td className="p-3">
                          <OrderStatusBadge status={ord.status} />
                        </td>
                        <td className="p-3 font-mono text-slate-400">
                          {formatDateTime(ord.created_at)}
                        </td>
                        <td className="p-3 text-right space-x-1.5">
                          {ord.status === 'PAID' && (
                            <Button
                              size="sm"
                              variant="outline"
                              onClick={() => updateOrderStatusMutation.mutate({ orderId: ord.id, status: 'SHIPPED' })}
                              className="text-purple-600 h-7 text-[11px]"
                            >
                              <Truck className="h-3 w-3 mr-1" />
                              Giao hàng (SHIPPED)
                            </Button>
                          )}

                          {ord.status === 'SHIPPED' && (
                            <Button
                              size="sm"
                              onClick={() => updateOrderStatusMutation.mutate({ orderId: ord.id, status: 'DELIVERED' })}
                              className="h-7 text-[11px]"
                            >
                              <PackageCheck className="h-3 w-3 mr-1" />
                              Đã giao (DELIVERED)
                            </Button>
                          )}

                          {(ord.status === 'PENDING' || ord.status === 'PAID') && (
                            <Button
                              size="sm"
                              variant="outline"
                              onClick={() => updateOrderStatusMutation.mutate({ orderId: ord.id, status: 'CANCELLED' })}
                              className="text-rose-600 h-7 text-[11px]"
                            >
                              Hủy
                            </Button>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        )}

        {/* Tab 3: Products CRUD */}
        {activeTab === 'products' && (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <div>
                <h1 className="text-2xl font-bold tracking-tight text-slate-900 dark:text-white">
                  Quản lý Sản phẩm (CRUD)
                </h1>
                <p className="text-xs text-slate-500 mt-1">
                  Thêm mới, sửa đổi thông tin giá, kho hoặc lưu trữ sản phẩm trong cơ sở dữ liệu
                </p>
              </div>

              <Button onClick={handleOpenCreateProduct} className="gap-1.5">
                <Plus className="h-4 w-4" />
                <span>Thêm sản phẩm</span>
              </Button>
            </div>

            <div className="rounded-card border border-slate-200 bg-white overflow-hidden shadow-sm dark:border-slate-800 dark:bg-slate-900">
              <div className="overflow-x-auto">
                <table className="w-full text-left text-xs">
                  <thead className="bg-slate-50 border-b border-slate-200 dark:bg-slate-800/50 dark:border-slate-800 text-slate-600 dark:text-slate-400 font-semibold">
                    <tr>
                      <th className="p-3">Ảnh</th>
                      <th className="p-3">Tên sản phẩm</th>
                      <th className="p-3">Giá bán</th>
                      <th className="p-3">Kho</th>
                      <th className="p-3">Trạng thái</th>
                      <th className="p-3 text-right">Thao tác</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
                    {productsData?.items?.map((p) => (
                      <tr key={p.id} className="hover:bg-slate-50/50 dark:hover:bg-slate-800/50">
                        <td className="p-3">
                          <img
                            src={p.image_path || '/static/img/products/p01.jpg'}
                            alt=""
                            className="h-10 w-10 rounded object-cover"
                          />
                        </td>
                        <td className="p-3">
                          <div className="font-semibold text-slate-900 dark:text-white">{p.name}</div>
                          <div className="font-mono text-[10px] text-slate-400">{p.slug}</div>
                        </td>
                        <td className="p-3 font-mono font-bold">{formatVND(p.price)}</td>
                        <td className="p-3 font-mono">{p.stock}</td>
                        <td className="p-3">
                          <span
                            className={`rounded px-2 py-0.5 text-[10px] font-bold ${
                              p.status === 'active'
                                ? 'bg-teal-100 text-teal-800 dark:bg-teal-950 dark:text-teal-300'
                                : 'bg-slate-100 text-slate-800'
                            }`}
                          >
                            {p.status}
                          </span>
                        </td>
                        <td className="p-3 text-right space-x-2">
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => handleOpenEditProduct(p)}
                            className="h-7 px-2 text-xs"
                          >
                            <Edit2 className="h-3 w-3" />
                          </Button>
                          <Button
                            size="sm"
                            variant="destructive"
                            onClick={() => deleteProductMutation.mutate(p.id)}
                            className="h-7 px-2 text-xs"
                          >
                            <Trash2 className="h-3 w-3" />
                          </Button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        )}

        {/* Tab 4: Tickets Management */}
        {activeTab === 'tickets' && (
          <div className="space-y-6">
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-slate-900 dark:text-white">
                Quản lý Yêu cầu Hỗ trợ (Tickets)
              </h1>
              <p className="text-xs text-slate-500 mt-1">
                Phản hồi khách hàng và cập nhật đóng yêu cầu
              </p>
            </div>

            <div className="rounded-card border border-slate-200 bg-white overflow-hidden shadow-sm dark:border-slate-800 dark:bg-slate-900">
              <div className="divide-y divide-slate-100 dark:divide-slate-800">
                {ticketsData?.items?.map((t: any) => (
                  <div key={t.id} className="p-4 space-y-3">
                    <div className="flex items-center justify-between">
                      <div>
                        <div className="font-bold text-sm text-slate-900 dark:text-white">
                          {t.subject}
                        </div>
                        <div className="text-xs text-slate-400 font-mono">
                          Người gửi: {t.user_email} • {formatDateTime(t.created_at)}
                        </div>
                      </div>
                      <span
                        className={`rounded px-2 py-0.5 text-xs font-bold ${
                          t.status === 'OPEN'
                            ? 'bg-amber-100 text-amber-800'
                            : t.status === 'RESOLVED'
                            ? 'bg-teal-100 text-teal-800'
                            : 'bg-slate-100 text-slate-800'
                        }`}
                      >
                        {t.status}
                      </span>
                    </div>

                    {/* Messages in ticket */}
                    <div className="space-y-2 max-h-40 overflow-y-auto bg-slate-50 dark:bg-slate-950/40 p-3 rounded text-xs">
                      {t.messages?.map((m: any) => (
                        <div key={m.id} className="border-b border-slate-100 dark:border-slate-800 pb-1">
                          <strong className={m.sender_role === 'admin' ? 'text-teal-600' : 'text-slate-800 dark:text-slate-200'}>
                            {m.sender_role === 'admin' ? '[Admin]' : '[Khách hàng]'}:
                          </strong>{' '}
                          <span>{m.message}</span>
                        </div>
                      ))}
                    </div>

                    {t.status !== 'CLOSED' && (
                      <div className="flex justify-end gap-2">
                        <Button
                          size="sm"
                          onClick={() => {
                            setReplyTicketId(t.id);
                            setReplyContent('');
                          }}
                        >
                          Trả lời & Xử lý
                        </Button>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}

        {/* Tab 5: Users List */}
        {activeTab === 'users' && (
          <div className="space-y-6">
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-slate-900 dark:text-white">
                Danh sách Người dùng
              </h1>
              <p className="text-xs text-slate-500 mt-1">
                Tài khoản khách hàng và quản trị viên đã đăng ký
              </p>
            </div>

            <div className="rounded-card border border-slate-200 bg-white overflow-hidden shadow-sm dark:border-slate-800 dark:bg-slate-900">
              <table className="w-full text-left text-xs">
                <thead className="bg-slate-50 border-b border-slate-200 dark:bg-slate-800/50 dark:border-slate-800 text-slate-600 font-semibold">
                  <tr>
                    <th className="p-3">Mã ID</th>
                    <th className="p-3">Họ và tên</th>
                    <th className="p-3">Email</th>
                    <th className="p-3">Vai trò</th>
                    <th className="p-3">Ngày tạo</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100 dark:divide-slate-800 font-mono">
                  {usersData?.items?.map((u: any) => (
                    <tr key={u.id} className="hover:bg-slate-50/50">
                      <td className="p-3">{u.id}</td>
                      <td className="p-3 font-sans font-semibold">{u.name}</td>
                      <td className="p-3">{u.email}</td>
                      <td className="p-3">
                        <span
                          className={`rounded px-2 py-0.5 text-[10px] font-bold ${
                            u.role === 'admin' ? 'bg-rose-100 text-rose-800' : 'bg-slate-100 text-slate-800'
                          }`}
                        >
                          {u.role}
                        </span>
                      </td>
                      <td className="p-3 text-slate-400">{formatDateTime(u.created_at)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </main>

      {/* Product CRUD Modal */}
      <Dialog
        open={productModalOpen}
        onOpenChange={setProductModalOpen}
        title={editingProduct ? 'Chỉnh sửa Sản phẩm' : 'Thêm Sản phẩm Mới'}
        description="Điền thông tin chi tiết sản phẩm cho danh mục bán lẻ"
      >
        <form
          onSubmit={(e) => {
            e.preventDefault();
            saveProductMutation.mutate();
          }}
          className="space-y-4 text-xs"
        >
          <div>
            <label className="font-semibold block mb-1">Tên sản phẩm *</label>
            <Input
              required
              value={prodForm.name}
              onChange={(e) => setProdForm({ ...prodForm, name: e.target.value })}
            />
          </div>

          <div>
            <label className="font-semibold block mb-1">Slug URL *</label>
            <Input
              required
              value={prodForm.slug}
              onChange={(e) => setProdForm({ ...prodForm, slug: e.target.value })}
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="font-semibold block mb-1">Giá bán (VNĐ) *</label>
              <Input
                type="number"
                required
                value={prodForm.price}
                onChange={(e) => setProdForm({ ...prodForm, price: Number(e.target.value) })}
              />
            </div>
            <div>
              <label className="font-semibold block mb-1">Số lượng kho *</label>
              <Input
                type="number"
                required
                value={prodForm.stock}
                onChange={(e) => setProdForm({ ...prodForm, stock: Number(e.target.value) })}
              />
            </div>
          </div>

          <div>
            <label className="font-semibold block mb-1">Đường dẫn ảnh *</label>
            <Input
              required
              value={prodForm.image_path}
              onChange={(e) => setProdForm({ ...prodForm, image_path: e.target.value })}
            />
          </div>

          <div>
            <label className="font-semibold block mb-1">Mô tả sản phẩm *</label>
            <Textarea
              required
              rows={3}
              value={prodForm.description}
              onChange={(e) => setProdForm({ ...prodForm, description: e.target.value })}
            />
          </div>

          <div className="flex justify-end gap-2 pt-3 border-t">
            <Button type="button" variant="outline" onClick={() => setProductModalOpen(false)}>
              Hủy
            </Button>
            <Button type="submit" loading={saveProductMutation.isPending}>
              Lưu sản phẩm
            </Button>
          </div>
        </form>
      </Dialog>

      {/* Reply Ticket Modal */}
      <Dialog
        open={!!replyTicketId}
        onOpenChange={(open) => !open && setReplyTicketId(null)}
        title="Trả lời Ticket Hỗ trợ"
        description="Gửi phản hồi chính thức từ quản trị viên tới khách hàng"
      >
        <form
          onSubmit={(e) => {
            e.preventDefault();
            replyTicketMutation.mutate();
          }}
          className="space-y-4 text-xs"
        >
          <div>
            <label className="font-semibold block mb-1">Nội dung phản hồi *</label>
            <Textarea
              required
              rows={4}
              placeholder="Nhập câu trả lời hoặc hướng dẫn xử lý..."
              value={replyContent}
              onChange={(e) => setReplyContent(e.target.value)}
            />
          </div>

          <label className="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              checked={closeOnReply}
              onChange={(e) => setCloseOnReply(e.target.checked)}
            />
            <span>Đồng thời đóng ticket này (Mark as Closed)</span>
          </label>

          <div className="flex justify-end gap-2 pt-3 border-t">
            <Button type="button" variant="outline" onClick={() => setReplyTicketId(null)}>
              Hủy
            </Button>
            <Button type="submit" loading={replyTicketMutation.isPending}>
              Gửi phản hồi
            </Button>
          </div>
        </form>
      </Dialog>
    </div>
  );
}
