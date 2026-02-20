package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditLog 审计日志
type AuditLog struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"index" json:"user_id,omitempty"` // 用户 ID（可选）
	Method    string    `gorm:"index;not null" json:"method"`   // HTTP 方法
	Path      string    `gorm:"index;not null" json:"path"`     // 请求路径
	Status    int       `json:"status"`                         // 响应状态码
	IP        string    `json:"ip"`                             // 客户端 IP
	UserAgent string    `json:"user_agent"`                     // User-Agent
	Latency   int64     `json:"latency"`                        // 响应时间（毫秒）
	Message   string    `json:"message"`                        // 日志消息
	CreatedAt time.Time `json:"created_at"`
}

// BeforeCreate 创建前钩子
func (l *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	return nil
}
