package handler

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/shen060606/rag_koowledge_go/internal/store"
)

func ScanFile(vs *store.VectorStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		entries, err := os.ReadDir("uploads")
		if err != nil {
			c.JSON(500, gin.H{"msg": "读取失败"})
			return
		}
		var files []string
		for _, e := range entries {
			if !e.IsDir() {
				files = append(files, e.Name())
			}
		}
		c.JSON(200, gin.H{"files": files})
	}
}
