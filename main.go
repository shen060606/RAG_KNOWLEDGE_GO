package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/shen060606/rag_koowledge_go/config"
	"github.com/shen060606/rag_koowledge_go/internal/api"
	"github.com/shen060606/rag_koowledge_go/internal/database"
	"github.com/shen060606/rag_koowledge_go/internal/embedder"
	"github.com/shen060606/rag_koowledge_go/internal/store"
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
	var vs store.Store
	var err error

	vs, err = store.NewQdrantStore("127.0.0.1", 6333)
	if err != nil {
		slog.Warn("Qdrant不可用，降级为内存存储", "err", err)
		vs = store.NewMemoryStore()
	}

	//4 web服务
	r := api.Setup(vs)
	slog.Info("服务已启动", "port", config.Cfg.Server.Port)
	r.Run(fmt.Sprintf(":%d", config.Cfg.Server.Port))

}
