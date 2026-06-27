package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/shen060606/rag_koowledge_go/config"
	"github.com/shen060606/rag_koowledge_go/internal/api"
	"github.com/shen060606/rag_koowledge_go/internal/database"
	"github.com/shen060606/rag_koowledge_go/internal/embedder"
	"github.com/shen060606/rag_koowledge_go/internal/rag"
	"github.com/shen060606/rag_koowledge_go/internal/store"
	"github.com/shen060606/rag_koowledge_go/internal/uploads"
)

func main() {
	//0 加载配置
	if err := config.Load("config.yaml"); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	//1 初始化mysql
	slog.Info("正在连接数据库...")
	dsn := config.Cfg.MySQL.DSN()
	if err := database.InitDB(dsn); err != nil {
		slog.Error("MYSQL连接失败", "err", err)
		os.Exit(1)
	}

	//2 初始化redis
	embedder.InitRedis(config.Cfg.Redis.Addr, config.Cfg.Redis.DB)

	//3 初始化向量存储器
	vs := store.NewVectorStore()

	//4  导入知识库
	err := uploads.WalkDir("uploads", func(path string) error {
		content, err := uploads.ProcessFile(path)
		if err != nil {
			slog.Warn("跳过文件", "path", path, "err", err)
			return nil
		}

		chunkcount, err := rag.ImportDoc(vs, content)
		if err != nil {
			slog.Error("导入文档失败", "path", path, "err", err)
			return nil
		}

		// 同时写数据库（只存文件名，不存完整路径）
		if _, err := database.CreateDocument(filepath.Base(path), int64(len(content)), chunkcount, "ready"); err != nil {
			slog.Error("写入数据库失败", "path", path, "err", err)
		}
		slog.Info("文档已导入", "path", path, "chunks", chunkcount)
		return nil
	})

	if err != nil {
		slog.Warn("遍历uploads失败", "err", err)
	}

	//5 web服务
	r := api.Setup(vs)
	slog.Info("服务已启动", "port", config.Cfg.Server.Port)
	r.Run(fmt.Sprintf(":%d", config.Cfg.Server.Port))

}
