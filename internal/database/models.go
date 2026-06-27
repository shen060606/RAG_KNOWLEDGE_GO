package database

import "time"

//Document 上传文件记录
type Document struct {
	ID         uint   `gorm:"primary_key"`
	Filename   string `gorm:"size:255"`
	FileSize   int64
	ChunkCount int
	Status     string `gorm:"size:20;default:ready"` // ready / processing
	CreatedAt  time.Time
}

//ChatHistory 对话记录
type ChatHistory struct {
	ID        uint   `gorm:"primary_key"`
	SessionID string `gorm:"size:255"`  //会话ID
	Role      string `gorm:"size:20"`   //user / assistant
	Content   string `gorm:"type:text"` //对话内容
	CreatedAt time.Time
}
