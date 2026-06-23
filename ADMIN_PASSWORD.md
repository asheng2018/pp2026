# 修改 admin 后台密码

## 方法一：在服务器上执行（推荐）

```bash
# 1. 生成新的 bcrypt 密码哈希（替换你想用的密码）
python3 -c "import bcrypt; print(bcrypt.hashpw(b'你的新密码', bcrypt.gensalt(rounds=10)).decode())"

# 2. 更新到数据库（注意：用 heredoc 防止 $ 被 shell 展开）
docker exec -i docker-postgres-1 psql -U abpay -d ab_payment <<SQL
UPDATE admin_users SET password_hash = '上面生成的哈希值' WHERE username = 'admin';
SQL
```

## 完整示例（改密码为 myNewPass456）

```bash
# 生成哈希
HASH=$(python3 -c "import bcrypt; print(bcrypt.hashpw(b'myNewPass456', bcrypt.gensalt(rounds=10)).decode())")
echo "New hash: $HASH"

# 更新数据库
docker exec -i docker-postgres-1 psql -U abpay -d ab_payment <<SQL
UPDATE admin_users SET password_hash = '$HASH' WHERE username = 'admin';
SQL

# 验证
curl -X POST http://localhost:8080/api/v1/admin/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"myNewPass456"}'
```

## 方法二：直接一行命令

```bash
docker exec -i docker-postgres-1 psql -U abpay -d ab_payment <<SQL
UPDATE admin_users SET password_hash = '\$2b\$10\$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy' WHERE username = 'admin';
SQL
```

> ⚠️ 上面的 hash 是 `admin123`，$ 需要转义 `\$`

## 创建新管理员

```bash
docker exec -i docker-postgres-1 psql -U abpay -d ab_payment <<SQL
INSERT INTO admin_users (username, password_hash, role) 
VALUES ('operator1', '新用户的bcrypt哈希', 'operator');
SQL
```
