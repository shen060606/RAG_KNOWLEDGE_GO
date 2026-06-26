package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/shen060606/rag_koowledge_go/internal/llm"
	"github.com/shen060606/rag_koowledge_go/internal/rag"
	"github.com/shen060606/rag_koowledge_go/internal/store"
)

func ChatStream(vs *store.VectorStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := c.Query("q")
		if q == "" {
			c.JSON(400, gin.H{
				"msg": "请输入问题",
			})
			return
		}
		prompt := rag.AskThreeSteps(vs, q)
		// 4. 设 SSE headers
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		// 5. 流式调用llm发给前端,llm.StreamChat(prompt, func(token){ c.SSEvent(...) })
		llm.CallDeepseekAPI(prompt, func(token string) {
			c.SSEvent("token", token)
			c.Writer.Flush() //接受到一个token就发到前端，不阻塞
		})
		// 6. 结束信号，向前端发送一条事件名为 done 的自定义事件，携带的数据为空字符串。
		// 业务含义：通知前端，当前流式任务全部执行完成（比如 RAG 问答回答完毕、文件解析进度走完、扫描任务结束）。
		c.SSEvent("done", "")
	}
}
