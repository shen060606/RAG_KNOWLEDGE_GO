# RAG Knowledge Base - 知识库问答系统

基于 Go 语言从零实现的 RAG（检索增强生成）知识库系统，支持文档导入、向量检索、LLM 智能问答与 Web 交互。

## ✨ 功能

- **多格式文档导入**：支持 TXT、Markdown、PDF 文件
- **拖拽上传**：Web 端拖拽文件即可上传，自动入库
- **智能文本切分**：固定长度切分 + overlap 重叠窗口（中文友好）
- **向量化存储**：调用硅基流动 Embedding API（BAAI/bge-large-zh-v1.5）
- **相似度检索**：余弦相似度 + TopK 召回
- **SSE 流式问答**：调用 DeepSeek API，前端逐字流式显示
- **知识库面板**：左侧实时展示已导入文档列表
- **持久化存储**：文档落盘保存，重启后自动加载

## 🖼️ 界面预览

| 主界面                                      | 问答效果                                |
| ------------------------------------------- | --------------------------------------- |
| ![主界面](/img/rag-index.png) <!-- TODO --> | ![问答](/img/rag-ask.png) <!-- TODO --> |


## 📁 项目结构

```
rag_knowledge/
├── main.go                    # 入口：加载文档 + 启动 Web 服务
├── go.mod
├── uploads/                   # 用户上传文档存放目录
│   ├── 1.txt
│   └── example.pdf
├── web/                       # 前端资源
│   ├── templates/
│   │   └── index.html         # 分栏式聊天界面
│   └── static/
│       └── style.css          # 页面样式
└── internal/
    ├── chunker/
    │   └── chunker.go         # 文本切分（Chunk）
    ├── embedder/
    │   └── embedder.go        # Embedding API 调用
    ├── llm/
    │   └── llm.go             # DeepSeek API 调用（支持回调式流式）
    ├── store/
    │   └── store.go           # 向量存储 + 余弦相似度检索
    ├── rag/
    │   └── rag.go             # RAG 核心流程（ImportDoc + Ask）
    ├── uploads/
    │   └── upload.go          # 文件类型识别 + 文本提取（txt/md/pdf）
    └── api/
        ├── router.go          # Gin 路由注册 + 静态文件
        └── handler/
            ├── upload.go      # POST /api/upload   - 文件上传
            ├── chat.go        # GET  /api/chat/stream - SSE 流式问答
            └── scanfile.go    # GET  /api/file      - 已导入文件列表
```

## 🔄 数据流

```
浏览器 ──HTTP──→ Gin Server ──→ RAG 核心模块 ──→ DeepSeek / 硅基流动 API
   │                                  │
   └── SSE 流式接收回答 ←──────────────┘

问答流程：
  用户提问 → Embedding API → 向量检索(TopK=5) → 拼 Prompt → LLM 流式回答 → SSE 逐字推送前端

文档导入流程：
  拖拽文件 → 保存到 uploads/ → 提取纯文本 → 切分 Chunk → Embedding → 存入向量库
```

## 🔌 API 接口

| 方法   | 路径                     | 说明           | 请求                       | 响应                                  |
| ------ | ------------------------ | -------------- | -------------------------- | ------------------------------------- |
| `GET`  | `/`                      | 首页           | -                          | HTML 页面                             |
| `POST` | `/api/upload`            | 文件上传       | multipart form `file` 字段 | `{"ok":true, "filename":"..."}`       |
| `GET`  | `/api/chat/stream?q=xxx` | SSE 流式问答   | query 参数 `q`             | SSE 事件流 `token` / `done` / `error` |
| `GET`  | `/api/file`              | 已导入文件列表 | -                          | `{"files":["1.txt","xss.pdf"]}`       |
| `GET`  | `/static/*filepath`      | 静态资源       | -                          | CSS / JS 文件                         |

SSE 事件协议：

```
data: {"token":"C"}
data: {"token":"S"}
data: {"token":"R"}
data: {"token":"F"}
data: {"done":""}
```

## 🚀 快速开始

### 环境要求

- Go 1.26+
- DeepSeek API Key（LLM 调用）
- 硅基流动 API Key（Embedding 调用）

### 安装

```bash
git clone <your-repo-url>
cd rag_knowledge
go mod tidy
```

### 配置 API Key

```bash
# Linux / macOS
export DEEPSEEK_API_KEY="your-deepseek-api-key"
export SILICONFLOW_API_KEY="your-siliconflow-api-key"

# Windows (PowerShell)
$env:DEEPSEEK_API_KEY="your-deepseek-api-key"
$env:SILICONFLOW_API_KEY="your-siliconflow-api-key"
```

### 准备知识库（可选）

启动时会自动加载 `uploads/` 目录下已有的文档。也可以启动后通过 Web 页面上传。

```bash
mkdir -p uploads
cp ~/Documents/网络安全笔记.md uploads/
cp ~/Documents/owasp_top10.pdf uploads/
```

### 运行

```bash
go run .
```

浏览器打开 `http://localhost:8088`。

### 使用

1. **上传文档**：左侧拖拽文件到上传区 / 点击「选择文件」
2. **提问**：右下角输入问题，回车发送
3. **查看回答**：AI 逐字流式输出，自动滚动

## 📦 模块说明

| 模块       | 职责           | 关键函数                                                   |
| ---------- | -------------- | ---------------------------------------------------------- |
| `chunker`  | 文本切分       | `SplitText(text, chunkSize, overlap)`                      |
| `embedder` | 文本向量化     | `GetEmbedding(text) → []float64`                           |
| `store`    | 向量存储与检索 | `Add()`, `Search()`, `CosineSimilarity()`                  |
| `llm`      | 大模型调用     | `CallDeepseekAPI(prompt, onToken) → (string, error)`       |
| `rag`      | RAG 核心流程   | `ImportDoc()`, `Ask()`, `AskThreeSteps()`                  |
| `uploads`  | 文件处理       | `DetectType()`, `ExtractText()`, `ProcessFile()`           |
| `api`      | HTTP 层        | `Setup()`, `UploadHandler()`, `ChatStream()`, `ScanFile()` |

## 🛠 技术选型

| 组件           | 选择                        | 说明                                         |
| -------------- | --------------------------- | -------------------------------------------- |
| Embedding 模型 | BAAI/bge-large-zh-v1.5      | 硅基流动 API，中文效果好                     |
| LLM            | DeepSeek-Chat               | 流式输出，性价比高                           |
| Web 框架       | Gin                         | 轻量高性能，Go 生态主流                      |
| 流式传输       | SSE (Server-Sent Events)    | 前端 EventSource 原生支持，比 WebSocket 简单 |
| 向量存储       | 内存切片（`[]VectorChunk`） | 当前规模适用，后续可扩展 Qdrant/Milvus       |
| PDF 解析       | ledongthuc/pdf              | 纯 Go 实现，零 CGo 依赖                      |
| 前端           | 原生 HTML/CSS/JS            | 零框架依赖，简单直接                         |

## 📋 版本规划

- **V1** ✅ 核心 RAG 链路（命令行版）
- **V2** ✅ Gin Web 服务 + 分栏式前端 + SSE 流式问答 + 拖拽上传
- **V3** 🚧 多轮对话记忆 + SQLite 持久化
- **V4** 🚧 Eino 框架重构（对比手动实现）
- **V5** 🚧 混合检索（BM25 + 向量检索）+ Rerank

## 📄 License

MIT
