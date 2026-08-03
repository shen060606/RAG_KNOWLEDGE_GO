package handler

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/shen060606/rag_koowledge_go/internal/database"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const sessionCookieName = "session"

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Register(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误"})
		return
	}

	//去除头和尾的空白字符
	req.Username = strings.TrimSpace(req.Username)
	usernameLen := utf8.RuneCountInString(req.Username)

	//对用户名和密码进行限制（符合数据库中的存储条件）
	if usernameLen < 3 || usernameLen > 50 {
		c.JSON(400, gin.H{"msg": "用户名长度必须为 3 到 50 个字符"})
		return
	}

	if len([]byte(req.Password)) < 8 {
		c.JSON(400, gin.H{"msg": "密码至少需要 8 个字符"})
		return
	}

	if len([]byte(req.Password)) > 72 {
		c.JSON(400, gin.H{"msg": "密码不能超过 72 字节"})
		return
	}

	if req.Username == "" || req.Password == "" {
		c.JSON(400, gin.H{"msg": "用户名或密码不能为空"})
		return
	}

	//看看用户名是否已经存在
	user, err := database.GetUserByUsername(req.Username)
	if err == nil && user != nil {
		c.JSON(409, gin.H{"msg": "用户名已存在"})
		return
	}

	//这个错误是不是?"数据库没有找到记录" 这个错误
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(500, gin.H{"msg": "查询用户失败"})
		return
	}

	//密码加密
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"msg": "密码加密失败"})
		return
	}

	//获取数据库中注册用户数量
	count, err := database.CountUsers()
	if err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误"})
		return
	}

	role := "user"
	if count == 0 {
		role = "admin"
	}
	//创建用户
	if _, err := database.CreateUser(req.Username, string(passwordHash), role); err != nil {
		c.JSON(500, gin.H{"msg": "用户注册失败"})
		return
	}

	c.JSON(200, gin.H{"ok": true})
}

func Login(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误"})
		return
	}

	req.Username = strings.TrimSpace(req.Username)

	if req.Username == "" || req.Password == "" {
		c.JSON(400, gin.H{"msg": "用户名或密码不能为空"})
		return
	}

	user, err := database.GetUserByUsername(req.Username)
	if err != nil {
		c.JSON(401, gin.H{"msg": "用户名或密码错误"})
		return
	}

	//密码验证
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(401, gin.H{"msg": "用户名或密码错误"})
		return
	}

	// 登录成功后随机生成一个sessionid
	sessionID, err := generateSessionID()
	if err != nil {
		c.JSON(500, gin.H{"msg": "生成sessionid失败"})
		return
	}

	//session 7 天过期
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err := database.CreateSession(sessionID, user.ID, expiresAt); err != nil {
		c.JSON(500, gin.H{"msg": "创建session失败"})
		return
	}

	// 设置session cookie
	c.SetCookie(
		sessionCookieName,
		sessionID,
		int(time.Until(expiresAt).Seconds()),
		"/",
		"",
		false,
		true,
	)

	c.JSON(200, gin.H{"ok": true})

}

// logout 退出登录
func Logout(c *gin.Context) {
	sessionID, err := c.Cookie(sessionCookieName)
	if err == nil && sessionID != "" {
		if err := database.DeleteSession(sessionID); err != nil {
			c.JSON(500, gin.H{"msg": "退出登录失败"})
			return
		}
	}

	// 清除session cookie
	c.SetCookie(sessionCookieName, "", -1, "/", "", false, true)
	c.JSON(200, gin.H{"ok": true})
}

// authmiddleware 鉴权中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//1 先从cookie里面获取sessionid
		sessionID, err := c.Cookie(sessionCookieName)
		if err != nil || sessionID == "" {
			c.JSON(401, gin.H{"msg": "请先登录"})
			c.Abort()
			return
		}

		//2 去数据库里面查看sessionid是否存在
		session, err := database.GetSessionByID(sessionID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(401, gin.H{"msg": "session 不存在，请重新登录"})
				c.Abort()
				return
			}

			c.JSON(500, gin.H{"msg": "查询 session 失败"})
			c.Abort()
			return
		}

		if time.Now().After(session.ExpiresAt) {
			if err := database.DeleteSession(sessionID); err != nil {
				c.JSON(500, gin.H{"msg": "清理过期 session 失败"})
				c.Abort()
				return
			}

			c.JSON(401, gin.H{"msg": "登录已过期，请重新登录"})
			c.Abort()
			return
		}

		//3 根据session找到用户
		user, err := database.GetUserByID(session.UserID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(401, gin.H{"msg": "用户不存在"})
				c.Abort()
				return
			}

			c.JSON(500, gin.H{"msg": "查询用户失败"})
			c.Abort()
			return
		}

		//4 设置用户信息到上下文，后面的handler就能拿到了
		c.Set("userID", user.ID)
		c.Set("username", user.Username)
		c.Set("role", user.Role)

		c.Next() //继续执行后续的handler

	}
}

// 生成随机sessionid
func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// getCurrentUserID 从业务handler里面提取userid
func getCurrentUserID(c *gin.Context) (uint, bool) {
	userIDval, ok := c.Get("userID")
	if !ok {
		c.JSON(401, gin.H{"msg": "请先登录"})
		return 0, false
	}

	userID, ok := userIDval.(uint)
	if !ok {
		c.JSON(500, gin.H{"msg": "用户信息异常"})
		return 0, false
	}

	return userID, true
}

func IsAdmin(c *gin.Context) bool {
	role, ok := c.Get("role")
	if !ok {
		return false
	}

	roleStr, ok := role.(string)
	if !ok {
		return false
	}

	return roleStr == "admin"
}

func GetCurrentUserRole(c *gin.Context) string {
	role, ok := c.Get("role")
	if !ok {
		return "user"
	}

	roleStr, ok := role.(string)
	if !ok {
		return "user"
	}

	return roleStr
}
