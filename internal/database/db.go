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
	if err := DB.AutoMigrate(&Document{}, &ChatHistory{}); err != nil {
		return err
	}

	slog.Info("MYSQL连接成功，表结构已同步")
	return nil
}

// ===== 文档相关 =====

// CreateDocument 创建文档记录
func CreateDocument(filename string, filesize int64, chunkcount int, status string) (*Document, error) {
	doc := &Document{
		Filename:   filename,
		FileSize:   filesize,
		ChunkCount: chunkcount,
		Status:     status,
		CreatedAt:  time.Now(),
	}
	if err := DB.Create(doc).Error; err != nil {
		return nil, err
	}

	return doc, nil
}

// ListDocuments 查询所有已就绪的文档
func ListDocuments() ([]Document, error) {
	var docs []Document
	err := DB.Where("status = ?", "ready").Order("created_at DESC").Find(&docs).Error
	return docs, err
}

// ===== 对话相关 =====

// SaveMessage 保存一条对话记录
func SaveMessage(sessionid, role, content string) error {
	msg := &ChatHistory{
		SessionID: sessionid,
		Role:      role,
		Content:   content,
		CreatedAt: time.Now(),
	}

	return DB.Create(msg).Error
}

// GetSessionHistory 获取某个session会话历史记录
func GetSessionHistory(sessionid string) ([]ChatHistory, error) {
	var history []ChatHistory
	err := DB.Where("session_id = ?", sessionid).Order("created_at DESC").Find(&history).Error
	return history, err
}

// DocumentExists 检查文件是否已导入
func DocumentExists(filename string) bool {
	var count int64
	DB.Model(&Document{}).Where("filename = ? AND status = ?", filename, "ready").Count(&count)
	return count > 0
}

// GetDocumentByFilename 根据文件名查一条文档记录
func GetDocumentByFilename(filename string) (*Document, error) {
	var doc Document
	err := DB.Where("filename = ? AND status = ?", filename, "ready").First(&doc).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func DeleteDocument(filename string) error {
	return DB.Where("filename = ?", filename).Delete(&Document{}).Error
}
