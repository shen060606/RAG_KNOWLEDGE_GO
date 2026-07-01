package handler

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/shen060606/rag_koowledge_go/internal/database"
	"github.com/shen060606/rag_koowledge_go/internal/rag"
	"github.com/shen060606/rag_koowledge_go/internal/store"
)

func DeleteHandler(vs store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		filename := c.Param("filename")
		if filename == "" {
			c.JSON(400, gin.H{"msg": "缺少文件名"})
			return
		}

		//1 查找数据库，拿到chunkcount算chunkid范围
		doc, err := database.GetDocumentByFilename(filename)
		if err != nil || doc == nil {
			c.JSON(404, gin.H{"msg": "文档不存在"})
			return
		}

		//2 删除向量存储里面的chunk
		if err := rag.DeleteDoc(vs, filename, doc.ChunkCount); err != nil {
			c.JSON(500, gin.H{"msg": "删除向量数据失败: " + err.Error()})
			return
		}

		//3 删除数据库记录
		if err := database.DeleteDocument(filename); err != nil {
			c.JSON(500, gin.H{"msg": "删除数据库记录失败: " + err.Error()})
			return
		}

		// 4. 删除物理文件
		os.Remove("./uploads/" + filename)

		c.JSON(200, gin.H{"ok": true})
	}
}
