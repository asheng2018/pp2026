import React from 'react';
import { Card, Typography, Descriptions, Divider } from 'antd';
import { useAuth } from '../../contexts/AuthContext';

const { Title, Paragraph } = Typography;

const Settings: React.FC = () => {
  const { user } = useAuth();

  return (
    <div>
      <Title level={3} style={{ marginBottom: 24 }}>System Settings</Title>

      <Card title="Current Admin User" style={{ marginBottom: 16 }}>
        <Descriptions column={2} bordered size="small">
          <Descriptions.Item label="ID">{user?.id || '-'}</Descriptions.Item>
          <Descriptions.Item label="Username">{user?.username || '-'}</Descriptions.Item>
          <Descriptions.Item label="Role">{user?.role || '-'}</Descriptions.Item>
          <Descriptions.Item label="TOTP Enabled">{user?.totp_enabled ? 'Yes' : 'No'}</Descriptions.Item>
          <Descriptions.Item label="Last Login">{user?.last_login_at || '-'}</Descriptions.Item>
          <Descriptions.Item label="Created">{user?.created_at?.slice(0, 16) || '-'}</Descriptions.Item>
        </Descriptions>
      </Card>

      <Card title="Configuration">
        <Paragraph type="secondary">
          System settings are managed via environment variables in docker-compose.yml.
        </Paragraph>
        <Descriptions column={1} bordered size="small">
          <Descriptions.Item label="API Base URL">/api/v1</Descriptions.Item>
          <Descriptions.Item label="JWT Secret">JWT_SECRET (env var)</Descriptions.Item>
          <Descriptions.Item label="JWT Expiry">ADMIN_JWT_EXPIRY (default: 24h)</Descriptions.Item>
          <Descriptions.Item label="Database">PostgreSQL (postgres:5432)</Descriptions.Item>
          <Descriptions.Item label="Redis">Redis (redis:6379)</Descriptions.Item>
          <Descriptions.Item label="NATS">NATS (nats:4222)</Descriptions.Item>
          <Descriptions.Item label="Grafana">Port 3000</Descriptions.Item>
          <Descriptions.Item label="Gateway">Port 8081</Descriptions.Item>
          <Descriptions.Item label="Orchestrator">Port 8080</Descriptions.Item>
          <Descriptions.Item label="Admin UI">Port 8088</Descriptions.Item>
        </Descriptions>
      </Card>
    </div>
  );
};

export default Settings;
