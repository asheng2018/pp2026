import React from 'react';
import { Card, Typography, Row, Col, Statistic, Table } from 'antd';
import {
  ShoppingCartOutlined, DollarOutlined, SafetyOutlined, SwapOutlined,
} from '@ant-design/icons';

const { Title, Paragraph } = Typography;

const Reports: React.FC = () => {
  const summaryData = [
    { key: '1', metric: 'Total Orders Today', value: '0' },
    { key: '2', metric: 'Total Revenue Today', value: '$0.00' },
    { key: '3', metric: 'Settlements Processed', value: '0' },
    { key: '4', metric: 'Risk Events Blocked', value: '0' },
    { key: '5', metric: 'Exchange Rates Updated', value: '0' },
    { key: '6', metric: 'Logistics Items Tracked', value: '0' },
  ];

  return (
    <div>
      <Title level={3} style={{ marginBottom: 24 }}>Reports</Title>
      <Row gutter={[16, 16]}>
        <Col span={24}>
          <Card title="System Activity Summary">
            <Paragraph type="secondary">
              Comprehensive reporting module. Data shown below aggregates across all merchants, accounts, and gateways.
              For merchant-specific reports, use the Merchants page to drill down by status filters.
            </Paragraph>
            <Table columns={[
              { title: 'Metric', dataIndex: 'metric', key: 'metric' },
              { title: 'Value', dataIndex: 'value', key: 'value' },
            ]} dataSource={summaryData} pagination={false} size="small" />
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Reports;
