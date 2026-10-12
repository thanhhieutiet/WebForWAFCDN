// Core domain and application types for LabShop

export type UserRole = 'customer' | 'admin';

export interface User {
  id: string;
  email: string;
  name: string;
  phone?: string;
  address?: string;
  role: UserRole;
  created_at: string;
  updated_at?: string;
}

export interface Category {
  id: string;
  name: string;
  slug: string;
  description?: string;
  created_at: string;
}

export type ProductStatus = 'active' | 'archived';

export interface Product {
  id: string;
  category_id: string;
  category_name?: string;
  name: string;
  slug: string;
  description: string;
  price: number; // in VND
  stock: number;
  image_path: string;
  status: ProductStatus;
  rating_avg: number;
  rating_count: number;
  created_at: string;
  updated_at: string;
}

export interface ProductListResponse {
  items: Product[];
  total: number;
  page: number;
  page_size: number;
}

export interface Review {
  id: string;
  product_id: string;
  user_id: string;
  user_name: string;
  rating: number; // 1-5
  comment: string;
  created_at: string;
}

export interface CartItem {
  product_id: string;
  product_name: string;
  slug: string;
  price: number;
  quantity: number;
  image_path: string;
}

export interface CartView {
  user_id: string;
  items: CartItem[];
  subtotal: number;
  discount: number;
  shipping_fee: number;
  total: number;
  coupon_code: string;
  item_count: number;
}

export type OrderStatus = 'PENDING' | 'PAID' | 'SHIPPED' | 'DELIVERED' | 'CANCELLED';
export type PaymentMethod = 'mock_gateway' | 'cod';

export interface OrderItem {
  id?: string;
  product_id: string;
  product_name: string;
  price?: number;
  product_price?: number;
  quantity: number;
  subtotal?: number;
  image_path?: string;
}

export interface Order {
  id: string;
  user_id: string;
  status: OrderStatus;
  recipient_name?: string;
  recipient_phone?: string;
  shipping_name?: string;
  shipping_phone?: string;
  shipping_address: string;
  subtotal: number;
  discount: number;
  shipping_fee: number;
  total: number;
  coupon_code?: string;
  payment_method: PaymentMethod;
  idempotency_key?: string;
  created_at: string;
  updated_at: string;
  items: OrderItem[];
}

export type TicketStatus = 'OPEN' | 'RESOLVED' | 'CLOSED';

export interface TicketAttachment {
  id: string;
  filename: string;
  url: string;
  size?: number;
}

export interface TicketMessage {
  id: string;
  sender_role: 'customer' | 'admin' | 'staff';
  sender_name: string;
  content: string;
  created_at: string;
}

export interface Ticket {
  id: string;
  user_id: string;
  user_email?: string;
  user_name?: string;
  subject: string;
  status: TicketStatus;
  created_at: string;
  updated_at: string;
  messages: TicketMessage[];
  attachments?: TicketAttachment[];
}

// Lab & HTTP observability types
export type LabBadgeType = 'origin' | 'hit' | 'miss' | 'blocked' | 'stale' | 'compressed';

export interface LabRequestLog {
  id: string;
  timestamp: string;
  method: string;
  path: string;
  fullUrl: string;
  status: number;
  latencyMs: number;
  requestHeaders: Record<string, string>;
  responseHeaders: Record<string, string>;
  requestBodyPreview?: string;
  responseBodyPreview?: string;
  badges: LabBadgeType[];
  originInstance?: string;
  originTime?: string;
  requestId?: string;
  cacheStatus?: string;
  age?: string;
  contentEncoding?: string;
  server?: string;
  via?: string;
  speculatedEdgeBlocked: boolean;
  speculationEvidence?: string;
}

export type RoutingMode = 'direct' | 'waap' | 'cdn' | 'full' | 'custom' | 'unknown';

export interface RoutingHostConfig {
  direct: string;
  waap: string;
  cdn: string;
  full: string;
}

// Tutorial & Scenario models
export type ScenarioGroup = 'cdn' | 'waap' | 'shop';
export type ScenarioPriority = 'P0' | 'P1';

export interface ScenarioRequest {
  method: string;
  path: string;
  headers?: Record<string, string>;
  body?: string;
  curl?: string;
  runnable: boolean;
  nonRunnableReason?: string;
  isProbe?: boolean;
}

export interface ScenarioExpect {
  condition: string;
  expectedResult: string;
}

export interface Scenario {
  id: string;
  group: ScenarioGroup;
  title: string;
  goal: string;
  priority: ScenarioPriority;
  portalSteps: string[];
  requests: ScenarioRequest[];
  expect: ScenarioExpect[];
  notes?: string;
  docsUrl?: string;
}
