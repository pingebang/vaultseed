#!/bin/bash

# VaultSeed 部署脚本
# 域名: tg.zhwenxing.cn

set -e

echo "🚀 开始部署 VaultSeed (tg.zhwenxing.cn)"

# 检查 Docker 和 Docker Compose
if ! command -v docker &> /dev/null; then
    echo "❌ Docker 未安装，请先安装 Docker"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose 未安装，请先安装 Docker Compose"
    exit 1
fi

# 创建必要的目录
echo "📁 创建必要的目录..."
mkdir -p letsencrypt backups

# 设置目录权限
echo "🔒 设置目录权限..."
chmod 755 letsencrypt
chmod 755 backups
touch backend/vaultseed.db
# 停止并删除现有容器
echo "🛑 停止现有服务..."
docker-compose -f docker-compose.prod.yml down || true

# 构建并启动服务
echo "🔨 构建和启动服务..."
docker-compose -f docker-compose.prod.yml up -d --build

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 10

# 检查服务状态
echo "🔍 检查服务状态..."
docker-compose -f docker-compose.prod.yml ps

# 显示访问信息
echo ""
echo "✅ 部署完成！"
echo ""
echo "📊 服务访问信息："
echo "   • 前端应用: https://tg.zhwenxing.cn"
echo "   • 后端API: https://tg.zhwenxing.cn/api"
echo "   • Traefik Dashboard: http://服务器IP:8080"
echo ""
echo "🔧 管理命令："
echo "   • 查看日志: docker-compose -f docker-compose.prod.yml logs -f"
echo "   • 停止服务: docker-compose -f docker-compose.prod.yml down"
echo "   • 重启服务: docker-compose -f docker-compose.prod.yml restart"
echo "   • 更新服务: ./deploy.sh"
echo ""
echo "📝 证书信息："
echo "   • Let's Encrypt 证书会自动申请和续期"
echo "   • 证书存储在: ./letsencrypt/acme.json"
echo ""
echo "⚠️  重要提示："
echo "   1. 确保域名 tg.zhwenxing.cn 已解析到服务器IP"
echo "   2. 服务器必须开放 80 和 443 端口"
echo "   3. 首次访问可能需要等待证书申请完成（约1-2分钟）"
echo "   4. 检查防火墙设置，确保端口可访问"

# 显示初始证书申请状态
echo ""
echo "📋 检查证书申请状态..."
docker-compose -f docker-compose.prod.yml logs traefik --tail=20 | grep -i "certificate\|acme\|tls" || true

echo ""
echo "🎉 部署脚本执行完成！"
