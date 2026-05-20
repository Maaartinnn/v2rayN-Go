package service

import (
	"fmt"

	"v2rayn-go/database"
)

// StrategyGroupService 策略组业务逻辑层
type StrategyGroupService struct{}

// NewStrategyGroupService 创建策略组服务
func NewStrategyGroupService() *StrategyGroupService {
	return &StrategyGroupService{}
}

// List 获取所有策略组，按 sort_order 排序
func (s *StrategyGroupService) List() ([]database.StrategyGroup, error) {
	var groups []database.StrategyGroup
	if err := database.DB.Order("sort_order ASC").Find(&groups).Error; err != nil {
		return nil, fmt.Errorf("failed to list strategy groups: %w", err)
	}
	return groups, nil
}

// Get 根据 UUID 获取单个策略组
func (s *StrategyGroupService) Get(uuid string) (*database.StrategyGroup, error) {
	var group database.StrategyGroup
	if err := database.DB.Where("uuid = ?", uuid).First(&group).Error; err != nil {
		return nil, NewNotFound("strategy group not found", err)
	}
	return &group, nil
}

// Create 创建策略组
func (s *StrategyGroupService) Create(group *database.StrategyGroup) error {
	group.SortOrder = database.SortNew(&database.StrategyGroup{})
	group.UUID = database.GenerateUUID()
	if err := database.DB.Create(group).Error; err != nil {
		return NewValidation("failed to create strategy group", err)
	}
	return nil
}

// Update 更新策略组
func (s *StrategyGroupService) Update(uuid string, updated *database.StrategyGroup) (*database.StrategyGroup, error) {
	var group database.StrategyGroup
	if err := database.DB.Where("uuid = ?", uuid).First(&group).Error; err != nil {
		return nil, NewNotFound("strategy group not found", err)
	}
	updated.ID = group.ID
	if err := database.DB.Save(updated).Error; err != nil {
		return nil, fmt.Errorf("failed to update strategy group: %w", err)
	}
	return updated, nil
}

// Delete 删除策略组
func (s *StrategyGroupService) Delete(uuid string) error {
	var group database.StrategyGroup
	if err := database.DB.Where("uuid = ?", uuid).First(&group).Error; err != nil {
		return NewNotFound("strategy group not found", err)
	}
	if err := database.DB.Delete(&group).Error; err != nil {
		return fmt.Errorf("failed to delete strategy group: %w", err)
	}
	return nil
}
