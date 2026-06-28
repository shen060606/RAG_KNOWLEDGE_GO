package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/shen060606/rag_koowledge_go/internal/database"
	"github.com/shen060606/rag_koowledge_go/internal/store"
)

func ScanFile(vs store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		docs, err := database.ListDocuments()
		if err != nil {
			c.JSON(500, gin.H{
				"msg": "查询文档目录失败",
			})
			return
		}

		var files []gin.H //返回json切片
		for _, d := range docs {
			files = append(files, gin.H{
				"filename":    d.Filename,
				"chunk_count": d.ChunkCount,
				"filesize":    d.FileSize,
				"status":      d.Status,
				"created_at":  d.CreatedAt.Format("01-02 15:04"),
			})
		}

		c.JSON(200, gin.H{
			"files": files,
		})
	}
}
