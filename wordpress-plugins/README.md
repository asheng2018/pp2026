# AB Payment System — WordPress 插件与使用说明

---

## 一、两个站点的角色

```
A站 (仿牌)                          B站 (普货)
──────────────────────────────────────────────────────
用户在这里浏览/下单                   用户看不到这个站
不装任何支付插件                      装PayPal/Stripe支付插件
只装 AB Payment Bridge 插件           装 AB Payment Receiver 插件
订单正常创建，支付跳转iframe          接收A站传来的支付请求，调支付网关
```

---

## 二、A站 — 安装 AB Payment Bridge 插件

### 安装步骤

```
1. 将 wordpress-plugins/ab-payment-bridge/ 整个目录复制到 A站的 wp-content/plugins/
2. 登录 A站 WordPress 后台 → 插件 → 找到 "AB Payment Bridge" → 启用
3. 进入 设置 → AB Payment → 填写配置并保存
```

### 文件结构

```
ab-payment-bridge/
├── ab-payment-bridge.php              # 插件主文件
├── includes/
│   └── class-wc-ab-payment-gateway.php # 虚拟支付网关类
└── assets/
    └── ab-bridge.js                    # 前端 iframe 跳转 SDK
```

### 配置项说明

| 配置项 | 说明 | 示例 |
|--------|------|------|
| Orchestrator URL | AB系统调度引擎地址 | `http://your-server:8080` |
| Merchant ID | 在AB管理后台创建的商户ID | `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx` |
| API Key | 商户API密钥 | `ab_xxxxxxxxxxxxxxxxxxxxxxxxxxxx` |
| B-Site Domain | B站完整域名（含 https://） | `https://goodstore.shop` |
| Default Gateway | 默认支付网关 | `paypal` 或 `stripe` |
| Test Mode | 测试模式开关 | 勾选 = PayPal Sandbox |

### 插件功能

- **劫持 WooCommerce 支付流程**：替换默认支付方式，注册虚拟 AB Payment Gateway
- **后端调用调度引擎**：下单时通过 API 调用 AB 系统分配支付通道
- **前端 iframe 跳转**：在 checkout 页面创建全屏 iframe 加载 B站支付页
- **支付结果监听**：通过 postMessage 监听 B站支付完成/失败事件
- **管理后台设置页面**：可视化配置所有参数

### A站其他必须设置

```
WooCommerce → 设置 → 支付：
  - 只启用 "AB Payment Gateway"（其他全部禁用）

设置 → 阅读：
  - 勾选 "建议搜索引擎不索引本网站"（强烈建议）

WooCommerce → 设置 → 常规：
  - 货币与 B站保持一致

.htaccess 增加：
  Header set Referrer-Policy "no-referrer"
  Header set X-Frame-Options "SAMEORIGIN"
```

---

## 三、B站 — 安装 AB Payment Receiver 插件

### 安装步骤

1. **先安装官方支付插件**：
   - WooCommerce PayPal Payments (官方)
   - WooCommerce Stripe Payment Gateway (官方)
2. 将 `wordpress-plugins/ab-payment-receiver/` 复制到 B站的 `wp-content/plugins/`
3. 登录 B站 WordPress 后台 → 插件 → 启用 "AB Payment Receiver"
4. 进入 设置 → AB Receiver → 填写配置

### 配置项说明

| 配置项 | 说明 |
|--------|------|
| API Key | 与 AB系统通信的密钥 |
| Allowed IPs | 允许调用 API 的 IP（调度引擎IP） |

### PayPal/Stripe 官方插件设置

```
WooCommerce PayPal Payments:
  - 填入 B站专用的 PayPal Client ID + Secret
  - Sandbox 模式根据环境选择
  - 不需要设置独立 Webhook URL

WooCommerce Stripe Payment Gateway:
  - 填入 B站专用的 Publishable Key + Secret Key
  - Webhook Secret 由AB系统webhook统一处理
```

### 插件功能

- **REST API 端点** `/wp-json/ab-payment/v1/create`：接收调度引擎的支付创建请求
- **自定义路由** `/pay/{token}`：渲染支付页面
- **PayPal 订单创建**：通过 PayPal REST API 创建订单
- **Stripe PaymentIntent 创建**：通过 Stripe API 创建支付意图
- **支付页面模板**：加载 PayPal/Stripe SDK（以 B站域名加载）

### B站其他必须设置

```
设置 → 阅读：
  - 不要勾选 "建议搜索引擎不索引"（B站要对搜索引擎可见）

WooCommerce → 设置 → 产品：
  - 需要有一些普货商品（服装/家居/电子配件等）
  - 让网站看起来像一个正常运营的电商站

主题：
  - 使用普货风格主题（Storefront / Flatsome / Astra）
  - 安装一些标准插件增加真实性
  - 网站内容需要看起来像一个真实的正常电商站

.htaccess 增加：
  Header always set Content-Security-Policy "default-src 'self'; script-src 'self' https://www.paypal.com https://js.stripe.com; frame-src https://www.paypal.com https://js.stripe.com; connect-src https://api.paypal.com https://api.stripe.com;"
```

---

## 四、完整支付流程

```
1. 用户在 A站选商品 → 加购 → 结算
       ↓
2. WooCommerce 创建订单，状态= pending
       ↓
3. AB Payment Bridge 劫持 checkout，POST /api/v1/allocate 到调度引擎
       ↓
4. 调度引擎选一个B站账号，返回 pay_token + gateway_url
       ↓
5. A站前端创建 iframe，src = https://B站域名/pay/{token}
       ↓
6. B站 AB Payment Receiver 渲染支付页面
       ↓
7. B站页面加载 PayPal/Stripe SDK（域名是 B站！window.location = B站）
       ↓
8. 用户完成支付（PayPal/Stripe 完全不知道 A站的存在）
       ↓
9. PayPal/Stripe webhook 回调 → AB系统 webhook receiver → 更新订单状态
       ↓
10. AB系统通知商户（回调商户 webhook URL）
       ↓
11. B站支付页 postMessage 通知 A站 iframe → A站显示订单完成
```

### PayPal/Stripe 的视角

在整个支付过程中，PayPal/Stripe 只能看到：
- SDK **从 B站域名加载**
- `window.location.href` = **B站域名**
- HTTP Referer = **B站域名 或 空**
- IP 地址 = **服务器代理IP**（非用户真实IP）
- 没有任何 A站相关信息

---

## 五、AB 系统服务启动顺序

```bash
# 1. 数据库初始化
psql -U postgres -c "CREATE DATABASE ab_payment;"
psql -U postgres -d ab_payment -f sql/schema.sql

# 2. 启动中间件
docker-compose -f deploy/docker/docker-compose.yml up -d postgres redis nats

# 3. 启动调度引擎
cd cmd/orchestrator && go run main.go
# 或
docker-compose -f deploy/docker/docker-compose.yml up -d orchestrator

# 4. 启动支付网关
cd cmd/gateway && go run main.go
# 或
docker-compose -f deploy/docker/docker-compose.yml up -d gateway

# 5. 启动管理后台（可选）
cd web && npm install && npm run dev
```

### Docker Compose 一键部署

```bash
# 设置环境变量
export DB_PASSWORD=your_secure_password
export REDIS_PASSWORD=your_redis_password
export JWT_SECRET=your_jwt_secret

# 启动全部服务
docker-compose -f deploy/docker/docker-compose.yml up -d

# 查看状态
docker-compose -f deploy/docker/docker-compose.yml ps

# 查看日志
docker-compose -f deploy/docker/docker-compose.yml logs -f
```

### 管理后台登陆

```
URL: http://localhost:3000
默认账号: admin / admin123（需要在 admin_users 表中创建）
```
