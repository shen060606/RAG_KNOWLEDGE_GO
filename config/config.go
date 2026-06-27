package config

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Server    ServerCfg    `yaml:"server"`
	MySQL     MySQLCfg     `yaml:"mysql"`
	Redis     RedisCfg     `yaml:"redis"`
	LLM       LLMCfg       `yaml:"llm"`
	Embedding EmbeddingCfg `yaml:"embedding"`
	Chunk     ChunkCfg     `yaml:"chunk"`
	Search    SearchCfg    `yaml:"search"`
}

type ServerCfg struct {
	Port int `yaml:"port"`
}

type MySQLCfg struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type RedisCfg struct {
	Addr              string `yaml:"addr"`
	DB                int    `yaml:"db"`
	EmbeddingCacheTTL int    `yaml:"embedding_cache_ttl"`
}

type LLMCfg struct {
	Model       string  `yaml:"model"`
	Temperature float64 `yaml:"temperature"`
	MaxTokens   int     `yaml:"max_tokens"`
}

type EmbeddingCfg struct {
	Model string `yaml:"model"`
}

type ChunkCfg struct {
	Size    int `yaml:"size"`
	Overlap int `yaml:"overlap"`
}

type SearchCfg struct {
	TopK int `yaml:"top_k"`
}

var Cfg *Config

func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	//初始化全局单例 cfg 结构体（分配内存）
	Cfg = &Config{}

	//yaml.Unmarshal 将文本data映射填充到cfg结构体
	if err := yaml.Unmarshal(data, Cfg); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 默认值
	if Cfg.Server.Port == 0 {
		Cfg.Server.Port = 8088
	}
	if Cfg.LLM.Temperature == 0 {
		Cfg.LLM.Temperature = 0.7
	}
	if Cfg.LLM.MaxTokens == 0 {
		Cfg.LLM.MaxTokens = 2048
	}
	if Cfg.Chunk.Size == 0 {
		Cfg.Chunk.Size = 500
	}
	if Cfg.Chunk.Overlap == 0 {
		Cfg.Chunk.Overlap = 50
	}
	if Cfg.Search.TopK == 0 {
		Cfg.Search.TopK = 5
	}

	return nil
}

func (m *MySQLCfg) DSN() string {
	// root:123456@tcp(127.0.0.1:3306)/rag_knowledge?charset=utf8mb4&parseTime=True
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True", m.User, m.Password, m.Host, m.Port, m.Database)
}
