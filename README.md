# RAG Knowledge Base - 知识库问答系统

基于 Go + Gin + MySQL + Redis + Qdrant 实现的 RAG（检索增强生成）知识库问答系统。项目支持用户注册登录、Session 鉴权、文档上传、向量检索、多轮对话、SSE 流式回答和用户级数据隔离。

## ✨ 功能

- **登录注册**：支持用户注册、登录、退出登录，密码使用 bcrypt 哈希保存
- **Session 鉴权**：登录后服务端创建 Session，并通过 Cookie 识别当前用户
- **用户数据隔离**：文档、聊天记录、上传文件、向量数据都按 `userID` 隔离
- **多格式文档导入**：支持 TXT、Markdown、PDF 文件上传并自动入库
- **文件安全校验**：限制上传文件类型，避免任意类型文件上传
- **文档管理**：支持查看当前用户已上传文档，并删除自己的文档
- **智能文本切分**：固定长度切分 + overlap 重叠窗口，适合中文文档
- **向量存储与检索**：基于 Store 接口抽象，支持 MemoryStore 和 QdrantStore
- **Redis 缓存**：使用 Cache-Aside 模式缓存 Embedding 向量
- **多轮对话记忆**：按用户和会话保存历史消息，让 LLM 感知上下文
- **SSE 流式问答**：DeepSeek API 流式调用，前端逐字显示回答
- **MySQL 持久化**：保存用户、Session、文档元数据和聊天记录
- **YAML 配置管理**：统一使用 `config.yaml` 管理项目配置
- **请求耗时统计**：使用 `slog` 输出 RAG 链路耗时，方便定位性能问题

## 🖼️ 页面说明

### 1. 登录 / 注册页面

访问：

```text
http://localhost:8088/login
```

登录页面使用一个简约的未来感界面，支持：

- 用户登录
- 用户注册
- 登录成功后跳转到首页 `/`
- 未登录访问首页时，前端请求接口会收到 `401`，然后自动跳转到 `/login`

### 2. 知识库问答页面

访问：

```text
http://localhost:8088/
```

首页用于文档上传、文件列表展示、文档删除和知识库问答。首页中的 `/api/upload`、`/api/file`、`/api/chat/stream`、`/api/file/:filename`、`/api/logout` 都需要登录后才能使用。

## 🏗️ 系统架构

```text
Browser
  │
  ├── GET /login
  │      └── login.html：登录 / 注册
  │
  └── GET /
         └── index.html：知识库问答页面

API
  │
  ├── POST /api/register    注册
  ├── POST /api/login       登录，创建 Session Cookie
  │
  └── AuthMiddleware        校验 Session，读取当前 userID
         │
         ├── POST   /api/upload              上传文档
         ├── GET    /api/file                查看当前用户文档
         ├── DELETE /api/file/:filename      删除当前用户文档
         ├── GET    /api/chat/stream         SSE 流式问答
         └── POST   /api/logout              退出登录

RAG Service
  │
  ├── uploads/<userID>/filename              用户文件隔离
  ├── MySQL documents/chat_histories         用户数据隔离
  ├── Qdrant payload user_id                 向量数据隔离
  ├── Redis Embedding Cache                  向量缓存
  └── DeepSeek / SiliconFlow                 LLM 与 Embedding
```

## 📁 项目结构

```text
rag_knowledge/
├── main.go                     # 入口：加载配置 → 初始化 MySQL/Redis/Qdrant → 启动 Web
├── go.mod
├── config.yaml                 # 配置文件
├── start.bat                   # 一键启动 Redis + Qdrant
├── web/
│   ├── templates/
│   │   ├── index.html          # 知识库问答页面
│   │   └── login.html          # 登录 / 注册页面
│   └── static/
│       └── style.css           # 页面样式
└── internal/
    ├── api/
    │   ├── router.go           # Gin 路由注册
    │   └── handler/
    │       ├── auth.go         # 注册、登录、退出、鉴权中间件
    │       ├── upload.go       # 文件上传
    │       ├── delete.go       # 文件删除
    │       ├── chat.go         # SSE 流式问答
    │       └── scanfile.go     # 文件列表
    ├── database/
    │   ├── db.go               # GORM 初始化 + 数据库函数
    │   └── models.go           # User / Session / Document / ChatHistory 模型
    ├── rag/
    │   └── rag.go              # RAG 核心流程：导入、检索、问答、删除向量
    ├── store/
    │   ├── store.go            # Store 接口
    │   ├── memory.go           # MemoryStore 实现
    │   └── qdrant.go           # QdrantStore 实现
    ├── uploads/
    │   └── upload.go           # 文件类型识别与文本提取
    ├── chunker/
    ├── embedder/
    └── llm/
```

## 🔄 核心流程

### 登录流程

```text
用户输入账号密码
  → POST /api/login
  → 根据 username 查询用户
  → bcrypt 校验密码
  → 创建 Session 记录到 MySQL
  → 设置 session Cookie
  → 前端跳转到 /
```

### 鉴权流程

```text
请求受保护接口
  → AuthMiddleware 读取 session Cookie
  → 查询 sessions 表，确认 Session 存在且未过期
  → 查询 users 表，确认用户存在
  → 把 userID / username 写入 Gin Context
  → 后续 handler 通过 getCurrentUserID(c) 获取当前用户
```

### 文档导入流程

```text
用户上传文件
  → 从 Context 获取 userID
  → 检查文件类型
  → 检查当前用户下是否已存在同名文档
  → 保存到 uploads/<userID>/<filename>
  → 提取文本
  → 切分 Chunk
  → Embedding，优先走 Redis 缓存
  → 写入 Qdrant/MemoryStore，并带上 userID
  → MySQL 写入 documents 记录
```

### 问答流程

```text
用户提问
  → 从 Context 获取 userID
  → 根据 userID + session_id 查询历史对话
  → 对问题做 Embedding
  → 只检索当前用户的向量数据
  → 拼接 prompt
  → DeepSeek 流式回答
  → SSE 推送给前端
  → MySQL 保存 user 和 assistant 消息
```

### 删除流程

```text
用户删除文档
  → 从 Context 获取 userID
  → 查询当前用户的文档记录
  → 删除当前文档对应的向量数据
  → 删除 MySQL documents 记录
  → 删除 uploads/<userID>/<filename> 本地文件
```

## 🔌 API 接口

| 方法     | 路径                  | 是否需要登录       | 说明                   |
| -------- | --------------------- | ------------------ | ---------------------- |
| `GET`    | `/`                   | 页面会检查登录状态 | 知识库首页             |
| `GET`    | `/login`              | 否                 | 登录 / 注册页面        |
| `POST`   | `/api/register`       | 否                 | 注册用户               |
| `POST`   | `/api/login`          | 否                 | 登录并创建 Session     |
| `POST`   | `/api/logout`         | 是                 | 退出登录并删除 Session |
| `POST`   | `/api/upload`         | 是                 | 上传文档               |
| `GET`    | `/api/file`           | 是                 | 获取当前用户的文档列表 |
| `DELETE` | `/api/file/:filename` | 是                 | 删除当前用户的指定文档 |
| `GET`    | `/api/chat/stream`    | 是                 | SSE 流式问答           |
| `GET`    | `/static/*filepath`   | 否                 | 静态资源               |

## 🛠 核心设计

### 1. 用户与 Session

`users` 表保存用户信息，密码不明文保存，而是保存 bcrypt 哈希值。

`sessions` 表保存登录态，浏览器通过 Cookie 携带 Session ID。后端每次访问受保护接口时都会通过中间件校验 Session。

### 2. 用户级数据隔离

系统通过 `userID` 隔离不同用户的数据：

- MySQL `documents` 表带 `user_id`
- MySQL `chat_histories` 表带 `user_id`
- 上传文件保存到 `uploads/<userID>/`
- Qdrant 向量 payload 带 `user_id`
- 向量检索时带 `user_id` 过滤条件

这样可以避免 A 用户看到、检索或删除 B 用户的数据。

### 3. Store 接口抽象

```go
type Store interface {
    Add(userID uint, chunkID int, text string, vector []float64) error
    Search(userID uint, queryVec []float64, topK int) ([]VectorChunk, error)
    Delete(chunkIDs []int) error
}
```

MemoryStore 适合开发调试，QdrantStore 适合持久化向量检索。业务层只依赖 `Store` 接口，不关心底层实现。

### 4. Chunk ID 设计

导入文档时，chunk ID 会结合 `userID + filename` 生成，避免不同用户上传同名文件时向量 ID 冲突。

## 🚀 快速开始

### 环境要求

- Go 1.26+
- MySQL 8.0+
- Redis
- Qdrant
- DeepSeek API Key
- 硅基流动 API Key

### 前置准备

```bash
mysql -u root -p
CREATE DATABASE rag_knowledge DEFAULT CHARACTER SET utf8mb4;

.\start.bat
```

### 配置

编辑 `config.yaml`，填写 MySQL、Redis、Qdrant、模型等配置。

API Key 建议通过环境变量设置：

```powershell
$env:DEEPSEEK_API_KEY="your-deepseek-api-key"
$env:SILICONFLOW_API_KEY="your-siliconflow-api-key"
```

### 运行

```bash
./start.bat
go run .
```

浏览器打开：

```text
http://localhost:8088/login
```

注册或登录成功后会进入首页：

```text
http://localhost:8088/
```

## 📦 模块说明

| 模块                      | 职责           | 关键函数                                                               |
| ------------------------- | -------------- | ---------------------------------------------------------------------- |
| `api/router.go`           | 路由注册       | `Setup()`                                                              |
| `api/handler/auth.go`     | 登录注册与鉴权 | `Register()`, `Login()`, `Logout()`, `AuthMiddleware()`                |
| `api/handler/upload.go`   | 文件上传       | `UploadHandler()`                                                      |
| `api/handler/delete.go`   | 文件删除       | `DeleteHandler()`                                                      |
| `api/handler/chat.go`     | 流式问答       | `ChatStream()`                                                         |
| `api/handler/scanfile.go` | 文件列表       | `ScanFile()`                                                           |
| `database`                | MySQL 持久化   | `CreateUser()`, `CreateSession()`, `CreateDocument()`, `SaveMessage()` |
| `rag`                     | RAG 核心流程   | `ImportDoc()`, `AskThreeSteps()`, `DeleteDoc()`                        |
| `store`                   | 向量存储接口   | `Add()`, `Search()`, `Delete()`                                        |
| `uploads`                 | 文件解析       | `DetectType()`, `ExtractText()`, `ProcessFile()`, `IsAllowedFile()`    |
| `embedder`                | 向量化与缓存   | `GetEmbedding()`, `EmbedWithCache()`                                   |
| `llm`                     | LLM 调用       | `CallDeepseekAPIHistory()`                                             |

## 🛠 技术选型

| 组件           | 选择                   | 说明                                       |
| -------------- | ---------------------- | ------------------------------------------ |
| Web 框架       | Gin                    | Go 生态常用 Web 框架                       |
| 数据库         | MySQL + GORM           | 保存用户、Session、文档和聊天记录          |
| 密码哈希       | bcrypt                 | 慢哈希算法，适合保存用户密码               |
| 登录态         | Session + Cookie       | 服务端保存 Session，浏览器通过 Cookie 携带 |
| 缓存           | Redis                  | 缓存 Embedding 向量                        |
| 向量存储       | Memory / Qdrant        | Memory 用于开发，Qdrant 用于持久化检索     |
| Embedding 模型 | BAAI/bge-large-zh-v1.5 | 中文向量效果较好                           |
| LLM            | DeepSeek-Chat          | 支持流式输出                               |
| 流式传输       | SSE                    | 适合服务端向前端单向推送回答               |
| 前端           | 原生 HTML/CSS/JS       | 简单直接，方便展示项目能力                 |

## 📋 版本规划

- **V1** ✅ 核心 RAG 链路（命令行版）
- **V2** ✅ Gin Web 服务 + 前端页面 + SSE 流式问答 + 文件上传
- **V3** ✅ 多轮对话记忆 + MySQL 持久化 + Redis Embedding 缓存 + 配置管理
- **V4** ✅ Store 接口抽象 + Qdrant 向量库 + 文件上传安全修复 + 文档删除
- **V5** ✅ 登录注册 + Session 鉴权 + 用户级数据隔离 + 登录页面
- **后续优化** 🚧 管理后台、公共知识库、安全增强、混合检索、Rerank

## 📄 License

MIT
