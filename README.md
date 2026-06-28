# RAG Knowledge Base - 知识库问答系统

基于 Go 语言从零实现的 RAG（检索增强生成）知识库系统。支持文档导入、向量检索、多轮对话、Redis 缓存、Qdrant 向量库与 MySQL 持久化。

## ✨ 功能

- **多格式文档导入**：支持 TXT、Markdown、PDF 文件，拖拽上传自动入库
- **智能文本切分**：固定长度切分 + overlap 重叠窗口（中文友好）
- **向量存储与检索**：基于 Store 接口抽象，支持内存（MemoryStore）和 Qdrant 两种实现，余弦相似度 TopK 召回
- **Redis 缓存**：Cache-Aside 模式缓存 Embedding 向量，降级策略保障服务高可用
- **多轮对话记忆**：Session 级别对话历史，LLM 感知上下文
- **SSE 流式问答**：DeepSeek API 流式调用，前端逐字显示
- **知识库面板**：左侧实时展示已导入文档（文件名 · chunk 数 · 导入时间）
- **MySQL 持久化**：GORM + MySQL，文档元数据 + 对话历史落盘
- **YAML 配置管理**：统一 config.yaml，消除硬编码
- **请求耗时统计**：全链路 slog 耗时日志，快速定位性能瓶颈

## 🖼️ 界面预览

### 1. Web 对话界面
![Web 对话界面](img/v3-rag-ask.png)

### 2. MySQL 数据库（文档表 + 对话历史表）
![MySQL 表结构](img/v3-mysql-doc.png)
![MySQL 表结构](img/v3-mysql-chat.png)

### 3. Redis 缓存（Embedding 向量缓存）
![Redis 缓存](img/v3-redis.png)

## 🏗️ 系统架构

```
                   Browser
                      │
               HTTP / SSE
                      │
                 Gin Router
                      │
               API Handler
        ┌─────────────┼─────────────┐
        │             │             │
   UploadHandler  ChatStream   ScanFile
        │             │             │
        ▼             ▼             ▼
   ┌─────────────────────────────────────┐
   │            RAG Service              │
   │   ImportDoc() / AskThreeSteps()     │
   └──────┬──────┬──────┬──────┬────────┘
          │      │      │      │
     Chunker  Embedder  Store   LLM
     (固定长度  (Cache-  (接口)  (DeepSeek
      +overlap) Aside)    │      SSE流式)
                    ┌─────┴─────┐
                    │           │
              MemoryStore  QdrantStore
              (内存切片)   (REST API)
                    │           │
              Redis Cache   Qdrant HTTP
              (向量缓存)    (持久化检索)
                    │           │
              SiliconFlow    MySQL
              (Embedding)   (GORM)
```

## 📁 项目结构

```
rag_knowledge/
├── main.go                     # 入口：加载配置 → 初始化 MySQL/Redis/Qdrant → 启动 Web
├── go.mod
├── config.yaml                 # 配置文件（端口、数据库、Redis、模型参数）
├── start.bat                   # 一键启动 Redis + Qdrant
├── config/
│   └── config.go               # 配置解析（yaml → struct）
├── uploads/                    # 用户上传文档存放目录
├── web/                        # 前端资源
│   ├── templates/
│   │   └── index.html          # 分栏式聊天界面
│   └── static/
│       └── style.css           # 页面样式
└── internal/
    ├── chunker/
    │   └── chunker.go           # 文本切分（Chunk）
    ├── embedder/
    │   ├── embedder.go          # Embedding API 调用
    │   └── cache.go             # Redis 缓存层（Cache-Aside，带降级）
    ├── llm/
    │   └── llm.go               # DeepSeek API 调用（单轮 / 多轮 / 流式）
    ├── store/
    │   ├── store.go             # Store 接口 + VectorChunk + CosineSimilarity
    │   ├── memory.go            # MemoryStore：内存切片实现
    │   └── qdrant.go            # QdrantStore：REST API 实现（持久化 + HNSW 检索）
    ├── rag/
    │   └── rag.go               # RAG 核心流程（全局唯一 Chunk ID）
    ├── uploads/
    │   └── upload.go            # 文件类型识别 + 文本提取（txt/md/pdf）
    ├── database/
    │   ├── db.go                # GORM 初始化 + CRUD + DocumentExists
    │   └── models.go            # Document / ChatHistory 模型
    └── api/
        ├── router.go            # Gin 路由注册 + 静态文件
        └── handler/
            ├── upload.go        # POST /api/upload     - 文件上传（含去重）
            ├── chat.go          # GET  /api/chat/stream - SSE 流式问答（多轮 + 耗时统计）
            └── scanfile.go      # GET  /api/file        - 已导入文件列表（查 MySQL）
```

## 🔄 数据流

```
问答流程：
  用户提问 → Redis 缓存? → Embedding API → Qdrant/Memory 向量检索(TopK=5)
  → MySQL 查历史对话 → 拼 Messages → LLM 流式回答
  → SSE 逐字推送前端 → MySQL 存回答

  每步耗时通过 slog 输出：
  [Ask] embedding+search=115ms LLM首token=520ms total=1.65s

文档导入流程：
  拖拽文件 → 保存到 uploads/ → DocumentExists 去重检查
  → 提取纯文本 → 切分 Chunk → 文件名 hash 生成全局唯一 ID
  → Embedding（走 Redis 缓存）→ 存入 Qdrant/MemoryStore → MySQL 记录元数据
```

## 🔌 API 接口

| 方法 | 路径 | 说明 | 关键参数 |
|------|------|------|---------|
| `GET` | `/` | 首页 | - |
| `POST` | `/api/upload` | 文件上传 | multipart form `file` |
| `GET` | `/api/chat/stream` | SSE 流式问答 | `q` (问题), `session_id` (会话ID) |
| `GET` | `/api/file` | 已导入文件列表 | - |
| `GET` | `/static/*filepath` | 静态资源 | - |

## 🛠 核心设计

### Store 接口抽象

```
Store 接口：Add() + Search()
    ├── MemoryStore   — 开发阶段，内存切片，遍历 + 余弦相似度
    └── QdrantStore   — 生产环境，REST API，HNSW 索引检索

切换方式：main.go 一行代码
  vs := store.NewMemoryStore()                    // 内存
  vs, _ := store.NewQdrantStore("127.0.0.1", 6333) // Qdrant
```

**设计理由**：调用方只依赖接口，不依赖具体实现。后续可无缝扩展 Milvus、Pinecone 等向量库，业务代码零改动。

### Embedding 缓存（Cache-Aside 模式）

```
文本 → MD5 → Redis GET
              ├── 命中 → 直接返回（<1ms）
              └── 未命中 → 调 API（~300ms）→ Redis SET（TTL 24h）
              └── Redis 不可用 → 降级跳过缓存，调 API
```

**设计理由**：相同文本不重复调 API，降级策略保证缓存故障不影响主流程。

### 多轮对话（Session 隔离）

```
前端加载 → 生成唯一 SessionID
每轮对话 → MySQL 存 user + assistant 消息
下次提问 → 按 SessionID 查历史 → 拼接 Messages → LLM 感知上下文
```

### SSE vs WebSocket

选用 SSE 的理由：问答场景是单向流（服务端 → 前端），SSE 基于 HTTP，前端 EventSource 原生支持，无需引入 WebSocket 库。HTTP/2 下 SSE 可复用连接，性能不输 WebSocket。

### Chunk 全局唯一 ID

```
filename → MD5 → 前 4 字节 → 文档序号
chunkID = 文档序号 × 100000 + chunk 本地编号

不同文档的 chunk ID 永不冲突
```

**设计理由**：Qdrant 用 point ID 做向量主键，不同文档的 chunk 从 0 开始编号会导致相互覆盖。用文件名 hash 生成全局唯一 ID 彻底解决冲突。

## 🚀 快速开始

### 环境要求

- Go 1.26+
- MySQL 8.0+
- Redis（本地运行）
- Qdrant（本地运行，可选）
- DeepSeek API Key（LLM 调用）
- 硅基流动 API Key（Embedding 调用）

### 前置准备

```bash
# 1. 启动 MySQL，创建数据库
mysql -u root -p
CREATE DATABASE rag_knowledge DEFAULT CHARACTER SET utf8mb4;

# 2. 一键启动 Redis + Qdrant
#    编辑 start.bat，修改 Redis 和 Qdrant 的安装路径
.\start.bat
```

### 安装

```bash
git clone <your-repo-url>
cd rag_knowledge
go mod tidy
```

### 配置

编辑 `config.yaml` 填入你的 MySQL、Redis 参数。

```bash
# API Key 通过环境变量设置
export DEEPSEEK_API_KEY="your-deepseek-api-key"
export SILICONFLOW_API_KEY="your-siliconflow-api-key"

# Windows (PowerShell)
$env:DEEPSEEK_API_KEY="your-deepseek-api-key"
$env:SILICONFLOW_API_KEY="your-siliconflow-api-key"
```

### 运行

```bash
go run .
# 浏览器打开 http://localhost:8088
```

## 📦 模块说明

| 模块 | 职责 | 关键函数 |
|------|------|---------|
| `chunker` | 文本切分 | `SplitText(text, chunkSize, overlap)` |
| `embedder` | 向量化 + Redis 缓存 | `GetEmbedding()`, `EmbedWithCache()`, `InitRedis()` |
| `store` | 向量存储接口 + 双实现 | `Store interface`, `MemoryStore`, `QdrantStore` |
| `llm` | DeepSeek API 调用 | `CallDeepseekAPI()`, `CallDeepseekAPIHistory()`, `dorequest()` |
| `rag` | RAG 核心流程 | `ImportDoc()`, `Ask()`, `AskThreeSteps()` |
| `uploads` | 文件解析 | `DetectType()`, `ExtractText()`, `ProcessFile()` |
| `database` | MySQL 持久化 | `InitDB()`, `CreateDocument()`, `SaveMessage()`, `DocumentExists()` |
| `api` | HTTP 层 + 路由 | `Setup()`, `UploadHandler()`, `ChatStream()`, `ScanFile()` |
| `config` | 配置管理 | `Load()`, `Cfg` 全局实例 |

## 🛠 技术选型

| 组件 | 选择 | 说明 |
|------|------|------|
| Embedding 模型 | BAAI/bge-large-zh-v1.5 | 硅基流动 API，中文效果好，1024 维 |
| LLM | DeepSeek-Chat | 流式输出，支持多轮对话 |
| Web 框架 | Gin | 轻量高性能，Go 生态主流 |
| 流式传输 | SSE (Server-Sent Events) | 单向推送，前端 EventSource 原生支持 |
| 数据库 | MySQL 8.0 + GORM | 文档元数据与对话历史持久化 |
| 缓存 | Redis | Cache-Aside 模式，Embedding 向量缓存，TTL 24h |
| 向量存储 | Memory / Qdrant (REST API) | Store 接口抽象，HNSW 索引检索 |
| PDF 解析 | ledongthuc/pdf | 纯 Go 实现，零 CGo 依赖 |
| 前端 | 原生 HTML/CSS/JS | 零框架依赖 |
| 配置 | YAML + struct | 消除硬编码，统一管理 |
| 日志 | log/slog | Go 标准库，结构化日志 |

## 📋 版本规划

- **V1** ✅ 核心 RAG 链路（命令行版）
- **V2** ✅ Gin Web 服务 + 分栏前端 + SSE 流式问答 + 拖拽上传
- **V3** ✅ 多轮对话记忆 + MySQL 持久化 + Redis Embedding 缓存 + 配置管理
- **V4** ✅ Store 接口抽象 + Qdrant 向量库 + 请求耗时统计 + 一键启动脚本
- **V5** 🚧 Eino Workflow 编排 + 混合检索（BM25 + 向量）+ Rerank

## 📄 License

MIT
