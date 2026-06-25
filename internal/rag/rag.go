package rag

import (
	"fmt"
	"strings"

	"github.com/shen060606/rag_koowledge_go/internal/chunker"
	"github.com/shen060606/rag_koowledge_go/internal/embedder"
	"github.com/shen060606/rag_koowledge_go/internal/llm"
	"github.com/shen060606/rag_koowledge_go/internal/store"
)

// Importdoc 导入文档
func ImportDoc(vs *store.VectorStore, content string) error {
	chunks := chunker.SplitText(content, 500, 50)
	for _, c := range chunks {
		vec, err := embedder.GetEmbedding(c.Text)
		if err != nil {
			return err
		}

		vs.Add(c.ID, c.Text, vec)
	}
	return nil
}

// ask 提问
func Ask(vs *store.VectorStore, question string) (string, error) {
	//1 问题向量化
	queryVec, err := embedder.GetEmbedding(question)
	if err != nil {
		return "", err
	}
	//2 检索topk
	results := vs.Search(queryVec, 5)

	//3 拼接答案
	var contextBuilder strings.Builder
	for _, c := range results {
		contextBuilder.WriteString(fmt.Sprintf("- %s\n", c.Text))
	}

	//4 构造prompt
	prompt := fmt.Sprintf(`你是一个安全知识助手，请根据以下参考资料回答问题。

参考资料：
%s

问题：%s

请用中文回答。`, contextBuilder.String(), question)

	//5调用llm
	answer, err := llm.CallDeepseekAPI(prompt)
	if err != nil {
		return "", fmt.Errorf("api 调用 error: %v", err)
	}
	return answer, nil
}
