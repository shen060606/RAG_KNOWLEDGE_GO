package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/shen060606/rag_koowledge_go/config"
)

const apiEndpoint = "https://api.deepseek.com/v1/chat/completions"

// api调用所用结构体
type Message struct {
	Role    string `json:"role"`    //对话角色，如user或assistant或system
	Content string `json:"content"` //文本对话内容
}

type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`              //对话上下文数组，支持多轮聊天
	Temperature float64   `json:"temperature,omiteppty"` //随机性，0~1，越大回答越天马行空；0 = 严谨、无创造
	MaxTokens   int       `json:"max_tokens,omitempty"`  //限制 AI 输出最大 token 长度
	Stream      bool      `json:"stream"`
}

func BuildRequest(prompt string) *ChatRequest {
	return &ChatRequest{
		Model: config.Cfg.LLM.Model,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
		Temperature: config.Cfg.LLM.Temperature,
		MaxTokens:   config.Cfg.LLM.MaxTokens,
		Stream:      true,
	}
}

// BuildRequestHistory 构建历史对话请求体
func BuildRequestHistory(messages []Message) *ChatRequest {
	return &ChatRequest{
		Model:       config.Cfg.LLM.Model,
		Messages:    messages,
		Temperature: config.Cfg.LLM.Temperature,
		MaxTokens:   config.Cfg.LLM.MaxTokens,
		Stream:      true,
	}
}

// 封装生成调用 DeepSeek API 所需的请求头 Header
func CreateAuthHeader() http.Header {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	// fmt.Println("程序读取到的完整密钥：|", apiKey, "|")
	return http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", apiKey)},
		"Content-Type":  []string{"application/json"},
	}
}

// dorequest  发起 HTTP 流式请求到 DeepSeek
func dorequest(reqbody *ChatRequest, onToken func(string)) (string, error) {
	jsonData, err := json.Marshal(reqbody)
	if err != nil {
		return "", fmt.Errorf("json marshal error: %v", err)
	}

	// 2. 创建 HTTP 请求
	req, err := http.NewRequest("POST", apiEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("http request error: %v", err)
	}
	req.Header = CreateAuthHeader()

	// 3. 发送请求
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request error: %v", err)
	}
	defer resp.Body.Close()

	// 4. 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// 5. 流式读取 SSE 事件流，逐字打印
	// 	DeepSeek 流式返回的数据长这样，每行一个独立的事件：

	// data: {"choices":[{"delta":{"content":"你"}}]}

	// data: {"choices":[{"delta":{"content":"好"}}]}

	// data: [DONE]结束标志

	var fullAnswer strings.Builder

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// 跳过非 data 行
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		payload := strings.TrimPrefix(line, "data: ")

		// 流结束标志
		if payload == "[DONE]" {
			break
		}

		// 解析当前 chunk
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}

		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			continue
		}

		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {

			if onToken != nil {
				onToken(chunk.Choices[0].Delta.Content)
			} else {
				fmt.Print(chunk.Choices[0].Delta.Content)
			}
			fullAnswer.WriteString(chunk.Choices[0].Delta.Content) //保存下来
		}
	}
	if onToken == nil {
		fmt.Println() // 流式输出结束后换行
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scanner error: %v", err)
	}

	return fullAnswer.String(), nil
}

// CallDeepseekAPI 单轮对话（CLI / 简单调用）
func CallDeepseekAPI(prompt string, onToken func(string)) (string, error) {
	return dorequest(BuildRequest(prompt), onToken)
}

// 多轮对话，外面拼好历史对话内容，传入
func CallDeepseekAPIHistory(messages []Message, onToken func(string)) (string, error) {
	return dorequest(BuildRequestHistory(messages), onToken)
}
