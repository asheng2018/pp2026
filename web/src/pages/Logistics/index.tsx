import React, { useEffect, useState } from 'react';
import { Table, Tag, Typography } from 'antd';
import { logisticsService, LogisticsRow } from '../../services/dataService';
import dayjs from 'dayjs';

const { Title } = Typography;
const statusColors: Record<string, string> = { pending: 'orange', in_transit: 'blue', delivered: 'green', failed: 'red' };

const Logistics: React.FC = () => {
  const [data, setData] = useState<LogisticsRow[]>([]);
  const [loading, setLoading] = useState(true);
  useEffect(() => { logisticsService.list().then(r => { setData(r.tracks); setLoading(false); }).catch(() => setLoading(false)); }, []);

  const columns = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 280, ellipsis: true },
    { title: 'Order ID', dataIndex: 'order_id', key: 'oid', width: 280, ellipsis: true, render: (v: string) => v || '-' },
    { title: 'Tracking #', dataIndex: 'tracking_number', key: 'tracking', width: 200, render: (v: string) => v || '-' },
    { title: 'Carrier', dataIndex: 'carrier', key: 'carrier', width: 120, render: (v: string) => v || '-' },
    { title: 'Status', dataIndex: 'status', key: 'status', width: 100, render: (s: string) => <Tag color={statusColors[s] || 'default'}>{s}</Tag> },
    { title: 'Synced to B-Site', dataIndex: 'synced_to_b_site', key: 'synced', width: 120, render: (v: boolean) => v ? <Tag color="green">Yes</Tag> : <Tag>No</Tag> },
    { title: 'Created', dataIndex: 'created_at', key: 'created', width: 160, render: (v: string) => dayjs(v).format('YYYY-MM-DD HH:mm') },
  ];

  return (
    <div>
      <Title level={3} style={{ marginBottom: 16 }}>Logistics Tracking</Title>
      <Table columns={columns} dataSource={data} rowKey="id" loading={loading} size="small" pagination={{ pageSize: 20 }} scroll={{ x: 1300 }} />
    </div>
  );
};

export default Logistics;
