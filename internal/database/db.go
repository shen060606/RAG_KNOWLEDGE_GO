package database

import (
	"log/slog"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化数据库连接并自动建表
func InitDB(dsn string) error {
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn), //只打印警告和错误信息
	})
	if err != nil {
		//logger.Error("数据库连接失败", err)
		return err
	}

	//自动建表（只创建不存在的表/加不存在的列）
	if err := DB.AutoMigrate(&Document{}, &ChatHistory{}, &User{}, &Session{}); err != nil {
		return err
	}

	slog.Info("MYSQL连接成功，表结构已同步")
	return nil
}

// ===== 文档相关 =====

// CreateDocument 创建文档记录
func CreateDocument(userID uint, filename string, filesize int64, chunkcount int, status string, isPublic bool) (*Document, error) {
	doc := &Document{
		UserID:     userID,
		Filename:   filename,
		FileSize:   filesize,
		ChunkCount: chunkcount,
		Status:     status,
		IsPublic:   isPublic,
		CreatedAt:  time.Now(),
	}
	if err := DB.Create(doc).Error; err != nil {
		return nil, err
	}

	return doc, nil
}

// ListDocuments 查询所有已就绪的文档
func ListDocuments(userID uint) ([]Document, error) {
	var docs []Document
	err := DB.Where("(user_id=? or is_public=?) and status = ?", userID, true, "ready").Order("created_at DESC").Find(&docs).Error
	return docs, err
}

// 查看文档是否是公共文档
func GetPublicDocumentByFilename(filename string) (*Document, error) {
	var doc Document
	err := DB.Where("filename = ? AND is_public = ? AND status = ?", filename, true, "ready").First(&doc).Error

	if err != nil {
		return nil, err
	}

	return &doc, nil
}

// ===== 对话相关 =====

// SaveMessage 保存一条对话记录
func SaveMessage(userID uint, sessionid, role, content string) error {
	msg := &ChatHistory{
		UserID:    userID,
		SessionID: sessionid,
		Role:      role,
		Content:   content,
		CreatedAt: time.Now(),
	}

	return DB.Create(msg).Error
}

// GetSessionHistory 获取某个session会话历史记录
func GetSessionHistory(userID uint, sessionid string) ([]ChatHistory, error) {
	var history []ChatHistory
	err := DB.Where("user_id = ? and session_id = ?", userID, sessionid).Order("created_at ASC").Find(&history).Error
	return history, err
}

// DocumentExists 检查文件是否已导入
func DocumentExists(userID uint, filename string) bool {
	var count int64
	DB.Model(&Document{}).Where("user_id=? and filename = ? AND status = ?", userID, filename, "ready").Count(&count)
	return count > 0
}

// GetDocumentByFilename 根据文件名查一条文档记录
func GetDocumentByFilename(userID uint, filename string) (*Document, error) {
	var doc Document
	err := DB.Where("user_id=? and filename = ? AND status = ?", userID, filename, "ready").First(&doc).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func DeleteDocument(userID uint, filename string) error {
	return DB.Where("user_id=? and filename = ?", userID, filename).Delete(&Document{}).Error
}

// ===== 用户相关 =====

// 注册接口会调用
func CreateUser(username, passwordHash, role string) (*User, error) {
	user := &User{
		Username:     username,
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    time.Now(),
	}

	if err := DB.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// getuserbyusername登录时根据用户名查用户是否存在，然后比对密码哈希。
func GetUserByUsername(username string) (*User, error) {
	var user User
	err := DB.Where("username=?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// 通过session里面的user_id获得user
func GetUserByID(userID uint) (*User, error) {
	var user User

	err := DB.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// 获取表内总共多少个序号
func CountUsers() (int64, error) {
	var count int64
	err := DB.Model(&User{}).Count(&count).Error
	return count, err
}

// 个人中心修改用户名
func UpdateUsername(userID uint, newUsername string) error {
	return DB.Model(&User{}).Where("id=?", userID).Update("username", newUsername).Error
}

// 个人中心修改密码,修改完密码之后配合删除session来重新登录,,,这个是吧两个数据库操作封装成一个事务来写
func ChangePasswordAndRevokeSessions(userID uint, passwordHash string) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&User{}).
			Where("id = ?", userID).
			Update("password_hash", passwordHash).Error; err != nil {
			return err
		}

		if err := tx.Where("user_id = ?", userID).
			Delete(&Session{}).Error; err != nil {
			return err
		}

		return nil
	})
}

// ===== session相关 =====

// 登录成功后生成一个随机 sessionID，存进数据库。
func CreateSession(sessionID string, userID uint, expiresAt time.Time) error {
	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	return DB.Create(session).Error
}

// 这个函数后面鉴权中间件必须用。没有它，就没法根据 cookie 里的 session_id 判断用户是谁
func GetSessionByID(sessionID string) (*Session, error) {
	var session Session
	err := DB.Where("id = ?", sessionID).First(&session).Error

	if err != nil {
		return nil, err
	}

	return &session, nil
}

// 退出登录时用
func DeleteSession(sessionID string) error {
	return DB.Where("id=?", sessionID).Delete(&Session{}).Error
}

// 可以后面定时清理用
func DeleteExpiredSessions() error {
	return DB.Where("expires_at <= ?", time.Now()).Delete(&Session{}).Error
}

// ===== 个人中心相关 =====
// 统计个人的文档数量、chunk数量、提问数量
type UserStatistics struct {
	DocumentCount int64 `json:"document_count"`
	ChunkCount    int64 `json:"chunk_count"`
	QuestionCount int64 `json:"question_count"`
}

func GetUserStatistics(userID uint) (*UserStatistics, error) {
	var result UserStatistics

	// 当前用户上传的文档数量
	if err := DB.Model(&Document{}).
		Where("user_id = ?", userID).
		Count(&result.DocumentCount).Error; err != nil {
		return nil, err
	}

	// 当前用户上传文档的 Chunk 总数
	if err := DB.Model(&Document{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(chunk_count), 0)").
		Scan(&result.ChunkCount).Error; err != nil {
		return nil, err
	}

	// role=user 表示用户发出的问题，不统计 AI 回答
	if err := DB.Model(&ChatHistory{}).
		Where("user_id = ? AND role = ?", userID, "user").
		Count(&result.QuestionCount).Error; err != nil {
		return nil, err
	}

	return &result, nil
}
