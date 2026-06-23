# Ubuntu 24 服务器 — AB支付系统部署指南

> 两种方案：方案一纯 Docker，方案二用 1Panel 面板

---

## 方案一：Docker + Docker Compose 原生部署

### 1.1 全新 Ubuntu 24 一键装 Docker

```bash
# SSH 登录服务器后，复制整段执行
# 全程自动，不需要任何手动干预

# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装依赖
sudo apt install -y ca-certificates curl gnupg lsb-release

# 添加 Docker 官方 GPG 密钥
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

# 添加 Docker 源
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# 安装 Docker
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# 验证安装
docker --version
docker compose version

# 把当前用户加入 docker 组（不用每次 sudo）
sudo usermod -aG docker $USER

# 设置 Docker 开机自启
sudo systemctl enable docker

# 退出 SSH 重新登录，使 docker 组权限生效
exit
```

### 1.2 重新登录后，部署 AB 系统

```bash
# SSH 重新连上服务器

# 上传项目到服务器
# 在本地电脑执行（不是在服务器上）：
# scp -r "e:/AI code/pp/ab-payment-system" user@你的服务器IP:/opt/

# 在服务器上：
cd /opt/ab-payment-system

# 设置三个密码
export DB_PASSWORD="自己设一个复杂密码"
export REDIS_PASSWORD="自己设另一个密码"
export JWT_SECRET="$(openssl rand -hex 32)"

# 启动！一行命令
docker compose -f deploy/docker/docker-compose.yml up -d

# 查看启动状态（应该全部是 Up/running）
docker compose -f deploy/docker/docker-compose.yml ps

# 验证
curl http://localhost:8080/health
```

### 1.3 开放端口

```bash
# Ubuntu 24 默认用 ufw 防火墙
sudo ufw allow 8080/tcp   # 调度引擎 API
sudo ufw allow 3000/tcp   # 管理后台
sudo ufw allow 443/tcp    # 如果 WordPress 也在这台服务器上

# 查看状态
sudo ufw status
```

### 1.4 完成

```
调度引擎:  http://服务器IP:8080
管理后台:  http://服务器IP:3000
支付网关:  http://服务器IP:8081
```

---

## 方案二：用 1Panel 面板部署（更简单，有图形界面）

### 2.1 安装 1Panel

```bash
# SSH 登录服务器，一行命令安装 1Panel
curl -sSL https://resource.fit2cloud.com/1panel/package/quick_start.sh -o quick_start.sh && sudo bash quick_start.sh

# 安装过程中会提示：
# 1. 选择安装目录 → 直接回车（默认 /opt）
# 2. 选择端口 → 直接回车（默认 10086）
# 3. 设置面板密码 → 输入你的密码

# 安装完成后会显示：
# ========================================
# 面板地址: http://你的IP:10086/安全入口码
# 用户名: admin
# 密码: 你设置的密码
# ========================================

# 记下这三样东西！
```

> 1Panel 会自动安装 Docker 和 Docker Compose，不用手动装

### 2.2 登录 1Panel，安装运行环境

```
1. 浏览器打开 http://你的服务器IP:10086/安全入口码
2. 输入用户名 admin 和密码

3. 左侧菜单 → 应用商店 → 搜索安装：
   - PostgreSQL（版本选 16）
   - Redis（版本选 7）
   - OpenResty（用作反向代理，自动装好 Nginx）
```

### 2.3 通过 1Panel 部署 AB 系统

#### 第一步：上传项目

```bash
# 在服务器上
mkdir -p /opt/ab-payment-system

# 从你的电脑上传项目文件到服务器 /opt/ab-payment-system/
# 可以用 1Panel 自带的文件管理器上传
# 或者用 scp 命令
```

#### 第二步：修改 docker-compose.yml

把 `deploy/docker/docker-compose.yml` 里自带的 postgres/redis/nginx 删掉（因为 1Panel 已经装了），只保留 orchestrator 和 gateway 服务。创建一个新的精简版：

**`deploy/docker/docker-compose-1panel.yml`**：

```yaml
version: '3.8'

services:
  orchestrator:
    build:
      context: ../..
      dockerfile: deploy/docker/Dockerfile.orchestrator
    container_name: ab-orchestrator
    environment:
      ENV: production
      SERVER_PORT: "8080"
      GRPC_PORT: "9090"
      DB_HOST: ${DB_HOST:-postgres}
      DB_PORT: ${DB_PORT:-5432}
      DB_USER: ${DB_USER:-abpay}
      DB_PASSWORD: ${DB_PASSWORD}
      DB_NAME: ${DB_NAME:-ab_payment}
      DB_SSLMODE: disable
      REDIS_ADDR: ${REDIS_ADDR:-redis:6379}
      REDIS_PASSWORD: ${REDIS_PASSWORD}
      NATS_URL: ${NATS_URL:-nats://nats:4222}
      JWT_SECRET: ${JWT_SECRET}
    ports:
      - "8080:8080"
      - "9090:9090"
    restart: unless-stopped
    networks:
      - ab-network

  gateway:
    build:
      context: ../..
      dockerfile: deploy/docker/Dockerfile.gateway
    container_name: ab-gateway
    environment:
      ENV: production
      SERVER_PORT: "8080"
      DB_HOST: ${DB_HOST:-postgres}
      DB_PORT: ${DB_PORT:-5432}
      DB_USER: ${DB_USER:-abpay}
      DB_PASSWORD: ${DB_PASSWORD}
      DB_NAME: ${DB_NAME:-ab_payment}
      DB_SSLMODE: disable
      REDIS_ADDR: ${REDIS_ADDR:-redis:6379}
      REDIS_PASSWORD: ${REDIS_PASSWORD}
      JWT_SECRET: ${JWT_SECRET}
    ports:
      - "8081:8080"
    restart: unless-stopped
    networks:
      - ab-network

networks:
  ab-network:
    external: true
```

> ⚠️ 注意：需要把 `DB_HOST` 和 `REDIS_ADDR` 改成 1Panel 安装的 PostgreSQL 和 Redis 的容器名或 IP。在 1Panel → 数据库 里可以查看连接地址。

#### 第三步：创建数据库

```
1Panel → 数据库 → PostgreSQL → 创建数据库
  - 数据库名: ab_payment
  - 用户名: abpay
  - 密码: 设置一个密码

2. 导入表结构：
  1Panel → 数据库 → ab_payment → SQL 执行
  把 sql/schema.sql 的内容粘贴进去，点执行
```

#### 第四步：构建并启动

```bash
cd /opt/ab-payment-system

# 构建镜像
docker compose -f deploy/docker/docker-compose-1panel.yml build

# 启动
docker compose -f deploy/docker/docker-compose-1panel.yml up -d

# 查看状态
docker ps | grep ab-
```

#### 第五步：用 OpenResty 配置反向代理和域名

```
1Panel → 网站 → 创建网站 →
  - 类型: 反向代理
  - 域名: admin.你的域名.com
  - 代理地址: http://127.0.0.1:3000 （管理后台）

再创建一个：
  - 域名: api.你的域名.com
  - 代理地址: http://127.0.0.1:8080 （调度引擎 API）

1Panel 会自动帮你申请 SSL 证书并配置 HTTPS！
```

---

## 方案三：最最最简单的方式（纯 1Panel + 已安装环境）

如果你已经用 1Panel 装好了 PostgreSQL 和 Redis，最省事的方式：

### 直接命令行启动 Go 程序

```bash
cd /opt/ab-payment-system

# 设置环境变量（改成你 1Panel 里的实际值）
export DB_HOST=127.0.0.1
export DB_PORT=5432
export DB_USER=abpay
export DB_PASSWORD=1panel里设置的密码
export DB_NAME=ab_payment
export REDIS_ADDR=127.0.0.1:6379
export REDIS_PASSWORD=1panel里设置的Redis密码
export JWT_SECRET=$(openssl rand -hex 32)

# 启动调度引擎（后台运行）
nohup go run cmd/orchestrator/main.go > /var/log/ab-orchestrator.log 2>&1 &

# 启动支付网关（后台运行，用不同端口）
export SERVER_PORT=8081
nohup go run cmd/gateway/main.go > /var/log/ab-gateway.log 2>&1 &

# 验证
curl http://localhost:8080/health
```

> 需要先装 Go：`sudo snap install go --classic`

---

## 三种方案对比

| 方案 | 难度 | 时间 | 适合场景 |
|------|------|------|----------|
| **方案一 Docker原生** | ⭐⭐ | 10分钟 | 纯命令行，一台服务器全搞定 |
| **方案二 1Panel+Docker** | ⭐ | 5分钟 | 有面板，图形化操作方便 |
| **方案三 1Panel+裸Go** | ⭐ | 3分钟 | 最轻量，不需要编译Docker镜像 |

---

## 不管哪种方案，最后都需要做的

### 1. 初始化数据库

```bash
# 连接 PostgreSQL
psql -h 127.0.0.1 -U abpay -d ab_payment -f sql/schema.sql
```

### 2. 创建管理员账号

```sql
-- 连上数据库后执行
INSERT INTO admin_users (username, password_hash, role)
VALUES ('admin', crypt('你的密码', gen_salt('bf')), 'superadmin');
```

> 然后访问 `http://服务器IP:3000` 或 `https://admin.你的域名.com` 登录后台

### 3. 装 WordPress 插件

按 `wordpress-plugins/README.md` 里的说明操作：
- A站 → 装 `ab-payment-bridge`
- B站 → 装 `ab-payment-receiver` + PayPal/Stripe 官方插件

### 4. 管理后台创建商户 → 生成 API Key → 填到 A站插件

---

## 推荐

**如果追求快**：用方案一，Docker 原生，10 分钟搞定。

**如果长期运营**：用方案二，1Panel 面板管理方便，自动备份、SSL 续签、日志查看都有图形界面。

**如果服务器配置低**（1核2G）：用方案三直接跑 Go 二进制，省资源。
