import api from './api';

// --- Auth ---
export interface LoginRequest { username: string; password: string; }
export interface LoginResponse { token: string; admin_id: string; username: string; role: string; expires_at: string; }
export interface AdminUser { id: string; username: string; role: string; totp_enabled: boolean; last_login_at: string | null; created_at: string; }

export const authService = {
  login: (data: LoginRequest) => api.post<LoginResponse>('/admin/login', data).then(r => r.data),
  logout: () => api.post('/admin/logout'),
  getMe: () => api.get<AdminUser>('/admin/me').then(r => r.data),
};

// --- Dashboard ---
export interface DashboardStats {
  total_orders: number; today_orders: number; today_revenue: string;
  active_accounts: number; active_merchants: number; success_rate: number;
  pending_orders: number; failed_orders: number;
}
export interface DailyRevenue { date: string; orders: number; revenue: string; }

export const dashboardService = {
  getStats: () => api.get<DashboardStats>('/admin/dashboard/stats').then(r => r.data),
  getRevenue: () => api.get<DailyRevenue[]>('/admin/dashboard/revenue').then(r => r.data),
};

// --- Orders ---
export interface OrderRow {
  id: string; order_no: string; merchant_id: string; account_id: string;
  gateway: string; amount: string; currency: string; status: string;
  customer_email: string; customer_country: string; risk_level: string;
  risk_score: number; created_at: string; updated_at: string;
}
export interface OrderDetail extends OrderRow {
  customer_ip: string; gateway_order_id: string; a_site_referer: string;
  paid_at: string | null; expired_at: string | null; canceled_at: string | null;
  callback_data: any; metadata: any;
}
export interface OrderListRes { orders: OrderRow[]; total: number; }

export const orderService = {
  list: (status?: string) => api.get<OrderListRes>('/admin/orders', { params: status ? { status } : {} }).then(r => r.data),
  getById: (id: string) => api.get<OrderDetail>('/admin/orders/' + id).then(r => r.data),
};

// --- Accounts ---
export interface AccountRow {
  id: string; gateway: string; alias: string; status: string;
  b_site_id: string; merchant_id: string; weight: number; priority: number;
  tags: string[]; limit_config: any; created_at: string;
}
export interface AccountCreateData {
  gateway: string; alias: string; b_site_id?: string; merchant_id?: string;
  paypal_client_id?: string; paypal_secret?: string;
  stripe_publishable_key?: string; stripe_secret_key?: string;
  single_min?: string; single_max?: string; daily_max?: string; monthly_max?: string;
  weight?: number; priority?: number; tags?: string[];
}
export const accountService = {
  list: () => api.get<{ accounts: AccountRow[]; total: number }>('/admin/accounts').then(r => r.data),
  create: (data: AccountCreateData) => api.post('/admin/accounts', data).then(r => r.data),
  update: (id: string, data: any) => api.put('/admin/accounts/' + id, data).then(r => r.data),
  delete: (id: string) => api.delete('/admin/accounts/' + id).then(r => r.data),
  updateStatus: (id: string, status: string) => api.patch('/admin/accounts/' + id + '/status', { status }).then(r => r.data),
};

// --- Merchants ---
export interface MerchantRow { id: string; name: string; email: string; status: string; routing_mode: string; created_at: string; }
export interface APIToken { key_id: string; api_key: string; key_prefix: string; permissions: string[]; }
export interface MerchantDetail extends MerchantRow {
  risk_profile?: any; fee_config?: any; settlement_config?: any; metadata?: any;
}
export const merchantService = {
  list: () => api.get<{ merchants: MerchantRow[]; total: number }>('/admin/merchants').then(r => r.data),
  create: (data: { name: string; email?: string; routing_mode?: string; fee_rate?: number; }) => api.post('/admin/merchants', data).then(r => r.data),
  getById: (id: string) => api.get<MerchantDetail>('/admin/merchants/' + id).then(r => r.data),
  update: (id: string, data: any) => api.put('/admin/merchants/' + id, data).then(r => r.data),
  delete: (id: string) => api.delete('/admin/merchants/' + id).then(r => r.data),
  generateAPIKey: (id: string, data?: any) => api.post<APIToken>('/admin/merchants/' + id + '/apikeys', data || {}).then(r => r.data),
  listAPIKeys: (id: string) => api.get<{ keys: any[]; total: number }>('/admin/merchants/' + id + '/apikeys').then(r => r.data),
  revokeAPIKey: (mid: string, kid: string) => api.delete('/admin/merchants/' + mid + '/apikeys/' + kid).then(r => r.data),
};

// --- Settlements ---
export interface SettlementRow { id: string; merchant_id: string; cycle_start: string; cycle_end: string; total_orders: number; total_amount: string; total_fee: string; net_amount: string; payout_method: string; status: string; created_at: string; }
export const settlementService = { list: () => api.get<{ settlements: SettlementRow[]; total: number }>('/admin/settlements').then(r => r.data) };

// --- Proxies ---
export interface ProxyRow { id: string; proxy_type: string; host: string; port: number; username?: string; protocol?: string; country: string; city?: string; isp: string; status: string; latency: number; success_rate: number; bound_account_id?: string; is_dedicated?: boolean; created_at: string; }
export const proxyService = {
  list: () => api.get<{ proxies: ProxyRow[]; total: number }>('/admin/proxies').then(r => r.data),
  create: (data: any) => api.post('/admin/proxies', data).then(r => r.data),
  batchImport: (proxies: any[]) => api.post<{ imported: number; total: number }>('/admin/proxies/batch', { proxies }).then(r => r.data),
  update: (id: string, data: any) => api.put('/admin/proxies/' + id, data).then(r => r.data),
  delete: (id: string) => api.delete('/admin/proxies/' + id).then(r => r.data),
  updateStatus: (id: string, status: string) => api.patch('/admin/proxies/' + id + '/status', { status }).then(r => r.data),
};

// --- B-Sites ---
export interface BSiteRow { id: string; domain: string; name: string; hosting_ip: string; status: string; created_at: string; }
export const bSiteService = {
  list: () => api.get<{ b_sites: BSiteRow[]; total: number }>('/admin/b-sites').then(r => r.data),
  create: (data: any) => api.post('/admin/b-sites', data).then(r => r.data),
  getById: (id: string) => api.get('/admin/b-sites/' + id).then(r => r.data),
  update: (id: string, data: any) => api.put('/admin/b-sites/' + id, data).then(r => r.data),
  delete: (id: string) => api.delete('/admin/b-sites/' + id).then(r => r.data),
  updateStatus: (id: string, status: string) => api.patch('/admin/b-sites/' + id + '/status', { status }).then(r => r.data),
};

// --- Risk Events ---
export interface RiskRow { id: string; merchant_id: string; order_id: string; rule_name: string; risk_level: string; risk_score: number; action: string; reason: string; created_at: string; }
export const riskService = { list: () => api.get<{ events: RiskRow[]; total: number }>('/admin/risk-events').then(r => r.data) };

// --- Exchange Rates ---
export interface ExchangeRow { id: string; base_currency: string; target_currency: string; rate: string; source: string; fetched_at: string; }
export const exchangeService = { list: () => api.get<{ rates: ExchangeRow[]; total: number }>('/admin/exchange-rates').then(r => r.data) };

// --- Logistics ---
export interface LogisticsRow { id: string; order_id: string; tracking_number: string; carrier: string; status: string; synced_to_b_site: boolean; created_at: string; }
export const logisticsService = { list: () => api.get<{ tracks: LogisticsRow[]; total: number }>('/admin/logistics').then(r => r.data) };
