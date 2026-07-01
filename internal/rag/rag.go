package rag

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"log/slog"
	"strings"

	"github.com/shen060606/rag_koowledge_go/config"
	"github.com/shen060606/rag_koowledge_go/internal/chunker"
	"github.com/shen060606/rag_koowledge_go/internal/embedder"
	"github.com/shen060606/rag_koowledge_go/internal/llm"
	"github.com/shen060606/rag_koowledge_go/internal/store"
)

// Importdoc 导入文档。filename 用于生成全局唯一的 chunk ID，防止不同文档的 chunk 互相覆盖。
func ImportDoc(vs store.Store, filename string, content string) (int, error) {
	// 用文件名 hash 的前 4 字节作为文档编号，乘 100000 保证不同文档的 chunk ID 不冲突
	hash := md5.Sum([]byte(filename))
	docBase := int(binary.BigEndian.Uint32(hash[:4])) * 100000

	chunks := chunker.SplitText(content, config.Cfg.Chunk.Size, config.Cfg.Chunk.Overlap)
	for _, c := range chunks {
		vec, err := embedder.EmbedderCache(c.Text)
		if err != nil {
			return len(chunks), err
		}

		vs.Add(docBase+c.ID, c.Text, vec) // 全局唯一 ID
	}
	return len(chunks), nil
}

// ask 提问
func Ask(vs store.Store, question string) (string, error) {
	// 	//1 问题向量化
	// 	queryVec, err := embedder.GetEmbedding(question)
	// 	if err != nil {
	// 		return "", err
	// 	}
	// 	//2 检索topk
	// 	results := vs.Search(queryVec, 5)

	// 	//3 拼接答案
	// 	var contextBuilder strings.Builder
	// 	for _, c := range results {
	// 		contextBuilder.WriteString(fmt.Sprintf("- %s\n", c.Text))
	// 	}

	// 	//4 构造prompt
	// 	prompt := fmt.Sprintf(`你是一个知识助手，请根据以下参考资料回答问题。

	// 参考资料：
	// %s

	// 问题：%s

	// 请用中文回答。`, contextBuilder.String(), question)

	prompt := AskThreeSteps(vs, question)

	//5调用llm
	answer, err := llm.CallDeepseekAPI(prompt, nil)
	if err != nil {
		return "", fmt.Errorf("api 调用 error: %v", err)
	}
	return answer, nil
}

// ask的前三步给抽象出来->使用eino抽象出来
func AskThreeSteps(vs store.Store, question string) string {
	//1 问题向量化
	queryVec, err := embedder.EmbedderCache(question)
	if err != nil {
		return ""
	}
	//2 检索topk
	results, err := vs.Search(queryVec, config.Cfg.Search.TopK)
	if err != nil {
		slog.Error(err.Error())
		return ""
	}

	//3 拼接答案
	var contextBuilder strings.Builder
	for _, c := range results {
		contextBuilder.WriteString(fmt.Sprintf("- %s\n", c.Text))
	}

	//4 构造prompt
	prompt := fmt.Sprintf(`你是一个知识助手。请根据以下参考资料回答用户问题。如果资料中有相关信息，优先使用；如果资料中没有，可以基于你自己的知识补充。

参考资料：
%s

问题：%s`, contextBuilder.String(), question)

	return prompt
}
