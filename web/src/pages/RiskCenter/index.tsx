import React, { useEffect, useState } from 'react';
import { Table, Tag, Typography } from 'antd';
import { riskService, RiskRow } from '../../services/dataService';
import dayjs from 'dayjs';

const { Title } = Typography;
const levelColors: Record<string, string> = { low: 'green', medium: 'orange', high: 'red', block: 'magenta' };

const RiskCenter: React.FC = () => {
  const [data, setData] = useState<RiskRow[]>([]);
  const [loading, setLoading] = useState(true);
  useEffect(() => { riskService.list().then(r => { setData(r.events); setLoading(false); }).catch(() => setLoading(false)); }, []);

  const columns = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 280, ellipsis: true },
    { title: 'Merchant', dataIndex: 'merchant_id', key: 'mid', width: 200, ellipsis: true },
    { title: 'Order', dataIndex: 'order_id', key: 'oid', width: 200, ellipsis: true, render: (v: string) => v || '-' },
    { title: 'Rule', dataIndex: 'rule_name', key: 'rule', width: 120 },
    { title: 'Level', dataIndex: 'risk_level', key: 'level', width: 80, render: (v: string) => <Tag color={levelColors[v] || 'default'}>{v}</Tag> },
    { title: 'Score', dataIndex: 'risk_score', key: 'score', width: 70 },
    { title: 'Action', dataIndex: 'action', key: 'action', width: 100, render: (v: string) => <Tag>{v}</Tag> },
    { title: 'Reason', dataIndex: 'reason', key: 'reason', ellipsis: true },
    { title: 'Time', dataIndex: 'created_at', key: 'time', width: 160, render: (v: string) => dayjs(v).format('YYYY-MM-DD HH:mm') },
  ];

  return (
    <div>
      <Title level={3} style={{ marginBottom: 16 }}>Risk Events</Title>
      <Table columns={columns} dataSource={data} rowKey="id" loading={loading} size="small" pagination={{ pageSize: 20 }} scroll={{ x: 1400 }} />
    </div>
  );
};

export default RiskCenter;
