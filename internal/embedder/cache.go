package embedder

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// rdb 全局redis客户端
var rdb *redis.Client

// ctx 上下文,所有redis操作都要传
var ctx = context.Background()

// InitRedis 初始化
func InitRedis(addr string, db int) {
	rdb = redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   db, //使用的内置的逻辑数据库编号
	})

	//发送ping测试连接，不通就panic
	if err := rdb.Ping(ctx).Err(); err != nil {
		panic("Redis连接失败： " + err.Error())
	}
}

// md5hash 把文本转为固定长度的key
func md5hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// EmbedderCache 带redis缓存 的embedding
// 流程：文本 → md5 → 查 Redis → 命中就返回 → 没命中就调 API → 存 Redis → 返回
func EmbedderCache(text string) ([]float64, error) {
	key := "emb:" + md5hash(text)

	//1 查redis
	val, err := rdb.Get(ctx, key).Result()
	// rdb.Get 返回两个值：val（字符串）和 err
	// 如果 key 存在 → err == nil，val 是缓存的值
	// 如果 key 不存在 → err == redis.Nil，需要调 API
	if err == nil {
		//命中 ，把json字符串转为float64切片
		var vec []float64
		if err := json.Unmarshal([]byte(val), &vec); err != nil {
			rdb.Del(ctx, key)
		} else {
			return vec, nil
		}

	}

	if err != redis.Nil {
		// Redis 真的出错了（不是"key 不存在"），打日志但继续走 API
		// 降级：不因为 Redis 挂了导致整个功能不能用\
		slog.Warn("Redis 读取失败，降级到直接调用api", "err", err)
	}

	// ===== 第2步：缓存未命中，调 API =====
	vec, err := GetEmbedding(text)
	if err != nil {
		return nil, err
	}

	// ===== 第3步：存入 Redis，24 小时后自动过期 =====
	data, err := json.Marshal(vec)
	if err != nil {
		return vec, err
	}

	ttl := time.Duration(23) * time.Hour
	rdb.Set(ctx, key, data, ttl)

	return vec, nil
}
