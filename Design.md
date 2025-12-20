# **在线共享白板系统 \- 完整架构与开发设计文档**

## **1\. 系统概览 (System Overview)**

本系统旨在构建一个高性能、可扩展的在线协作白板平台。系统采用 **前后端分离** 架构，后端基于 **Go-zero** 微服务框架，前端采用 **Vue 3 \+ Fabric.js**，全栈环境通过 **Docker Compose** 进行容器化编排。

### **1.1 技术栈选型**

| 领域 | 技术/工具 | 说明 |
| :---- | :---- | :---- |
| **前端** | Vue 3, TypeScript, Vite | 现代 Web 开发框架 |
| **白板引擎** | **Fabric.js** | 处理 Canvas 对象模型、序列化与交互 |
| **状态管理** | Pinia | 前端状态管理 |
| **后端框架** | **Go-zero** | 高性能微服务框架 (API \+ RPC) |
| **实时通信** | WebSocket (Gorilla) | 双向实时信令传输 |
| **数据库** | MySQL 8.0 | 核心业务数据持久化 |
| **缓存/消息** | **Redis** | 缓存、Pub/Sub 消息总线、在线状态存储 |
| **对象存储** | **RustFS** | 高性能分布式文件存储 (图片, 快照) |
| **网关/代理** | Nginx | 反向代理、负载均衡、静态资源托管 |
| **容器编排** | Docker Compose | 开发与私有化部署编排 |

### **1.2 系统架构图**

```mermaid
graph TD  
    User\[用户 Browser\]  
      
    subgraph Frontend \[前端容器\]  
        View\[Vue 页面\]  
        Canvas\[Fabric.js 画板\]  
    end

    subgraph Infrastructure \[基础设施 (Docker Compose)\]  
        Nginx\[Nginx 网关\]  
          
        subgraph Microservices \[Go-zero 微服务集群\]  
            direction TB  
            UserSvc\[用户服务 (API+RPC)\]  
            RoomSvc\[房间服务 (API+RPC)\]  
            FileSvc\[文件服务 (API+RPC)\]  
            HistSvc\[历史服务 (API+RPC)\]  
            WSSvc\[WebSocket 服务 (Hub)\]  
        end

        subgraph DataLayer \[数据存储层\]  
            Redis\[(Redis Pub/Sub)\]  
            MySQL\[(MySQL DB)\]  
            RustFS\[(RustFS Storage)\]  
        end  
    end

    %% 交互流  
    User \<--\>|HTTP/WS| Nginx  
    View \--\> Canvas  
      
    Nginx \--\>|/api/v1/user| UserSvc  
    Nginx \--\>|/api/v1/room| RoomSvc  
    Nginx \--\>|/api/v1/upload| FileSvc  
    Nginx \--\>|/api/v1/history| HistSvc  
    Nginx \<--\>|/ws| WSSvc

    %% 服务间调用 (RPC)  
    WSSvc \-.-\>|gRPC Check Token| UserSvc  
    WSSvc \-.-\>|gRPC Check Room| RoomSvc  
    WSSvc \-.-\>|gRPC Save Msg| HistSvc  
    FileSvc \--\> RustFS  
      
    %% 数据层交互  
    UserSvc & RoomSvc & HistSvc \--\> MySQL  
    WSSvc \<--\> Redis
```

## **2\. 后端微服务详细设计 (Microservice Design)**

后端采用 **Go-zero** 框架，遵循 **API (对外 HTTP)** 与 **RPC (对内 业务逻辑)** 分离的设计原则。

### **2.1 服务模块划分**

#### **A. 用户服务 (User Service)**

负责用户的身份认证与信息管理。

* **API 职责**: 登录、注册、用户信息查询。  
* **RPC 职责**: Token 校验、用户数据 CRUD。  
* **Proto 定义**: GetUser, CheckUser。

#### **B. 房间服务 (Room Service)**

负责房间的生命周期管理。

* **API 职责**: 创建房间、获取房间列表、加入权限检查。  
* **RPC 职责**: 房间状态管理、密码校验。  
* **Proto 定义**: GetRoomInfo, CreateRoom。

#### **C. 文件服务 (File Service)**

负责对接底层存储 RustFS，处理二进制文件。

* **API 职责**: 头像上传、聊天图片上传、白板截图上传。  
* **RPC 职责**: 文件元数据落库 (files 表)。  
* **交互**: 接收 Multipart Form \-\> 流式上传至 RustFS \-\> 返回 URL。

#### **D. 历史服务 (History Service)**

负责高频/大容量数据的持久化。

* **API 职责**: 获取历史聊天记录、保存/获取白板快照。  
* **RPC 职责**: 异步写入聊天消息、白板 JSON 数据存取。  
* **Proto 定义**: AddChatMessage, GetSnapshot。

#### **E. 实时通信服务 (WebSocket Service)**

独立运行的 Go 服务，不使用 go-zero 代码生成，但调用其他服务的 RPC。

* **核心功能**: 维护长连接、心跳保活、Redis 消息订阅与广播。

### **2.2 Nginx 网关配置**
```
http {  
    upstream user\_api { server user-api:8888; }  
    upstream room\_api { server room-api:8889; }  
    upstream file\_api { server file-api:8891; }  
    upstream hist\_api { server hist-api:8892; }  
    upstream ws\_server { server ws-server:8890; }

    server {  
        listen 80;  
          
        \# API 路由分发  
        location /api/v1/user/ { proxy\_pass http://user\_api; }  
        location /api/v1/room/ { proxy\_pass http://room\_api; }  
        location /api/v1/upload {   
            proxy\_pass http://file\_api;   
            client\_max\_body\_size 50M; \# 允许大文件  
        }  
        location /api/v1/history/ { proxy\_pass http://hist\_api; }

        \# WebSocket 升级  
        location /ws {  
            proxy\_pass http://ws\_server;  
            proxy\_http\_version 1.1;  
            proxy\_set\_header Upgrade $http\_upgrade;  
            proxy\_set\_header Connection "upgrade";  
        }  
          
        \# 前端静态资源  
        location / { root /usr/share/nginx/html; try\_files $uri $uri/ /index.html; }  
    }  
}
```

## **3\. 核心业务流程与逻辑**

### **3.1 实时协同 (绘画同步)**

核心原则：**后端不处理图像像素，只转发 Fabric.js 的 JSON 指令对象**。

1. **动作捕获**: 用户在 Canvas 绘制一条线，Fabric.js 触发 object:added。  
2. **序列化**: 前端将对象转为 JSON: {"type":"path", "path":\[...\], "stroke":"red"}。  
3. **发送**: 通过 WebSocket 发送 DRAW 类型消息。  
4. **广播**: WebSocket 服务收到消息 \-\> Publish 到 Redis Channel room:{id}。  
5. **接收与渲染**: 其他用户的 WS 服务收到 Redis 消息 \-\> Push 给前端 \-\> 前端调用 canvas.loadFromJSON() 增量渲染。

### **3.2 白板持久化 (快照机制)**

当新用户加入或房间关闭时，需要保存/恢复白板状态。

1. **保存 (Snapshot)**:  
   * 房主手动点击保存或定时触发。  
   * 前端调用 canvas.toJSON() 获取全量 JSON。  
   * 调用 History API \-\> 存入 MySQL snapshots 表 (若数据过大则存入 RustFS 并在 DB 存 URL)。  
2. **恢复 (Restore)**:  
   * 用户进入房间 \-\> 调用 History API 获取最新快照。  
   * 前端渲染快照 \-\> 建立 WS 连接 \-\> 接收快照之后产生的新增指令。

## **4\. 数据库设计 (MySQL)**

### **4.1 用户表 (users)**

| 字段 | 类型 | 说明 |
| :---- | :---- | :---- |
| id | BIGINT | PK |
| username | VARCHAR | 账号 |
| password | VARCHAR | Hash |
| avatar | VARCHAR | RustFS URL |

### **4.2 房间表 (rooms)**

| 字段 | 类型 | 说明 |
| :---- | :---- | :---- |
| id | CHAR(36) | UUID |
| name | VARCHAR | 房间名 |
| type | TINYINT | 0:公开, 1:加密 |
| password | VARCHAR | 房间密码 |
| status | TINYINT | 1:活跃, 0:关闭 |

### **4.3 快照表 (snapshots) \- 归属 History 服务**

| 字段 | 类型 | 说明 |
| :---- | :---- | :---- |
| id | BIGINT | PK |
| room\_id | CHAR(36) | FK |
| data\_json | LONGTEXT | 全量 JSON 数据 |
| preview\_url | VARCHAR | 缩略图 (RustFS URL) |
| version | INT | 版本号 |

### **4.4 聊天与文件表**

* chat\_messages: 存储文本消息内容。  
* files: 存储上传到 RustFS 的文件元数据（文件名、大小、上传者、URL）。

## **5\. 通信协议设计 (WebSocket)**

**Endpoint**: ws://host/ws?room\_id={uuid}\&token={jwt}

### **5.1 Redis 结构**

* **Pub/Sub**: room\_channel:{room\_id} (广播通道)  
* **Set**: room\_users:{room\_id} (在线用户列表)

### **5.2 消息 Payload 定义**

**1\. 绘图 (DRAW)**
```json
{  
  "type": "DRAW",  
  "payload": {  
    "action": "add",   
    "object": { "type": "rect", "left": 100, "top": 100, "fill": "red" }   
  }  
}
```

**2\. 聊天 (CHAT)**
```json
{  
  "type": "CHAT",  
  "payload": {  
    "content": "http://rustfs/bucket/image.png",  
    "msg\_type": "image"   
  }  
}
```

## **6\. 基础设施与部署 (Docker Compose)**

所有服务通过 Docker Compose 一键拉起，处于同一网络 whiteboard-net 中。

### **服务编排清单**

1. **mysql**: 业务数据存储，挂载 mysql\_data 卷。  
2. **redis**: 消息总线与缓存，挂载 redis\_data 卷。  
3. **rustfs**: 对象存储服务，提供 S3 兼容 API 或 HTTP API，挂载 rustfs\_data 卷。  
4. **user-rpc, room-rpc, file-rpc, hist-rpc**: Go-zero RPC 服务集群。  
5. **user-api, room-api, file-api, hist-api**: Go-zero API 服务集群。  
6. **ws-server**: 独立的 WebSocket 服务。  
7. **frontend**: 包含 Nginx 配置的前端静态资源容器。  
8. **nginx-gateway**: 主网关，暴露 80/443 端口，代理上述所有服务。