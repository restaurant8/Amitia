# Amitia · Linux Docker 一键部署

一条命令在 Linux 服务器上跑起 Amitia(网页对话版)。大模型 / Embedding 走云端 API，**无需 GPU**。

## 架构（4 个容器）

| 容器 | 说明 | 镜像 |
|------|------|------|
| `qdrant` | 向量库（长期记忆） | 官方 `qdrant/qdrant` |
| `surrealdb` | 图数据库（知识图谱） | 官方 `surrealdb/surrealdb` |
| `backend` | Go 后端（本仓库源码编译） | 自建 |
| `frontend` | Vue 构建产物 + nginx（含 /api 反代、WebSocket） | 自建 |

后端以「外部引擎模式」运行（`SKIP_ENGINE_LAUNCH=1`），连接 qdrant / surrealdb 容器，不再自己拉起二进制。

## 前置条件

- Linux 服务器，装好 **Docker Engine** 与 **docker compose** 插件
- 最低配置：1C / 2G 内存 / 10G 磁盘（单会话网页对话）
- 出站网络可达你的大模型 / Embedding API

## 国内服务器网络（重要）

若构建时卡在 `TLS handshake timeout` 或拉不动镜像（`golang`/`node`/`qdrant`/`surrealdb` 都来自 Docker Hub），给 Docker 配置镜像加速器：

```bash
sudo mkdir -p /etc/docker
sudo tee /etc/docker/daemon.json >/dev/null <<'EOF'
{
  "registry-mirrors": ["https://docker.m.daocloud.io", "https://dockerproxy.com"]
}
EOF
sudo systemctl restart docker
```

（镜像源按你服务器实际可用的替换。）配置好再执行下面的一键部署。

## 一键部署

```bash
cd deploy
./deploy.sh
```

或手动：

```bash
cd deploy
docker compose up -d --build
```

完成后浏览器访问 **http://<服务器IP>:5178**：
1. 首次进入 → 创建管理员账号；
2. 登录后进「设置 / 模型配置」→ 填入大模型 API Key、Base URL、模型名，以及 Embedding 的 Key。

## 常用命令

```bash
docker compose logs -f backend      # 看后端日志
docker compose ps                   # 状态
docker compose down                 # 停止（保留数据）
docker compose down -v              # 停止并删除所有数据卷（重置）
docker compose up -d --build        # 改代码后重建
```

## 数据与备份

数据全部在命名卷里：`amitia_backend_data`（SQLite/密钥/媒体）、`amitia_qdrant_data`、`amitia_surreal_data`。备份用 `docker run --rm -v amitia_backend_data:/d -v $PWD:/b alpine tar czf /b/backend_data.tgz -C /d .`。

## 注意事项

- **向量维度**：`config.docker.yml` 里 `vectorDim: 1536`，必须与你所选 Embedding 模型输出维度一致。若用豆包 embedding（2560 维）等，需同步改 `vectorDim` 后 `down -v` 重建向量卷。
- **端口**：仅前端 `5178` 暴露到宿主；后端 8899、qdrant、surreal 只在内部网络。要放公网请自行加反向代理 + HTTPS。
- **JWT 密钥**：首次启动自动生成并持久化到数据卷（`data/jwt_secret.key`）。
- **微信 / QQ 桥接未包含在本 compose**（当前为网页对话版）。如需微信桥接，可另加一个 node 侧车容器跑 `release/sidecar/bundle.mjs`，并设置 `CORE_URL=http://backend:8899`、`BRIDGE_API_TOKEN`；服务器（无本机代理干扰）上 openclaw 的 `ilinkai` 长轮询通常可正常工作。
- 修改 `config.docker.yml` 后需 `docker compose up -d --build backend` 重建后端镜像。
