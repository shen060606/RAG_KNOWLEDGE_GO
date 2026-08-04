package api

import (
	"github.com/gin-gonic/gin"
	"github.com/shen060606/rag_koowledge_go/internal/api/handler"
	"github.com/shen060606/rag_koowledge_go/internal/store"
)

func Setup(vs store.Store) *gin.Engine {
	// gin.SetMode("release")
	r := gin.Default()

	r.Static("/static", "web/static")
	r.LoadHTMLGlob("web/templates/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	r.GET("/login", func(c *gin.Context) {
		c.HTML(200, "login.html", nil)
	})

	r.GET("/console", func(c *gin.Context) {
		c.HTML(200, "console.html", nil)
	})

	r.POST("/api/register", handler.Register)
	r.POST("/api/login", handler.Login)

	//需要登录之后才能进入的界面，使用路由组来使用中间件
	auth := r.Group("/api")
	auth.Use(handler.AuthMiddleware())

	auth.GET("/user/me", handler.Me) // 获取当前登录用户信息
	auth.GET("/user/statistics", handler.UserStatistics)
	auth.PATCH("/user/username", handler.UpdateUsername)
	auth.PATCH("/user/password", handler.UpdatePassword)

	auth.POST("/upload", handler.UploadHandler(vs))

	auth.GET("/chat/stream", handler.ChatStream(vs))

	auth.GET("/file", handler.ScanFile(vs))

	auth.DELETE("/file/:filename", handler.DeleteHandler(vs))

	auth.POST("/logout", handler.Logout)
	return r
}
