package repository

import (
	"rubick/internal/database"
	"rubick/internal/model"
)

// ListComposeProjects 获取 Compose 项目列表
func ListComposeProjects(hostID string) ([]model.ComposeProject, error) {
	var projects []model.ComposeProject
	query := database.GetDB().Preload("Host")
	if hostID != "" {
		query = query.Where("host_id = ?", hostID)
	}
	if err := query.Order("name ASC").Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

// GetComposeProjectByID 根据 ID 获取 Compose 项目
func GetComposeProjectByID(id string) (*model.ComposeProject, error) {
	var project model.ComposeProject
	if err := database.GetDB().Preload("Host").First(&project, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

// CreateComposeProject 创建 Compose 项目
func CreateComposeProject(project *model.ComposeProject) error {
	return database.GetDB().Create(project).Error
}

// UpdateComposeProject 更新 Compose 项目
func UpdateComposeProject(id string, updates *model.ComposeProject) error {
	return database.GetDB().Model(&model.ComposeProject{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteComposeProject 删除 Compose 项目
func DeleteComposeProject(id string) error {
	return database.GetDB().Delete(&model.ComposeProject{}, "id = ?", id).Error
}

// UpdateComposeProjectStatus 更新 Compose 项目状态
func UpdateComposeProjectStatus(id string, status string) error {
	return database.GetDB().Model(&model.ComposeProject{}).Where("id = ?", id).Update("status", status).Error
}
