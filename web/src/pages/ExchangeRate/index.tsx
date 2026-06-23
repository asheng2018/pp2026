import React, { useEffect, useState } from 'react';
import { Table, Typography } from 'antd';
import { exchangeService, ExchangeRow } from '../../services/dataService';
import dayjs from 'dayjs';

const { Title } = Typography;

const ExchangeRate: React.FC = () => {
  const [data, setData] = useState<ExchangeRow[]>([]);
  const [loading, setLoading] = useState(true);
  useEffect(() => { exchangeService.list().then(r => { setData(r.rates); setLoading(false); }).catch(() => setLoading(false)); }, []);

  const columns = [
    { title: 'Base', dataIndex: 'base_currency', key: 'base', width: 100 },
    { title: 'Target', dataIndex: 'target_currency', key: 'target', width: 100 },
    { title: 'Rate', dataIndex: 'rate', key: 'rate', width: 180, render: (v: string) => Number(v).toFixed(8) },
    { title: 'Source', dataIndex: 'source', key: 'source', width: 100, render: (v: string) => v || '-' },
    { title: 'Fetched At', dataIndex: 'fetched_at', key: 'fetched', width: 180, render: (v: string) => dayjs(v).format('YYYY-MM-DD HH:mm:ss') },
  ];

  return (
    <div>
      <Title level={3} style={{ marginBottom: 16 }}>Exchange Rates</Title>
      <Table columns={columns} dataSource={data} rowKey="id" loading={loading} size="small" pagination={{ pageSize: 20 }} />
    </div>
  );
};

export default ExchangeRate;
