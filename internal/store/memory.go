package store

import "sort"

// VectorStore 文档块向量存储
type MemoryStore struct {
	Chunks []VectorChunk
}

// NewVectorStore 创建一个新的向量存储实例
func NewMemoryStore() Store {
	return &MemoryStore{
		Chunks: make([]VectorChunk, 0),
	}
}

// Add 添加一个文档块到向量存储中
func (vs *MemoryStore) Add(chunkID int, text string, vector []float64) error {
	vs.Chunks = append(vs.Chunks, VectorChunk{
		ID:     chunkID,
		Text:   text,
		Vector: vector,
	})

	return nil
}

// Search 检索Topk个最相似的文档块，TopK = 相似度最高的前 K 个结果
func (vs *MemoryStore) Search(queryVec []float64, topK int) ([]VectorChunk, error) {
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

	return topChunks, nil
}
