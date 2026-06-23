import React, { useEffect, useState } from 'react';
import { Table, Tag, Typography, Select, Space, message, Button, Modal, Form, Input, InputNumber, Popconfirm, Descriptions, Alert } from 'antd';
import { PlusOutlined, DeleteOutlined, EyeOutlined, ImportOutlined } from '@ant-design/icons';
import { proxyService, ProxyRow } from '../../services/dataService';

const { Title, Text } = Typography;
const statusColors: Record<string, string> = { online: 'green', offline: 'red', testing: 'blue', banned: 'magenta' };

const ProxyPool: React.FC = () => {
  const [data, setData] = useState<ProxyRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [createOpen, setCreateOpen] = useState(false);
  const [batchOpen, setBatchOpen] = useState(false);
  const [detailOpen, setDetailOpen] = useState(false);
  const [detail, setDetail] = useState<ProxyRow | null>(null);
  const [form] = Form.useForm();
  const [batchText, setBatchText] = useState('');

  const load = async () => {
    setLoading(true);
    try { setData((await proxyService.list()).proxies); } catch { /* */ } finally { setLoading(false); }
  };
  useEffect(() => { load(); }, []);

  const updateStatus = async (id: string, status: string) => {
    try { await proxyService.updateStatus(id, status); message.success('Status updated'); load(); } catch { message.error('Update failed'); }
  };

  const handleCreate = async (values: any) => {
    try { await proxyService.create(values); message.success('Proxy added!'); setCreateOpen(false); form.resetFields(); load(); } catch (e: any) { message.error(e?.response?.data?.error || 'Create failed'); }
  };

  const handleBatchImport = async () => {
    const lines = batchText.trim().split('\n').filter(l => l.trim());
    const proxies = lines.map(line => {
      const parts = line.split(':');
      return { host: parts[0]?.trim(), port: parseInt(parts[1]?.trim()) || 1080, username: parts[2]?.trim() || '', password: parts[3]?.trim() || '', proxy_type: 'residential', protocol: 'socks5' };
    }).filter(p => p.host && p.port);
    if (proxies.length === 0) { message.error('No valid proxies found in text'); return; }
    try { const res = await proxyService.batchImport(proxies); message.success(`Imported ${res.imported}/${res.total} proxies`); setBatchOpen(false); setBatchText(''); load(); } catch (e: any) { message.error('Batch import failed'); }
  };

  const handleDelete = async (id: string) => {
    try { await proxyService.delete(id); message.success('Deleted'); load(); } catch { message.error('Delete failed'); }
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 250, ellipsis: true },
    { title: 'Type', dataIndex: 'proxy_type', key: 'type', width: 90 },
    { title: 'Host:Port', key: 'host', width: 180, render: (_: any, r: ProxyRow) => `${r.host}:${r.port}` },
    { title: 'Country', dataIndex: 'country', key: 'country', width: 80, render: (v: string) => v || '-' },
    { title: 'ISP', dataIndex: 'isp', key: 'isp', width: 120, render: (v: string) => v || '-' },
    { title: 'Status', dataIndex: 'status', key: 'status', width: 80, render: (s: string) => <Tag color={statusColors[s] || 'default'}>{s}</Tag> },
    { title: 'Latency', dataIndex: 'latency', key: 'latency', width: 80, render: (v: number) => v > 0 ? `${v}ms` : '-' },
    { title: 'Success%', dataIndex: 'success_rate', key: 'sr', width: 80, render: (v: any) => `${(Number(v) || 0).toFixed(0)}%` },
    { title: 'Created', dataIndex: 'created_at', key: 'created', width: 140, render: (v: string) => v?.slice(0, 16) },
    {
      title: 'Actions', key: 'actions', width: 180,
      render: (_: any, r: ProxyRow) => (
        <Space>
          <Select size="small" defaultValue={r.status} style={{ width: 90 }} onChange={(v: string) => updateStatus(r.id, v)}
            options={['online', 'offline', 'testing', 'banned'].map(s => ({ value: s, label: s }))} />
          <Button size="small" icon={<EyeOutlined />} onClick={() => { setDetail(r); setDetailOpen(true); }} />
          <Popconfirm title="Delete this proxy?" onConfirm={() => handleDelete(r.id)}>
            <Button size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Space style={{ marginBottom: 16, justifyContent: 'space-between', width: '100%' }}>
        <Title level={3} style={{ margin: 0 }}>Proxy Pool ({data.length})</Title>
        <Space>
          <Button icon={<ImportOutlined />} onClick={() => setBatchOpen(true)}>Batch Import</Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateOpen(true)}>Add Proxy</Button>
        </Space>
      </Space>
      <Table columns={columns} dataSource={data} rowKey="id" loading={loading} size="small" pagination={{ pageSize: 20 }} scroll={{ x: 1400 }} />

      {/* Create Proxy */}
      <Modal title="Add Proxy" open={createOpen} onOk={() => form.submit()} onCancel={() => setCreateOpen(false)} width={560} okText="Create">
        <Form form={form} layout="vertical" onFinish={handleCreate} initialValues={{ proxy_type: 'residential', protocol: 'socks5' }}>
          <Space style={{ display: 'flex', gap: 16 }}>
            <Form.Item name="proxy_type" label="Type" style={{ width: 140 }}>
              <Select options={[{ value: 'residential', label: 'Residential' }, { value: 'datacenter', label: 'Datacenter' }, { value: 'mobile', label: 'Mobile' }, { value: 'fixed_isp', label: 'Fixed ISP' }]} />
            </Form.Item>
            <Form.Item name="protocol" label="Protocol" style={{ width: 120 }}>
              <Select options={[{ value: 'socks5', label: 'SOCKS5' }, { value: 'http', label: 'HTTP' }]} />
            </Form.Item>
            <Form.Item name="country" label="Country" style={{ width: 100 }}>
              <Input placeholder="US" maxLength={4} />
            </Form.Item>
          </Space>
          <Space style={{ display: 'flex', gap: 16 }}>
            <Form.Item name="host" label="Host" rules={[{ required: true }]} style={{ flex: 1 }}>
              <Input placeholder="proxy-1.residential.example.com" />
            </Form.Item>
            <Form.Item name="port" label="Port" rules={[{ required: true }]} style={{ width: 100 }}>
              <InputNumber min={1} max={65535} />
            </Form.Item>
          </Space>
          <Space style={{ display: 'flex', gap: 16 }}>
            <Form.Item name="username" label="Username" style={{ flex: 1 }}>
              <Input placeholder="proxyuser" />
            </Form.Item>
            <Form.Item name="password" label="Password" style={{ flex: 1 }}>
              <Input.Password placeholder="proxypass" />
            </Form.Item>
          </Space>
          <Space style={{ display: 'flex', gap: 16 }}>
            <Form.Item name="isp" label="ISP" style={{ flex: 1 }}><Input placeholder="Comcast / AT&T" /></Form.Item>
            <Form.Item name="bound_account_id" label="Bind to Account" style={{ flex: 1 }}><Input placeholder="UUID of payment account" /></Form.Item>
          </Space>
        </Form>
      </Modal>

      {/* Batch Import */}
      <Modal title="Batch Import Proxies" open={batchOpen} onOk={handleBatchImport} onCancel={() => setBatchOpen(false)} width={560} okText="Import">
        <Alert message="Paste proxy list (one per line): host:port:username:password" type="info" showIcon style={{ marginBottom: 12 }} />
        <Input.TextArea rows={12} value={batchText} onChange={e => setBatchText(e.target.value)}
          placeholder={`185.199.100.1:1234:user:pass\n185.199.100.2:5678:user2:pass2\nproxy.example.com:9050`} />
      </Modal>

      {/* Detail */}
      <Modal title="Proxy Detail" open={detailOpen} onCancel={() => setDetailOpen(false)} footer={null} width={500}>
        {detail && <Descriptions column={2} bordered size="small">
          <Descriptions.Item label="ID" span={2}>{detail.id}</Descriptions.Item>
          <Descriptions.Item label="Type">{detail.proxy_type}</Descriptions.Item>
          <Descriptions.Item label="Protocol">{detail.protocol || 'socks5'}</Descriptions.Item>
          <Descriptions.Item label="Host">{detail.host}</Descriptions.Item>
          <Descriptions.Item label="Port">{detail.port}</Descriptions.Item>
          <Descriptions.Item label="Country">{detail.country || '-'}</Descriptions.Item>
          <Descriptions.Item label="ISP">{detail.isp || '-'}</Descriptions.Item>
          <Descriptions.Item label="Status"><Tag color={statusColors[detail.status]}>{detail.status}</Tag></Descriptions.Item>
          <Descriptions.Item label="Latency">{detail.latency > 0 ? `${detail.latency}ms` : '-'}</Descriptions.Item>
          <Descriptions.Item label="Created">{detail.created_at?.slice(0, 19)}</Descriptions.Item>
        </Descriptions>}
      </Modal>
    </div>
  );
};

export default ProxyPool;
