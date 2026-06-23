import React, { useEffect, useState } from 'react';
import { Table, Tag, Typography, Select, Space, message, Button, Modal, Form, Input, InputNumber, Descriptions, Popconfirm } from 'antd';
import { PlusOutlined, DeleteOutlined, EyeOutlined, ImportOutlined } from '@ant-design/icons';
import { bSiteService, BSiteRow } from '../../services/dataService';

const { Title } = Typography;
const statusColors: Record<string, string> = { active: 'green', inactive: 'default', suspended: 'orange' };

const BSites: React.FC = () => {
  const [data, setData] = useState<BSiteRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [createOpen, setCreateOpen] = useState(false);
  const [detailOpen, setDetailOpen] = useState(false);
  const [detail, setDetail] = useState<any>(null);
  const [form] = Form.useForm();

  const load = async () => {
    setLoading(true);
    try { setData((await bSiteService.list()).b_sites); } catch { /* */ } finally { setLoading(false); }
  };
  useEffect(() => { load(); }, []);

  const updateStatus = async (id: string, status: string) => {
    try { await bSiteService.updateStatus(id, status); message.success('Status updated'); load(); } catch { message.error('Update failed'); }
  };

  const handleCreate = async (values: any) => {
    try {
      await bSiteService.create(values);
      message.success('B-Site created!');
      setCreateOpen(false); form.resetFields(); load();
    } catch (e: any) { message.error(e?.response?.data?.error || 'Create failed'); }
  };

  const handleDelete = async (id: string) => {
    try { await bSiteService.delete(id); message.success('Deleted'); load(); } catch { message.error('Delete failed'); }
  };

  const viewDetail = async (r: BSiteRow) => {
    try {
      const d = await bSiteService.getById(r.id);
      setDetail(d);
      setDetailOpen(true);
    } catch { setDetail(r); setDetailOpen(true); }
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 250, ellipsis: true },
    { title: 'Domain', dataIndex: 'domain', key: 'domain', width: 200 },
    { title: 'Name', dataIndex: 'name', key: 'name', width: 160 },
    { title: 'IP', dataIndex: 'hosting_ip', key: 'ip', width: 140, render: (v: string) => v || '-' },
    { title: 'Status', dataIndex: 'status', key: 'status', width: 100, render: (s: string) => <Tag color={statusColors[s] || 'default'}>{s}</Tag> },
    { title: 'Created', dataIndex: 'created_at', key: 'created', width: 140, render: (v: string) => v?.slice(0, 16) },
    {
      title: 'Actions', key: 'actions', width: 180,
      render: (_: any, r: BSiteRow) => (
        <Space>
          <Select size="small" defaultValue={r.status} style={{ width: 100 }} onChange={(v: string) => updateStatus(r.id, v)}
            options={['active', 'inactive', 'suspended'].map(s => ({ value: s, label: s }))} />
          <Button size="small" icon={<EyeOutlined />} onClick={() => viewDetail(r)} />
          <Popconfirm title="Delete this B-site?" onConfirm={() => handleDelete(r.id)}>
            <Button size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Space style={{ marginBottom: 16, justifyContent: 'space-between', width: '100%' }}>
        <Title level={3} style={{ margin: 0 }}>B-Sites ({data.length})</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateOpen(true)}>Add B-Site</Button>
      </Space>
      <Table columns={columns} dataSource={data} rowKey="id" loading={loading} size="small" pagination={{ pageSize: 20 }} scroll={{ x: 1200 }} />

      {/* Create B-Site */}
      <Modal title="Add B-Site" open={createOpen} onOk={() => form.submit()} onCancel={() => setCreateOpen(false)} width={640} okText="Create">
        <Form form={form} layout="vertical" onFinish={handleCreate}>
          <Space style={{ display: 'flex', gap: 16 }}>
            <Form.Item name="domain" label="Domain" rules={[{ required: true, message: 'e.g. goodstore.shop' }]} style={{ flex: 1 }}>
              <Input placeholder="goodstore.shop" />
            </Form.Item>
            <Form.Item name="name" label="Store Name" style={{ flex: 1 }}>
              <Input placeholder="Good Store" />
            </Form.Item>
          </Space>
          <Space style={{ display: 'flex', gap: 16 }}>
            <Form.Item name="hosting_ip" label="Hosting IP" style={{ flex: 1 }}>
              <Input placeholder="1.2.3.4" />
            </Form.Item>
            <Form.Item name="hosting_provider" label="Hosting Provider" style={{ flex: 1 }}>
              <Input placeholder="Cloudflare / AWS" />
            </Form.Item>
          </Space>
          <Form.Item name="woocommerce_url" label="WooCommerce URL">
            <Input placeholder="https://goodstore.shop" />
          </Form.Item>
          <Space style={{ display: 'flex', gap: 16 }}>
            <Form.Item name="woocommerce_key" label="WC Consumer Key" style={{ flex: 1 }}>
              <Input placeholder="ck_..." />
            </Form.Item>
            <Form.Item name="woocommerce_secret" label="WC Consumer Secret" style={{ flex: 1 }}>
              <Input.Password placeholder="cs_..." />
            </Form.Item>
          </Space>
          <Space style={{ display: 'flex', gap: 16 }}>
            <Form.Item name="ssl_provider" label="SSL Provider" style={{ flex: 1 }}>
              <Input placeholder="Let's Encrypt" />
            </Form.Item>
            <Form.Item name="ssl_expires_at" label="SSL Expires" style={{ flex: 1 }}>
              <Input placeholder="2026-12-31" />
            </Form.Item>
          </Space>
        </Form>
      </Modal>

      {/* Detail Modal */}
      <Modal title="B-Site Detail" open={detailOpen} onCancel={() => setDetailOpen(false)} footer={null} width={600}>
        {detail && <Descriptions column={2} bordered size="small">
          <Descriptions.Item label="ID" span={2}>{detail.id}</Descriptions.Item>
          <Descriptions.Item label="Domain">{detail.domain}</Descriptions.Item>
          <Descriptions.Item label="Name">{detail.name || '-'}</Descriptions.Item>
          <Descriptions.Item label="Hosting IP">{detail.hosting_ip || '-'}</Descriptions.Item>
          <Descriptions.Item label="Provider">{detail.hosting_provider || '-'}</Descriptions.Item>
          <Descriptions.Item label="WooCommerce URL">{detail.woocommerce_url || '-'}</Descriptions.Item>
          <Descriptions.Item label="SSL Provider">{detail.ssl_provider || '-'}</Descriptions.Item>
          <Descriptions.Item label="SSL Expires">{detail.ssl_expires_at || '-'}</Descriptions.Item>
          <Descriptions.Item label="Status"><Tag color={statusColors[detail.status]}>{detail.status}</Tag></Descriptions.Item>
          <Descriptions.Item label="Created">{detail.created_at?.slice(0, 19)}</Descriptions.Item>
        </Descriptions>}
      </Modal>
    </div>
  );
};

export default BSites;
