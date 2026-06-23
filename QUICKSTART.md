# AB Payment System — 快速部署指南

> **5分钟启动，一行命令部署**

---

## 最简方式：Docker 一键启动

### 前提

- 服务器已安装 Docker 和 Docker Compose
- 一台服务器即可（Linux，推荐 Ubuntu 22.04）

### 步骤1：准备项目

```bash
# 把 ab-payment-system 目录整个上传到服务器
scp -r ab-payment-system user@your-server:/opt/
cd /opt/ab-payment-system
```

### 步骤2：设置三个密码

```bash
# 直接写在命令行里，三行搞定
export DB_PASSWORD="你的数据库密码"
export REDIS_PASSWORD="你的Redis密码"
export JWT_SECRET="你的JWT密钥随便打一串字母数字"
```

### 步骤3：启动

```bash
docker-compose -f deploy/docker/docker-compose.yml up -d
```

### 完成！验证一下：

```bash
# 查状态，全绿就OK
docker-compose -f deploy/docker/docker-compose.yml ps

# 测一下
curl http://localhost:8080/health
# 返回 {"status":"healthy"} 就说明跑起来了
```

---

## 服务端口一览

启动后这些服务就都有了：

| 服务 | 地址 | 用途 |
|------|------|------|
| 调度引擎 | `http://你的服务器IP:8080` | A站插件连这个 |
| 支付网关 | `http://你的服务器IP:8081` | B站支付页 |
| 管理后台 | `http://你的服务器IP:3000` | 运营管理界面 |
| 数据库 | `localhost:5432` | 自动初始化 |
| Redis | `localhost:6379` | 缓存 |

---

## WordPress 插件安装（A站和B站）

### A站（仿牌站）：三步

```bash
# 1. 把插件复制过去
cp -r wordpress-plugins/ab-payment-bridge /var/www/仿牌站/wp-content/plugins/

# 2. 后台启用插件

# 3. 设置 → AB Payment → 填三样东西：
#    Orchestrator URL: http://你的服务器IP:8080
#    Merchant ID:      (管理后台创建商户后会得到)
#    API Key:          (管理后台生成)
#    B-Site Domain:    https://你的B站域名.com
```

### B站（普货站）：三步

```bash
# 1. 安装 PayPal 和 Stripe 官方插件
#    WordPress后台 → 插件 → 搜索安装：
#    - "WooCommerce PayPal Payments"
#    - "WooCommerce Stripe Payment Gateway"

# 2. 复制 AB 插件
cp -r wordpress-plugins/ab-payment-receiver /var/www/普货站/wp-content/plugins/

# 3. 后台启用 → 设置 → AB Receiver → 填写：
#    API Key: (和调度引擎用同一个)
```

---

## 首次使用：4步走通

### ① 创建管理员（数据库里直接插）

```bash
# 连上数据库
docker exec -it ab-postgres psql -U abpay -d ab_payment

# 创建管理员（密码是 admin123，上线后立刻改）
INSERT INTO admin_users (username, password_hash, role)
VALUES ('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'superadmin');
# 退出: \q
```

> 然后访问 `http://你的服务器IP:3000`，用 `admin / admin123` 登录

### ② 在管理后台添加支付账号

```
登录后台 → Accounts → Add Account →
  输入 PayPal Client ID 和 Secret
  设置限额（单笔最小$1，最大$5000，日限额$50000）
  保存
```

### ③ 在管理后台创建商户

```
Merchants → Create Merchant →
  填写商户名
  生成 API Key（复制下来！）
```

### ④ 把 API Key 填到 A站插件里

```
登录A站后台 → 设置 → AB Payment →
  粘贴刚刚复制的 API Key
  保存
```

---

## 全部完成，测试一下

1. 浏览器打开 A站 → 随便买个东西 → 结算
2. 应该看到弹出支付页面（PayPal/Stripe）
3. 完成支付
4. 管理后台 Orders 里能看到订单

---

## 常用命令

```bash
# 查看日志
docker-compose -f deploy/docker/docker-compose.yml logs -f

# 重启
docker-compose -f deploy/docker/docker-compose.yml restart

# 停止
docker-compose -f deploy/docker/docker-compose.yml down

# 完全重来（删数据）
docker-compose -f deploy/docker/docker-compose.yml down -v
docker-compose -f deploy/docker/docker-compose.yml up -d
```

---

## 没有 Docker？手动启动也可以

```bash
# 1. 装数据库
# 装个 PostgreSQL，创建数据库 ab_payment
# 执行 sql/schema.sql

# 2. 装 Redis
# 装好启动就行

# 3. 启动后端（两个终端窗口）
cd /opt/ab-payment-system

# 窗口1：调度引擎
export DB_HOST=localhost DB_USER=postgres DB_PASSWORD=你的密码 REDIS_ADDR=localhost:6379 JWT_SECRET=随便写
go run cmd/orchestrator/main.go

# 窗口2：支付网关
export DB_HOST=localhost DB_USER=postgres DB_PASSWORD=你的密码 REDIS_ADDR=localhost:6379 JWT_SECRET=随便写 SERVER_PORT=8081
go run cmd/gateway/main.go

# 4. 启动管理后台（第三个窗口）
cd web
npm install && npm run dev
```

---

## 常见问题

**Q: 启动报错连不上数据库？**
```bash
# 检查数据库是不是在运行
docker ps | grep postgres
# 检查密码对不对
docker exec -it ab-postgres psql -U abpay -d ab_payment
```

**Q: A站点了结算没反应？**
```bash
# 检查调度引擎能不能访问
curl http://你的服务器IP:8080/health
# 如果连不上，检查防火墙是否开放了 8080 端口
```

**Q: 支付页面显示不出来？**
```bash
# 确认B站域名的SSL证书正常
# 确认B站插件已启用并配置了 API Key
curl https://你的B站域名/wp-json/ab-payment/v1/health
```

**Q: 没有在线账号？**
```
管理后台 → Accounts → 找到你的账号 → 把状态改成 online
```

---

**一句话总结**：Docker 装好 → 三个环境变量 → `docker-compose up -d` → 搞定。
