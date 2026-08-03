package handler

import "github.com/gin-gonic/gin"

func Me(c *gin.Context) {
	userID, ok := getCurrentUserID(c)

	if !ok {
		c.JSON(401, gin.H{"msg": "未登录"})
		return
	}

	username, _ := c.Get("username")
	role, _ := c.Get("role")

	c.JSON(200, gin.H{
		"userID":   userID,
		"username": username,
		"role":     role,
	})
}
