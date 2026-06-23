import React, { useEffect, useState } from 'react';
import { Row, Col, Card, Statistic, Table, Spin, Typography } from 'antd';
import {
  ShoppingCartOutlined, CreditCardOutlined,
  TeamOutlined, CheckCircleOutlined, WarningOutlined,
} from '@ant-design/icons';
import { dashboardService, DashboardStats, DailyRevenue } from '../../services/dataService';
import dayjs from 'dayjs';

const { Title } = Typography;

const Dashboard: React.FC = () => {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [revenue, setRevenue] = useState<DailyRevenue[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const load = async () => {
      try {
        const [s, r] = await Promise.all([dashboardService.getStats(), dashboardService.getRevenue()]);
        setStats(s); setRevenue(r);
      } catch { /* ignore */ } finally { setLoading(false); }
    };
    load();
  }, []);

  if (loading) return <div style={{ display: 'flex', justifyContent: 'center', padding: 100 }}><Spin size="large" /></div>;

  const revenueColumns = [
    { title: 'Date', dataIndex: 'date', key: 'date', render: (v: string) => dayjs(v).format('MMM DD') },
    { title: 'Orders', dataIndex: 'orders', key: 'orders' },
    { title: 'Revenue', dataIndex: 'revenue', key: 'revenue', render: (v: string) => `$${Number(v).toFixed(2)}` },
  ];

  return (
    <div>
      <Title level={3} style={{ marginBottom: 24 }}>Dashboard Overview</Title>
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}><Card><Statistic title="Total Orders" value={stats?.total_orders ?? 0} prefix={<ShoppingCartOutlined />} /></Card></Col>
        <Col xs={24} sm={12} lg={6}><Card><Statistic title="Today Orders" value={stats?.today_orders ?? 0} prefix={<ShoppingCartOutlined />} /></Card></Col>
        <Col xs={24} sm={12} lg={6}><Card><Statistic title="Today Revenue" value={Number(stats?.today_revenue ?? 0)} precision={2} prefix="$" /></Card></Col>
        <Col xs={24} sm={12} lg={6}><Card><Statistic title="Success Rate" value={stats?.success_rate ?? 100} precision={1} suffix="%" prefix={<CheckCircleOutlined />} /></Card></Col>
        <Col xs={24} sm={12} lg={6}><Card><Statistic title="Active Accounts" value={stats?.active_accounts ?? 0} prefix={<CreditCardOutlined />} /></Card></Col>
        <Col xs={24} sm={12} lg={6}><Card><Statistic title="Active Merchants" value={stats?.active_merchants ?? 0} prefix={<TeamOutlined />} /></Card></Col>
        <Col xs={24} sm={12} lg={6}><Card><Statistic title="Pending Orders" value={stats?.pending_orders ?? 0} prefix={<WarningOutlined />} /></Card></Col>
        <Col xs={24} sm={12} lg={6}><Card><Statistic title="Failed / Canceled" value={stats?.failed_orders ?? 0} prefix={<WarningOutlined />} /></Card></Col>
      </Row>
      <Card title="Revenue (Last 7 Days)" style={{ marginTop: 16 }}>
        <Table columns={revenueColumns} dataSource={revenue} rowKey="date" pagination={false} size="small" />
      </Card>
    </div>
  );
};

export default Dashboard;
