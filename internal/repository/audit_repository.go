package repository

import (
	"rubick/internal/database"
	"rubick/internal/model"
)

// CreateAuditLog 创建审计日志
func CreateAuditLog(log *model.AuditLog) error {
	return database.GetDB().Create(log).Error
}

// ListAuditLogs 获取审计日志列表
func ListAuditLogs(page, pageSize int, method, path string) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64

	query := database.GetDB().Model(&model.AuditLog{})

	if method != "" {
		query = query.Where("method = ?", method)
	}
	if path != "" {
		query = query.Where("path LIKE ?", "%"+path+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
