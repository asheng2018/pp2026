import React, { useState } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { Layout, Menu, Button, Avatar, Dropdown, theme } from 'antd';
import {
  DashboardOutlined, ShoppingCartOutlined, CreditCardOutlined,
  TeamOutlined, DollarOutlined, SafetyOutlined, GlobalOutlined,
  CloudServerOutlined, SwapOutlined, TruckOutlined,
  BarChartOutlined, SettingOutlined, MenuFoldOutlined, MenuUnfoldOutlined,
  UserOutlined, LogoutOutlined,
} from '@ant-design/icons';
import { useAuth } from '../../contexts/AuthContext';

const { Header, Sider, Content } = Layout;

const menuItems = [
  { key: '/dashboard', icon: <DashboardOutlined />, label: 'Dashboard' },
  { key: '/orders', icon: <ShoppingCartOutlined />, label: 'Orders' },
  { key: '/accounts', icon: <CreditCardOutlined />, label: 'Accounts' },
  { key: '/merchants', icon: <TeamOutlined />, label: 'Merchants' },
  { key: '/settlements', icon: <DollarOutlined />, label: 'Settlements' },
  { key: '/risk', icon: <SafetyOutlined />, label: 'Risk Center' },
  { key: '/proxies', icon: <GlobalOutlined />, label: 'Proxy Pool' },
  { key: '/b-sites', icon: <CloudServerOutlined />, label: 'B-Sites' },
  { key: '/exchange', icon: <SwapOutlined />, label: 'Exchange Rate' },
  { key: '/logistics', icon: <TruckOutlined />, label: 'Logistics' },
  { key: '/reports', icon: <BarChartOutlined />, label: 'Reports' },
  { key: '/settings', icon: <SettingOutlined />, label: 'Settings' },
];

const MainLayout: React.FC = () => {
  const [collapsed, setCollapsed] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { token: { colorBgContainer } } = theme.useToken();
  const { user, logout } = useAuth();

  const userMenuItems = [
    { key: 'profile', icon: <UserOutlined />, label: 'Profile' },
    { key: 'logout', icon: <LogoutOutlined />, label: 'Logout', danger: true },
  ];

  const handleUserMenuClick = async ({ key }: { key: string }) => {
    if (key === 'logout') {
      await logout();
      navigate('/login', { replace: true });
    }
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        trigger={null}
        collapsible
        collapsed={collapsed}
        style={{ background: colorBgContainer }}
        width={240}
      >
        <div style={{
          height: 64, display: 'flex', alignItems: 'center', justifyContent: 'center',
          borderBottom: '1px solid #f0f0f0', fontWeight: 700, fontSize: collapsed ? 14 : 18,
        }}>
          {collapsed ? 'AB' : 'AB Payment System'}
        </div>
        <Menu
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
          style={{ borderRight: 0 }}
        />
      </Sider>
      <Layout>
        <Header style={{
          padding: '0 24px', background: colorBgContainer,
          display: 'flex', alignItems: 'center', justifyContent: 'space-between',
          boxShadow: '0 1px 4px rgba(0,0,0,0.08)',
        }}>
          <Button
            type="text"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setCollapsed(!collapsed)}
          />
          <Dropdown menu={{ items: userMenuItems, onClick: handleUserMenuClick }}>
            <div style={{ cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 8 }}>
              <Avatar icon={<UserOutlined />} />
              <span>{user?.username || 'Admin'}</span>
            </div>
          </Dropdown>
        </Header>
        <Content>
          <div className="content-wrapper">
            <Outlet />
          </div>
        </Content>
      </Layout>
    </Layout>
  );
};

export default MainLayout;
