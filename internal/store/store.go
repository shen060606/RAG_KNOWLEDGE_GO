package store

import (
	"math"
)

// VectorChunk 一条文档块+它的向量
type VectorChunk struct {
	ID     int
	Text   string
	Vector []float64
}

// 定义接口，方便以后改向量数据库等
type Store interface {
	Add(chunkID int, text string, vector []float64) error
	Search(queryVec []float64, topK int) ([]VectorChunk, error)
}

// CosineSimilarity 计算两个向量的余弦相似度
func CosineSimilarity(a, b []float64) float64 {
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]   //点积和
		normA += a[i] * a[i] //a向量模
		normB += b[i] * b[i] // b向量模
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
