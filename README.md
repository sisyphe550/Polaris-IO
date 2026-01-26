# Polaris-IO (北极星) 分布式云存储平台

一个基于 Go-Zero 微服务架构的企业级私有云盘系统，支持 PB 级数据存储、高并发上传下载。

> 本项目参考了 [go-zero-looklook](https://github.com/Mikaelemmmm/go-zero-looklook.git) 最佳实践，学习其全栈微服务架构设计与工程化规范，并在此基础上实现了分布式云存储业务场景。

## 核心特性

- **文件秒传**：基于 SHA256 哈希的秒传机制，秒传率 30%+
- **大文件支持**：支持 1GB+ 大文件上传，前端直传 S3
- **高性能**：核心接口 P99 延迟 < 200ms
- **文件分享**：支持提取码、过期时间、转存功能
- **全文搜索**：基于 Elasticsearch 的文件名搜索
- **异步任务**：基于 Asynq 的延迟队列（回收站清理、配额退还等）

## 技术栈

| 类别 | 技术 |
|------|------|
| 微服务框架 | Go-Zero |
| 关系数据库 | MySQL 8.0 |
| 文档数据库 | MongoDB 7.0 |
| 缓存 | Redis 7.0 |
| 搜索引擎 | Elasticsearch 8.x |
| 消息队列 | Kafka |
| 对象存储 | Garage (S3 兼容) |
| 延迟队列 | Asynq (Redis) |
| 链路追踪 | Jaeger |
| 监控 | Prometheus + Grafana |

## 系统架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         Docker (polaris-net)                            │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐      │
│  │  Nginx   │ │  MySQL   │ │  Redis   │ │  Kafka   │ │ Garage   │ ...  │
│  │  :8888   │ │  :3306   │ │  :6379   │ │  :9092   │ │  :3900   │      │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘      │
└───────┼────────────┼────────────┼────────────┼────────────┼────────────┘
        │            │ 33069      │ 36379      │ 39092      │ 33900
        │ host.docker│            │            │            │
        │ .internal  ▼            ▼            ▼            ▼
┌───────┼─────────────────────────────────────────────────────────────────┐
│       ▼                    宿主机 (开发模式)                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │ user-api │  │ file-api │  │share-api │  │search-api│  │search-job│  │
│  │  :1001   │  │  :1002   │  │  :1003   │  │  :1004   │  │          │  │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘  └──────────┘  │
│       │             │             │             │                       │
│  ┌────┴─────┐  ┌────┴─────┐  ┌────┴─────┐  ┌────┴─────┐  ┌──────────┐  │
│  │ user-rpc │  │ file-rpc │  │share-rpc │  │search-rpc│  │mqueue-job│  │
│  │  :2001   │  │  :2002   │  │  :2003   │  │  :2004   │  │          │  │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘  └──────────┘  │
└─────────────────────────────────────────────────────────────────────────┘
```

## 服务端口

### 后端服务

| 服务 | 类型 | 端口 | 说明 |
|------|------|------|------|
| user-api | HTTP | 1001 | 用户认证、配额管理 |
| user-rpc | gRPC | 2001 | 用户服务 RPC |
| file-api | HTTP | 1002 | 文件上传下载、目录管理 |
| file-rpc | gRPC | 2002 | 文件服务 RPC |
| share-api | HTTP | 1003 | 分享管理 |
| share-rpc | gRPC | 2003 | 分享服务 RPC |
| search-api | HTTP | 1004 | 文件搜索 |
| search-rpc | gRPC | 2004 | 搜索服务 RPC |
| search-job | Job | - | Kafka 消费者，同步 ES 索引 |
| mqueue-job | Job | - | Asynq 消费者，异步任务处理 |

### 基础设施 (Docker 端口映射)

| 服务 | 容器内端口 | 宿主机端口 | 说明 |
|------|-----------|-----------|------|
| MySQL | 3306 | 33069 | 关系数据库 |
| MongoDB | 27017 | 37017 | 文档数据库 |
| Redis | 6379 | 36379 | 缓存 & Asynq |
| Elasticsearch | 9200 | 39200 | 搜索引擎 |
| Kafka | 9092 | 39092 | 消息队列 |
| Garage S3 | 3900 | 33900 | 对象存储 |
| Jaeger | 16686 | 16686 | 链路追踪 UI |
| Prometheus | 9090 | 9090 | 监控 |
| Grafana | 3000 | 3000 | 监控面板 |

---

## 开发模式启动

开发模式下，Go 服务直接在宿主机运行，通过端口映射连接 Docker 内的基础设施。

### 1. 启动基础设施

```bash
cd backend

# 创建 Docker 网络（首次需要）
docker network create polaris-net

# 启动基础设施容器
docker compose -f docker-compose-env.yml up -d
```

这将启动：MySQL、MongoDB、Redis、Elasticsearch、Kafka、Garage S3、Jaeger、Prometheus、Grafana 等。

### 2. 配置 Kafka Topic（首次需要）

```bash
# 进入 Kafka 容器
docker exec -it io-kafka bash

# 创建文件事件 Topic
/opt/kafka/bin/kafka-topics.sh --create \
  --topic polaris-file-event \
  --partitions 3 \
  --replication-factor 1 \
  --bootstrap-server localhost:9092

# 验证 Topic 创建成功
/opt/kafka/bin/kafka-topics.sh --list --bootstrap-server localhost:9092

# 退出容器
exit
```

### 3. 配置 Garage S3（首次需要）

Garage 是 S3 兼容的对象存储，首次启动需要初始化集群和创建 Bucket。

```bash
# 1. 获取节点 ID
docker exec io-storage /garage status

# 2. 配置节点布局（将 <NODE_ID> 替换为上一步获取的 ID）
docker exec io-storage /garage layout assign -z dc1 -c 1G <NODE_ID>
docker exec io-storage /garage layout apply --version 1

# 3. 创建 API Key
docker exec io-storage /garage key create polaris-app-key

# 输出示例：
# Key ID: GK...
# Secret key: ...
# 记录这两个值，稍后配置到 file.yaml

# 4. 创建 Bucket
docker exec io-storage /garage bucket create polaris-files

# 5. 授权 Key 访问 Bucket
docker exec io-storage /garage bucket allow polaris-files --read --write --key polaris-app-key

# 6. 验证配置
docker exec io-storage /garage bucket info polaris-files
```

将获取的 Key ID 和 Secret Key 配置到 `app/file/cmd/rpc/etc/file.yaml`：

```yaml
S3:
  Endpoint: "http://127.0.0.1:3900"
  AccessKey: "GK..."        # 上面获取的 Key ID
  SecretKey: "..."          # 上面获取的 Secret key
  Bucket: "polaris-files"
  Region: "us-east-1"
```

### 4. 验证基础设施

```bash
# 检查容器状态
docker ps

# 验证 MySQL 连接
mysql -h 127.0.0.1 -P 33069 -u root -proot -e "SHOW DATABASES;"

# 验证 Redis 连接
redis-cli -h 127.0.0.1 -p 36379 -a polaris PING

# 验证 Kafka
docker exec io-kafka /opt/kafka/bin/kafka-topics.sh --list --bootstrap-server localhost:9092

# 验证 Garage S3
curl http://127.0.0.1:3900
```

### 5. 启动后端服务

需要按依赖顺序启动：**RPC 服务必须先于 API 服务启动**。

```bash
cd backend

# ===== 终端 1: User RPC =====
go run app/user/cmd/rpc/usercenter.go -f app/user/cmd/rpc/etc/usercenter.yaml

# ===== 终端 2: User API =====
go run app/user/cmd/api/usercenter.go -f app/user/cmd/api/etc/usercenter.yaml

# ===== 终端 3: File RPC =====
go run app/file/cmd/rpc/file.go -f app/file/cmd/rpc/etc/file.yaml

# ===== 终端 4: File API =====
go run app/file/cmd/api/file.go -f app/file/cmd/api/etc/file.yaml

# ===== 终端 5: Share RPC =====
go run app/share/cmd/rpc/share.go -f app/share/cmd/rpc/etc/share.yaml

# ===== 终端 6: Share API =====
go run app/share/cmd/api/share.go -f app/share/cmd/api/etc/share.yaml

# ===== 终端 7: Search RPC =====
go run app/search/cmd/rpc/search.go -f app/search/cmd/rpc/etc/search.yaml

# ===== 终端 8: Search API =====
go run app/search/cmd/api/search.go -f app/search/cmd/api/etc/search.yaml

# ===== 终端 9: Search Job (Kafka Consumer) =====
go run app/search/cmd/job/search-job.go -f app/search/cmd/job/etc/search-job.yaml

# ===== 终端 10: Mqueue Job (Asynq Consumer) =====
go run app/mqueue/cmd/job/mqueue.go -f app/mqueue/cmd/job/etc/mqueue.yaml
```

### 6. 验证服务

```bash
# 用户注册
curl -X POST http://127.0.0.1:1001/usercenter/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{"mobile":"13800138000","password":"123456","name":"test"}'

# 返回示例：{"code":200,"msg":"OK","data":{"accessToken":"xxx","accessExpire":1234567890,"refreshAfter":1234567890}}
# 或已注册：{"code":200002,"msg":"该手机号已注册"}

# 用户登录
curl -X POST http://127.0.0.1:1001/usercenter/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{"mobile":"13800138000","password":"123456"}'

# 返回示例：{"code":200,"msg":"OK","data":{"accessToken":"eyJhbGc...","accessExpire":1234567890,"refreshAfter":1234567890}}

# 使用 Token 获取用户信息
TOKEN="上一步返回的 accessToken"
curl -H "Authorization: $TOKEN" http://127.0.0.1:1001/usercenter/v1/user/info
```

---

## 生产部署

生产环境下，所有服务（包括 Go 后端）都运行在 Docker 容器内，通过 `polaris-net` 网络互相通信。

### 1. 配置修改

需要将配置文件中的 `127.0.0.1:端口` 改为 Docker 容器名：

| 开发模式 | 生产模式 |
|----------|----------|
| `127.0.0.1:33069` | `io-mysql:3306` |
| `127.0.0.1:37017` | `io-mongo:27017` |
| `127.0.0.1:36379` | `io-redis:6379` |
| `127.0.0.1:39200` | `io-elasticsearch:9200` |
| `127.0.0.1:39092` | `io-kafka:9092` |
| `127.0.0.1:33900` | `io-storage:3900` |

### 2. 启动所有服务

```bash
cd backend

# 创建网络
docker network create polaris-net

# 启动基础设施
docker compose -f docker-compose-env.yml up -d

# 启动 Nginx 网关 + 应用容器
docker compose up -d
```

### 3. 访问方式

生产模式下，所有请求通过 Nginx 网关统一入口：

```bash
# 通过网关访问 (端口 8888)
curl -X POST http://127.0.0.1:8888/usercenter/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{"mobile":"13800138000","password":"123456"}'
```

---

## 项目结构

```
polaris-io/
├── backend/
│   ├── app/                          # 微服务应用
│   │   ├── user/                     # 用户服务
│   │   │   ├── cmd/api/              # HTTP API
│   │   │   ├── cmd/rpc/              # gRPC 服务
│   │   │   └── model/                # 数据模型
│   │   ├── file/                     # 文件服务
│   │   ├── share/                    # 分享服务
│   │   ├── search/                   # 搜索服务
│   │   │   ├── cmd/api/              # HTTP API
│   │   │   ├── cmd/rpc/              # gRPC 服务
│   │   │   ├── cmd/job/              # Kafka 消费者
│   │   │   ├── es/                   # Elasticsearch 客户端
│   │   │   └── types/                # 类型定义
│   │   └── mqueue/                   # 异步任务服务
│   │       └── cmd/job/              # Asynq 消费者
│   ├── pkg/                          # 公共包
│   │   ├── xerr/                     # 错误处理
│   │   ├── ctxdata/                  # 上下文工具
│   │   ├── asynqjob/                 # Asynq 任务定义
│   │   ├── kafka/                    # Kafka 生产者
│   │   ├── s3client/                 # S3 客户端
│   │   ├── filecache/                # 文件哈希缓存
│   │   └── quotacache/               # 配额缓存
│   ├── deploy/                       # 部署配置
│   │   ├── sql/                      # 数据库初始化脚本
│   │   ├── nginx/                    # Nginx 配置
│   │   ├── prometheus/               # 监控配置
│   │   └── goctl/                    # 代码生成模板
│   ├── docs/                         # 设计文档
│   ├── docker-compose-env.yml        # 基础设施容器
│   └── docker-compose.yml            # 网关 + 应用容器
├── go.mod
├── go.sum
└── README.md
```

---

## 监控与运维

| 服务 | 地址 | 说明 |
|------|------|------|
| Prometheus | http://127.0.0.1:9090 | 指标监控 |
| Grafana | http://127.0.0.1:3000 | 监控面板 (admin/admin) |
| Jaeger | http://127.0.0.1:16686 | 链路追踪 |
| Kibana | http://127.0.0.1:5601 | 日志查询 |
| Asynqmon | http://127.0.0.1:8980 | 异步任务监控 |

---

## API 端点一览

### User 服务 (1001)

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/usercenter/v1/user/register` | 用户注册 | 否 |
| POST | `/usercenter/v1/user/login` | 用户登录 | 否 |
| GET | `/usercenter/v1/user/info` | 获取用户信息 | JWT |
| GET | `/usercenter/v1/user/quota` | 获取配额信息 | JWT |

### File 服务 (1002)

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/file/v1/upload/instant` | 秒传检测 | JWT |
| POST | `/file/v1/upload/presign` | 获取上传预签名 | JWT |
| POST | `/file/v1/upload/callback` | 上传完成回调 | JWT |
| GET | `/file/v1/file/list` | 文件列表 | JWT |
| POST | `/file/v1/folder/create` | 创建文件夹 | JWT |
| ... | ... | ... | ... |

### Share 服务 (1003)

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/share/v1/share/create` | 创建分享 | JWT |
| GET | `/share/v1/share/list` | 分享列表 | JWT |
| POST | `/share/v1/share/validate` | 验证分享 | 否 |
| POST | `/share/v1/share/save` | 转存文件 | JWT |
| ... | ... | ... | ... |

### Search 服务 (1004)

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/search/v1/search/files` | 搜索文件 | JWT |
| GET | `/search/v1/search/stats` | 用户统计 | JWT |

---

## 文档

- [架构设计](backend/docs/01_architecture_design.md)
- [用户服务设计](backend/docs/user_service_design.md)
- [文件服务设计](backend/docs/file_service_design.md)
- [分享服务设计](backend/docs/share_service_design.md)
- [搜索服务设计](backend/docs/search_service_design.md)

---

## License

MIT
