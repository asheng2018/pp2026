# AB支付系统 — A站和B站 WordPress/WooCommerce 接入完整指南

---

## 概述

AB支付系统需要两个独立的 WordPress 站点：

| 站点 | 角色 | 用户可见 | 插件 | 支付方式 |
|------|------|---------|------|---------|
| **A站** | 仿牌站，用户在这里浏览下单 | 是 | AB Payment Bridge | 只启AB Payment Gateway |
| **B站** | 普货站，支付真正发生的地方 | 否 | AB Payment Receiver + PayPal/Stripe | PayPal + Stripe |

**关键原则**：A站和B站必须使用不同的域名、不同的服务器IP、不同的WordPress安装。

---

## 第一步：A站设置

### 1.1 安装 WordPress + WooCommerce

```bash
# 标准 WordPress 安装
# 安装 WooCommerce 插件
# 设置商店货币、运费等基础参数
```

### 1.2 安装 AB Payment Bridge 插件

```bash
# 把 ab-payment-bridge 目录上传到 A站
cp -r wordpress-plugins/ab-payment-bridge/ /var/www/a-site/wp-content/plugins/
```

```bash
# 通过 WP-CLI 激活
wp plugin activate ab-payment-bridge --path=/var/www/a-site
```

或在 WordPress 后台 → 插件 → 找到 "AB Payment Bridge" → 启用。

### 1.3 配置 AB Payment Bridge

A站后台 → 设置 → AB Payment：

| 配置项 | 值 | 说明 |
|--------|-----|------|
| Orchestrator URL | `http://你的服务器IP:8080` | AB调度引擎地址 |
| Merchant ID | `0c1afe12-2a9c-4c4c-8250-5bf6b126a34b` | 在AB管理后台创建的商户ID |
| API Key | `（在管理后台生成）` | 商户API密钥 |
| B-Site Domain | `https://你的B站域名.com` | B站完整URL |
| Default Gateway | `paypal` | 默认支付网关 |

> **注意**：Orchestrator URL 必须使用 HTTP（不是 HTTPS），并且端口是 8080。

### 1.4 配置 WooCommerce 支付方式

A站后台 → WooCommerce → 设置 → 支付：

- **启用** "AB Payment Gateway"（标题显示为 "Pay with Card / PayPal"）
- **禁用** 所有其他支付方式（Direct Bank Transfer、Check Payments、Cash on Delivery 等全部关闭）

![支付方式设置：只保留 AB Payment Gateway](只留一个)

### 1.5 A站 SEO 和安全设置

**设置 → 阅读**：
- ✅ 勾选 "建议搜索引擎不索引本网站"

**.htaccess 或 Nginx 配置**：
```nginx
# Nginx
add_header Referrer-Policy "no-referrer" always;
add_header X-Robots-Tag "noindex, nofollow" always;
```

```apache
# Apache .htaccess
Header always set Referrer-Policy "no-referrer"
Header always set X-Robots-Tag "noindex, nofollow"
```

---

## 第二步：B站设置

### 2.1 安装 WordPress + WooCommerce

B站需要看起来像一个**真实的正常电商网站**。需要：

- 安装 WordPress + WooCommerce
- 选择普货风格主题（Storefront、Astra、Flatsome 等）
- **添加至少 10-20 个普货商品**（服装、家居、电子配件等，价格 $5-$200）
- 设置网站 Logo、favicon、页脚公司信息
- 添加 About Us、Contact、Shipping Policy、Return Policy 页面
- 安装一些常见插件增加真实性（Yoast SEO、Contact Form 7 等）

### 2.2 安装 PayPal 和 Stripe 官方插件

B站后台 → 插件 → 安装新插件 → 搜索安装：

| 插件 | 说明 |
|------|------|
| **WooCommerce PayPal Payments** | 官方 PayPal 支付插件 |
| **WooCommerce Stripe Payment Gateway** | 官方 Stripe 支付插件 |

### 2.3 配置 PayPal Payments

WooCommerce → 设置 → 支付 → PayPal → 管理：

1. 点击 "Activate PayPal" 按钮连接你的 PayPal Business 账号
2. 或手动输入 API 凭证：
   - **Client ID**：B站专用的 PayPal Client ID
   - **Secret Key**：B站专用的 PayPal Secret
   - **Sandbox 模式**：测试时开启，上线时关闭
3. **Webhook URL 留空**（由 AB 系统的 webhook 统一处理）

### 2.4 配置 Stripe

WooCommerce → 设置 → 支付 → Stripe → 管理：

1. 输入 Publishable Key 和 Secret Key
2. **Webhook Secret 留空**（由 AB 系统 webhook 统一处理）

### 2.5 安装 AB Payment Receiver 插件

```bash
# 把 ab-payment-receiver 目录上传到 B站
cp -r wordpress-plugins/ab-payment-receiver/ /var/www/b-site/wp-content/plugins/
```

```bash
# 激活
wp plugin activate ab-payment-receiver --path=/var/www/b-site
```

### 2.6 配置 AB Payment Receiver

B站后台 → 设置 → AB Receiver：

| 配置项 | 值 | 说明 |
|--------|-----|------|
| API Key | `（和AB调度引擎一致）` | 共享密钥 |
| Allowed IPs | `（AB调度引擎服务器IP）` | 留空允许所有 |
| PayPal Webhook ID | `（可选）` | 留空 |
| Tracking Sync Delay | `24` | 物流同步延迟（小时） |

### 2.7 B站 SEO 设置

**设置 → 阅读**：
- ❌ **不要**勾选 "建议搜索引擎不索引"（B站要对搜索引擎可见）

**设置 → 固定链接**：
- 选择 "文章名" 格式（让 URL 看起来自然）

### 2.8 B站 Nginx 安全头配置

```nginx
# 支付页面 CSP
location /pay/ {
    add_header Content-Security-Policy "default-src 'self'; 
      script-src 'self' 'unsafe-inline' https://www.paypal.com https://www.paypalobjects.com https://js.stripe.com; 
      frame-src https://www.paypal.com https://js.stripe.com https://hooks.stripe.com; 
      connect-src 'self' https://api.paypal.com https://api-m.paypal.com https://api.stripe.com; 
      img-src 'self' data: https://www.paypalobjects.com https://*.stripe.com;" always;
    
    add_header Cache-Control "no-store, no-cache, must-revalidate";
    add_header Referrer-Policy "no-referrer" always;
    add_header X-Robots-Tag "noindex, nofollow" always;
}
```

---

## 第三步：验证两个站点连接

### 3.1 验证 A站 → AB调度引擎 连通

```bash
# 在 A站服务器上执行
curl http://你的AB服务器IP:8080/health
# 应该返回 {"status":"healthy"}
```

### 3.2 验证 AB调度引擎 → B站 连通

```bash
# 在服务器上执行
curl -X POST http://你的AB服务器IP:8080/api/v1/allocate \
  -H 'Content-Type: application/json' \
  -d '{"order_id":"TEST-001","amount":"9.99","currency":"USD","merchant_id":"你的商户ID","gateway":"paypal"}'
# 应该返回 PayToken 和 GatewayURL
```

### 3.3 验证 B站 API 正常工作

```bash
curl https://你的B站域名/wp-json/ab-payment/v1/health \
  -H 'X-API-Key: 你的API密钥'
# 应该返回 {"status":"ok"}
```

---

## 第四步：端到端测试

1. 浏览器打开 A站 → 浏览商品 → 加入购物车 → 结算
2. 在结算页应该看到 "Pay with Card / PayPal" 支付方式
3. 填写收货信息，点击下单
4. 页面应该显示一个全屏支付 iframe（加载 B站支付页）
5. iframe 内加载 PayPal/Stripe 支付表单
6. 完成支付
7. 自动跳转回 A站订单完成页

---

## 常见问题排查

### A站结算页没有显示支付方式

1. 检查 AB Payment Bridge 插件是否已激活
2. 检查 WooCommerce → 设置 → 支付 → "AB Payment Gateway" 是否已启用
3. 检查 Orchestrator URL 配置是否正确

### 点击下单后报错

1. 打开浏览器 F12 → Network 标签
2. 点击下单，找到 `/api/v1/allocate` 请求
3. 查看 Response，根据错误信息排查：
   - `NO_ACCOUNT`：AB管理后台没有创建支付账号
   - `ACCOUNT_NOT_FOUND`：merchant_id 对应的账号不存在
   - `Connection refused`：AB调度引擎没有运行

### 支付 iframe 加载不出来

1. 检查 B站域名是否可访问（HTTPS 证书是否有效）
2. 检查 B站 `/pay/` 路由是否生效
3. 检查 B站 AB Payment Receiver 插件是否已激活

---

## 两个站点的对比清单

| 检查项 | A站 | B站 |
|--------|------|------|
| WooCommerce 已安装 | ✅ | ✅ |
| PayPal Payments 插件 | ❌ | ✅ |
| Stripe Gateway 插件 | ❌ | ✅ |
| AB Payment Bridge 插件 | ✅ | ❌ |
| AB Payment Receiver 插件 | ❌ | ✅ |
| 支付方式 | 只开启 AB Payment Gateway | 开启 PayPal + Stripe |
| 商品类型 | 任意 | 普货商品（10+个） |
| 搜索引擎可见 | ❌ noindex | ✅ |
| SSL (HTTPS) | ✅ | ✅ 必须 |
| 独立域名 | 是 | 是（与A站不同） |
| 独立服务器IP | 是 | 是（与A站不同） |

---

*最后更新: 2026-06-23*
