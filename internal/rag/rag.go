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

// 把参考来源定义成一个结构体
type Source struct {
	Index    int    `json:"index"`    //引用编号，比如 [1]
	Filename string `json:"filename"` //来源文件名
	IsPublic bool   `json:"is_public"`
}

// Importdoc 导入文档。filename 用于生成全局唯一的 chunk ID，防止不同文档的 chunk 互相覆盖。
func ImportDoc(vs store.Store, userID uint, filename string, content string, isPublic bool) (int, error) {
	docKey := fmt.Sprintf("%d/%s", userID, filename)
	// 用文件名 hash 的前 4 字节作为文档编号，乘 100000 保证不同文档的 chunk ID 不冲突
	hash := md5.Sum([]byte(docKey))
	docBase := int(binary.BigEndian.Uint32(hash[:4])) * 100000

	chunks := chunker.SplitText(content, config.Cfg.Chunk.Size, config.Cfg.Chunk.Overlap)
	for _, c := range chunks {
		vec, err := embedder.EmbedderCache(c.Text)
		if err != nil {
			return len(chunks), err
		}
		// 全局唯一 ID
		if err := vs.Add(userID, docBase+c.ID, filename, c.Text, vec, isPublic); err != nil {
			return len(chunks), err
		}
	}
	return len(chunks), nil
}

// ask 提问
func Ask(vs store.Store, userID uint, question string) (string, error) {
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

	prompt, _ := AskThreeSteps(vs, userID, question)

	//5调用llm
	answer, err := llm.CallDeepseekAPI(prompt, nil)
	if err != nil {
		return "", fmt.Errorf("api 调用 error: %v", err)
	}
	return answer, nil
}

// ask的前三步给抽象出来->使用eino抽象出来
func AskThreeSteps(vs store.Store, userID uint, question string) (string, []Source) {
	//1 问题向量化
	queryVec, err := embedder.EmbedderCache(question)
	if err != nil {
		slog.Error("问题向量化失败", "err", err)
		return "", nil
	}
	//2 检索topk
	results, err := vs.Search(userID, queryVec, config.Cfg.Search.TopK)
	if err != nil {
		slog.Error("向量检索失败", "err", err)
		return "", nil
	}

	//3 给来源文件编号，filename去重
	sourceIndex := make(map[string]int)
	var sources []Source

	var contextBuilder strings.Builder

	for _, c := range results {
		filename := c.Filename
		if filename == "" {
			filename = "未知来源"
		}

		index, ok := sourceIndex[filename]
		if !ok {
			index = len(sources) + 1
			sourceIndex[filename] = index
			sources = append(sources, Source{
				Index:    index,
				Filename: filename,
				IsPublic: c.IsPublic,
			})
		}

		contextBuilder.WriteString(fmt.Sprintf("[来源 %d] 文件：%s\n内容：%s\n\n", index, filename, c.Text))
	}

	//4 构造prompt
	prompt := fmt.Sprintf(`你是一个知识助手。请根据以下参考资料回答用户问题。

要求：
1. 如果资料中有相关信息，优先基于资料回答。
2. 如果资料中没有相关内容，就明确说“文档库中没有相关内容”。
3. 回答时，如果使用了某段资料，请在对应句子后面标注来源编号，例如 [1]、[2]。
4. 不要编造不存在的来源编号。
5. 请用中文回答。

参考资料：
%s

问题：%s`, contextBuilder.String(), question)

	return prompt, sources
}

func DeleteDoc(vs store.Store, userID uint, filename string, chunkcount int) error {
	docKey := fmt.Sprintf("%d/%s", userID, filename)
	hash := md5.Sum([]byte(docKey))
	docBase := int(binary.BigEndian.Uint32(hash[:4])) * 100000

	//生成该文档 的所有chunkid
	ids := make([]int, chunkcount)
	for i := 0; i < chunkcount; i++ {
		ids[i] = docBase + i
	}

	return vs.Delete(ids) //删除该文档的所有chunk
}
