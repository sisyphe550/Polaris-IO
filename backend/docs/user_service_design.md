# User 服务完整设计解析

## 1. 整体架构

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              User 服务                                       │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────────┐         RPC 调用        ┌─────────────────────┐    │
│  │     user-api        │  ───────────────────>   │    user-rpc         │    │
│  │   (HTTP Gateway)    │                         │   (业务核心)         │    │
│  │   Port: 1001        │                         │   Port: 2001        │    │
│  └─────────────────────┘                         └─────────────────────┘    │
│           │                                               │                 │
│           │                                               │                 │
│           ▼                                               ▼                 │
│  ┌─────────────────────┐                         ┌─────────────────────┐    │
│  │   JWT 认证中间件     │                         │      MySQL          │    │
│  └─────────────────────┘                         │   polaris_user      │    │
│                                                  │   ├── user          │    │
│                                                  │   └── user_quota    │    │
│                                                  └─────────────────────┘    │
│                                                           │                 │
│                                                           ▼                 │
│                                                  ┌─────────────────────┐    │
│                                                  │       Redis         │    │
│                                                  │   (sqlc 缓存)       │    │
│                                                  └─────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. 数据模型

### 2.1 user 表

```sql
CREATE TABLE `user` (
    `id`          bigint unsigned NOT NULL AUTO_INCREMENT,
    `mobile`      varchar(11)  NOT NULL DEFAULT '' COMMENT '手机号',
    `password`    varchar(64)  NOT NULL DEFAULT '' COMMENT '密码(MD5)',
    `name`        varchar(64)  NOT NULL DEFAULT '' COMMENT '昵称',
    `avatar`      varchar(255) NOT NULL DEFAULT '' COMMENT '头像URL',
    `info`        varchar(255) NOT NULL DEFAULT '' COMMENT '个人简介',
    `version`     bigint unsigned NOT NULL DEFAULT '0' COMMENT '乐观锁',
    `del_state`   tinyint NOT NULL DEFAULT '0' COMMENT '0:正常 1:删除',
    `delete_time` bigint unsigned NOT NULL DEFAULT '0',
    `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_mobile` (`mobile`, `del_state`)
);
```

### 2.2 user_quota 表

```sql
CREATE TABLE `user_quota` (
    `id`          bigint unsigned NOT NULL AUTO_INCREMENT,
    `user_id`     bigint unsigned NOT NULL COMMENT '用户ID',
    `total_size`  bigint unsigned NOT NULL DEFAULT '0' COMMENT '总容量(字节)',
    `used_size`   bigint unsigned NOT NULL DEFAULT '0' COMMENT '已用容量(字节)',
    `version`     bigint unsigned NOT NULL DEFAULT '0' COMMENT '乐观锁',
    `del_state`   tinyint NOT NULL DEFAULT '0',
    `delete_time` bigint unsigned NOT NULL DEFAULT '0',
    `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_delete` (`user_id`, `delete_time`)
);
```

---

## 3. API 接口设计

### 3.1 对外 HTTP 接口 (user-api)

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| POST | `/usercenter/v1/user/register` | ❌ | 用户注册 |
| POST | `/usercenter/v1/user/login` | ❌ | 用户登录 |
| GET | `/usercenter/v1/user/info` | ✅ JWT | 获取当前用户信息 |
| GET | `/usercenter/v1/user/quota` | ✅ JWT | 获取当前用户配额 |

### 3.2 内部 RPC 接口 (user-rpc)

| 方法 | 调用者 | 说明 |
|------|--------|------|
| `Register` | user-api | 用户注册（事务：创建用户+配额） |
| `Login` | user-api | 用户登录 |
| `GenerateToken` | 内部 | 生成 JWT Token |
| `GetUserInfo` | user-api, file-rpc | 获取用户信息 |
| `GetUserQuota` | user-api, file-api | 获取用户配额 |
| `DeductQuota` | file-api | 扣减配额（上传前） |
| `RefundQuota` | file-api | 退还配额（删除时） |

---

## 4. 核心业务流程

### 4.1 注册流程

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           Register 流程                                  │
└─────────────────────────────────────────────────────────────────────────┘

用户请求 ──> user-api ──> user-rpc
                              │
                              ▼
                    ┌─────────────────┐
                    │ 1. 检查手机号    │  SELECT * FROM user WHERE mobile = ?
                    │    是否已注册    │
                    └────────┬────────┘
                             │ 未注册
                             ▼
                    ┌─────────────────┐
                    │ 2. 开启事务     │
                    │    BEGIN        │
                    └────────┬────────┘
                             │
                    ┌────────┴────────┐
                    │                 │
                    ▼                 ▼
           ┌──────────────┐  ┌──────────────┐
           │ 插入 user    │  │ 插入 quota   │
           │ password=MD5 │  │ total=10GB   │
           │ name=随机    │  │ used=0       │
           └──────────────┘  └──────────────┘
                    │                 │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │ 3. COMMIT       │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │ 4. 生成 JWT     │  userId 写入 claims
                    │    7天有效期    │
                    └────────┬────────┘
                             │
                             ▼
                       返回 Token
```

### 4.2 登录流程

```
用户请求 ──> user-api ──> user-rpc
                              │
                              ▼
                    ┌─────────────────┐
                    │ 1. 根据手机号   │  SELECT * FROM user WHERE mobile = ?
                    │    查询用户     │
                    └────────┬────────┘
                             │ 找到用户
                             ▼
                    ┌─────────────────┐
                    │ 2. 校验密码     │  MD5(input) == stored_password
                    └────────┬────────┘
                             │ 密码正确
                             ▼
                    ┌─────────────────┐
                    │ 3. 生成 JWT     │
                    └────────┬────────┘
                             │
                             ▼
                       返回 Token
```

### 4.3 配额扣减流程（被 file-api 调用）

```sql
-- DeductQuota 的原子 SQL（单条语句，无需先 SELECT）
UPDATE user_quota 
SET used_size = used_size + ?, 
    version = version + 1 
WHERE user_id = ? 
  AND del_state = 0 
  AND used_size + ? <= total_size  -- 防止超额
```

**关键点**：
- 单条 SQL，原子操作
- `used_size + ? <= total_size` 确保不超额
- 返回 `RowsAffected = 0` 表示配额不足

---

## 5. 安全设计

### 5.1 密码存储

```go
// 使用 MD5 加密存储（简化版）
user.Password = tool.Md5ByString(in.Password)
```

### 5.2 JWT Token

```go
claims := make(jwt.MapClaims)
claims["exp"] = iat + 604800    // 7天过期
claims["iat"] = iat             // 签发时间
claims["userId"] = userId       // 用户ID
token := jwt.New(jwt.SigningMethodHS256)
```

---

## 6. 可优化点分析

### 🔴 高优先级（建议尽快优化）

#### 6.1 密码加密方式不安全

**当前问题**：
```go
user.Password = tool.Md5ByString(in.Password)  // 纯 MD5，容易被彩虹表破解
```

**优化方案**：使用 bcrypt 或 Argon2

```go
import "golang.org/x/crypto/bcrypt"

// 注册时
hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// 登录时
err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(inputPassword))
```

#### 6.2 缺少登录限流/防暴力破解

**当前问题**：没有限制登录尝试次数

**优化方案**：
```go
// 使用 Redis 记录失败次数
key := fmt.Sprintf("login:fail:%s", mobile)
count, _ := redis.Incr(key)
redis.Expire(key, 15*time.Minute)

if count > 5 {
    return errors.New("登录失败次数过多，请15分钟后重试")
}
```

#### 6.3 缺少 Token 刷新机制

**当前问题**：Token 7天后直接过期，用户需要重新登录

**优化方案**：双 Token 机制
```
AccessToken:  有效期 2 小时，用于接口认证
RefreshToken: 有效期 7 天，用于刷新 AccessToken

新增接口：POST /usercenter/v1/user/refresh
```

---

### 🟡 中优先级（后续版本优化）

#### 6.4 用户信息缺少缓存

**当前问题**：每次 `GetUserInfo` 都查数据库

**优化方案**：

```go
// 使用 go-zero 的 sqlc 缓存（配置中已有，但 model 未启用）
func NewUserModel(conn sqlx.SqlConn, c cache.CacheConf) UserModel {
    return &customUserModel{
        defaultUserModel: newUserModel(conn, c),  // 传入缓存配置
    }
}
```

或使用 Redis 手动缓存：
```go
func (l *GetUserInfoLogic) GetUserInfo(in *pb.GetUserInfoReq) (*pb.GetUserInfoResp, error) {
    key := fmt.Sprintf("user:info:%d", in.UserId)
    
    // 1. 先查缓存
    cached, err := l.svcCtx.Redis.Get(key)
    if err == nil && cached != "" {
        var user pb.User
        json.Unmarshal([]byte(cached), &user)
        return &pb.GetUserInfoResp{User: &user}, nil
    }
    
    // 2. 查数据库
    user, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(in.UserId))
    
    // 3. 写入缓存（5分钟过期）
    l.svcCtx.Redis.Setex(key, json.Marshal(user), 300)
    
    return &pb.GetUserInfoResp{User: user}, nil
}
```

#### 6.5 配额操作缺少缓存（你之前问的问题）

**当前问题**：每次上传都查数据库

**优化方案**：Redis 缓存配额
```go
// 缓存结构
HSET user:quota:123 total 10737418240
HSET user:quota:123 used  1073741824

// 扣减配额（Lua 脚本保证原子性）
EVAL "
    local used = tonumber(redis.call('HGET', KEYS[1], 'used'))
    local total = tonumber(redis.call('HGET', KEYS[1], 'total'))
    if used + ARGV[1] > total then return -1 end
    redis.call('HINCRBY', KEYS[1], 'used', ARGV[1])
    return 0
" 1 user:quota:123 1024
```

#### 6.6 缺少用户状态管理

**当前问题**：没有用户状态字段（正常/禁用/待验证）

**优化方案**：
```sql
ALTER TABLE user ADD COLUMN `status` tinyint NOT NULL DEFAULT 1 
COMMENT '状态: 0=禁用 1=正常 2=待验证';
```

---

### 🟢 低优先级（锦上添花）

#### 6.7 缺少手机号验证

**当前问题**：注册时不验证手机号真实性

**优化方案**：集成短信验证码
```
1. POST /usercenter/v1/sms/send   - 发送验证码
2. POST /usercenter/v1/user/register - 携带验证码注册
```

#### 6.8 缺少用户信息修改接口

**当前问题**：注册后无法修改昵称、头像等

**优化方案**：
```go
// 新增 RPC 方法
rpc UpdateUserInfo(UpdateUserInfoReq) returns(UpdateUserInfoResp);

message UpdateUserInfoReq {
    int64  userId = 1;
    string name = 2;
    string avatar = 3;
    string info = 4;
}
```

#### 6.9 缺少用户注销功能

**当前问题**：用户无法注销账号

**优化方案**：软删除用户和相关数据

---

## 7. 优化优先级总结

| 优先级 | 优化项 | 影响 | 复杂度 |
|--------|--------|------|--------|
| 🔴 高 | 密码加密升级 bcrypt | 安全性 | 低 |
| 🔴 高 | 登录限流/防暴力破解 | 安全性 | 低 |
| 🔴 高 | Token 刷新机制 | 用户体验 | 中 |
| 🟡 中 | 用户信息缓存 | 性能 | 低 |
| 🟡 中 | 配额 Redis 缓存 | 性能 | 中 |
| 🟡 中 | 用户状态管理 | 功能完整性 | 低 |
| 🟢 低 | 手机号验证 | 安全性 | 高（需短信服务） |
| 🟢 低 | 用户信息修改 | 功能完整性 | 低 |
| 🟢 低 | 用户注销 | 功能完整性 | 中 |

---

## 8. 当前设计的优点

| 优点 | 说明 |
|------|------|
| ✅ 事务保证 | 注册时用户和配额在同一事务中创建 |
| ✅ 乐观锁 | 配额更新使用 version 字段防止并发问题 |
| ✅ 原子配额操作 | DeductQuota 单条 SQL 防止超额 |
| ✅ 软删除 | 支持数据恢复，符合企业需求 |
| ✅ 链路追踪 | 集成 Jaeger，便于问题排查 |
| ✅ 监控就绪 | Prometheus metrics 已配置 |

---
