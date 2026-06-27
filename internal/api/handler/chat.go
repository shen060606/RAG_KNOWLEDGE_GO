package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/shen060606/rag_koowledge_go/internal/database"
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

		sessionID := c.Query("session_id")
		if sessionID == "" {
			sessionID = "default" // 默认session_id
		}

		//1 查看历史对话
		history, _ := database.GetSessionHistory(sessionID)
		var messages []llm.Message
		for _, h := range history {
			messages = append(messages, llm.Message{Role: h.Role, Content: h.Content})

		}

		//2 rag检索，拼接prompt
		prompt := rag.AskThreeSteps(vs, q)
		messages = append(messages, llm.Message{Role: "user", Content: prompt})

		//3 存用户消息到数据库
		database.SaveMessage(sessionID, "user", q)

		// 4. 设 SSE headers
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		// 5. 流式调用llm发给前端,多轮回答
		answer, err := llm.CallDeepseekAPIHistory(messages, func(token string) {
			c.SSEvent("token", token)
			c.Writer.Flush() //接受到一个token就发到前端，不阻塞

		})
		if err != nil {
			c.SSEvent("error", err.Error())
			return
		}

		//6 存ai回答到数据库
		database.SaveMessage(sessionID, "assistant", answer)

		// 7. 结束信号，向前端发送一条事件名为 done 的自定义事件，携带的数据为空字符串。
		// 业务含义：通知前端，当前流式任务全部执行完成（比如 RAG 问答回答完毕、文件解析进度走完、扫描任务结束）。
		c.SSEvent("done", "")
	}
}
