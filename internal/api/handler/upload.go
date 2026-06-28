package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/shen060606/rag_koowledge_go/internal/database"
	"github.com/shen060606/rag_koowledge_go/internal/rag"
	"github.com/shen060606/rag_koowledge_go/internal/store"
	"github.com/shen060606/rag_koowledge_go/internal/uploads"
)

func UploadHandler(vs store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. c.FormFile("file") 拿文件
		avator, err := c.FormFile("file")
		if err != nil {
			c.JSON(400, gin.H{
				"msg": "文件上传失败",
			})
			return
		}
		// 2. 存到 uploads/
		dst := "./uploads/" + avator.Filename

		if err := c.SaveUploadedFile(avator, dst); err != nil {
			c.JSON(400, gin.H{
				"msg": "保存文件失败",
			})
			return
		}
		// 3. 检查重复
		if database.DocumentExists(avator.Filename) {
			c.JSON(409, gin.H{"msg": "文件已存在，请勿重复上传"})
			return
		}

		// 4. 读内容，调 rag.ImportDoc
		content, err := uploads.ProcessFile(dst)
		if err != nil {
			c.JSON(500, gin.H{
				"msg": "保存文件失败",
			})
			return
		}

		chunkcount, err := rag.ImportDoc(vs, avator.Filename, content)
		if err != nil {
			c.JSON(500, gin.H{
				"msg": "导入知识库失败",
			})
			return
		}

		// 5. 保存到数据库
		database.CreateDocument(
			avator.Filename,
			avator.Size,
			chunkcount,
			"ready",
		)

		// 5. 返回 {"ok":true, "filename":"..."}
		c.JSON(200, gin.H{
			"ok":       true,
			"filename": avator.Filename,
		})
	}
}
