package handler

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/shen060606/rag_koowledge_go/internal/database"
	"github.com/shen060606/rag_koowledge_go/internal/rag"
	"github.com/shen060606/rag_koowledge_go/internal/store"
	"gorm.io/gorm"
)

func DeleteHandler(vs store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetCurrentUserID(c)
		if !ok {
			return
		}

		filename := c.Param("filename")
		if filename == "" {
			c.JSON(400, gin.H{"msg": "缺少文件名"})
			return
		}

		//1 查找数据库，拿到chunkcount算chunkid范围
		doc, err := database.GetDocumentByFilename(userID, filename)
		if err != nil || doc == nil {
			// 检查是否是公共文档
			publicDoc, publicErr := database.GetPublicDocumentByFilename(filename)
			if publicErr == nil && publicDoc != nil {
				c.JSON(403, gin.H{"msg": "无权限删除公共文档"})
				return
			}

			if publicErr != nil && !errors.Is(publicErr, gorm.ErrRecordNotFound) {
				c.JSON(500, gin.H{"msg": "查询公共文档失败"})
				return
			}

			c.JSON(404, gin.H{"msg": "文档不存在"})
			return
		}

		//2 删除向量存储里面的chunk
		if err := rag.DeleteDoc(vs, userID, filename, doc.ChunkCount); err != nil {
			c.JSON(500, gin.H{"msg": "删除向量数据失败: " + err.Error()})
			return
		}

		//3 删除数据库记录
		if err := database.DeleteDocument(userID, filename); err != nil {
			c.JSON(500, gin.H{"msg": "删除数据库记录失败: " + err.Error()})
			return
		}

		// 4. 删除物理文件
		filePath := filepath.Join("uploads", fmt.Sprint(userID), filename)
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			c.JSON(500, gin.H{"msg": "删除物理文件失败"})
			return
		}

		c.JSON(200, gin.H{"ok": true})
	}
}
