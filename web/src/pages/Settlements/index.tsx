import React, { useEffect, useState } from 'react';
import { Table, Tag, Typography } from 'antd';
import { settlementService, SettlementRow } from '../../services/dataService';
import dayjs from 'dayjs';

const { Title } = Typography;
const statusColors: Record<string, string> = { pending: 'orange', processing: 'processing', completed: 'success', failed: 'error' };

const Settlements: React.FC = () => {
  const [data, setData] = useState<SettlementRow[]>([]);
  const [loading, setLoading] = useState(true);
  useEffect(() => { settlementService.list().then(r => { setData(r.settlements); setLoading(false); }).catch(() => setLoading(false)); }, []);

  const columns = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 280, ellipsis: true },
    { title: 'Merchant', dataIndex: 'merchant_id', key: 'mid', width: 200, ellipsis: true },
    { title: 'Cycle', key: 'cycle', width: 200, render: (_: any, r: SettlementRow) => `${r.cycle_start?.slice(0,10)} ~ ${r.cycle_end?.slice(0,10)}` },
    { title: 'Orders', dataIndex: 'total_orders', key: 'orders', width: 80 },
    { title: 'Amount', dataIndex: 'total_amount', key: 'amount', width: 100, render: (v: string) => `$${Number(v).toFixed(2)}` },
    { title: 'Fee', dataIndex: 'total_fee', key: 'fee', width: 100, render: (v: string) => `$${Number(v).toFixed(2)}` },
    { title: 'Net', dataIndex: 'net_amount', key: 'net', width: 100, render: (v: string) => `$${Number(v).toFixed(2)}` },
    { title: 'Method', dataIndex: 'payout_method', key: 'method', width: 80, render: (v: string) => v || '-' },
    { title: 'Status', dataIndex: 'status', key: 'status', width: 100, render: (s: string) => <Tag color={statusColors[s] || 'default'}>{s}</Tag> },
    { title: 'Created', dataIndex: 'created_at', key: 'created', width: 160, render: (v: string) => dayjs(v).format('YYYY-MM-DD HH:mm') },
  ];

  return (
    <div>
      <Title level={3} style={{ marginBottom: 16 }}>Settlements</Title>
      <Table columns={columns} dataSource={data} rowKey="id" loading={loading} size="small" pagination={{ pageSize: 20 }} scroll={{ x: 1300 }} />
    </div>
  );
};

export default Settlements;
