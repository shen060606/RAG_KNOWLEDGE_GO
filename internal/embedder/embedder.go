package embedder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const embeddingEndpoint = "https://api.siliconflow.cn/v1/embeddings"

// embedding调用所用结构体
type EmbeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type EmbeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}

func BuildEmbeddingRequest(text string) *EmbeddingRequest {
	return &EmbeddingRequest{
		Model: "BAAI/bge-large-zh-v1.5",
		Input: text,
	}
}

// 封装硅基流动的key用来调用embedding模型
func CreateEmbeddingAuthHeader() http.Header {
	apiKey := os.Getenv("SILICONFLOW_API_KEY")
	// fmt.Println("程序读取到的完整密钥：|", apiKey, "|")
	return http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", apiKey)},
		"Content-Type":  []string{"application/json"},
	}
}

func GetEmbedding(text string) ([]float64, error) {
	// 1. 构造请求（json.Marshal）
	reqBody := BuildEmbeddingRequest(text)
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("json marshal error: %v", err)
	}

	// 2. 创建 HTTP 请求

	req, err := http.NewRequest("POST", embeddingEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("http request error: %v", err)
	}
	req.Header = CreateEmbeddingAuthHeader()

	// 3. 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request error: %v", err)
	}
	defer resp.Body.Close()

	// 4. 检查状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// 5. 解析响应，拿出 data[0].embedding
	var result EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("json decode error: %v", err)
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	// 6. 返回 []float64
	return result.Data[0].Embedding, nil

}
