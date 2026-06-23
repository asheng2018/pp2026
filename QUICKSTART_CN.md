# AB Payment System — 入门教程

## 架构概览

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│    A站 (仿牌站)   │────▶│  AB Orchestrator │────▶│   B站 (普货站)   │
│  WooCommerce     │     │   port 8080      │     │  WooCommerce     │
│  + Bridge 插件   │     │  (支付网关+管理)   │     │  + Receiver 插件 │
└─────────────────┘     └──────────────────┘     └─────────────────┘
        │                        │                        │
        │ 1. 用户下单             │                        │
        │ 2. POST /allocate      │                        │
        │                        │ 3. 轮询选择账号+代理    │
        │                        │ 4. 生成 pay_token     │
        │ 5. iframe 跳转         │                        │
        │──────────────────────────────────────────────▶│
        │                        │                        │ 6. 渲染 PayPal/Stripe
        │◀──────────────────────────────────────────────│
        │ 7. postMessage 回调     │                        │ 8. 用户付款
        │ 9. 跳转 checkout/success                        │
```

## 第一步：登录管理后台

浏览器打开 `http://你的服务器IP:8088`，用 `admin` / `admin123` 登录。

---

## 第二步：添加 B-Site（钱款接收站）

B-Site 是你的**普货站**，上面部署了真实 WooCommerce + PayPal/Stripe 支付插件。

1. 点击左侧菜单 **B-Sites**
2. 点击右上角 **Add B-Site**
3. 填写表单：

| 字段 | 说明 | 示例 |
|------|------|------|
| Domain | B站域名 | `goodstore.shop` |
| Store Name | 店铺名称 | `Good Store` |
| Hosting IP | B站服务器 IP | `23.45.67.89` |
| Hosting Provider | 托管商 | `Cloudflare` |
| WooCommerce URL | B站网址（完整） | `https://goodstore.shop` |
| WC Consumer Key | WooCommerce REST API Key | `ck_xxxx` |
| WC Consumer Secret | WooCommerce REST API Secret | `cs_xxxx` |
| SSL Provider | SSL 证书提供商 | `Let's Encrypt` |
| SSL Expires | 到期日 | `2026-12-31` |

4. 点击 **Create**

### B 站前置条件

确保 B 站已经：
- ✅ 安装了 WordPress + WooCommerce
- ✅ 安装了 PayPal Payments 或 Stripe 插件（B站后台配置真实商户）
- ✅ 安装了 **AB Payment Receiver** 插件（`wordpress-plugins/ab-payment-receiver/`）
- ✅ 启用 HTTPS（必须！）

#### 安装 Receiver 插件

```bash
# 把插件文件夹上传到 B站
scp -r wordpress-plugins/ab-payment-receiver/ user@b-site-host:/var/www/html/wp-content/plugins/
```

然后在 B站 WordPress 后台 **Plugins** → 激活 **AB Payment Receiver**。  
进入 **Settings → AB Receiver**，配置一个 API Key（自己设一个长随机字符串，后面会用到）。

---

## 第三步：添加 Payment Account（收款号）

1. 点击 **Accounts** → **Add Account**
2. 填写：

| 字段 | 说明 | 示例 |
|------|------|------|
| Gateway | 选择 paypal 或 stripe | `paypal` |
| Alias | 账号别名，方便识别 | `PP-US-01` |
| Merchant ID | 所属商户 ID（可选） | 复制商户的 UUID |
| B-Site ID | 绑定的 B 站 ID（可选） | 复制 B-Site 的 UUID |
| Weight | 权重，越大分配越多 | `100` |
| Priority | 优先级，越大越优先 | `10` |

**PayPal 模式：**
| 字段 | 说明 |
|------|------|
| PayPal Client ID | PayPal REST API Client ID |
| PayPal Secret | PayPal REST API Secret |

**Stripe 模式：**
| 字段 | 说明 |
|------|------|
| Stripe Publishable Key | `pk_live_...` |
| Stripe Secret Key | `sk_live_...` |

**限额设置：**
| 字段 | 说明 | 建议值 |
|------|------|--------|
| Single Min | 单笔最小金额 | `1` |
| Single Max | 单笔最大金额 | `5000` |
| Daily Max | 单日最大金额 | `50000` |
| Monthly Max | 单月最大金额 | `500000` |

3. 点击 **Create**

可以添加多个账号实现**轮询收款**。

---

## 第四步：创建 Merchant（商户/通道商）

1. 点击 **Merchants** → **Add Merchant**
2. 填写：

| 字段 | 说明 |
|------|------|
| Merchant Name | 商户名称 |
| Email | 商户邮箱 |
| Routing Mode | 路由策略（推荐 Weighted Round Robin） |
| Fee Rate (%) | 向商户收取的费率（你自己赚的差价） |

3. 点击 **Create**
4. 在列表中找到新建的商户，点击 **API Key** 按钮
5. **立刻复制生成的 API Key！（只显示一次）**

这个 API Key 要给到 A 站的 Bridge 插件使用。

---

## 第五步：A 站安装 Bridge 插件

A 站是你的**仿牌站**，用户在这里浏览和下单，但不会直接看到支付页面。

### 前提条件
- ✅ WordPress + WooCommerce
- ✅ 启用 HTTPS

### 安装

```bash
# 把插件文件夹上传到 A站
scp -r wordpress-plugins/ab-payment-bridge/ user@a-site-host:/var/www/html/wp-content/plugins/
```

在 A 站 WordPress 后台激活 **AB Payment Bridge**，然后进入 **Settings → AB Payment**：

| 设置 | 说明 | 示例 |
|------|------|------|
| Orchestrator URL | 网关服务器地址 | `http://你的服务器IP:8080` |
| Merchant ID | 商户 UUID | 从管理后台复制 |
| API Key | 上一步生成的 Key | 粘贴 |
| B-Site Domain | B 站完整域名 | `https://goodstore.shop` |
| Default Gateway | paypal 或 stripe | `paypal` |

---

## 第六步：测试第一笔订单

### 6.1 直接 curl 测试支付分配

```bash
curl -X POST http://localhost:8080/api/v1/allocate \
  -H 'Content-Type: application/json' \
  -d '{
    "order_id": "TEST-001",
    "amount": "9.99",
    "currency": "USD",
    "merchant_id": "你的MERCHANT_UUID",
    "gateway": "paypal",
    "strategy": "weighted_round_robin"
  }'
```

成功返回：
```json
{
  "pay_token": "xxx...",
  "gateway_url": "/pay/xxx...",
  "account_ref": "2dadbe85",
  "expires_at": "2026-06-21T11:00:00Z",
  "gateway": "paypal"
}
```

### 6.2 浏览器模拟完整流程

1. 用浏览器打开 B 站的支付页面：
   ```
   http://你的服务器IP:8081/pay/{上面返回的pay_token}
   ```

2. 应该能看到 PayPal/Stripe 支付按钮

3. 点击付款（沙盒模式测试）

### 6.3 A 站真实下单测试

1. 打开 A 站，随便加一个商品到购物车
2. 去结算页面
3. 应该看到唯一支付方式 **AB Payment Gateway**
4. 点击下单 → 会在 iframe 中加载 B 站的支付页面
5. 完成付款 → 自动跳转回 success 页

---

## 第七步：配置代理池（反追踪）

1. 点击 **Proxy Pool** → **Batch Import**
2. 粘贴代理列表，每行一个，格式：`host:port:username:password`

```
res-proxy1.example.com:1234:user1:pass1
res-proxy2.example.com:5678:user2:pass2
185.199.100.10:9050:user3:pass3
```

3. 点击 **Import**
4. 每个代理初始状态为 `testing`，手动改为 `online` 后即可参与分配

代理分配 API：
```bash
curl -X POST http://localhost:8080/api/v1/proxy-allocate \
  -H 'Content-Type: application/json' \
  -d '{"gateway":"paypal","merchant_id":"你的MERCHANT_UUID"}'
```

---

## 第八步：监控与运维

### 查看 Dashboard
`http://你的服务器IP:8088` → Dashboard，实时显示：
- 今日订单数 + 收入
- 活跃账号数 + 商户数
- 成功率
- 7 天收入趋势曲线

### 查看 Grafana
`http://你的服务器IP:3000`，默认账号 `admin/admin`

### 日常运营
| 任务 | 操作 |
|------|------|
| 新增收款号 | Accounts → Add Account |
| 某号风控被封 | 改为 offline，自动轮询跳过 |
| 额度即将触顶 | 改 Daily Max 或切换账号 |
| B 站 SSL 快过期 | B-Sites → Edit → 更新 SSL 信息 |
| 代理 IP 挂了 | Proxy Pool → 改为 offline |
| 商户要 API Key | Merchants → API Key → 生成 → 发给对方 |

### 风控建议
- 每个 PayPal/Stripe 号**每天不超过 50 笔**，$50000
- 每个 B 站域名**每周换一个**（或准备好多个备用）
- 代理 IP **必须绑定账号**，做到一号一 IP
- 物流单号**延迟 24-48 小时同步**到 B 站（Receiver 插件设置）
