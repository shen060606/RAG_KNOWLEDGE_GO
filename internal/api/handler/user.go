package handler

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/shen060606/rag_koowledge_go/internal/database"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Me(c *gin.Context) {
	userID, ok := GetCurrentUserID(c)

	if !ok {
		return
	}

	user, err := database.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(401, gin.H{"msg": "用户不存在"})
			return
		}

		c.JSON(500, gin.H{"msg": "查询用户信息失败"})
		return
	}

	c.JSON(200, gin.H{
		"userID":    userID,
		"username":  user.Username,
		"role":      user.Role,
		"createdAt": user.CreatedAt,
	})
}

func UserStatistics(c *gin.Context) {
	userID, ok := GetCurrentUserID(c)
	if !ok {
		return
	}

	stats, err := database.GetUserStatistics(userID)
	if err != nil {
		c.JSON(500, gin.H{"msg": "查询个人统计失败"})
		return
	}
	c.JSON(200, stats)

}

// 改名的限制路由
type UpdateUsernameRequest struct {
	Username string `json:"username"`
}

func UpdateUsername(c *gin.Context) {
	userID, ok := GetCurrentUserID(c)

	if !ok {
		return
	}

	var req UpdateUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误"})
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	usernameLength := utf8.RuneCountInString(req.Username)

	//对新用户名的一些限制
	if usernameLength < 3 || usernameLength > 50 {
		c.JSON(400, gin.H{
			"msg": "用户名长度必须为 3 到 50 个字符",
		})
		return
	}

	existingUser, err := database.GetUserByUsername(req.Username)
	if err == nil {
		if existingUser.ID == userID {
			c.JSON(200, gin.H{
				"ok":       true,
				"username": req.Username,
			})
			return
		}

		c.JSON(409, gin.H{
			"msg": "用户名已被使用",
		})
		return
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(500, gin.H{
			"msg": "检查用户名失败",
		})
		return
	}

	if err := database.UpdateUsername(userID, req.Username); err != nil {
		c.JSON(500, gin.H{
			"msg": "修改用户名失败",
		})
		return
	}

	c.JSON(200, gin.H{
		"ok":       true,
		"username": req.Username,
	})

}

// 改密码的限制路由
type UpdatePasswordRequest struct {
	OldPassword     string `json:"oldPassword"`
	NewPassword     string `json:"newPassword"`
	ConfirmPassword string `json:"confirmPassword"`
}

func UpdatePassword(c *gin.Context) {
	userID, ok := GetCurrentUserID(c)
	if !ok {
		return
	}

	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"msg": "参数错误",
		})
		return
	}

	if req.OldPassword == "" ||
		req.NewPassword == "" ||
		req.ConfirmPassword == "" {
		c.JSON(400, gin.H{
			"msg": "密码不能为空",
		})
		return
	}

	if req.NewPassword != req.ConfirmPassword {
		c.JSON(400, gin.H{
			"msg": "两次输入的新密码不一致",
		})
		return
	}

	if len([]byte(req.NewPassword)) < 8 {
		c.JSON(400, gin.H{
			"msg": "新密码至少需要 8 个字符",
		})
		return
	}

	// bcrypt 最多处理 72 字节
	if len([]byte(req.NewPassword)) > 72 {
		c.JSON(400, gin.H{
			"msg": "新密码不能超过 72 字节",
		})
		return
	}

	if req.NewPassword == req.OldPassword {
		c.JSON(400, gin.H{
			"msg": "新密码不能与原密码相同",
		})
		return
	}

	user, err := database.GetUserByID(userID)
	if err != nil {
		c.JSON(500, gin.H{
			"msg": "查询用户失败",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(req.OldPassword),
	); err != nil {
		c.JSON(400, gin.H{
			"msg": "原密码错误",
		})
		return
	}

	newPasswordHash, err := bcrypt.GenerateFromPassword(
		[]byte(req.NewPassword),
		bcrypt.DefaultCost,
	)
	if err != nil {
		c.JSON(500, gin.H{
			"msg": "密码加密失败",
		})
		return
	}

	if err := database.ChangePasswordAndRevokeSessions(
		userID,
		string(newPasswordHash),
	); err != nil {
		c.JSON(500, gin.H{
			"msg": "修改密码失败",
		})
		return
	}

	c.SetCookie(
		sessionCookieName,
		"",
		-1,
		"/",
		"",
		false,
		true,
	)

	c.JSON(200, gin.H{
		"ok":              true,
		"require_relogin": true,
	})
}
