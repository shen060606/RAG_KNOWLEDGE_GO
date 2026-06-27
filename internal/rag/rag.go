package rag

import (
	"fmt"
	"strings"

	"github.com/shen060606/rag_koowledge_go/config"
	"github.com/shen060606/rag_koowledge_go/internal/chunker"
	"github.com/shen060606/rag_koowledge_go/internal/embedder"
	"github.com/shen060606/rag_koowledge_go/internal/llm"
	"github.com/shen060606/rag_koowledge_go/internal/store"
)

// Importdoc 导入文档
func ImportDoc(vs *store.VectorStore, content string) (int, error) {
	chunks := chunker.SplitText(content, config.Cfg.Chunk.Size, config.Cfg.Chunk.Overlap)
	for _, c := range chunks {
		vec, err := embedder.EmbedderCache(c.Text)
		if err != nil {
			return len(chunks), err
		}

		vs.Add(c.ID, c.Text, vec)
	}
	return len(chunks), nil
}

// ask 提问
func Ask(vs *store.VectorStore, question string) (string, error) {
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

// ask的前三步给抽象出来
func AskThreeSteps(vs *store.VectorStore, question string) string {
	//1 问题向量化
	queryVec, err := embedder.EmbedderCache(question)
	if err != nil {
		return ""
	}
	//2 检索topk
	results := vs.Search(queryVec, config.Cfg.Search.TopK)

	//3 拼接答案
	var contextBuilder strings.Builder
	for _, c := range results {
		contextBuilder.WriteString(fmt.Sprintf("- %s\n", c.Text))
	}

	//4 构造prompt
	prompt := fmt.Sprintf(`你是一个知识助手。请按以下规则回答：

1. 优先根据参考资料回答问题
2. 如果参考资料中有答案，回答格式：
   "【来自知识库】
   你的回答内容..."

3. 如果参考资料中信息不足，回答格式：
   "【来自知识库】
   未在知识库文档中找到相关内容。

   【来自外部知识】
   根据现有知识，...（你的补充回答）"

现在开始，参考资料如下：
---
%s
---
问题：%s
`, contextBuilder.String(), question)

	return prompt
}
