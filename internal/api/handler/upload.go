package handler

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/shen060606/rag_koowledge_go/internal/database"
	"github.com/shen060606/rag_koowledge_go/internal/rag"
	"github.com/shen060606/rag_koowledge_go/internal/store"
	"github.com/shen060606/rag_koowledge_go/internal/uploads"
)

func UploadHandler(vs store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getCurrentUserID(c)
		if !ok {
			c.JSON(401, gin.H{"msg": "未登录"})
			return
		}

		isPublic := IsAdmin(c)

		// 1. c.FormFile("file") 拿文件
		avator, err := c.FormFile("file")
		if err != nil {
			c.JSON(400, gin.H{
				"msg": "文件上传失败",
			})
			return
		}

		//1.2 清理文件名，防止路径穿越
		filename := filepath.Base(avator.Filename)
		if filename == "." || filename == string(filepath.Separator) || filename != avator.Filename {
			c.JSON(400, gin.H{"msg": "非法文件名"})
			return
		}

		//1.5 检查文件类型
		if !uploads.IsAllowedFile(filename) {
			c.JSON(400, gin.H{"msg": "只支持pdf,txt,md 类型"})
			return
		}

		// 2. 检查重复
		if database.DocumentExists(userID, filename) {
			c.JSON(409, gin.H{"msg": "文件已存在，请勿重复上传"})
			return
		}

		//2.5 把文件保存路径修改成userid+文件名
		userUploadDir := filepath.Join("uploads", fmt.Sprint(userID))
		if err := os.MkdirAll(userUploadDir, 0755); err != nil {
			c.JSON(500, gin.H{"msg": "创建用户上传目录失败"})
			return
		}

		// 3. 存到 uploads/里面的各自的user目录
		dst := filepath.Join(userUploadDir, filename)

		if err := c.SaveUploadedFile(avator, dst); err != nil {
			c.JSON(400, gin.H{
				"msg": "保存文件失败",
			})
			return
		}

		// 4. 读内容，调 rag.ImportDoc
		content, err := uploads.ProcessFile(dst)
		if err != nil {
			c.JSON(500, gin.H{
				"msg": "解析文件失败",
			})
			return
		}

		chunkcount, err := rag.ImportDoc(vs, userID, filename, content, isPublic)
		if err != nil {
			c.JSON(500, gin.H{
				"msg": "导入知识库失败",
			})
			return
		}

		// 5. 保存到数据库
		if _, err := database.CreateDocument(
			userID,
			filename,
			avator.Size,
			chunkcount,
			"ready",
			isPublic,
		); err != nil {
			c.JSON(500, gin.H{"msg": "保存文档记录失败"})
			return
		}

		// 5. 返回 {"ok":true, "filename":"..."}
		c.JSON(200, gin.H{
			"ok":          true,
			"filename":    filename,
			"chunk_count": chunkcount,
			"is_public":   isPublic,
		})
	}
}
