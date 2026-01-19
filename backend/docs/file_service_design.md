# File 服务完整设计解析

## 1. 整体架构

```
┌───────────────────────────────────────────────────────────────────────────────────────────┐
│                                    File 服务                                               │
├───────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                           │
│  ┌─────────────────────┐         RPC 调用        ┌─────────────────────────────────────┐  │
│  │     file-api        │  ───────────────────>   │            file-rpc                 │  │
│  │   (HTTP Gateway)    │                         │          (业务核心)                  │  │
│  │   Port: 1002        │                         │          Port: 2002                 │  │
│  └─────────────────────┘                         └─────────────────────────────────────┘  │
│           │                                                      │                        │
│           │ 调用 usercenter-rpc                                   │                        │
│           │ (配额扣减/退还)                                        │                        │
│           ▼                                                      │                        │
│  ┌─────────────────────┐                         ┌───────────────┴───────────────┐        │
│  │  usercenter-rpc     │                         │                               │        │
│  │   (配额管理)         │                         ▼                               ▼        │
│  └─────────────────────┘               ┌─────────────────┐             ┌─────────────────┐│
│                                        │      MySQL      │             │    MongoDB      ││
│                                        │  polaris_file   │             │  polaris_file   ││
│                                        │ user_repository │             │   file_meta     ││
│                                        │   (目录树)      │              │  (文件元数据)    ││
│                                        └─────────────────┘             └─────────────────┘│
│                                                │                               │          │
│                                                ▼                               │          │
│                                        ┌─────────────────┐                     │          │
│                                        │      Redis      │                     │          │
│                                        │   (sqlc 缓存)   │                     │          │
│                                        └─────────────────┘                     │          │
│                                                                                │          │
│           ┌────────────────────────────────────────────────────────────────────┘          │
│           │                                                                               │
│           ▼                                                                               │
│  ┌─────────────────┐                                       ┌─────────────────┐            │
│  │   Garage S3     │  <─────── 预签名URL ─────────────────  │     Kafka       │            │
│  │  (对象存储)      │                                       │  (事件通知)      │            │
│  │  文件实体存储    │                                       │ polaris-file-   │            │
│  └─────────────────┘                                       │    event        │            │
│                                                            └─────────────────┘            │
└───────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. 数据模型

### 2.1 MySQL: user_repository 表（目录树）

```sql
CREATE TABLE `user_repository` (
  `id`          bigint unsigned NOT NULL AUTO_INCREMENT,
  `identity`    varchar(36)  NOT NULL DEFAULT '' COMMENT '文件唯一标识(UUID)',
  `hash`        varchar(64)  NOT NULL DEFAULT '' COMMENT '文件指纹(SHA256)',
  `user_id`     bigint unsigned NOT NULL DEFAULT '0' COMMENT '用户ID',
  `parent_id`   bigint unsigned NOT NULL DEFAULT '0' COMMENT '父目录ID(0=根目录)',
  `name`        varchar(255) NOT NULL DEFAULT '' COMMENT '文件名',
  `ext`         varchar(30)  NOT NULL DEFAULT '' COMMENT '扩展名(文件夹为空)',
  `size`        bigint unsigned NOT NULL DEFAULT '0' COMMENT '文件大小(字节)',
  `path`        varchar(255) NOT NULL DEFAULT '' COMMENT 'S3存储路径',
  `version`     bigint unsigned NOT NULL DEFAULT '0' COMMENT '乐观锁',
  `del_state`   tinyint(1)   NOT NULL DEFAULT '0' COMMENT '0:正常 1:已删除',
  `delete_time` bigint unsigned NOT NULL DEFAULT '0' COMMENT '删除时间戳',
  `create_time` timestamp DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_parent` (`user_id`, `parent_id`),
  UNIQUE KEY `idx_identity` (`identity`)
);
```

**设计要点**：
- `identity`：UUID，对外暴露，不暴露自增 ID
- `hash`：文件夹为空，文件有值（用于关联 MongoDB 元数据）
- `parent_id=0` 表示根目录
- `del_state` 实现软删除（回收站功能）

### 2.2 MongoDB: file_meta 集合（文件元数据）

```javascript
// 集合: polaris_file.file_meta
{
  "_id":         ObjectId,
  "hash":        "sha256...",     // 文件 SHA256，唯一索引
  "size":        1024,            // 文件大小(字节)
  "s3_key":      "uploads/2026/01/15/xxx.pdf",  // S3 存储路径
  "ext":         "pdf",           // 扩展名
  "mime_type":   "application/pdf",
  "ref_count":   3,               // 引用计数（秒传核心）
  "ext_attr":    {},              // 扩展属性（图片宽高、视频时长等）
  "create_time": ISODate,
  "update_time": ISODate
}

// 索引
{ "hash": 1 }        // 唯一索引，用于秒传查询
{ "create_time": -1 }
```

**为什么用 MongoDB？**
- 秒传查询频繁，需要高性能
- `ext_attr` 扩展属性灵活
- 与用户目录树解耦

---

## 3. 存储分层设计

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                              文件存储三层架构                                             │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                         │
│  ┌─────────────────────────────────────────────────────────────────────────────────┐    │
│  │                        Layer 1: 用户目录树 (MySQL)                               │    │
│  │                                                                                  │    │
│  │   用户A的文件                          用户B的文件                                │    │
│  │   ├── 文档/                           ├── 工作/                                  │    │
│  │   │   ├── report.pdf (hash: abc123)   │   └── report.pdf (hash: abc123) ←同一文件│    │
│  │   │   └── notes.txt                   └── 照片/                                  │    │
│  │   └── 照片/                               └── photo.jpg                         │    │
│  │       └── photo.jpg                                                             │    │
│  │                                                                                  │    │
│  │   每个用户有独立的目录结构，文件通过 hash 关联到 Layer 2                           │    │
│  └─────────────────────────────────────────────────────────────────────────────────┘    │
│                                          │                                              │
│                                          │ hash                                         │
│                                          ▼                                              │
│  ┌─────────────────────────────────────────────────────────────────────────────────┐    │
│  │                        Layer 2: 文件元数据 (MongoDB)                             │    │
│  │                                                                                  │    │
│  │   file_meta 集合（全局唯一，按 hash 去重）                                        │    │
│  │                                                                                  │    │
│  │   { hash: "abc123", s3_key: "uploads/.../xxx.pdf", ref_count: 2 }               │    │
│  │   { hash: "def456", s3_key: "uploads/.../yyy.jpg", ref_count: 1 }               │    │
│  │                                                                                  │    │
│  │   ref_count=2 表示有 2 个用户引用了这个文件（秒传的核心）                          │    │
│  └─────────────────────────────────────────────────────────────────────────────────┘    │
│                                          │                                              │
│                                          │ s3_key                                       │
│                                          ▼                                              │
│  ┌─────────────────────────────────────────────────────────────────────────────────┐    │
│  │                        Layer 3: 文件实体 (S3/Garage)                             │    │
│  │                                                                                  │    │
│  │   polaris-bucket/                                                               │    │
│  │   └── uploads/                                                                  │    │
│  │       └── 2026/01/15/                                                           │    │
│  │           ├── uuid1.pdf   ← 实际存储的文件二进制                                  │    │
│  │           └── uuid2.jpg                                                         │    │
│  │                                                                                  │    │
│  │   相同内容的文件只存一份！                                                        │    │
│  └─────────────────────────────────────────────────────────────────────────────────┘    │
│                                                                                         │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 4. API 接口设计

### 4.1 对外 HTTP 接口 (file-api)

| 分类 | 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|------|
| **上传** | POST | `/file/v1/upload/check` | ✅ | 秒传检查 |
| | POST | `/file/v1/upload/presign` | ✅ | 获取预签名上传URL |
| | POST | `/file/v1/upload/complete` | ✅ | 上传完成回调 |
| **目录** | POST | `/file/v1/folder/create` | ✅ | 创建文件夹 |
| **文件** | GET | `/file/v1/list` | ✅ | 列出目录内容 |
| | GET | `/file/v1/detail` | ✅ | 获取文件详情 |
| | GET | `/file/v1/download` | ✅ | 获取下载URL |
| | POST | `/file/v1/move` | ✅ | 移动文件/文件夹 |
| | POST | `/file/v1/rename` | ✅ | 重命名 |
| | POST | `/file/v1/copy` | ✅ | 复制文件/文件夹 |
| | POST | `/file/v1/delete` | ✅ | 删除（移入回收站） |
| **回收站** | GET | `/file/v1/trash/list` | ✅ | 回收站列表 |
| | POST | `/file/v1/trash/restore` | ✅ | 恢复文件 |
| | POST | `/file/v1/trash/delete` | ✅ | 彻底删除 |
| | POST | `/file/v1/trash/clear` | ✅ | 清空回收站 |

### 4.2 内部 RPC 接口 (file-rpc)

| 分类 | 方法 | 调用者 | 说明 |
|------|------|--------|------|
| **上传** | `CheckInstantUpload` | file-api | 查询 MongoDB 检查秒传 |
| | `GetPresignedUploadUrl` | file-api | 生成 S3 预签名URL |
| | `CreateFile` | file-api | 创建文件记录 |
| **查询** | `GetFileInfo` | share-rpc | 获取文件信息（供分享服务） |
| | `GetFilesByIds` | search-rpc | 批量获取文件 |
| | `ListFiles` | file-api | 列出目录内容 |
| **操作** | `MoveFiles` | file-api | 移动 |
| | `RenameFile` | file-api | 重命名 |
| | `CopyFiles` | file-api | 复制 |
| | `SoftDeleteFiles` | file-api | 软删除 |
| **回收站** | `ListTrash` | file-api | 回收站列表 |
| | `RestoreFiles` | file-api | 恢复 |
| | `HardDeleteFiles` | file-api | 彻底删除 |
| | `ClearTrash` | file-api | 清空回收站 |

---

## 5. 核心业务流程

### 5.1 文件上传流程（完整版）

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                              文件上传完整流程                                             │
└─────────────────────────────────────────────────────────────────────────────────────────┘

客户端                    file-api                file-rpc                外部服务
  │                          │                       │                       │
  │  1. POST /upload/check   │                       │                       │
  │  {hash, size, name}      │                       │                       │
  │ ─────────────────────────>                       │                       │
  │                          │  CheckInstantUpload   │                       │
  │                          │ ──────────────────────>                       │
  │                          │                       │                       │
  │                          │       查询 MongoDB    │                       │
  │                          │       file_meta       │──────────────────────>│ MongoDB
  │                          │                       │<──────────────────────│
  │                          │                       │                       │
  │                          │<───────────────────── │                       │
  │<───────────────────────── │                       │                       │
  │  {exists: true/false}    │                       │                       │
  │                          │                       │                       │
  ├──────────────────────────┼───────────────────────┼───────────────────────┤
  │  如果 exists=true (秒传) │                       │                       │
  │  直接跳到步骤 4          │                       │                       │
  ├──────────────────────────┼───────────────────────┼───────────────────────┤
  │                          │                       │                       │
  │  2. POST /upload/presign │                       │                       │
  │  {hash, size, name}      │                       │                       │
  │ ─────────────────────────>                       │                       │
  │                          │                       │                       │
  │                          │  先扣减配额            │                       │
  │                          │ ──────────────────────────────────────────────>│ usercenter
  │                          │<──────────────────────────────────────────────│ DeductQuota
  │                          │                       │                       │
  │                          │  GetPresignedUploadUrl│                       │
  │                          │ ──────────────────────>                       │
  │                          │                       │  生成 S3 预签名URL    │
  │                          │                       │──────────────────────>│ S3
  │                          │                       │<──────────────────────│
  │                          │<───────────────────── │                       │
  │<───────────────────────── │                       │                       │
  │  {uploadUrl, uploadKey}  │                       │                       │
  │                          │                       │                       │
  │  3. PUT uploadUrl        │                       │                       │
  │  (直接上传到 S3)          │                       │                       │
  │ ─────────────────────────────────────────────────────────────────────────>│ S3
  │<─────────────────────────────────────────────────────────────────────────│
  │                          │                       │                       │
  │  4. POST /upload/complete│                       │                       │
  │  {uploadKey, hash, ...}  │                       │                       │
  │ ─────────────────────────>                       │                       │
  │                          │      CreateFile       │                       │
  │                          │ ──────────────────────>                       │
  │                          │                       │                       │
  │                          │                       │  1. 创建/更新 MongoDB │
  │                          │                       │     file_meta         │
  │                          │                       │  2. 创建 MySQL        │
  │                          │                       │     user_repository   │
  │                          │                       │  3. 发送 Kafka 事件   │
  │                          │                       │──────────────────────>│ Kafka
  │                          │                       │                       │
  │                          │<───────────────────── │                       │
  │<───────────────────────── │                       │                       │
  │  {identity}              │                       │                       │
  │                          │                       │                       │
```

### 5.2 秒传原理

```
用户A 上传 report.pdf (hash=abc123)
                │
                ▼
┌──────────────────────────────────────┐
│  1. 检查 MongoDB file_meta          │
│     是否存在 hash=abc123 ?          │
└──────────────────────────────────────┘
                │
       不存在 (首次上传)
                │
                ▼
┌──────────────────────────────────────┐
│  2. 上传到 S3                        │
│  3. 创建 file_meta (ref_count=1)    │
│  4. 创建 user_repository            │
└──────────────────────────────────────┘

─────────────────────────────────────────

用户B 上传 同一个 report.pdf (hash=abc123)
                │
                ▼
┌──────────────────────────────────────┐
│  1. 检查 MongoDB file_meta          │
│     hash=abc123 存在！               │
└──────────────────────────────────────┘
                │
         存在 (秒传！)
                │
                ▼
┌──────────────────────────────────────┐
│  2. 不上传到 S3！                    │
│  3. ref_count++ (变成 2)            │
│  4. 创建 user_repository            │
│     (指向同一个 file_meta)          │
└──────────────────────────────────────┘
```

### 5.3 删除流程与引用计数

```
用户删除文件流程：

1. 软删除 (移入回收站)
   └── user_repository.del_state = 1
   └── 配额退还
   └── 发送 Kafka file_deleted 事件

2. 彻底删除 (从回收站删除)
   └── file_meta.ref_count--
   └── 删除 user_repository 记录
   
3. 当 ref_count = 0 时
   └── 删除 MongoDB file_meta 记录
   └── 发送 Asynq 任务删除 S3 文件（通过 mqueue）
```

---

## 6. Kafka 事件设计

### 6.1 事件类型

```json
// file_uploaded - 文件上传完成
{
  "event_type": "file_uploaded",
  "user_id": 123,
  "file_id": 456,
  "identity": "uuid-xxx",
  "name": "report.pdf",
  "hash": "sha256...",
  "size": 1024,
  "ext": "pdf",
  "timestamp": "2026-01-15T10:30:00Z"
}

// file_deleted - 文件删除
{
  "event_type": "file_deleted",
  "user_id": 123,
  "file_id": 456,
  "identity": "uuid-xxx",
  "timestamp": "2026-01-15T11:00:00Z"
}

// file_updated - 文件更新（移动/重命名）
{
  "event_type": "file_updated",
  "user_id": 123,
  "file_id": 456,
  "identity": "uuid-xxx",
  "name": "new_name.pdf",
  "parent_id": 789,
  "timestamp": "2026-01-15T11:30:00Z"
}
```

### 6.2 消费者（后续开发）

```
┌─────────────────┐     polaris-file-event     ┌─────────────────┐
│   file-rpc      │  ─────────────────────────> │   search-rpc    │
│   (生产者)       │                            │   (消费者)       │
└─────────────────┘                            └─────────────────┘
                                                       │
                                                       ▼
                                               ┌─────────────────┐
                                               │  Elasticsearch  │
                                               │   (文件搜索)     │
                                               └─────────────────┘
```

---

## 7. 可优化点分析

### 🔴 高优先级

#### 7.1 上传中断未处理（你之前问的问题）

**当前问题**：
```
presign 成功 → 配额已扣减 → 用户上传中断 → complete 未调用 → 配额永久占用
```

**优化方案**：Asynq 延迟任务
```go
// presign 时创建延迟任务
task := asynq.NewTask("upload:verify", payload)
asynqClient.Enqueue(task, asynq.ProcessIn(30*time.Minute))

// 30分钟后检查：如果 complete 未调用，退还配额
```

#### 7.2 彻底删除时未删除 S3 文件

**当前问题**：
```go
// hardDeleteFilesLogic.go
// ref_count-- 后如果为 0，应该删除 S3 文件
// 当前只是记录日志：TODO: 如果引用计数为 0，删除 S3 文件
```

**优化方案**：
```go
// 检查 ref_count
meta, _ := l.svcCtx.FileMetaModel.FindByHash(ctx, hash)
if meta.RefCount <= 0 {
    // 方案1: 同步删除 S3
    l.svcCtx.S3Client.DeleteObject(ctx, meta.S3Key)
    l.svcCtx.FileMetaModel.DeleteByHash(ctx, hash)
    
    // 方案2: 异步删除（推荐，更可靠）
    asynqClient.Enqueue(asynq.NewTask("s3:cleanup", payload))
}
```

#### 7.3 大文件夹递归删除性能问题

**当前问题**：
```go
// clearTrashLogic.go
for {
    files, _ := FindTrashList(...)  // 每次查 100 条
    for _, file := range files {
        HardDelete(...)  // 逐条删除
    }
}
```

**优化方案**：
```go
// 方案1: 批量删除 SQL
DELETE FROM user_repository WHERE user_id = ? AND del_state = 1 LIMIT 1000

// 方案2: 异步任务分批处理
if count > 100 {
    asynqClient.Enqueue(asynq.NewTask("trash:clear", payload))
}
```

---

### 🟡 中优先级

#### 7.4 秒传查询缺少 Redis 缓存

**当前问题**：每次秒传检查都查 MongoDB

**优化方案**：
```go
func (l *CheckInstantUploadLogic) CheckInstantUpload(in *pb.CheckInstantUploadReq) (*pb.CheckInstantUploadResp, error) {
    // 1. 先查 Redis
    key := fmt.Sprintf("file:meta:%s", in.Hash)
    cached, err := l.svcCtx.Redis.Get(key)
    if err == nil && cached != "" {
        // 命中缓存
        return &pb.CheckInstantUploadResp{Exists: true, ...}, nil
    }
    
    // 2. 查 MongoDB
    meta, err := l.svcCtx.FileMetaModel.FindByHash(ctx, in.Hash)
    if err == nil {
        // 3. 写入缓存（24小时过期）
        l.svcCtx.Redis.Setex(key, json.Marshal(meta), 86400)
    }
    
    return &pb.CheckInstantUploadResp{...}, nil
}
```

#### 7.5 文件列表查询缺少缓存

**当前问题**：每次列出目录都查数据库

**优化方案**：
```go
// 缓存目录结构（短时间缓存，如 5 分钟）
key := fmt.Sprintf("file:list:%d:%d", userId, parentId)

// 目录内容变更时清除缓存
// 在 CreateFile, Move, Rename, Delete 后调用：
l.svcCtx.Redis.Del(key)
```

#### 7.6 缺少文件类型限制

**当前问题**：没有限制上传文件类型

**优化方案**：
```go
// 配置允许的文件类型
AllowedMimeTypes:
  - image/*
  - application/pdf
  - text/*
  
// 检查
if !isAllowedMimeType(in.MimeType) {
    return nil, xerr.NewErrCode(xerr.FILE_TYPE_NOT_ALLOWED)
}
```

#### 7.7 缺少文件大小限制

**当前问题**：没有限制单文件大小

**优化方案**：
```yaml
# 配置
MaxFileSize: 5368709120  # 5GB

# 检查
if in.Size > c.MaxFileSize {
    return nil, xerr.NewErrCode(xerr.FILE_SIZE_EXCEEDED)
}
```

---

### 🟢 低优先级

#### 7.8 缺少分片上传支持

**当前问题**：大文件一次性上传，容易失败

**优化方案**：S3 Multipart Upload
```go
// 1. 初始化分片上传
uploadId := s3.CreateMultipartUpload(bucket, key)

// 2. 上传各分片（可并行）
for i, part := range parts {
    s3.UploadPart(bucket, key, uploadId, i+1, part)
}

// 3. 完成分片上传
s3.CompleteMultipartUpload(bucket, key, uploadId, parts)
```

#### 7.9 缺少文件预览支持

**当前问题**：只能下载，不能预览

**优化方案**：
```go
// 对于图片、PDF、Office 等
// 生成预览图/转换格式
// 存储到 S3 的 previews/ 目录
```

#### 7.10 缺少文件历史版本

**当前问题**：覆盖上传会丢失旧版本

**优化方案**：
```sql
CREATE TABLE `file_version` (
    `id` bigint NOT NULL,
    `file_id` bigint NOT NULL,
    `version` int NOT NULL,
    `hash` varchar(64) NOT NULL,
    `size` bigint NOT NULL,
    `create_time` timestamp,
    ...
);
```

---

## 8. 优化优先级总结

| 优先级 | 优化项 | 影响 | 复杂度 |
|--------|--------|------|--------|
| 🔴 高 | 上传中断处理（Asynq 延迟任务） | 数据一致性 | 中 |
| 🔴 高 | 彻底删除时清理 S3 文件 | 存储成本 | 低 |
| 🔴 高 | 大文件夹批量删除优化 | 性能 | 中 |
| 🟡 中 | 秒传 Redis 缓存 | 性能 | 低 |
| 🟡 中 | 文件列表缓存 | 性能 | 低 |
| 🟡 中 | 文件类型限制 | 安全性 | 低 |
| 🟡 中 | 文件大小限制 | 安全性 | 低 |
| 🟢 低 | 分片上传 | 大文件体验 | 高 |
| 🟢 低 | 文件预览 | 用户体验 | 高 |
| 🟢 低 | 文件历史版本 | 功能完整性 | 高 |

---

## 9. 当前设计的优点

| 优点 | 说明 |
|------|------|
| ✅ 三层存储分离 | MySQL 目录树 + MongoDB 元数据 + S3 实体，各司其职 |
| ✅ 秒传设计 | 通过 hash + ref_count 实现，节省存储和带宽 |
| ✅ 预签名上传 | 客户端直传 S3，服务端不过数据流 |
| ✅ 软删除 + 回收站 | 符合用户习惯，防止误删 |
| ✅ Kafka 事件 | 为搜索、统计等服务预留扩展点 |
| ✅ 乐观锁 | 防止并发修改冲突 |
| ✅ 配额联动 | 上传扣减、删除退还，保证数据一致 |

---