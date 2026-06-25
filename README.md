# RAG Knowledge Base - 网络安全知识库

基于 Go 语言从零实现的 RAG（检索增强生成）知识库系统，支持文档导入、向量检索与 LLM 智能问答。

## ✨ 功能

- **多格式文档导入**：支持 TXT、Markdown、PDF 文件
- **智能文本切分**：固定长度切分 + overlap 重叠窗口（中文友好）
- **向量化存储**：调用硅基流动 Embedding API（BAAI/bge-large-zh-v1.5）
- **相似度检索**：余弦相似度 + TopK 召回
- **LLM 问答**：调用 DeepSeek API，流式输出回答
- **命令行交互**：支持提问和运行时导入新文件

## 📁 项目结构

```
rag_knowledge/
├── main.go                    # 入口：初始化 + 交互循环
├── go.mod
├── uploads/                   # 知识库文档存放目录
│   └── upload.go             # 文件类型识别 + 文本提取
└── internal/
    ├── chunker/
    │   └── chunker.go         # 文本切分（Chunk）
    ├── embedder/
    │   └── embedder.go        # Embedding API 调用
    ├── llm/
    │   └── llm.go             # DeepSeek LLM API 调用
    ├── store/
    │   └── store.go           # 向量存储 + 余弦相似度检索
    └── rag/
        └── rag.go             # RAG 核心流程（ImportDoc + Ask）
```

## 🔄 数据流

```
用户提问（问题）
      │
      ▼
┌─────────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
│ Embedding   │ →  │ 向量检索  │ →  │ 拼接 Prompt │ →  │ LLM 回答 │
│ API         │    │ TopK=5   │    │ 上下文+问题 │    │ (流式)   │
└─────────────┘    └──────────┘    └──────────┘    └──────────┘
      ↑                ↑
      │                │
  问题向量         向量库中的文档向量

文档导入流程：
  文件(txt/md/pdf) → 提取纯文本 → 切分Chunk → Embedding → 存入向量库
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
export DEEPSEEK_API_KEY="your-deepseek-api-key"
export SILICONFLOW_API_KEY="your-siliconflow-api-key"
```

### 准备知识库

将文档放入 `uploads/` 目录：

```bash
mkdir -p uploads
cp ~/Documents/网络安全笔记.md uploads/
cp ~/Documents/owasp_top10.pdf uploads/
```

### 运行

```bash
go run .
```

### 使用

```
===== RAG Knowledge System =====
请输入您的问题（输入 exit 退出）：

> 什么是CSRF攻击？
...（流式输出回答）...

> upload
请输入文件路径: uploads/xss.md
✓ 已导入: uploads/xss.md

> exit
```

## 📦 模块说明

| 模块 | 职责 | 关键函数 |
|------|------|---------|
| `chunker` | 文本切分 | `SplitText(text, chunkSize, overlap)` |
| `embedder` | 文本向量化 | `GetEmbedding(text) → []float64` |
| `store` | 向量存储与检索 | `Add()`, `Search()`, `CosineSimilarity()` |
| `llm` | 大模型调用 | `CallDeepseekAPI(prompt) → (string, error)` |
| `rag` | RAG 核心流程 | `ImportDoc()`, `Ask()` |
| `uploads` | 文件处理 | `DetectType()`, `ExtractText()`, `ProcessFile()` |

## 🛠 技术选型

| 组件 | 选择 | 说明 |
|------|------|------|
| Embedding 模型 | BAAI/bge-large-zh-v1.5 | 硅基流动 API，中文效果好 |
| LLM | DeepSeek-Chat | 流式输出，性价比高 |
| 向量存储 | 内存切片（`[]VectorChunk`） | 当前规模适用，后续可扩展 Qdrant/Milvus |
| PDF 解析 | ledongthuc/pdf | 纯 Go 实现，零 CGo 依赖 |

## 📋 版本规划

- **V1** ✅ 核心 RAG 链路（命令行版）
- **V2** 🚧 Gin Web 服务 + 前端页面
- **V3** 🚧 多轮对话记忆
- **V4** 🚧 Eino 框架重构
- **V5** 🚧 混合检索（BM25 + 向量检索）+ Rerank

## 📄 License

MIT
