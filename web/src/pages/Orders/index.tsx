import React, { useEffect, useState } from 'react';
import { Table, Tag, Card, Typography, Space, Select, Drawer, Descriptions, Spin } from 'antd';
import { orderService, OrderRow, OrderDetail } from '../../services/dataService';
import dayjs from 'dayjs';

const { Title } = Typography;

const statusColors: Record<string, string> = {
  pending: 'default', processing: 'processing', paid: 'success', failed: 'error',
  canceled: 'warning', expired: 'warning', refunding: 'processing', refunded: 'purple',
  partially_refunded: 'purple', disputed: 'orange', dispute_won: 'green', dispute_lost: 'red', completed: 'success',
};

const Orders: React.FC = () => {
  const [orders, setOrders] = useState<OrderRow[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [statusFilter, setStatusFilter] = useState<string | undefined>();
  const [selected, setSelected] = useState<OrderDetail | null>(null);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [detailLoading, setDetailLoading] = useState(false);

  const load = async (s?: string) => {
    setLoading(true);
    try {
      const res = await orderService.list(s);
      setOrders(res.orders); setTotal(res.total);
    } finally { setLoading(false); }
  };
  useEffect(() => { load(statusFilter); }, [statusFilter]);

  const viewDetail = async (id: string) => {
    setDetailLoading(true); setDrawerOpen(true);
    try { setSelected(await orderService.getById(id)); } finally { setDetailLoading(false); }
  };

  const columns = [
    { title: 'Order #', dataIndex: 'order_no', key: 'order_no', width: 180 },
    { title: 'Gateway', dataIndex: 'gateway', key: 'gateway', width: 80 },
    { title: 'Amount', dataIndex: 'amount', key: 'amount', width: 100, render: (v: string) => `$${Number(v).toFixed(2)}` },
    { title: 'Currency', dataIndex: 'currency', key: 'currency', width: 80 },
    { title: 'Status', dataIndex: 'status', key: 'status', width: 100, render: (s: string) => <Tag color={statusColors[s] || 'default'}>{s}</Tag> },
    { title: 'Risk', dataIndex: 'risk_level', key: 'risk', width: 80, render: (v: string) => <Tag color={v === 'low' ? 'green' : v === 'medium' ? 'orange' : 'red'}>{v}</Tag> },
    { title: 'Customer', dataIndex: 'customer_email', key: 'email', ellipsis: true },
    { title: 'Country', dataIndex: 'customer_country', key: 'country', width: 80 },
    { title: 'Created', dataIndex: 'created_at', key: 'created', width: 160, render: (v: string) => dayjs(v).format('YYYY-MM-DD HH:mm') },
    { title: '', key: 'action', width: 60, render: (_: any, r: OrderRow) => <a onClick={() => viewDetail(r.id)}>Detail</a> },
  ];

  return (
    <div>
      <Space style={{ marginBottom: 16, justifyContent: 'space-between', width: '100%' }}>
        <Title level={3} style={{ margin: 0 }}>Orders ({total})</Title>
        <Select allowClear placeholder="Filter by status" style={{ width: 180 }} value={statusFilter} onChange={setStatusFilter}
          options={['pending','processing','paid','failed','canceled','expired','refunding','refunded','partially_refunded','disputed','completed'].map(s => ({ value: s, label: s }))} />
      </Space>
      <Table columns={columns} dataSource={orders} rowKey="id" loading={loading} size="small" pagination={{ pageSize: 20, total }} />

      <Drawer title="Order Detail" open={drawerOpen} onClose={() => setDrawerOpen(false)} width={640}>
        {detailLoading ? <Spin /> : selected ? (
          <Descriptions column={2} bordered size="small">
            <Descriptions.Item label="ID">{selected.id}</Descriptions.Item>
            <Descriptions.Item label="Order No">{selected.order_no}</Descriptions.Item>
            <Descriptions.Item label="Gateway">{selected.gateway}</Descriptions.Item>
            <Descriptions.Item label="Status"><Tag color={statusColors[selected.status]}>{selected.status}</Tag></Descriptions.Item>
            <Descriptions.Item label="Amount">${Number(selected.amount).toFixed(2)}</Descriptions.Item>
            <Descriptions.Item label="Currency">{selected.currency}</Descriptions.Item>
            <Descriptions.Item label="Merchant ID">{selected.merchant_id}</Descriptions.Item>
            <Descriptions.Item label="Account ID">{selected.account_id || '-'}</Descriptions.Item>
            <Descriptions.Item label="Customer Email">{selected.customer_email || '-'}</Descriptions.Item>
            <Descriptions.Item label="Customer IP">{selected.customer_ip || '-'}</Descriptions.Item>
            <Descriptions.Item label="Country">{selected.customer_country || '-'}</Descriptions.Item>
            <Descriptions.Item label="Gateway Order ID">{selected.gateway_order_id || '-'}</Descriptions.Item>
            <Descriptions.Item label="Risk"><Tag color={selected.risk_level === 'low' ? 'green' : 'orange'}>{selected.risk_level}</Tag></Descriptions.Item>
            <Descriptions.Item label="Risk Score">{selected.risk_score}</Descriptions.Item>
            <Descriptions.Item label="Paid At">{selected.paid_at ? dayjs(selected.paid_at).format('YYYY-MM-DD HH:mm') : '-'}</Descriptions.Item>
            <Descriptions.Item label="Expired At">{selected.expired_at ? dayjs(selected.expired_at).format('YYYY-MM-DD HH:mm') : '-'}</Descriptions.Item>
          </Descriptions>
        ) : null}
      </Drawer>
    </div>
  );
};

export default Orders;
