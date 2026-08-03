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
func (q *QdrantStore) Add(userID uint, chunkID int, text string, vector []float64, isPublic bool) error {
	body, _ := json.Marshal(map[string]any{
		"points": []map[string]any{
			{
				"id":     chunkID,
				"vector": vector,
				"payload": map[string]any{
					"text":      text,
					"user_id":   userID,
					"is_public": isPublic,
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
func (q *QdrantStore) Search(userID uint, queryVec []float64, topK int) ([]VectorChunk, error) {
	//Qdrant 检索时只返回当前用户的点
	body, _ := json.Marshal(map[string]any{
		"vector": queryVec, //用于相似度搜索的查询向量
		"limit":  topK,     //最多返回多少条结果
		"filter": map[string]any{ //附加过滤条件
			"should": []map[string]any{ //里面的条件满足其中一个就行
				{
					"key": "user_id", //要过滤的 payload 字段名
					"match": map[string]any{ //使用精确匹配条件
						"value": userID, //要匹配的具体值
					},
				},
				{
					"key": "is_public", //要过滤的 payload 字段名
					"match": map[string]any{
						"value": true,
					},
				},
			},
		},
		"with_payload": true, //是否把 point 的 payload 一起返回
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

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("搜索向量失败: HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Result []struct {
			ID      int            `json:"id"`
			Payload map[string]any `json:"payload"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析 Qdrant 搜索结果失败: %w", err)
	}

	chunks := make([]VectorChunk, len(result.Result))
	for i, r := range result.Result {
		// 提取 payload 中的 text 和 user_id
		text, _ := r.Payload["text"].(string)

		//类型判断
		var payloadUserID uint
		switch v := r.Payload["user_id"].(type) {
		case float64:
			payloadUserID = uint(v)
		case int:
			payloadUserID = uint(v)
		}

		isPublic, _ := r.Payload["is_public"].(bool)

		chunks[i] = VectorChunk{
			ID:       int(r.ID),
			UserID:   payloadUserID,
			Text:     text,
			IsPublic: isPublic,
		}
	}

	return chunks, nil
}

func (q *QdrantStore) Delete(chunkIDs []int) error {
	body, _ := json.Marshal(map[string]any{
		"points": chunkIDs,
	})

	req, _ := http.NewRequest("POST",
		q.baseURL+"/collections/"+q.collection+"/points/delete",
		bytes.NewBuffer(body))

	req.Header.Set("Content-Type", "application/json")

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("删除向量失败: HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
