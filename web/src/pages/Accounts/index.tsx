import React, { useEffect, useState } from 'react';
import { Table, Tag, Typography, Select, Space, message, Button, Modal, Form, Input, InputNumber, Descriptions, Popconfirm } from 'antd';
import { PlusOutlined, DeleteOutlined, EyeOutlined } from '@ant-design/icons';
import { accountService, AccountRow, AccountCreateData } from '../../services/dataService';

const { Title } = Typography;

const statusColors: Record<string, string> = { online: 'green', offline: 'red', cooling: 'orange', draining: 'blue', warming: 'cyan' };

const Accounts: React.FC = () => {
  const [data, setData] = useState<AccountRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [createOpen, setCreateOpen] = useState(false);
  const [detailOpen, setDetailOpen] = useState(false);
  const [detail, setDetail] = useState<AccountRow | null>(null);
  const [form] = Form.useForm();

  const load = async () => {
    setLoading(true);
    try { setData((await accountService.list()).accounts); } catch { } finally { setLoading(false); }
  };
  useEffect(() => { load(); }, []);

  const updateStatus = async (id: string, status: string) => {
    try { await accountService.updateStatus(id, status); message.success('Status updated'); load(); } catch { message.error('Update failed'); }
  };

  const handleCreate = async (values: AccountCreateData) => {
    try {
      await accountService.create(values);
      message.success('Account created');
      setCreateOpen(false);
      form.resetFields();
      load();
    } catch (e: any) { message.error(e?.response?.data?.error || 'Create failed'); }
  };

  const handleDelete = async (id: string) => {
    try { await accountService.delete(id); message.success('Deleted'); load(); } catch { message.error('Delete failed'); }
  };

  const viewDetail = async (r: AccountRow) => {
    setDetail(r);
    setDetailOpen(true);
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 250, ellipsis: true },
    { title: 'Gateway', dataIndex: 'gateway', key: 'gateway', width: 80 },
    { title: 'Alias', dataIndex: 'alias', key: 'alias', width: 120 },
    { title: 'Status', dataIndex: 'status', key: 'status', width: 100, render: (s: string) => <Tag color={statusColors[s] || 'default'}>{s}</Tag> },
    { title: 'B-Site', dataIndex: 'b_site_id', key: 'bs', width: 180, ellipsis: true, render: (v: string) => v || '-' },
    { title: 'Merchant', dataIndex: 'merchant_id', key: 'mid', width: 180, ellipsis: true, render: (v: string) => v || '-' },
    { title: 'Weight', dataIndex: 'weight', key: 'weight', width: 70 },
    { title: 'Priority', dataIndex: 'priority', key: 'priority', width: 70 },
    { title: 'Tags', dataIndex: 'tags', key: 'tags', render: (t: string[]) => t?.map((tg: string) => <Tag key={tg}>{tg}</Tag>) },
    { title: 'Created', dataIndex: 'created_at', key: 'created', width: 140, render: (v: string) => v?.slice(0, 16) },
    {
      title: 'Actions', key: 'actions', width: 180,
      render: (_: any, r: AccountRow) => (
        <Space>
          <Select size="small" defaultValue={r.status} style={{ width: 100 }} onChange={(v: string) => updateStatus(r.id, v)}
            options={['online', 'offline', 'cooling', 'draining', 'warming'].map(s => ({ value: s, label: s }))} />
          <Button size="small" icon={<EyeOutlined />} onClick={() => viewDetail(r)} />
          <Popconfirm title="Delete this account?" onConfirm={() => handleDelete(r.id)}>
            <Button size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Space style={{ marginBottom: 16, justifyContent: 'space-between', width: '100%' }}>
        <Title level={3} style={{ margin: 0 }}>Payment Accounts ({data.length})</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateOpen(true)}>Add Account</Button>
      </Space>
      <Table columns={columns} dataSource={data} rowKey="id" loading={loading} size="small" pagination={{ pageSize: 20 }} scroll={{ x: 1400 }} />

      {/* Create Account Modal */}
      <Modal title="Add Payment Account" open={createOpen} onOk={() => form.submit()} onCancel={() => setCreateOpen(false)} width={640} okText="Create">
        <Form form={form} layout="vertical" onFinish={handleCreate}>
          <Space style={{ display: 'flex', gap: 16 }}>
            <Form.Item name="gateway" label="Gateway" rules={[{ required: true }]} style={{ width: 150 }}>
              <Select options={[{ value: 'paypal', label: 'PayPal' }, { value: 'stripe', label: 'Stripe' }]} />
            </Form.Item>
            <Form.Item name="alias" label="Alias" rules={[{ required: true }]} style={{ flex: 1 }}>
              <Input placeholder="e.g. PP-US-01" />
            </Form.Item>
            <Form.Item name="weight" label="Weight" style={{ width: 80 }}>
              <InputNumber min={1} max={1000} />
            </Form.Item>
            <Form.Item name="priority" label="Priority" style={{ width: 80 }}>
              <InputNumber min={0} max={100} />
            </Form.Item>
          </Space>
          <Space style={{ display: 'flex', gap: 16 }}>
            <Form.Item name="merchant_id" label="Merchant ID" style={{ flex: 1 }}>
              <Input placeholder="UUID of merchant" />
            </Form.Item>
            <Form.Item name="b_site_id" label="B-Site ID" style={{ flex: 1 }}>
              <Input placeholder="UUID of B-site" />
            </Form.Item>
          </Space>
          <Form.Item noStyle shouldUpdate={(prev, cur) => prev.gateway !== cur.gateway}>
            {({ getFieldValue }) => {
              const gw = getFieldValue('gateway') || 'paypal';
              return gw === 'paypal' ? (
                <Space style={{ display: 'flex', gap: 16 }}>
                  <Form.Item name="paypal_client_id" label="PayPal Client ID" style={{ flex: 1 }} rules={[{ required: true }]}>
                    <Input placeholder="AQ7R..." />
                  </Form.Item>
                  <Form.Item name="paypal_secret" label="PayPal Secret" style={{ flex: 1 }} rules={[{ required: true }]}>
                    <Input.Password placeholder="EC..." />
                  </Form.Item>
                </Space>
              ) : (
                <Space style={{ display: 'flex', gap: 16 }}>
                  <Form.Item name="stripe_publishable_key" label="Stripe Publishable Key" style={{ flex: 1 }}>
                    <Input placeholder="pk_live_..." />
                  </Form.Item>
                  <Form.Item name="stripe_secret_key" label="Stripe Secret Key" style={{ flex: 1 }} rules={[{ required: true }]}>
                    <Input.Password placeholder="sk_live_..." />
                  </Form.Item>
                </Space>
              );
            }}
          </Form.Item>
          <Space style={{ display: 'flex', gap: 16 }}>
            <Form.Item name="single_min" label="Single Min ($)" style={{ width: 100 }}><Input placeholder="1" /></Form.Item>
            <Form.Item name="single_max" label="Single Max ($)" style={{ width: 100 }}><Input placeholder="5000" /></Form.Item>
            <Form.Item name="daily_max" label="Daily Max ($)" style={{ width: 100 }}><Input placeholder="50000" /></Form.Item>
            <Form.Item name="monthly_max" label="Monthly Max ($)" style={{ width: 100 }}><Input placeholder="500000" /></Form.Item>
          </Space>
        </Form>
      </Modal>

      {/* Detail Modal */}
      <Modal title="Account Detail" open={detailOpen} onCancel={() => setDetailOpen(false)} footer={null} width={600}>
        {detail && <Descriptions column={2} bordered size="small">
          <Descriptions.Item label="ID" span={2}>{detail.id}</Descriptions.Item>
          <Descriptions.Item label="Gateway">{detail.gateway}</Descriptions.Item>
          <Descriptions.Item label="Alias">{detail.alias}</Descriptions.Item>
          <Descriptions.Item label="Status"><Tag color={statusColors[detail.status]}>{detail.status}</Tag></Descriptions.Item>
          <Descriptions.Item label="Weight">{detail.weight}</Descriptions.Item>
          <Descriptions.Item label="B-Site">{detail.b_site_id || '-'}</Descriptions.Item>
          <Descriptions.Item label="Merchant">{detail.merchant_id || '-'}</Descriptions.Item>
          <Descriptions.Item label="Tags">{detail.tags?.join(', ') || '-'}</Descriptions.Item>
          <Descriptions.Item label="Created">{detail.created_at?.slice(0, 19)}</Descriptions.Item>
        </Descriptions>}
      </Modal>
    </div>
  );
};

export default Accounts;
