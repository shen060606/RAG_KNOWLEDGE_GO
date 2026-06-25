package store

import (
	"math"
	"sort"
)

// VectorChunk 一条文档块+它的向量
type VectorChunk struct {
	ID     int
	Text   string
	Vector []float64
}

// VectorStore 文档块向量存储
type VectorStore struct {
	Chunks []VectorChunk
}

// NewVectorStore 创建一个新的向量存储实例
func NewVectorStore() *VectorStore {
	return &VectorStore{
		Chunks: make([]VectorChunk, 0),
	}
}

// Add 添加一个文档块到向量存储中
func (vs *VectorStore) Add(chunkID int, text string, vector []float64) {
	vs.Chunks = append(vs.Chunks, VectorChunk{
		ID:     chunkID,
		Text:   text,
		Vector: vector,
	})
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

// Search 检索Topk个最相似的文档块，TopK = 相似度最高的前 K 个结果
func (vs *VectorStore) Search(queryVec []float64, topK int) []VectorChunk {
	//1 计算每个文档块与查询向量的余弦相似度
	type scored struct {
		chunk VectorChunk
		score float64
	}
	var results []scored

	for _, c := range vs.Chunks {
		results = append(results, scored{
			chunk: c,
			score: CosineSimilarity(queryVec, c.Vector),
		})
	}

	//2 按照相似度从高到低排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	//3 返回TopK个最相似的文档块
	if topK > len(results) {
		topK = len(results)
	}

	topChunks := make([]VectorChunk, topK)
	for i := 0; i < topK; i++ {
		topChunks[i] = results[i].chunk
	}

	return topChunks
}
