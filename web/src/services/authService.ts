import api from './api';

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  admin_id: string;
  username: string;
  role: string;
  expires_at: string;
}

export interface AdminUser {
  id: string;
  username: string;
  role: string;
  totp_enabled: boolean;
  last_login_at: string | null;
  created_at: string;
}

export const authService = {
  login: (data: LoginRequest) =>
    api.post<LoginResponse>('/admin/login', data).then((res) => res.data),

  logout: () => api.post('/admin/logout'),

  getMe: () => api.get<AdminUser>('/admin/me').then((res) => res.data),
};
