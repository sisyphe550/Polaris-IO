# Polaris-IO File 服务开发指南

> 此文档是 File 服务的补充开发指南，需配合主项目 Prompt 一起使用。

## 一、File 服务架构概述

### 1.1 存储分层设计

| 存储层 | 技术 | 用途 | 数据库/集合 |
|--------|------|------|-------------|
| **目录树** | MySQL 8.0 | 用户视角的文件/文件夹结构 | `polaris_file.user_repository` |
| **文件元数据** | MongoDB 7.0 | 文件实体信息（Hash、大小、S3路径）| `polaris_file.file_meta` |
| **秒传缓存** | Redis | Hash → file_meta._id 快速查询 | Key: `file:hash:{sha256}` |
| **文件实体** | Garage S3 | 实际文件内容存储 | Bucket: `polaris-bucket` |

### 1.2 服务角色

- **Kafka**: File 服务作为 **生产者**，发送文件变更事件给 search 服务
- **RPC 调用**: 调用 `usercenter-rpc` 进行配额扣减/退还

### 1.3 目录结构

```
app/file/
├── cmd/
│   ├── api/                          # HTTP 接口层
│   │   ├── desc/
│   │   │   ├── file/
│   │   │   │   └── file.api          # 类型定义
│   │   │   └── file.api              # 路由入口
│   │   ├── etc/
│   │   │   └── file.yaml
│   │   └── internal/
│   │       ├── config/
│   │       ├── handler/
│   │       ├── logic/
│   │       └── svc/                  # 初始化 S3Client, KafkaProducer, MongoDB
│   │
│   └── rpc/                          # gRPC 接口层
│       ├── etc/
│       │   └── file.yaml
│       ├── pb/
│       │   └── file.proto
│       └── internal/
│           ├── config/
│           ├── logic/
│           ├── server/
│           └── svc/
│
├── model/                            # MySQL Model (goctl 生成)
│   ├── userrepositorymodel.go
│   ├── userrepositorymodel_gen.go
│   └── vars.go
│
└── mongo/                            # MongoDB Model (手动编写)
    ├── filemeta.go
    └── types.go
```

---

## 二、数据库设计

### 2.1 MySQL - user_repository 表

位置: `deploy/sql/polaris_file.sql`

```sql
CREATE TABLE `user_repository` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `identity` varchar(36) NOT NULL DEFAULT '' COMMENT '文件唯一标识(UUID)',
  `hash` varchar(64) NOT NULL DEFAULT '' COMMENT '文件SHA256指纹',
  `user_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '用户ID',
  `parent_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '父目录ID(0=根目录)',
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '文件/文件夹名',
  `ext` varchar(30) NOT NULL DEFAULT '' COMMENT '扩展名(文件夹为空)',
  `size` bigint unsigned NOT NULL DEFAULT '0' COMMENT '文件大小(字节)',
  `path` varchar(255) NOT NULL DEFAULT '' COMMENT 'S3存储路径',
  
  `version` bigint unsigned NOT NULL DEFAULT '0' COMMENT '乐观锁版本号',
  `del_state` tinyint(1) NOT NULL DEFAULT '0' COMMENT '删除状态 0:正常 1:已删除',
  `delete_time` bigint unsigned NOT NULL DEFAULT '0' COMMENT '删除时间戳',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  PRIMARY KEY (`id`),
  KEY `idx_user_parent` (`user_id`, `parent_id`),
  UNIQUE KEY `idx_identity` (`identity`)
);
```

### 2.2 MongoDB - file_meta 集合

位置: `deploy/mongo/init.js`

```javascript
// 集合结构
{
  "_id": ObjectId,
  "hash": "sha256字符串",          // 文件 SHA256，唯一索引
  "size": NumberLong,              // 文件大小（字节）
  "s3_key": "path/to/file",        // S3 存储路径
  "ext": "pdf",                    // 文件扩展名
  "mime_type": "application/pdf",  // MIME 类型
  "ref_count": NumberInt,          // 引用计数
  "ext_attr": {                    // 扩展属性（可选）
    "width": 1920,
    "height": 1080,
    "duration": 120
  },
  "create_time": ISODate,
  "update_time": ISODate
}

// 索引
db.file_meta.createIndex({ "hash": 1 }, { unique: true });
```

---

## 三、基础设施配置

### 3.1 Garage S3 配置

```yaml
S3:
  Endpoint: "127.0.0.1:3900"           # 本地开发 | Docker: garage:3900
  Region: "garage"
  Bucket: "polaris-bucket"
  AccessKey: "GK1b745dac9808425c935ac639"
  SecretKey: "f16100350e1f496215a9edfd9bb5944abf4e3b7fb54e83da7f0fb108e39bf509"
  UseSSL: false
```

### 3.2 MongoDB 配置

```yaml
MongoDB:
  Uri: "mongodb://root:polaris-io-mongo-2026@127.0.0.1:27017"  # 本地 | Docker: mongo:27017
  Database: "polaris_file"
```

### 3.3 Kafka 配置 (生产者)

**已有 Topics：**
```bash
# 查看命令: docker exec -it io-kafka /opt/kafka/bin/kafka-topics.sh --list --bootstrap-server localhost:9094
polaris-file-event    # ← File 服务使用此 Topic 发送文件变更事件
polaris-log           # 日志收集用
```

```yaml
KafkaProducer:
  Brokers:
    - "127.0.0.1:9094"               # 本地开发 | Docker: kafka:9092
  Topic: "polaris-file-event"        # 已创建，无需额外操作
```

### 3.4 MySQL 配置

```yaml
DB:
  DataSource: "root:root@tcp(127.0.0.1:33069)/polaris_file?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"
  # Docker: mysql:3306
```

### 3.5 Redis 配置

```yaml
Redis:
  Host: "127.0.0.1:36379"            # 本地开发 | Docker: redis:6379
  Type: node
  Pass: "polaris"
```

---

## 四、核心 API 接口设计

### 4.1 上传流程

| 接口 | 方法 | 说明 |
|------|------|------|
| `/file/upload/check` | POST | 秒传检查（传入 hash + size） |
| `/file/upload/presign` | POST | 获取 S3 预签名上传 URL |
| `/file/upload/complete` | POST | 上传完成回调，创建记录 |

### 4.2 目录管理

| 接口 | 方法 | 说明 |
|------|------|------|
| `/file/folder/create` | POST | 创建文件夹 |
| `/file/list` | GET | 列出目录内容 |
| `/file/move` | POST | 移动文件/文件夹 |
| `/file/rename` | POST | 重命名 |
| `/file/copy` | POST | 复制文件/文件夹 |

### 4.3 文件操作

| 接口 | 方法 | 说明 |
|------|------|------|
| `/file/download` | GET | 获取下载预签名 URL |
| `/file/delete` | POST | 软删除（移入回收站） |
| `/file/detail` | GET | 获取文件详情 |

### 4.4 回收站

| 接口 | 方法 | 说明 |
|------|------|------|
| `/file/trash/list` | GET | 回收站列表 |
| `/file/trash/restore` | POST | 恢复文件 |
| `/file/trash/delete` | POST | 彻底删除 |
| `/file/trash/clear` | POST | 清空回收站 |

---

## 五、RPC 接口设计

供其他服务（如 share、search）调用：

| RPC 方法 | 说明 |
|----------|------|
| `GetFileInfo` | 获取文件信息 |
| `GetFilesByIds` | 批量获取文件信息 |
| `CheckFileExists` | 检查文件是否存在 |
| `GetDownloadUrl` | 获取下载 URL |
| `DeductUserQuota` | 扣减用户配额（调用 usercenter-rpc） |

---

## 六、开发流程

### 6.1 生成 MySQL Model

```bash
cd /Users/sisyphus/Documents/Code/Go/polaris-io/backend

# 生成 user_repository model
goctl model mysql ddl \
  --src ./deploy/sql/polaris_file.sql \
  --dir ./app/file/model \
  --home ./deploy/goctl \
  --style goZero \
  -c
```

### 6.2 设计并生成 API 代码

1. 编写 `app/file/cmd/api/desc/file/file.api` (类型定义)
2. 编写 `app/file/cmd/api/desc/file.api` (路由定义)

```bash
# 生成 API 代码
goctl api go \
  --api ./app/file/cmd/api/desc/file.api \
  --dir ./app/file/cmd/api \
  --home ./deploy/goctl \
  --style goZero
```

### 6.3 设计并生成 RPC 代码

1. 编写 `app/file/cmd/rpc/pb/file.proto`

```bash
# 生成 RPC 代码
goctl rpc protoc ./app/file/cmd/rpc/pb/file.proto \
  --go_out=./app/file/cmd/rpc \
  --go-grpc_out=./app/file/cmd/rpc \
  --zrpc_out=./app/file/cmd/rpc \
  --home ./deploy/goctl \
  --style goZero
```

### 6.4 手动编写部分

1. **MongoDB Model** (`app/file/mongo/`)
2. **配置文件** (`api/etc/file.yaml`, `rpc/etc/file.yaml`)
3. **ServiceContext** - 初始化 S3Client, MongoDB, KafkaProducer
4. **Logic 业务逻辑**

---

## 七、Kafka 事件设计

File 服务作为生产者，发送以下事件到 `polaris-file-event` Topic（已创建）：

```json
// 文件上传完成事件
{
  "event_type": "file_uploaded",
  "user_id": 123,
  "file_id": 456,
  "identity": "uuid-xxx",
  "name": "document.pdf",
  "hash": "sha256-xxx",
  "size": 1024000,
  "ext": "pdf",
  "timestamp": "2026-01-15T10:00:00Z"
}

// 文件删除事件
{
  "event_type": "file_deleted",
  "user_id": 123,
  "file_id": 456,
  "identity": "uuid-xxx",
  "timestamp": "2026-01-15T10:00:00Z"
}

// 文件移动/重命名事件
{
  "event_type": "file_updated",
  "user_id": 123,
  "file_id": 456,
  "identity": "uuid-xxx",
  "name": "new-name.pdf",
  "parent_id": 789,
  "timestamp": "2026-01-15T10:00:00Z"
}
```

Search 服务的 `cmd/mq` 消费这些事件并同步到 Elasticsearch。

---

## 八、依赖服务

| 服务 | 用途 | 调用方式 |
|------|------|----------|
| usercenter-rpc | 配额扣减/退还 | gRPC |
| Garage S3 | 文件存储 | AWS SDK |
| MongoDB | 文件元数据 | mongo-driver |
| Kafka | 事件发布 | kafka-go / sarama |

---

## 九、开发检查清单

- [ ] MySQL Model 生成 (user_repository)
- [ ] MongoDB Model 编写 (file_meta)
- [ ] API .api 文件设计
- [ ] RPC .proto 文件设计
- [ ] goctl 生成 API/RPC 框架代码
- [ ] 配置文件完善 (file.yaml)
- [ ] ServiceContext 初始化
- [ ] S3 工具类封装
- [ ] Kafka Producer 封装
- [ ] Logic 业务逻辑实现
- [ ] 错误码定义 (pkg/xerr)
- [ ] 单元测试
