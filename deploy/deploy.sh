#!/usr/bin/env bash
set -e

cd "$(dirname "$0")"

if ! command -v docker >/dev/null 2>&1; then
  echo "未检测到 docker，请先安装 Docker Engine 与 docker compose 插件。"
  exit 1
fi

echo "==> 构建并启动 Amitia (qdrant + surrealdb + backend + frontend)"
docker compose up -d --build

echo
echo "==> 当前状态"
docker compose ps

IP=$(hostname -I 2>/dev/null | awk '{print $1}')
echo
echo "部署完成。浏览器访问：  http://${IP:-<服务器IP>}:5178"
echo "首次访问会进入创建管理员向导；随后在设置里填入大模型 / Embedding 的 API Key。"
echo
echo "常用命令："
echo "  查看日志:   docker compose -f $(pwd)/docker-compose.yml logs -f backend"
echo "  停止:       docker compose -f $(pwd)/docker-compose.yml down"
echo "  彻底清数据: docker compose -f $(pwd)/docker-compose.yml down -v"
