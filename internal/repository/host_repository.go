package repository

import (
	"rubick/internal/crypto"
	"rubick/internal/database"
	"rubick/internal/model"

	"gorm.io/gorm"
)

// ListHosts 获取所有主机
func ListHosts() ([]model.Host, error) {
	var hosts []model.Host
	if err := database.GetDB().Order("is_default DESC, name ASC").Find(&hosts).Error; err != nil {
		return nil, err
	}
	return hosts, nil
}

// GetHostByID 根据 ID 获取主机
func GetHostByID(id string) (*model.Host, error) {
	var host model.Host
	if err := database.GetDB().First(&host, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &host, nil
}

// GetHostByName 根据名称获取主机
func GetHostByName(name string) (*model.Host, error) {
	var host model.Host
	if err := database.GetDB().First(&host, "name = ?", name).Error; err != nil {
		return nil, err
	}
	return &host, nil
}

// GetDefaultHost 获取默认主机
func GetDefaultHost() (*model.Host, error) {
	var host model.Host
	if err := database.GetDB().First(&host, "is_default = ?", true).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果没有默认主机，返回本地主机
			return GetHostByName("local")
		}
		return nil, err
	}
	return &host, nil
}

// CreateHost 创建主机
func CreateHost(host *model.Host) error {
	return database.GetDB().Create(host).Error
}

// UpdateHost 更新主机
func UpdateHost(id string, updates *model.Host) error {
	// 构建更新映射，只更新非空字段
	updateMap := map[string]interface{}{}

	if updates.Name != "" {
		updateMap["name"] = updates.Name
	}
	if updates.Host != "" {
		updateMap["host"] = updates.Host
	}
	if updates.Type != "" {
		updateMap["type"] = updates.Type
	}

	// 布尔字段直接更新
	updateMap["is_default"] = updates.IsDefault
	updateMap["is_active"] = updates.IsActive
	updateMap["skip_tls_verify"] = updates.SkipTLSVerify

	// 可选字段
	if updates.Description != "" {
		updateMap["description"] = updates.Description
	}
	if updates.DockerPort != 0 {
		updateMap["docker_port"] = updates.DockerPort
	}

	// SSH 配置 - 只在 SSH 类型且有值时更新
	if updates.SSHUser != "" {
		updateMap["ssh_user"] = updates.SSHUser
	}
	if updates.SSHAuthType != "" {
		updateMap["ssh_auth_type"] = updates.SSHAuthType
	}
	if updates.SSHPort != 0 {
		updateMap["ssh_port"] = updates.SSHPort
	}

	// 敏感字段 - 只在提供了新值时才更新（需要加密）
	if updates.SSHPassword != "" {
		encrypted, err := crypto.Encrypt(updates.SSHPassword)
		if err != nil {
			return err
		}
		updateMap["ssh_password"] = encrypted
	}
	if updates.SSHPrivateKey != "" {
		encrypted, err := crypto.Encrypt(updates.SSHPrivateKey)
		if err != nil {
			return err
		}
		updateMap["ssh_private_key"] = encrypted
	}

	// TLS 配置
	if updates.TLSCertID != "" {
		updateMap["tls_cert_id"] = updates.TLSCertID
	}

	return database.GetDB().Model(&model.Host{}).Where("id = ?", id).Updates(updateMap).Error
}

// DeleteHost 删除主机
func DeleteHost(id string) error {
	return database.GetDB().Delete(&model.Host{}, "id = ?", id).Error
}

// ClearDefaultHost 清除默认主机的默认标记
func ClearDefaultHost() error {
	return database.GetDB().Model(&model.Host{}).Where("is_default = ?", true).Update("is_default", false).Error
}

// SetDefaultHost 设置默认主机
func SetDefaultHost(id string) error {
	tx := database.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 清除所有默认标记
	if err := tx.Model(&model.Host{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 设置新的默认主机
	if err := tx.Model(&model.Host{}).Where("id = ?", id).Update("is_default", true).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
