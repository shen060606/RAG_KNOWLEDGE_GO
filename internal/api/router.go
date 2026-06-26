package api

import (
	"github.com/gin-gonic/gin"
	"github.com/shen060606/rag_koowledge_go/internal/api/handler"
	"github.com/shen060606/rag_koowledge_go/internal/store"
)

func Setup(vs *store.VectorStore) *gin.Engine {
	// gin.SetMode("release")
	r := gin.Default()

	r.Static("/static", "web/static")
	r.LoadHTMLGlob("web/templates/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	r.POST("/api/upload", handler.UploadHandler(vs))

	r.GET("/api/chat/stream", handler.ChatStream(vs))

	r.GET("/api/file", handler.ScanFile(vs))
	return r
}
