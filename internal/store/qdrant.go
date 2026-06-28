package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type QdrantStore struct {
	httpClient *http.Client
	baseURL    string
	collection string
}

func NewQdrantStore(host string, port int) (*QdrantStore, error) {
	baseURL := fmt.Sprintf("http://%s:%d", host, port)

	q := &QdrantStore{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    baseURL,
		collection: "rag_knowledge",
	}

	// 启动时自动创建 collection
	if err := q.createCollection(); err != nil {
		return nil, fmt.Errorf("创建 collection 失败: %w", err)
	}

	return q, nil
}

// createCollection 如果不存在就创建
func (q *QdrantStore) createCollection() error {
	body, _ := json.Marshal(map[string]any{
		"vectors": map[string]any{
			"size":     1024,
			"distance": "Cosine",
		},
	})

	req, _ := http.NewRequest("PUT",
		q.baseURL+"/collections/"+q.collection,
		bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 200/201 = 创建成功，409 = 已存在
	if resp.StatusCode != 200 && resp.StatusCode != 201 && resp.StatusCode != 409 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// Add 插入向量
func (q *QdrantStore) Add(chunkID int, text string, vector []float64) error {
	body, _ := json.Marshal(map[string]any{
		"points": []map[string]any{
			{
				"id":     chunkID,
				"vector": vector,
				"payload": map[string]string{
					"text": text,
				},
			},
		},
	})

	req, _ := http.NewRequest("PUT",
		q.baseURL+"/collections/"+q.collection+"/points",
		bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("插入向量失败: HTTP %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// Search 搜索向量
func (q *QdrantStore) Search(queryVec []float64, topK int) ([]VectorChunk, error) {
	body, _ := json.Marshal(map[string]any{
		"vector":       queryVec,
		"limit":        topK,
		"with_payload": true,
	})

	req, _ := http.NewRequest("POST",
		q.baseURL+"/collections/"+q.collection+"/points/search",
		bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Result []struct {
			ID      int            `json:"id"`
			Payload map[string]any `json:"payload"`
		} `json:"result"`
	}

	json.NewDecoder(resp.Body).Decode(&result)

	chunks := make([]VectorChunk, len(result.Result))
	for i, r := range result.Result {
		chunks[i] = VectorChunk{
			ID:   int(r.ID),
			Text: r.Payload["text"].(string),
		}
	}

	return chunks, nil
}
