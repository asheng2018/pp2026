import React, { useEffect, useState } from 'react';
import { Table, Tag, Typography, Select, Space, message, Button, Modal, Form, Input, InputNumber, Descriptions, Popconfirm, Divider, Alert } from 'antd';
import { PlusOutlined, KeyOutlined, DeleteOutlined, EyeOutlined, CopyOutlined } from '@ant-design/icons';
import { merchantService, MerchantRow, APIToken } from '../../services/dataService';

const { Title, Text } = Typography;
const statusColors: Record<string, string> = { active: 'green', suspended: 'orange', banned: 'red' };

const Merchants: React.FC = () => {
  const [data, setData] = useState<MerchantRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [createOpen, setCreateOpen] = useState(false);
  const [detailOpen, setDetailOpen] = useState(false);
  const [apiKeyOpen, setApiKeyOpen] = useState(false);
  const [detail, setDetail] = useState<MerchantRow | null>(null);
  const [apiKey, setApiKey] = useState<APIToken | null>(null);
  const [keyMerchantId, setKeyMerchantId] = useState('');
  const [form] = Form.useForm();

  const load = async () => {
    setLoading(true);
    try { setData((await merchantService.list()).merchants); } catch { } finally { setLoading(false); }
  };
  useEffect(() => { load(); }, []);

  const handleCreate = async (values: any) => {
    try { await merchantService.create(values); message.success('Merchant created!'); setCreateOpen(false); form.resetFields(); load(); } catch (e: any) { message.error(e?.response?.data?.error || 'Create failed'); }
  };

  const handleDelete = async (id: string) => {
    try { await merchantService.delete(id); message.success('Deleted'); load(); } catch { message.error('Delete failed'); }
  };

  const generateKey = async (mid: string) => {
    setKeyMerchantId(mid);
    try {
      const res = await merchantService.generateAPIKey(mid);
      setApiKey(res);
      setApiKeyOpen(true);
    } catch (e: any) { message.error(e?.response?.data?.error || 'Failed'); }
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 260, ellipsis: true },
    { title: 'Name', dataIndex: 'name', key: 'name', width: 160 },
    { title: 'Email', dataIndex: 'email', key: 'email', width: 200, ellipsis: true },
    { title: 'Status', dataIndex: 'status', key: 'status', width: 90, render: (s: string) => <Tag color={statusColors[s] || 'default'}>{s}</Tag> },
    { title: 'Routing', dataIndex: 'routing_mode', key: 'routing', width: 120 },
    { title: 'Created', dataIndex: 'created_at', key: 'created', width: 140, render: (v: string) => v?.slice(0, 16) },
    {
      title: 'Actions', key: 'actions', width: 180,
      render: (_: any, r: MerchantRow) => (
        <Space>
          <Button size="small" icon={<EyeOutlined />} onClick={() => { setDetail(r); setDetailOpen(true); }} />
          <Button size="small" type="primary" icon={<KeyOutlined />} onClick={() => generateKey(r.id)}>API Key</Button>
          <Popconfirm title="Delete this merchant?" onConfirm={() => handleDelete(r.id)}>
            <Button size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Space style={{ marginBottom: 16, justifyContent: 'space-between', width: '100%' }}>
        <Title level={3} style={{ margin: 0 }}>Merchants ({data.length})</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateOpen(true)}>Add Merchant</Button>
      </Space>
      <Table columns={columns} dataSource={data} rowKey="id" loading={loading} size="small" pagination={{ pageSize: 20 }} scroll={{ x: 1100 }} />

      {/* Create Merchant */}
      <Modal title="Add Merchant" open={createOpen} onOk={() => form.submit()} onCancel={() => setCreateOpen(false)} okText="Create">
        <Form form={form} layout="vertical" onFinish={handleCreate}>
          <Form.Item name="name" label="Merchant Name" rules={[{ required: true }]}>
            <Input placeholder="e.g. My Store" />
          </Form.Item>
          <Form.Item name="email" label="Email">
            <Input placeholder="merchant@example.com" />
          </Form.Item>
          <Space style={{ display: 'flex', gap: 16 }}>
            <Form.Item name="routing_mode" label="Routing Mode" style={{ width: 200 }}>
              <Select defaultValue="weighted_round_robin" options={[
                { value: 'weighted_round_robin', label: 'Weighted Round Robin' },
                { value: 'sequential', label: 'Sequential' },
                { value: 'random', label: 'Random' },
                { value: 'least_utilized', label: 'Least Utilized' },
              ]} />
            </Form.Item>
            <Form.Item name="fee_rate" label="Fee Rate (%)" style={{ width: 120 }}>
              <InputNumber min={0} max={20} step={0.1} placeholder="4.5" />
            </Form.Item>
          </Space>
        </Form>
      </Modal>

      {/* Merchant Detail */}
      <Modal title="Merchant Detail" open={detailOpen} onCancel={() => setDetailOpen(false)} footer={null} width={500}>
        {detail && <Descriptions column={1} bordered size="small">
          <Descriptions.Item label="ID">{detail.id}</Descriptions.Item>
          <Descriptions.Item label="Name">{detail.name}</Descriptions.Item>
          <Descriptions.Item label="Email">{detail.email || '-'}</Descriptions.Item>
          <Descriptions.Item label="Status"><Tag color={statusColors[detail.status]}>{detail.status}</Tag></Descriptions.Item>
          <Descriptions.Item label="Routing">{detail.routing_mode}</Descriptions.Item>
          <Descriptions.Item label="Created">{detail.created_at?.slice(0, 19)}</Descriptions.Item>
        </Descriptions>}
      </Modal>

      {/* API Key Display */}
      <Modal title="API Key Generated" open={apiKeyOpen} onCancel={() => setApiKeyOpen(false)} footer={null} width={520}>
        {apiKey && (
          <div>
            <Alert type="warning" message="Copy this API key now — it will NOT be shown again!" showIcon style={{ marginBottom: 16 }} />
            <Descriptions column={1} bordered size="small">
              <Descriptions.Item label="API Key">
                <Text copyable style={{ fontFamily: 'monospace', fontSize: 12, wordBreak: 'break-all' }}>{apiKey.api_key}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="Key ID"><Text code>{apiKey.key_id}</Text></Descriptions.Item>
              <Descriptions.Item label="Key Prefix"><Text code>{apiKey.key_prefix}</Text></Descriptions.Item>
              <Descriptions.Item label="Permissions">{apiKey.permissions?.join(', ')}</Descriptions.Item>
            </Descriptions>
            <div style={{ marginTop: 16, padding: 12, background: '#f5f5f5', borderRadius: 6 }}>
              <Text type="secondary">This key is used in the A-site (AB Payment Bridge) plugin to authenticate API requests. Paste it into the WordPress plugin Settings → AB Payment → API Key.</Text>
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
};

export default Merchants;
