-- AB Payment System - Database Schema
-- PostgreSQL

-- 1. Merchants
CREATE TABLE IF NOT EXISTS merchants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    risk_profile JSONB DEFAULT '{}',
    fee_config JSONB NOT NULL DEFAULT '{}',
    settlement_config JSONB NOT NULL DEFAULT '{}',
    routing_mode VARCHAR(64) DEFAULT 'weighted_round_robin',
    account_group VARCHAR(64),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Merchant API Keys
CREATE TABLE IF NOT EXISTS merchant_api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    key_hash VARCHAR(128) NOT NULL UNIQUE,
    key_prefix VARCHAR(16) NOT NULL,
    permissions VARCHAR(32)[] DEFAULT '{read,write}',
    ip_whitelist INET[] DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_api_keys_merchant ON merchant_api_keys(merchant_id);

-- 3. B-Sites
CREATE TABLE IF NOT EXISTS b_sites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    domain VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255),
    hosting_ip INET,
    hosting_provider VARCHAR(128),
    woocommerce_url VARCHAR(512),
    woocommerce_key VARCHAR(512),
    woocommerce_secret VARCHAR(512),
    ssl_provider VARCHAR(64),
    ssl_expires_at TIMESTAMPTZ,
    company_info JSONB DEFAULT '{}',
    bank_accounts JSONB DEFAULT '[]',
    social_media JSONB DEFAULT '{}',
    status VARCHAR(32) DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 4. Payment Accounts
CREATE TABLE IF NOT EXISTS payment_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    gateway VARCHAR(32) NOT NULL,
    alias VARCHAR(255),
    b_site_id UUID REFERENCES b_sites(id),
    merchant_id UUID REFERENCES merchants(id),
    encrypted_cred BYTEA NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'warming',
    weight INT DEFAULT 100,
    priority INT DEFAULT 0,
    tags VARCHAR(64)[] DEFAULT '{}',
    limit_config JSONB NOT NULL DEFAULT '{
        "single_min": 1.00,
        "single_max": 5000.00,
        "daily_max": 50000.00,
        "monthly_max": 500000.00,
        "lifetime_max": 5000000.00,
        "daily_count": 100
    }',
    supported_currencies VARCHAR(8)[] DEFAULT '{USD}',
    supported_countries VARCHAR(4)[] DEFAULT '{US}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 5. Proxies
CREATE TABLE IF NOT EXISTS proxies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    proxy_type VARCHAR(32) NOT NULL,
    protocol VARCHAR(16) NOT NULL DEFAULT 'socks5',
    host VARCHAR(255) NOT NULL,
    port INT NOT NULL,
    username VARCHAR(128),
    encrypted_password BYTEA,
    country VARCHAR(4),
    city VARCHAR(128),
    isp VARCHAR(128),
    bound_account_id UUID REFERENCES payment_accounts(id),
    is_dedicated BOOLEAN DEFAULT false,
    status VARCHAR(32) DEFAULT 'testing',
    latency INT,
    success_rate FLOAT DEFAULT 1.0,
    bandwidth_limit BIGINT,
    bandwidth_used BIGINT DEFAULT 0,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 6. Orders
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_no VARCHAR(64) NOT NULL UNIQUE,
    merchant_id UUID NOT NULL REFERENCES merchants(id),
    account_id UUID REFERENCES payment_accounts(id),
    gateway VARCHAR(32),
    amount NUMERIC(12,2) NOT NULL,
    currency VARCHAR(8) NOT NULL DEFAULT 'USD',
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    customer_email VARCHAR(255),
    customer_ip INET,
    customer_country VARCHAR(4),
    pay_token_hash VARCHAR(128),
    gateway_order_id VARCHAR(128),
    gateway_customer_id VARCHAR(128),
    risk_level VARCHAR(16) DEFAULT 'low',
    risk_score FLOAT DEFAULT 0,
    callback_data JSONB DEFAULT '{}',
    a_site_referer VARCHAR(512),
    metadata JSONB DEFAULT '{}',
    paid_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    canceled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_orders_merchant ON orders(merchant_id);
CREATE INDEX IF NOT EXISTS idx_orders_account ON orders(account_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_created ON orders(created_at);

-- 7. Payment Events
CREATE TABLE IF NOT EXISTS payment_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id),
    event_type VARCHAR(64) NOT NULL,
    gateway VARCHAR(32) NOT NULL,
    gateway_event_id VARCHAR(128),
    raw_data JSONB NOT NULL DEFAULT '{}',
    processed BOOLEAN DEFAULT false,
    retry_count INT DEFAULT 0,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 8. Settlements
CREATE TABLE IF NOT EXISTS settlements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL REFERENCES merchants(id),
    cycle_start TIMESTAMPTZ NOT NULL,
    cycle_end TIMESTAMPTZ NOT NULL,
    total_orders INT DEFAULT 0,
    total_amount NUMERIC(14,2) DEFAULT 0,
    total_fee NUMERIC(14,2) DEFAULT 0,
    net_amount NUMERIC(14,2) DEFAULT 0,
    payout_method VARCHAR(32),
    payout_address VARCHAR(512),
    payout_tx_id VARCHAR(128),
    status VARCHAR(32) DEFAULT 'pending',
    settled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 9. Reconciliation Records
CREATE TABLE IF NOT EXISTS reconciliation_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID REFERENCES orders(id),
    gateway VARCHAR(32) NOT NULL,
    gateway_tx_id VARCHAR(128),
    system_amount NUMERIC(12,2),
    gateway_amount NUMERIC(12,2),
    gateway_fee NUMERIC(12,2),
    diff_amount NUMERIC(12,2) DEFAULT 0,
    match_status VARCHAR(32) DEFAULT 'pending',
    mismatch_reason TEXT,
    resolved BOOLEAN DEFAULT false,
    reconciliation_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 10. Risk Events
CREATE TABLE IF NOT EXISTS risk_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID REFERENCES merchants(id),
    order_id UUID REFERENCES orders(id),
    account_id UUID REFERENCES payment_accounts(id),
    rule_name VARCHAR(128) NOT NULL,
    risk_level VARCHAR(16) NOT NULL,
    risk_score FLOAT DEFAULT 0,
    action VARCHAR(32) NOT NULL,
    reason TEXT,
    context JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 11. Exchange Rates
CREATE TABLE IF NOT EXISTS exchange_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    base_currency VARCHAR(8) NOT NULL,
    target_currency VARCHAR(8) NOT NULL,
    rate NUMERIC(18,8) NOT NULL,
    source VARCHAR(64),
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(base_currency, target_currency)
);

-- 12. Logistics Tracking
CREATE TABLE IF NOT EXISTS logistics_tracking (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID REFERENCES orders(id),
    tracking_number VARCHAR(128) NOT NULL,
    carrier VARCHAR(64) NOT NULL,
    status VARCHAR(32) DEFAULT 'pending',
    tracking_data JSONB DEFAULT '{}',
    synced_to_b_site BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 13. Account Activity Log
CREATE TABLE IF NOT EXISTS account_activity_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES payment_accounts(id),
    b_site_id UUID REFERENCES b_sites(id),
    activity_type VARCHAR(64) NOT NULL,
    amount NUMERIC(12,2),
    status VARCHAR(32),
    details JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 14. Admin Users
CREATE TABLE IF NOT EXISTS admin_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(128) UNIQUE NOT NULL,
    password_hash VARCHAR(256) NOT NULL,
    role VARCHAR(32) NOT NULL DEFAULT 'operator',
    totp_secret VARCHAR(64),
    totp_enabled BOOLEAN DEFAULT false,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 15. Audit Logs
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id UUID REFERENCES admin_users(id),
    action VARCHAR(128) NOT NULL,
    resource_type VARCHAR(64) NOT NULL,
    resource_id VARCHAR(64),
    old_value JSONB,
    new_value JSONB,
    ip_address INET,
    user_agent VARCHAR(512),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_merchants_updated') THEN
        CREATE TRIGGER trg_merchants_updated BEFORE UPDATE ON merchants FOR EACH ROW EXECUTE FUNCTION update_updated_at();
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_pa_updated') THEN
        CREATE TRIGGER trg_pa_updated BEFORE UPDATE ON payment_accounts FOR EACH ROW EXECUTE FUNCTION update_updated_at();
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_orders_updated') THEN
        CREATE TRIGGER trg_orders_updated BEFORE UPDATE ON orders FOR EACH ROW EXECUTE FUNCTION update_updated_at();
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_proxies_updated') THEN
        CREATE TRIGGER trg_proxies_updated BEFORE UPDATE ON proxies FOR EACH ROW EXECUTE FUNCTION update_updated_at();
    END IF;
END;
$$;
