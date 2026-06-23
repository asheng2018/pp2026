import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { authService, LoginResponse, AdminUser } from '../services/authService';

interface AuthState {
  user: AdminUser | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

interface AuthContextValue extends AuthState {
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

const TOKEN_KEY = 'ab_admin_token';

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [state, setState] = useState<AuthState>({
    user: null,
    token: localStorage.getItem(TOKEN_KEY),
    isAuthenticated: false,
    isLoading: true,
  });

  // On mount, verify any stored token
  useEffect(() => {
    const initAuth = async () => {
      const storedToken = localStorage.getItem(TOKEN_KEY);
      if (!storedToken) {
        setState({ user: null, token: null, isAuthenticated: false, isLoading: false });
        return;
      }

      try {
        const user = await authService.getMe();
        setState({ user, token: storedToken, isAuthenticated: true, isLoading: false });
      } catch {
        // Token is invalid or expired
        localStorage.removeItem(TOKEN_KEY);
        setState({ user: null, token: null, isAuthenticated: false, isLoading: false });
      }
    };

    initAuth();
  }, []);

  const login = useCallback(async (username: string, password: string) => {
    const result: LoginResponse = await authService.login({ username, password });
    localStorage.setItem(TOKEN_KEY, result.token);
    setState({
      user: {
        id: result.admin_id,
        username: result.username,
        role: result.role,
        totp_enabled: false,
        last_login_at: null,
        created_at: '',
      },
      token: result.token,
      isAuthenticated: true,
      isLoading: false,
    });
  }, []);

  const logout = useCallback(async () => {
    try {
      await authService.logout();
    } catch {
      // Even if the API call fails, clear local state
    }
    localStorage.removeItem(TOKEN_KEY);
    setState({ user: null, token: null, isAuthenticated: false, isLoading: false });
  }, []);

  return (
    <AuthContext.Provider value={{ ...state, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
};

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return ctx;
}
