package service

import (
	"fmt"

	"v2rayn-go/database"
)

// RoutingRuleService 路由规则业务逻辑层
type RoutingRuleService struct{}

// NewRoutingRuleService 创建路由规则服务
func NewRoutingRuleService() *RoutingRuleService {
	return &RoutingRuleService{}
}

// List 获取所有路由规则，按 sort_order 排序
func (s *RoutingRuleService) List() ([]database.RoutingRule, error) {
	var rules []database.RoutingRule
	if err := database.DB.Order("sort_order ASC").Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("failed to list routing rules: %w", err)
	}
	return rules, nil
}

// Get 根据 UUID 获取单个路由规则
func (s *RoutingRuleService) Get(uuid string) (*database.RoutingRule, error) {
	var rule database.RoutingRule
	if err := database.DB.Where("uuid = ?", uuid).First(&rule).Error; err != nil {
		return nil, NewNotFound("rule not found", err)
	}
	return &rule, nil
}

// Create 创建路由规则
func (s *RoutingRuleService) Create(rule *database.RoutingRule) error {
	rule.SortOrder = database.SortNew(&database.RoutingRule{})
	rule.UUID = database.GenerateUUID()
	if err := database.DB.Create(rule).Error; err != nil {
		return NewValidation("failed to create routing rule", err)
	}
	return nil
}

// Update 更新路由规则
func (s *RoutingRuleService) Update(uuid string, updated *database.RoutingRule) (*database.RoutingRule, error) {
	var rule database.RoutingRule
	if err := database.DB.Where("uuid = ?", uuid).First(&rule).Error; err != nil {
		return nil, NewNotFound("rule not found", err)
	}
	updated.ID = rule.ID
	if err := database.DB.Save(updated).Error; err != nil {
		return nil, fmt.Errorf("failed to update routing rule: %w", err)
	}
	return updated, nil
}

// Delete 删除路由规则
func (s *RoutingRuleService) Delete(uuid string) error {
	var rule database.RoutingRule
	if err := database.DB.Where("uuid = ?", uuid).First(&rule).Error; err != nil {
		return NewNotFound("rule not found", err)
	}
	if err := database.DB.Delete(&rule).Error; err != nil {
		return fmt.Errorf("failed to delete rule: %w", err)
	}
	return nil
}
