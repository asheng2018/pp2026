import React from 'react';
import { BrowserRouter, Routes, Route, Navigate, Outlet } from 'react-router-dom';
import { ConfigProvider, App as AntApp, Spin } from 'antd';
import { QueryClient, QueryClientProvider } from 'react-query';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import MainLayout from './components/Layout/MainLayout';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import Orders from './pages/Orders';
import Accounts from './pages/Accounts';
import Merchants from './pages/Merchants';
import Settlements from './pages/Settlements';
import RiskCenter from './pages/RiskCenter';
import ProxyPool from './pages/ProxyPool';
import BSites from './pages/BSites';
import ExchangeRate from './pages/ExchangeRate';
import Logistics from './pages/Logistics';
import Reports from './pages/Reports';
import Settings from './pages/Settings';

const queryClient = new QueryClient();

const ProtectedRoute: React.FC = () => {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return <Outlet />;
};

const App: React.FC = () => {
  return (
    <QueryClientProvider client={queryClient}>
      <ConfigProvider
        theme={{
          token: {
            colorPrimary: '#1677ff',
            borderRadius: 6,
          },
        }}
      >
        <AntApp>
          <AuthProvider>
            <BrowserRouter>
              <Routes>
                <Route path="/login" element={<Login />} />
                <Route element={<ProtectedRoute />}>
                  <Route path="/" element={<MainLayout />}>
                    <Route index element={<Navigate to="/dashboard" replace />} />
                    <Route path="dashboard" element={<Dashboard />} />
                    <Route path="orders" element={<Orders />} />
                    <Route path="accounts" element={<Accounts />} />
                    <Route path="merchants" element={<Merchants />} />
                    <Route path="settlements" element={<Settlements />} />
                    <Route path="risk" element={<RiskCenter />} />
                    <Route path="proxies" element={<ProxyPool />} />
                    <Route path="b-sites" element={<BSites />} />
                    <Route path="exchange" element={<ExchangeRate />} />
                    <Route path="logistics" element={<Logistics />} />
                    <Route path="reports" element={<Reports />} />
                    <Route path="settings" element={<Settings />} />
                  </Route>
                </Route>
              </Routes>
            </BrowserRouter>
          </AuthProvider>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>
  );
};

export default App;
