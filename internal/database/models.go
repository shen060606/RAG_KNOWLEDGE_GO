package database

import "time"

//Document 上传文件记录
type Document struct {
	ID         uint   `gorm:"primary_key"`
	UserID     uint   `gorm:"index;not null"` //用户ID
	Filename   string `gorm:"size:255"`
	FileSize   int64
	ChunkCount int
	Status     string `gorm:"size:20;default:ready"` // ready / processing
	CreatedAt  time.Time
}

//ChatHistory 对话记录
type ChatHistory struct {
	ID        uint   `gorm:"primary_key"`
	UserID    uint   `gorm:"index;not null"` //用户ID
	SessionID string `gorm:"size:255"`       //会话ID
	Role      string `gorm:"size:20"`        //user / assistant
	Content   string `gorm:"type:text"`      //对话内容
	CreatedAt time.Time
}

type User struct {
	ID           uint   `gorm:"primary_key"`
	Username     string `gorm:"size:50;uniqueIndex;not null"`
	PasswordHash string `gorm:"size:255;not null"`
	CreatedAt    time.Time
}

type Session struct {
	ID        string    `gorm:"primaryKey;size:64"`
	UserID    uint      `gorm:"index;not null"`
	ExpiresAt time.Time `gorm:"index;not null"` //会话过期时间
	CreatedAt time.Time
}
