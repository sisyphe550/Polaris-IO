# 项目名称：Polaris-IO (北极星) 分布式云存储平台
## 核心目标：
    - 构建一个支持 PB 级数据存储、高并发上传下载的私有云盘系统。
    - 实现基于 S3 协议的流式上传与断点续传。
    - 基于 Go-Zero 微服务架构，实现用户、文件、分享、搜索的解耦治理。
## 非功能性指标 (NFR)：
    - 支持 1GB+ 大文件上传。
    - 文件秒传率提升至 30% 以上。
    - 核心接口 P99 延迟 < 200ms。

## 系统设计
| **模块 (Module)**| **数据库与存储选型 (DB & Storage)**| **核心表/集合设计 (Key Tables/Collections)**| **核心功能 (Core Functions)**| **涉及的 Mqueue 延迟/异步任务**|
| ---------------------------------- | -------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **User**<br><br>  <br><br>(用户中心)   | **MySQL 8.0** (强事务)<br><br>  <br><br>**Redis** (缓存) | **1. `user` (用户主表):**<br><br>  <br><br>- `id`, `name`, `mobile/email`, `password_hash`, `status` (正常/封禁)<br><br>  <br><br>**2. `user_quota` (容量表):**<br><br>  <br><br>- `user_id` (PK)<br><br>  <br><br>- `total_size` (固定值, 如 100GB)<br><br>  <br><br>- `used_size` (已用, bytes)        | **1. 基础鉴权:** 注册/登录 (JWT 双 Token: Access/Refresh)。<br><br>  <br><br>**2. 空间管理:** 注册即送固定空间 (如100G)。<br><br>  <br><br>**3. 乐观锁更新:** `UPDATE set used = used + ? WHERE id=?` 防止并发超额。                                 | **1. `task:user:quota_compensate`**<br><br>  <br><br>(异步补偿)<br><br>  <br><br>文件上传成功但扣费（容量扣除）失败时，通过消息队列触发强制一致性检查。                                                                                                                          |
| **File**<br><br>  <br><br>(文件核心)   | **MongoDB** (元数据)<br><br>  <br><br>**MySQL** (目录树)<br><br>  <br><br>**Garage/S3** (实体)<br><br>  <br><br>**Redis** (秒传Hash) | **1. `file_meta` (Mongo 集合):**<br><br>  <br><br>- `_id`, `sha256` (秒传索引), `size`, `s3_key`, `ext_attr` (视频时长/图片宽高)<br><br>  <br><br>**2. `user_repository` (MySQL 目录表):**<br><br>  <br><br>- `id`, `user_id`, `parent_id` (父目录), `name`, `file_id` (关联Mongo), `status` (normal/deleted) | **1. 秒传 (Instant Upload):** 计算 Hash 查 Redis/Mongo，存在即直接生成引用。<br><br>  <br><br>**2. 预签名上传:** 生成 S3 `PutObject` 临时 URL，**前端直传 S3**。<br><br>  <br><br>**3. 目录树:** 文件夹创建、移动、重命名。<br><br>  <br><br>**4. 软删除:** 放入回收站。 | **1. `task:file:recycle_clean`**<br><br>  <br><br>(延迟 30 天)<br><br>  <br><br>回收站文件到期后，若未还原，则彻底物理删除 S3 对象。<br><br>  <br><br>**2. `task:file:upload_gc`**<br><br>  <br><br>(延迟 24 小时)<br><br>  <br><br>清理 S3 中初始化了但未完成（断网/取消）的分片上传碎片。 |
| **Share**<br><br>  <br><br>(分享分发)  | **MySQL** (持久化)<br><br>  <br><br>**Redis** (高频读) | **1. `share_link` (分享记录表):**<br><br>  <br><br>- `id`, `code` (提取码), `short_key` (唯一短链), `user_repository_id` (关联文件/目录), `expired_at` (过期时间), `click_count` | **1. 创建分享:** 生成唯一短链 Key，写入 Redis 缓存和 MySQL。<br><br>  <br><br>**2. 查看鉴权:** 校验提取码，判断是否过期。<br><br>  <br><br>**3. 转存文件:** 逻辑复制，在接收者的 `user_repository` 生成新记录，**引用**同一个 `file_meta`。| **1. `task:share:expire`**<br><br>  <br><br>(延迟 N 天)<br><br>  <br><br>分享有效期截止时，自动将分享状态置为 `expired`，并删除 Redis 中的缓存 Key 以释放内存。 |
| **Search**<br><br>  <br><br>(全文检索) | **Elasticsearch**<br><br>  <br><br>(倒排索引) | **1. `file_index` (ES 索引):**<br><br>  <br><br>- `doc_id` (对应 user_repo_id)<br><br>  <br><br>- `name` (文件名, ik_smart分词)<br><br>  <br><br>- `ext` (后缀)<br><br>  <br><br>- `owner_id` (归属人)<br><br>  <br><br>- `is_folder` (是否文件夹) | **1. 全文搜索:** 支持对文件名进行模糊搜索、高亮显示。<br><br>  <br><br>**2. 结果过滤:** 仅搜索当前登录用户拥有或已转存的文件。| **(无主动延迟任务)**<br><br>  <br><br>_注：Search 是纯消费者。由 Mqueue 监听 `FileUploaded` 事件后调用 SearchRPC 写入数据。_  |
| **Mqueue**<br><br>  <br><br>(异步治理) | **Redis** (Asynq 基座)<br><br>  <br><br>**Kafka** (高吞吐) | **(无独立业务表)**<br><br>  <br><br>利用 Redis 的 `Sorted Set` 和 `List` 实现延时和任务队列。| **1. 任务调度:** 接收并存储延时任务。<br><br>  <br><br>**2. 削峰填谷:** 消费 Kafka 中的 `FileUploaded` 等高频事件。<br><br>  <br><br>**3. 任务分发:** 调用 User/File/Search 的 RPC 接口执行具体逻辑。| **核心枢纽:**<br><br>  <br><br>负责所有“非实时响应”的逻辑，确保上传接口 < 100ms 响应。|


## 系统架构图

``` mermaid
graph TD
    %% =================样式定义=================
    classDef client fill:#f9f9f9,stroke:#333,stroke-width:2px;
    classDef gateway fill:#2E8B57,stroke:#333,stroke-width:2px,color:white;
    classDef api fill:#6495ED,stroke:#2b5eac,stroke-width:2px,color:white;
    classDef rpc fill:#FF7F50,stroke:#d45500,stroke-width:2px,color:white;
    classDef db fill:#FFD700,stroke:#daa520,stroke-width:2px,stroke-dasharray: 5 5;
    classDef mq fill:#9370DB,stroke:#4b0082,stroke-width:2px,color:white;

    %% =================1. 接入层=================
    User((用户 Client)):::client
    Nginx[Nginx Gateway]:::gateway
    
    User -- "HTTPS 请求" --> Nginx

    %% =================2. API 聚合层 (HTTP)=================
    subgraph "API Layer (BFF)"
        UserAPI[User API]:::api
        FileAPI[File API]:::api
        ShareAPI[Share API]:::api
        SearchAPI[Search API]:::api
    end

    Nginx -- "/user" --> UserAPI
    Nginx -- "/file" --> FileAPI
    Nginx -- "/share" --> ShareAPI
    Nginx -- "/search" --> SearchAPI

    %% =================3. RPC 服务层 (gRPC)=================
    subgraph "RPC Service Layer"
        UserRPC[User RPC]:::rpc
        FileRPC[File RPC]:::rpc
        ShareRPC[Share RPC]:::rpc
        SearchRPC[Search RPC]:::rpc
    end

    %% API -> RPC 同步调用
    UserAPI --> UserRPC
    FileAPI --> UserRPC & FileRPC
    ShareAPI --> ShareRPC & FileRPC & UserRPC
    SearchAPI --> SearchRPC

    %% =================4. 异步治理中心 (Mqueue)=================
    subgraph "Async Job Center"
        Kafka{{Kafka 消息队列}}:::mq
        Asynq{{Asynq 延时队列}}:::mq
        Mqueue[Mqueue Consumers]:::rpc
    end

    %% 生产者 (Production)
    FileRPC -. "1. 上传完成事件" .-> Kafka
    FileRPC -. "2. 30天回收任务" .-> Asynq

    %% 消费者 (Consumption)
    Kafka --> Mqueue
    Asynq --> Mqueue

    %% 消费行为 (Side Effects)
    Mqueue -- "3. 同步索引 (写)" --> SearchRPC
    Mqueue -- "4. 异步扣配额" --> UserRPC
    Mqueue -- "5. 生成缩略图" --> FileRPC

    %% =================5. 存储基础设施=================
    subgraph "Infrastructure"
        MySQL[(MySQL Cluster)]:::db
        Mongo[(MongoDB)]:::db
        Redis[(Redis Cache)]:::db
        ES[(Elasticsearch)]:::db
        Garage[(Garage S3)]:::db
    end

    %% RPC -> DB 数据流
    UserRPC --> MySQL
    UserRPC --> Redis
    ShareRPC --> MySQL
    FileRPC --> Mongo
    FileRPC --> Redis
    FileRPC -- "生成上传凭证" --> Garage
    
    %% 重点：SearchRPC 负责读写 ES
    SearchRPC -- "Index/Search" --> ES

    %% =================6. 特殊链路：文件直传=================
    User -. "Direct Upload (PutObject)" .-> Garage
```
