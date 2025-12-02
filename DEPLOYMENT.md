# VaultSeed 部署文档

VaultSeed 是一个端到端加密的内容存储应用，使用 Web3 钱包进行身份验证和加密。

## 系统架构

- **前端**: React + TypeScript + Tailwind CSS
- **后端**: Go (Gin) + SQLite
- **数据库**: SQLite（嵌入式）
- **加密**: AES-256-GCM + 钱包公钥加密

## 部署方式

### 1. 开发环境部署（HTTP）

#### 前提条件
- Docker 20.10+
- Docker Compose 2.0+

#### 部署步骤

1. **克隆项目**
   ```bash
   git clone <repository-url>
   cd vaultseed-backend
   ```

2. **构建并启动服务**
   ```bash
   docker-compose up -d
   ```

3. **验证服务状态**
   ```bash
   docker-compose ps
   ```

4. **查看日志**
   ```bash
   # 查看所有服务日志
   docker-compose logs -f
   
   # 查看特定服务日志
   docker-compose logs -f backend
   docker-compose logs -f frontend
   ```

5. **停止服务**
   ```bash
   docker-compose down
   ```

#### 服务访问
- 前端: http://localhost:8081
- 后端 API: http://localhost:8080/api
- 注意：开发环境使用HTTP，Web Crypto API可能受限

### 2. 生产环境部署（HTTPS + Traefik）

#### 前提条件
- Docker 20.10+
- Docker Compose 2.0+
- 域名 `tg.zhwenxing.cn` 已解析到服务器IP
- 服务器开放 80 和 443 端口

#### 快速部署脚本
```bash
# 1. 克隆项目
git clone <repository-url>
cd vaultseed-backend

# 2. 给部署脚本执行权限
chmod +x deploy.sh

# 3. 执行部署
./deploy.sh
```

#### 手动部署步骤

1. **创建必要目录**
   ```bash
   mkdir -p letsencrypt backups
   chmod 755 letsencrypt backups
   ```

2. **修改域名配置（如果需要）**
   ```bash
   # 编辑 docker-compose.prod.yml，将 tg.zhwenxing.cn 替换为你的域名
   sed -i 's/tg.zhwenxing.cn/your-domain.com/g' docker-compose.prod.yml
   sed -i 's/admin@zhwenxing.cn/your-email@example.com/g' docker-compose.prod.yml
   ```

3. **构建并启动服务**
   ```bash
   docker-compose -f docker-compose.prod.yml up -d --build
   ```

4. **验证服务状态**
   ```bash
   docker-compose -f docker-compose.prod.yml ps
   ```

5. **查看证书申请状态**
   ```bash
   docker-compose -f docker-compose.prod.yml logs traefik --tail=50 | grep -i "certificate\|acme\|tls"
   ```

#### 服务访问
- 前端应用: https://tg.zhwenxing.cn
- 后端 API: https://tg.zhwenxing.cn/api
- Traefik Dashboard: http://服务器IP:8080
- Let's Encrypt 证书自动申请和续期

#### 生产环境特点
- 自动HTTPS重定向（HTTP → HTTPS）
- Let's Encrypt SSL证书自动管理
- Traefik反向代理和负载均衡
- 自动备份数据库（每天备份，保留7天）
- 安全上下文：Web Crypto API完全支持

### 2. 手动构建 Docker 镜像

#### 构建后端镜像
```bash
cd backend
docker build -t vaultseed-backend:latest .
```

#### 构建前端镜像
```bash
cd frontend
docker build -t vaultseed-frontend:latest .
```

#### 运行容器
```bash
# 运行后端
docker run -d -p 8080:8080 --name vaultseed-backend vaultseed-backend:latest

# 运行前端
docker run -d -p 80:80 --name vaultseed-frontend vaultseed-frontend:latest
```

### 3. 源码部署

#### 后端部署
```bash
cd backend

# 安装依赖
go mod download

# 构建
go build -o vaultseed-backend ./cmd/main.go

# 运行
./vaultseed-backend
```

#### 前端部署
```bash
cd frontend

# 安装依赖
npm install

# 构建
npm run build

# 使用 nginx 或任何静态文件服务器
# 例如使用 serve
npx serve -s build -l 3000
```

## 环境配置

### 后端环境变量
```bash
# 生产模式
GIN_MODE=release

# 数据库路径（默认：vaultseed.db）
DB_PATH=/app/vaultseed.db

# 服务器端口（默认：8080）
PORT=8080

# CORS 允许的域名
CORS_ALLOW_ORIGIN=http://localhost:80
```

### 前端环境变量
```bash
# API 地址（默认：http://localhost:8080/api）
REACT_APP_API_URL=http://your-domain.com/api

# 钱包连接配置
REACT_APP_CHAIN_ID=1
REACT_APP_NETWORK_NAME=mainnet
```

## 数据库管理

### SQLite 数据库
- 位置: `backend/vaultseed.db`
- 自动迁移: 应用启动时自动创建表结构
- 备份: 建议定期备份数据库文件

### 数据库初始化
应用首次启动时会自动创建以下表：
- `users` - 用户表（钱包地址、公钥、nonce）
- `encrypted_contents` - 加密内容表

## 安全配置

### HTTPS 配置
1. **获取 SSL 证书**
   ```bash
   # 使用 Let's Encrypt
   certbot certonly --standalone -d your-domain.com
   ```

2. **更新 nginx 配置**
   修改 `frontend/nginx.conf` 添加 SSL 配置：
   ```nginx
   server {
       listen 443 ssl http2;
       server_name your-domain.com;
       
       ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
       ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;
       
       # ... 其他配置
   }
   ```

### 防火墙配置
```bash
# 开放必要端口
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 8080/tcp
sudo ufw enable
```

## 监控和维护

### 健康检查
- 后端健康检查: `GET http://localhost:8080/health`
- 前端健康检查: `GET http://localhost:80`

### 日志管理
```bash
# 查看 Docker 容器日志
docker-compose logs --tail=100 -f

# 日志轮转配置
# 在 /etc/docker/daemon.json 中添加
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
```

### 备份策略
```bash
# 备份数据库
cp backend/vaultseed.db vaultseed-backup-$(date +%Y%m%d).db

# 备份 Docker 卷
docker run --rm -v vaultseed_backend_data:/data -v $(pwd):/backup alpine tar czf /backup/backup-$(date +%Y%m%d).tar.gz /data
```

## 故障排除

### 常见问题

1. **端口冲突**
   ```bash
   # 检查端口占用
   sudo netstat -tulpn | grep :80
   sudo netstat -tulpn | grep :8080
   
   # 停止占用进程或修改端口
   ```

2. **数据库权限问题**
   ```bash
   # 确保数据库文件可写
   chmod 666 backend/vaultseed.db
   ```

3. **Docker 构建失败**
   ```bash
   # 清理 Docker 缓存
   docker system prune -a
   
   # 重新构建
   docker-compose build --no-cache
   ```

4. **前端无法连接后端**
   - 检查 `REACT_APP_API_URL` 配置
   - 检查 CORS 配置
   - 检查网络连通性

### 性能优化

1. **数据库优化**
   ```sql
   -- 创建索引
   CREATE INDEX idx_user_address ON encrypted_contents(user_address);
   CREATE INDEX idx_created_at ON encrypted_contents(created_at);
   ```

2. **Nginx 优化**
   ```nginx
   # 增加 worker 进程
   worker_processes auto;
   
   # 调整缓冲区大小
   client_body_buffer_size 10K;
   client_header_buffer_size 1k;
   client_max_body_size 8m;
   large_client_header_buffers 2 1k;
   ```

## 更新部署

### 滚动更新
```bash
# 拉取最新代码
git pull origin main

# 重新构建并重启
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

### 版本回滚
```bash
# 回滚到特定版本
git checkout v1.0.0
docker-compose down
docker-compose up -d
```

## 联系支持

如有问题，请参考：
- 项目文档: `README.md`
- API 文档: `http://your-domain.com:8080/swagger/index.html`
- 问题追踪: GitHub Issues

---

**安全提示**: 请妥善保管数据库备份和 SSL 证书私钥，定期更新依赖包以修复安全漏洞。
