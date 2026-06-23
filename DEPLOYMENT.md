# AB Payment System — 完整安装部署说明

---

## 目录

1. [系统概述](#1-系统概述)
2. [环境要求](#2-环境要求)
3. [项目结构](#3-项目结构)
4. [架构部署拓扑](#4-架构部署拓扑)
5. [基础设施部署](#5-基础设施部署)
6. [后端服务部署](#6-后端服务部署)
7. [WordPress 插件部署](#7-wordpress-插件部署)
8. [管理后台部署](#8-管理后台部署)
9. [Docker Compose 一键部署](#9-docker-compose-一键部署)
10. [Kubernetes 部署](#10-kubernetes-部署)
11. [初始配置向导](#11-初始配置向导)
12. [运维监控](#12-运维监控)
13. [故障排查](#13-故障排查)
14. [安全加固](#14-安全加固)

---

## 1. 系统概述

AB Payment System 是一套完整的 PayPal/Stripe AB轮询收款系统，包含15个核心模块：

| 模块 | 功能 | 服务 |
|------|------|------|
| **Orchestrator** | 核心调度引擎、账号选择、风控 | `cmd/orchestrator` |
| **Gateway** | AB跳转网关、支付页面渲染、反追踪 | `cmd/gateway` |
| **Admin Dashboard** | React 管理后台 | `web/` |
| **WordPress 插件** | A站跳转插件 + B站接收插件 | `wordpress-plugins/` |

### 核心数据流

```
用户 → A站(WP) → Orchestrator(调度) → B站(WP) → PayPal/Stripe
  │                  │                      │
  └── iframe支付 ────┴──────────────────────┘
```

---

## 2. 环境要求

### 2.1 开发/生产服务器

| 组件 | 最低版本 | 推荐 |
|------|---------|------|
| **Go** | 1.22+ | 1.22 |
| **Node.js** | 18+ | 20 LTS |
| **PostgreSQL** | 14+ | 16 |
| **Redis** | 6+ | 7 |
| **NATS** | 2.10+ | 2.10 |
| **Docker** | 24+ | 26+ |
| **Docker Compose** | 2.0+ | 2.24+ |
| **Kubernetes** (可选) | 1.28+ | 1.30+ |

### 2.2 WordPress 环境（A站和B站各需独立）

| 组件 | 要求 |
|------|------|
| **WordPress** | 5.8+（推荐 6.5+） |
| **WooCommerce** | 5.0+（推荐 8.0+） |
| **PHP** | 7.4+（推荐 8.2+） |
| **MySQL/MariaDB** | 5.7+ / 10.3+ |
| **SSL 证书** | 必须（B站必须HTTPS） |
| **独立域名** | A站和B站使用不同域名 |
| **独立IP** | A站和B站使用不同服务器IP |

---

## 3. 项目结构

```
ab-payment-system/
│
├── cmd/                            # 服务入口
│   ├── orchestrator/main.go        #   调度引擎入口 (端口 8080/9090)
│   └── gateway/main.go             #   支付网关入口 (端口 8081)
│
├── internal/                       # 15个核心业务模块
│   ├── scheduler/                  #   01 核心调度引擎
│   ├── gateway/                    #   02 AB跳转网关
│   ├── account/                    #   03 账号池管理
│   ├── proxy/                      #   04 IP代理池
│   ├── risk/                       #   05 风控引擎
│   ├── merchant/                   #   06 商户管理
│   ├── order/                      #   07 订单管理
│   ├── settlement/                 #   08 结算引擎
│   ├── webhook/                    #   09 回调通知
│   ├── reconciliation/             #   10 对账系统
│   ├── farming/                    #   11 养号自动化
│   ├── exchange/                   #   12 汇率管理
│   ├── logistics/                  #   13 物流管理
│   └── monitoring/                 #   14 监控告警
│
├── pkg/                            # 共享基础库
│   ├── config/config.go            #   配置加载
│   ├── crypto/crypto.go            #   AES-256-GCM 加密
│   ├── db/postgres.go              #   数据库连接
│   ├── redis/redis.go              #   Redis连接
│   ├── queue/nats.go               #   NATS消息队列
│   ├── errors/errors.go            #   统一错误定义
│   ├── logger/logger.go            #   日志配置
│   ├── types/types.go              #   共享类型定义
│   ├── utils/utils.go              #   工具函数
│   └── validator/validator.go      #   数据校验
│
├── api/proto/                      # gRPC 协议定义
│   ├── scheduler.proto
│   └── merchant.proto
│
├── sql/schema.sql                  # 数据库Schema（15张表）
│
├── web/                            # React 管理后台
│   ├── package.json
│   ├── vite.config.ts
│   └── src/
│       ├── main.tsx
│       ├── App.tsx
│       ├── index.css
│       ├── components/Layout/MainLayout.tsx
│       └── pages/Dashboard/index.tsx
│
├── wordpress-plugins/              # WordPress 插件
│   ├── README.md
│   ├── ab-payment-bridge/          #   A站插件（仿牌站装）
│   │   ├── ab-payment-bridge.php
│   │   ├── includes/class-wc-ab-payment-gateway.php
│   │   └── assets/ab-bridge.js
│   └── ab-payment-receiver/        #   B站插件（普货站装）
│       ├── ab-payment-receiver.php
│       └── templates/
│           ├── pay-page.php
│           └── result-page.php
│
├── deploy/                         # 部署配置
│   ├── docker/
│   │   ├── Dockerfile.orchestrator
│   │   ├── Dockerfile.gateway
│   │   ├── docker-compose.yml
│   │   └── prometheus.yml
│   └── k8s/
│       └── deployment.yaml
│
└── go.mod                          # Go模块定义
```

---

## 4. 架构部署拓扑

```
                          Internet
                             │
              ┌──────────────┼──────────────┐
              │              │              │
         ┌────▼────┐   ┌────▼────┐   ┌────▼────┐
         │  A站     │   │  B站     │   │ 管理后台 │
         │(仿牌域名) │   │(普货域名) │   │(内部访问) │
         │WP+AB插件 │   │WP+PayPal│   │React App│
         │:443      │   │:443     │   │:3000    │
         └────┬─────┘   └────┬─────┘   └────┬─────┘
              │              │              │
              │     ┌────────┼────────┐     │
              │     │  内部网络 (VPC)  │     │
              │     │                │     │
              │  ┌──▼──────┐  ┌──────▼──┐  │
              │  │Orchestrator│ │Gateway  │  │
              │  │ :8080     │  │ :8081   │  │
              │  │ :9090(gRPC)│  │         │  │
              │  └──┬───┬───┘  └────┬────┘  │
              │     │   │          │       │
              │  ┌──▼───▼──┐  ┌───▼───┐  │
              │  │PostgreSQL│  │ Redis  │  │
              │  │ :5432   │  │ :6379  │  │
              │  └─────────┘  └────────┘  │
              │     │                      │
              │  ┌──▼──────┐  ┌─────────┐ │
              │  │  NATS   │  │Prometheus│ │
              │  │ :4222   │  │ :9090   │ │
              │  └─────────┘  │Grafana  │ │
              │               │ :3000   │ │
              └───────────────┴─────────┘─┘
```

### 网络要求

| 服务 | 端口 | 协议 | 对外暴露 | 说明 |
|------|------|------|---------|------|
| Orchestrator HTTP | 8080 | TCP | 可选 | A站插件调用 |
| Orchestrator gRPC | 9090 | TCP | 否 | 内部服务通信 |
| Gateway HTTP | 8081 | TCP | 否 | 支付页面服务 |
| PostgreSQL | 5432 | TCP | 否 | 数据库 |
| Redis | 6379 | TCP | 否 | 缓存/状态 |
| NATS | 4222 | TCP | 否 | 消息队列 |
| Prometheus | 9090 | TCP | 否 | 监控 |
| Grafana | 3000 | TCP | 可选 | 运维面板 |
| A站 WordPress | 443 | TCP | 是 | 用户访问 |
| B站 WordPress | 443 | TCP | 是 | 支付页面域名 |
| Admin Dashboard | 3000 | TCP | 否 | 管理后台 |

---

## 5. 基础设施部署

### 5.1 PostgreSQL 数据库

#### 方式一：Docker
```bash
docker run -d \
  --name ab-postgres \
  --restart unless-stopped \
  -e POSTGRES_DB=ab_payment \
  -e POSTGRES_USER=abpay \
  -e POSTGRES_PASSWORD=<your-secure-password> \
  -v pgdata:/var/lib/postgresql/data \
  -p 127.0.0.1:5432:5432 \
  postgres:16-alpine
```

#### 方式二：直接安装
```bash
# Ubuntu/Debian
sudo apt install postgresql-16
sudo -u postgres psql -c "CREATE USER abpay WITH PASSWORD '<your-password>';"
sudo -u postgres psql -c "CREATE DATABASE ab_payment OWNER abpay;"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE ab_payment TO abpay;"
```

#### 初始化数据库表
```bash
# 执行 schema.sql 创建全部15张表
psql -h <DB_HOST> -p 5432 -U abpay -d ab_payment -f sql/schema.sql

# 验证
psql -h <DB_HOST> -U abpay -d ab_payment -c "\dt"
# 应该看到15张表:
#   merchants, merchant_api_keys, b_sites, payment_accounts,
#   proxies, orders, payment_events, settlements,
#   reconciliation_records, risk_events, exchange_rates,
#   logistics_tracking, account_activity_log, admin_users, audit_logs
```

### 5.2 Redis

```bash
# Docker 方式
docker run -d \
  --name ab-redis \
  --restart unless-stopped \
  --requirepass <your-redis-password> \
  -v redisdata:/data \
  -p 127.0.0.1:6379:6379 \
  redis:7-alpine redis-server --appendonly yes
```

### 5.3 NATS (消息队列)

```bash
docker run -d \
  --name ab-nats \
  --restart unless-stopped \
  -p 127.0.0.1:4222:4222 \
  -p 127.0.0.1:8222:8222 \
  nats:2.10-alpine -js -m 8222
```

---

## 6. 后端服务部署

### 6.1 配置环境变量

创建 `.env` 文件：

```bash
# ===== 环境 =====
export ENV=production
# development | production

# ===== 数据库 =====
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=abpay
export DB_PASSWORD=<your-db-password>
export DB_NAME=ab_payment
export DB_SSLMODE=disable
export DB_MAX_CONNS=50

# ===== Redis =====
export REDIS_ADDR=localhost:6379
export REDIS_PASSWORD=<your-redis-password>
export REDIS_DB=0
export REDIS_POOL_SIZE=20

# ===== NATS =====
export NATS_URL=nats://localhost:4222

# ===== 安全 =====
export JWT_SECRET=<generate-a-random-64-char-string>
# 生成方式: openssl rand -hex 32
export VAULT_ADDR=http://localhost:8200  # 可选: HashiCorp Vault
export VAULT_TOKEN=<vault-token>

# ===== 代理 =====
export PROXY_DEFAULT_TYPE=residential
export PROXY_MAX_FAIL_COUNT=3

# ===== 风控 =====
export RISK_MAX_FAILS=3
export RISK_MIN_SUCCESS_RATE=0.7

# ===== 调度 =====
export SCHEDULER_STRATEGY=weighted_round_robin
# weighted_round_robin | sequential | random | least_utilized

# ===== 服务端口 =====
export SERVER_PORT=8080
export GRPC_PORT=9090
```

```bash
# 加载环境变量
source .env

# 生成安全的 JWT_SECRET
export JWT_SECRET=$(openssl rand -hex 32)
echo "JWT_SECRET=$JWT_SECRET" >> .env
```

### 6.2 编译后端服务

```bash
cd ab-payment-system

# 安装依赖
go mod download
go mod tidy

# 编译调度引擎
go build -o bin/orchestrator ./cmd/orchestrator

# 编译支付网关
go build -o bin/gateway ./cmd/gateway
```

### 6.3 启动服务

```bash
# 终端1: 启动调度引擎
source .env
./bin/orchestrator

# 终端2: 启动支付网关
source .env
export SERVER_PORT=8081  # 网关使用不同端口
./bin/gateway
```

### 6.4 验证服务

```bash
# 检查调度引擎健康状态
curl http://localhost:8080/health
# 预期返回: {"status":"healthy","components":[...]}

# 检查就绪状态
curl http://localhost:8080/ready
# 预期返回: ready

# 检查支付网关
curl http://localhost:8081/health

# 查看 Prometheus 指标
curl http://localhost:8080/metrics
```

### 6.5 配置 systemd 服务（生产环境）

**`/etc/systemd/system/ab-orchestrator.service`**：

```ini
[Unit]
Description=AB Payment Orchestrator
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=abpay
Group=abpay
WorkingDirectory=/opt/ab-payment-system
EnvironmentFile=/opt/ab-payment-system/.env
ExecStart=/opt/ab-payment-system/bin/orchestrator
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

**`/etc/systemd/system/ab-gateway.service`**：

```ini
[Unit]
Description=AB Payment Gateway
After=network.target ab-orchestrator.service

[Service]
Type=simple
User=abpay
Group=abpay
WorkingDirectory=/opt/ab-payment-system
EnvironmentFile=/opt/ab-payment-system/.env
ExecStart=/opt/ab-payment-system/bin/gateway
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
# 启用服务
sudo systemctl daemon-reload
sudo systemctl enable ab-orchestrator ab-gateway
sudo systemctl start ab-orchestrator ab-gateway

# 查看状态
sudo systemctl status ab-orchestrator
sudo systemctl status ab-gateway

# 查看日志
sudo journalctl -u ab-orchestrator -f
sudo journalctl -u ab-gateway -f
```

---

## 7. WordPress 插件部署

### 7.1 A站（仿牌站）部署

#### 步骤1：安装 WordPress + WooCommerce
```bash
# 使用标准的 WordPress 安装流程即可
# 也可以通过 WP-CLI 快速安装

wp core install \
  --url=https://your-a-site.com \
  --title="Your Store" \
  --admin_user=admin \
  --admin_password=<secure-password> \
  --admin_email=admin@your-a-site.com

wp plugin install woocommerce --activate
wp theme install storefront --activate
```

#### 步骤2：安装 AB Payment Bridge 插件
```bash
# 复制插件到 A站
cp -r wordpress-plugins/ab-payment-bridge/ /var/www/a-site/wp-content/plugins/

# 通过 WP-CLI 激活
wp plugin activate ab-payment-bridge

# 配置插件
wp option update ab_orchestrator_url "http://<orchestrator-ip>:8080"
wp option update ab_merchant_id "<your-merchant-id>"
wp option update ab_api_key "<your-api-key>"
wp option update ab_b_site_domain "https://<b-site-domain>"
wp option update ab_default_gateway "paypal"
```

#### 步骤3：A站 WooCommerce 设置
```
WooCommerce → 设置 → 支付：
  - 只启用 "AB Secure Payment"
  - 禁用所有其他支付方式（包括默认的 PayPal、Stripe 等）

设置 → 阅读：
  - ✅ 勾选 "建议搜索引擎不索引本网站"

设置 → 固定链接：
  - 选择 "文章名" 格式
```

#### 步骤4：添加反追踪 HTTP 头
在 A站的 Nginx 配置或 `.htaccess` 中添加：

```nginx
# Nginx 配置
add_header Referrer-Policy "no-referrer" always;
add_header X-Frame-Options "SAMEORIGIN" always;
add_header X-Content-Type-Options "nosniff" always;
add_header X-Robots-Tag "noindex, nofollow" always;
```

```apache
# Apache .htaccess
Header always set Referrer-Policy "no-referrer"
Header always set X-Frame-Options "SAMEORIGIN"
Header always set X-Content-Type-Options "nosniff"
Header always set X-Robots-Tag "noindex, nofollow"
```

### 7.2 B站（普货站）部署

#### 步骤1：安装 WordPress + WooCommerce
```bash
wp core install \
  --url=https://your-b-site.com \
  --title="GoodStore - Quality Products" \
  --admin_user=admin \
  --admin_password=<secure-password> \
  --admin_email=admin@your-b-site.com

wp plugin install woocommerce --activate
wp theme install astra --activate
```

#### 步骤2：安装官方支付插件
```bash
# PayPal
wp plugin install woocommerce-paypal-payments --activate

# Stripe
wp plugin install woocommerce-gateway-stripe --activate
```

#### 步骤3：配置 PayPal Payments
```
WooCommerce → 设置 → 支付 → PayPal：
  1. 点击 "Activate PayPal"
  2. 使用B站专用的 PayPal Business 账号登录
  3. 或手动输入 API 凭证:
     - Client ID: <B站专用PayPal Client ID>
     - Secret Key: <B站专用PayPal Secret>
  4. Sandbox 模式: 测试时开启，生产时关闭
  5. Webhook URL 留空（由AB系统统一处理）
```

#### 步骤4：配置 Stripe
```
WooCommerce → 设置 → 支付 → Stripe：
  1. 输入 Publishable Key: <B站专用Stripe Publishable Key>
  2. 输入 Secret Key: <B站专用Stripe Secret Key>
  3. Webhook Secret 留空
```

#### 步骤5：安装 AB Payment Receiver 插件
```bash
# 复制插件到 B站
cp -r wordpress-plugins/ab-payment-receiver/ /var/www/b-site/wp-content/plugins/

# 激活
wp plugin activate ab-payment-receiver

# 配置
wp option update ab_receiver_api_key "<same-api-key-as-orchestrator>"
wp option update ab_receiver_allowed_ips "<orchestrator-ip>"
```

#### 步骤6：B站 SEO/内容设置
```
设置 → 阅读：
  - ❌ 不要勾选"建议搜索引擎不索引"（B站要对搜索引擎可见）

WooCommerce → 设置 → 产品 → 常规：
  - 添加至少 10-20 个普货商品（服装/家居/电子配件等）
  - 设置合理的价格范围 ($5 - $200)
  - 添加商品图片和描述

外观 → 自定义：
  - 设置网站 Logo
  - 设置网站图标 (favicon)
  - 配置页脚信息（公司名、地址等）
```

#### 步骤7：B站安全头配置
```nginx
# Nginx - B站 CSP 配置
add_header Content-Security-Policy "default-src 'self'; 
  script-src 'self' 'unsafe-inline' https://www.paypal.com https://www.paypalobjects.com https://js.stripe.com; 
  frame-src https://www.paypal.com https://js.stripe.com https://hooks.stripe.com; 
  connect-src 'self' https://api.paypal.com https://api-m.paypal.com https://api.stripe.com; 
  img-src 'self' data: https://www.paypalobjects.com https://*.stripe.com;" always;

add_header X-Frame-Options "SAMEORIGIN" always;
add_header X-Content-Type-Options "nosniff" always;

# 支付页面禁止缓存
location /pay/ {
    add_header Cache-Control "no-store, no-cache, must-revalidate, proxy-revalidate";
    add_header Pragma "no-cache";
    add_header Expires "0";
}
```

### 7.3 两个站点的对比清单

| 设置项 | A站 (仿牌) | B站 (普货) |
|--------|-----------|-----------|
| WooCommerce | ✅ | ✅ |
| PayPal Payments 插件 | ❌ 不装 | ✅ 必装 |
| Stripe Gateway 插件 | ❌ 不装 | ✅ 必装 |
| AB Payment Bridge | ✅ 必装 | ❌ 不装 |
| AB Payment Receiver | ❌ 不装 | ✅ 必装 |
| WooCommerce支付方式 | 只启用AB Payment | PayPal + Stripe |
| 商品类型 | 仿牌商品 | 普货商品 |
| 搜索引擎可见 | ❌ noindex | ✅ |
| SSL (HTTPS) | ✅ | ✅ 必须 |
| 对外访问 | 客户可访问 | AB系统可访问 |
| 服务器IP | 独立 | 独立（不同C段） |
| 域名 | 独立域名 | 独立域名 |

---

## 8. 管理后台部署

### 8.1 开发模式
```bash
cd web

# 安装依赖
npm install

# 启动开发服务器
npm run dev
# 访问 http://localhost:3000
```

### 8.2 生产构建
```bash
cd web

# 构建
npm run build

# 输出在 web/dist/ 目录
# 部署到 Nginx 或任何静态文件服务器
```

### 8.3 Nginx 配置
```nginx
server {
    listen 80;
    server_name admin.your-domain.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name admin.your-domain.com;

    ssl_certificate     /etc/ssl/admin.crt;
    ssl_certificate_key /etc/ssl/admin.key;

    root /var/www/ab-admin;
    index index.html;

    # SPA 路由支持
    location / {
        try_files $uri $uri/ /index.html;
    }

    # API 代理到调度引擎
    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # 安全头
    add_header X-Frame-Options "DENY" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
}
```

---

## 9. Docker Compose 一键部署

### 9.1 准备环境

```bash
# 确保安装了 Docker 和 Docker Compose
docker --version    # >= 24
docker-compose --version  # >= 2.0

# 创建必要的目录
mkdir -p /opt/ab-payment-system
cd /opt/ab-payment-system

# 复制整个项目到此目录
```

### 9.2 配置环境变量

```bash
# 创建 .env 文件
cat > deploy/docker/.env << EOF
DB_PASSWORD=<your-secure-db-password>
REDIS_PASSWORD=<your-redis-password>
JWT_SECRET=$(openssl rand -hex 32)
GRAFANA_PASSWORD=<your-grafana-password>
EOF
```

### 9.3 启动全部服务

```bash
# 从项目根目录执行
cd /opt/ab-payment-system

# 一键启动
docker-compose -f deploy/docker/docker-compose.yml up -d

# 查看启动状态
docker-compose -f deploy/docker/docker-compose.yml ps
```

预期输出：
```
NAME                STATUS    PORTS
ab-postgres         running   0.0.0.0:5432->5432/tcp
ab-redis            running   0.0.0.0:6379->6379/tcp
ab-nats             running   0.0.0.0:4222->4222/tcp, 0.0.0.0:8222->8222/tcp
ab-orchestrator     running   0.0.0.0:8080->8080/tcp, 0.0.0.0:9090->9090/tcp
ab-gateway          running   0.0.0.0:8081->8081/tcp
ab-prometheus       running   0.0.0.0:9090->9090/tcp
ab-grafana          running   0.0.0.0:3000->3000/tcp
```

### 9.4 管理命令

```bash
# 查看日志
docker-compose -f deploy/docker/docker-compose.yml logs -f orchestrator
docker-compose -f deploy/docker/docker-compose.yml logs -f gateway

# 重启单个服务
docker-compose -f deploy/docker/docker-compose.yml restart orchestrator

# 停止全部
docker-compose -f deploy/docker/docker-compose.yml down

# 停止并删除数据卷（慎用！）
docker-compose -f deploy/docker/docker-compose.yml down -v

# 更新后重新构建
docker-compose -f deploy/docker/docker-compose.yml build orchestrator gateway
docker-compose -f deploy/docker/docker-compose.yml up -d
```

---

## 10. Kubernetes 部署

### 10.1 构建 Docker 镜像

```bash
# 构建
docker build -f deploy/docker/Dockerfile.orchestrator -t ab-payment/orchestrator:latest .
docker build -f deploy/docker/Dockerfile.gateway -t ab-payment/gateway:latest .

# 推送到镜像仓库
docker tag ab-payment/orchestrator:latest your-registry/ab-payment/orchestrator:latest
docker push your-registry/ab-payment/orchestrator:latest
docker tag ab-payment/gateway:latest your-registry/ab-payment/gateway:latest
docker push your-registry/ab-payment/gateway:latest
```

### 10.2 修改 K8s 配置

编辑 `deploy/k8s/deployment.yaml`，替换以下内容：

- `your-registry/ab-payment/orchestrator:latest` → 你的镜像地址
- `your-registry/ab-payment/gateway:latest` → 你的镜像地址
- `changeme` → 真实密码（通过 Sealed Secrets 或 External Secrets 管理）

### 10.3 部署到 K8s

```bash
# 创建命名空间
kubectl create namespace ab-payment

# 部署
kubectl apply -f deploy/k8s/deployment.yaml

# 查看状态
kubectl -n ab-payment get all

# 查看日志
kubectl -n ab-payment logs -f deployment/orchestrator
kubectl -n ab-payment logs -f deployment/gateway

# 端口转发（本地访问）
kubectl -n ab-payment port-forward svc/orchestrator-svc 8080:8080
```

---

## 11. 初始配置向导

### 11.1 首次启动后的配置步骤

#### 1) 创建管理员账号
```sql
-- 在 PostgreSQL 中执行
INSERT INTO admin_users (username, password_hash, role)
VALUES ('admin', '<bcrypt-hash-of-password>', 'superadmin');
```

#### 2) 创建 B站记录
```sql
INSERT INTO b_sites (domain, name, hosting_ip, hosting_provider, status)
VALUES (
  'your-b-site.com',
  'GoodStore',
  '<b-site-server-ip>',
  'AWS EC2',
  'active'
);
```

#### 3) 创建支付账号
```sql
-- 将 PayPal/Stripe 凭证加密后存入
-- encrypted_cred 字段需要通过 AB 管理后台的 API 加密写入
-- 或者用以下方式：

-- 通过管理后台操作：
-- 登录管理后台 → Accounts → Add Account →
--   输入 PayPal Client ID + Secret
--   输入限额配置
--   选择关联的 B站
```

#### 4) 创建商户
```sql
INSERT INTO merchants (name, email, status, routing_mode, account_group)
VALUES ('Merchant-A', 'merchant@example.com', 'active', 'weighted_round_robin', 'group-a');
```

#### 5) 生成商户 API Key
```bash
# 通过管理后台操作：
# Merchants → 选择商户 → Generate API Key →
#   记录下显示的 API Key（只显示一次！）
#   将这个 Key 填入 A站的 AB Payment Bridge 插件配置
```

#### 6) 验证全链路
```bash
# 1. 检查调度引擎健康状态
curl http://localhost:8080/health

# 2. 测试支付分配
curl -X POST http://localhost:8080/api/v1/allocate \
  -H "Content-Type: application/json" \
  -H "X-API-Key: <merchant-api-key>" \
  -d '{
    "order_id": "test-001",
    "amount": "10.00",
    "currency": "USD",
    "merchant_id": "<merchant-id>",
    "gateway": "paypal",
    "country": "US"
  }'

# 3. 在浏览器访问 A站 → 添加商品 → 结算
#    确认支付 iframe 正确加载 B站的支付页面
```

---

## 12. 运维监控

### 12.1 关键指标

访问 Grafana: `http://<server>:3000`（默认账号 admin，密码见 .env）

内置 Prometheus 指标：

| 指标 | 含义 | 告警阈值 |
|------|------|---------|
| `ab_allocation_total` | 分配总数 | 突然下降 > 50% |
| `ab_allocation_latency_seconds` | 分配延迟 | P95 > 1s |
| `ab_account_health` | 账号在线状态 | 任何账号变为 0 |
| `ab_account_success_rate` | 账号成功率 | 低于 70% |
| `ab_order_paid_total` | 成功支付数 | 正常波动 |
| `ab_risk_actions_total` | 风控拦截数 | 突然飙升 |
| `ab_proxy_health` | 代理健康 | 任何代理变为 0 |
| `ab_webhook_received_total` | 回调数 | 突然下降 |

### 12.2 日志查看

```bash
# Docker Compose
docker-compose -f deploy/docker/docker-compose.yml logs -f orchestrator | grep ERROR

# Systemd
sudo journalctl -u ab-orchestrator --since "10 minutes ago" | grep ERROR

# K8s
kubectl -n ab-payment logs -f deployment/orchestrator --tail=100
```

### 12.3 日常运维任务

```bash
# ===== 每日检查 =====
# 1. 检查服务状态
curl http://localhost:8080/health

# 2. 检查在线账号数
curl http://localhost:8080/api/v1/health | jq .online_accounts

# 3. 检查今天订单量
# 管理后台 Dashboard 查看

# ===== 每周任务 =====
# 1. 对账
curl http://localhost:8080/api/v1/reconciliation/run

# 2. 查看结算状态
curl http://localhost:8080/api/v1/settlements/pending

# 3. 检查代理池健康
# 管理后台 Proxy Pool 页面

# ===== 每月任务 =====
# 1. 生成月度报表
# 管理后台 Reports → Monthly Report

# 2. 轮换部分代理IP
# 管理后台 Proxy Pool → Rotate

# 3. 养号进度检查
# 管理后台 Accounts → Farming Status
```

---

## 13. 故障排查

### 13.1 服务无法启动

```bash
# 检查端口占用
lsof -i :8080
lsof -i :8081

# 检查数据库连接
psql -h <DB_HOST> -U abpay -d ab_payment -c "SELECT 1;"

# 检查 Redis 连接
redis-cli -h <REDIS_HOST> -a <PASSWORD> PING
# 应返回: PONG

# 查看详细启动日志
# Docker:
docker logs ab-orchestrator
# Systemd:
sudo journalctl -u ab-orchestrator -n 50
```

### 13.2 支付分配失败

```bash
# 检查是否有在线账号
curl http://localhost:8080/api/v1/health
# online_accounts 应为 > 0

# 检查账号状态
# 管理后台 → Accounts → 查看状态列
# 如果全部 showing "cooling" 或 "offline":
#   等待冷却期结束（默认30分钟）
#   或手动将账号状态改为 online

# 检查账号限额是否用尽
# 管理后台 → Accounts → 查看 Daily Used 是否接近 Daily Max
```

### 13.3 WordPress 插件通信失败

```bash
# 从 A站服务器测试到调度引擎的连通性
curl -v http://<orchestrator-ip>:8080/health

# 检查 WordPress 错误日志
tail -f /var/www/a-site/wp-content/debug.log

# 确认 API Key 一致
# A站 AB Payment Bridge 设置中的 API Key
# 必须与 AB 管理后台中商户的 API Key 完全一致
```

### 13.4 支付页面加载失败

```bash
# 检查 B站是否可访问
curl -I https://<b-site-domain>/pay/test-token

# 检查 B站 SSL 证书
openssl s_client -connect <b-site-domain>:443 -servername <b-site-domain>

# 检查 B站 rewrite rules 是否生效
# 登录 B站 WP后台 → 设置 → 固定链接 → 重新保存一次

# 检查 B站插件是否启用
wp plugin list --status=active | grep ab-payment
```

### 13.5 Webhook 回调不到

```bash
# 检查 PayPal/Stripe Webhook URL 是否指向正确
# PayPal: https://<orchestrator-domain>/api/v1/webhook/paypal
# Stripe: https://<orchestrator-domain>/api/v1/webhook/stripe

# 查看 webhook 日志
docker logs ab-orchestrator | grep webhook

# 验证 webhook 签名
curl -X POST http://localhost:8080/api/v1/webhook/paypal/test
```

---

## 14. 安全加固

### 14.1 网络安全

```bash
# 使用防火墙限制端口访问
# 只开放必要的端口

# Orchestrator 只允许 A站服务器 IP 访问
sudo ufw allow from <a-site-ip> to any port 8080

# Gateway 只允许 Orchestrator 访问
sudo ufw allow from 127.0.0.1 to any port 8081

# PostgreSQL 只允许本地访问
sudo ufw deny 5432

# Redis 只允许本地访问
sudo ufw deny 6379
```

### 14.2 密钥管理

```bash
# 生产环境使用 HashiCorp Vault 存储密钥
# 不要在配置文件中硬编码密钥

# 最低要求：
# - JWT_SECRET 长度 >= 32 字节
# - DB_PASSWORD 长度 >= 16 字符
# - API Key 使用 ab_ 前缀 + UUID + 随机后缀
# - PayPal/Stripe 凭证使用 AES-256-GCM 加密存储
```

### 14.3 服务器加固

```bash
# 定期更新
sudo apt update && sudo apt upgrade -y

# 安装 fail2ban
sudo apt install fail2ban -y

# 配置 SSH 密钥登录，禁用密码登录
# /etc/ssh/sshd_config:
#   PasswordAuthentication no
#   PermitRootLogin no

# A站和B站使用不同的服务器提供商
# 使用不同的 IP 段
# 域名使用不同的注册商
```

---

## 附录 A: 快速启动检查清单

- [ ] PostgreSQL 安装并初始化 schema.sql
- [ ] Redis 启动并可通过密码连接
- [ ] NATS 启动并启用 JetStream
- [ ] Orchestrator 编译成功并启动（健康检查通过）
- [ ] Gateway 编译成功并启动
- [ ] A站 WordPress + WooCommerce 安装
- [ ] A站 AB Payment Bridge 插件安装并配置
- [ ] B站 WordPress + WooCommerce 安装
- [ ] B站 PayPal Payments + Stripe 插件安装
- [ ] B站 AB Payment Receiver 插件安装并配置
- [ ] B站至少添加 10 个普货商品
- [ ] 管理后台可登陆
- [ ] 创建至少一个商户
- [ ] 创建至少一个支付账号
- [ ] 创建至少一个 B站记录
- [ ] 生成商户 API Key 并填入 A站插件
- [ ] 端到端测试支付流程通过

## 附录 B: 常用配置文件路径

| 文件 | 路径 |
|------|------|
| 数据库 Schema | `sql/schema.sql` |
| Docker Compose | `deploy/docker/docker-compose.yml` |
| Dockerfile (Orchestrator) | `deploy/docker/Dockerfile.orchestrator` |
| Dockerfile (Gateway) | `deploy/docker/Dockerfile.gateway` |
| K8s 部署 | `deploy/k8s/deployment.yaml` |
| Prometheus 配置 | `deploy/docker/prometheus.yml` |
| A站插件 | `wordpress-plugins/ab-payment-bridge/` |
| B站插件 | `wordpress-plugins/ab-payment-receiver/` |
| Go 模块 | `go.mod` |
| 管理后台 | `web/` |

## 附录 C: 技术支持

- 完整架构设计文档：见项目根目录
- 模块拆分文档：见项目根目录
- WordPress 插件详细说明：`wordpress-plugins/README.md`
- 健康检查端点：`GET /health`（所有服务）
- Prometheus 指标：`GET /metrics`（所有服务）
