package service

import (
	"encoding/json"
	"fmt"

	"v2rayn-go/database"

	"gorm.io/gorm"
)

// GroupService 分组业务逻辑层
type GroupService struct{}

// NewGroupService 创建分组服务
func NewGroupService() *GroupService {
	return &GroupService{}
}

// List 获取所有分组，按 sort_order 排序，填充 node_count
func (s *GroupService) List() ([]database.NodeGroup, error) {
	var groups []database.NodeGroup
	if err := database.DB.Order("sort_order ASC").Find(&groups).Error; err != nil {
		return nil, NewNotFound("failed to list groups", err)
	}
	for i := range groups {
		var count int64
		database.DB.Model(&database.Profile{}).Where("group_uuid = ?", groups[i].UUID).Count(&count)
		groups[i].NodeCount = int(count)
	}
	return groups, nil
}

// Get 根据 UUID 获取单个分组，填充 node_count
func (s *GroupService) Get(uuid string) (*database.NodeGroup, error) {
	var group database.NodeGroup
	if err := database.DB.Where("uuid = ?", uuid).First(&group).Error; err != nil {
		return nil, NewNotFound("group not found", err)
	}
	var count int64
	database.DB.Model(&database.Profile{}).Where("group_uuid = ?", group.UUID).Count(&count)
	group.NodeCount = int(count)
	return &group, nil
}

// Create 创建分组
func (s *GroupService) Create(group *database.NodeGroup) error {
	if group.UUID == "" {
		group.UUID = database.GenerateUUID()
	}
	group.SortOrder = database.SortNew(&database.NodeGroup{})
	if err := database.DB.Create(group).Error; err != nil {
		return NewValidation("failed to create group", err)
	}
	return nil
}

// Update 更新分组
func (s *GroupService) Update(uuid string, updated *database.NodeGroup) (*database.NodeGroup, error) {
	var group database.NodeGroup
	if err := database.DB.Where("uuid = ?", uuid).First(&group).Error; err != nil {
		return nil, NewNotFound("group not found", err)
	}
	updated.ID = group.ID
	if updated.UUID == "" {
		updated.UUID = group.UUID
	}
	updated.SortOrder = group.SortOrder
	if err := database.DB.Save(updated).Error; err != nil {
		return nil, fmt.Errorf("failed to update group: %w", err)
	}
	return updated, nil
}

// Delete 删除分组（含事务保护：级联删除节点 + 清理策略组脏引用）
func (s *GroupService) Delete(uuid string) error {
	var group database.NodeGroup
	if err := database.DB.Where("uuid = ?", uuid).First(&group).Error; err != nil {
		return NewNotFound("group not found", err)
	}

	// 检查是否为最后一个分组
	var count int64
	if err := database.DB.Model(&database.NodeGroup{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to count groups: %w", err)
	}
	if count <= 1 {
		return NewConflict("cannot delete the last group", nil)
	}

	// 使用事务保证原子性
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 查出被删节点的 UUID 列表
		var deletedProfileUUIDs []string
		if err := tx.Model(&database.Profile{}).Where("group_uuid = ?", group.UUID).Pluck("uuid", &deletedProfileUUIDs).Error; err != nil {
			return fmt.Errorf("failed to query profiles: %w", err)
		}

		// 2. 删除该分组下的所有节点
		if err := tx.Where("group_uuid = ?", group.UUID).Delete(&database.Profile{}).Error; err != nil {
			return fmt.Errorf("failed to delete profiles: %w", err)
		}

		// 3. 清理 StrategyGroup 中的脏引用
		if len(deletedProfileUUIDs) > 0 {
			deletedSet := make(map[string]bool, len(deletedProfileUUIDs))
			for _, uid := range deletedProfileUUIDs {
				deletedSet[uid] = true
			}
			var strategyGroups []database.StrategyGroup
			tx.Find(&strategyGroups)
			for _, sg := range strategyGroups {
				if sg.ProfileUUIDs == "" {
					continue
				}
				var uuids []string
				if err := json.Unmarshal([]byte(sg.ProfileUUIDs), &uuids); err != nil {
					continue
				}
				var cleaned []string
				for _, uid := range uuids {
					if !deletedSet[uid] {
						cleaned = append(cleaned, uid)
					}
				}
				if len(cleaned) != len(uuids) {
					newJSON, _ := json.Marshal(cleaned)
					if err := tx.Model(&sg).Update("profile_uuids", string(newJSON)).Error; err != nil {
						return fmt.Errorf("failed to clean strategy group refs: %w", err)
					}
				}
			}
		}

		// 4. 删除分组本身
		if err := tx.Delete(&group).Error; err != nil {
			return fmt.Errorf("failed to delete group: %w", err)
		}

		return nil
	})
}

// Reorder 重排序分组，返回新的 sort_order
func (s *GroupService) Reorder(uuid string, beforeUUID, afterUUID string) (int, error) {
	if uuid == "" {
		return 0, NewValidation("uuid is required", nil)
	}

	var beforeOrder, afterOrder *int

	if beforeUUID != "" {
		var group database.NodeGroup
		if err := database.DB.Where("uuid = ?", beforeUUID).First(&group).Error; err == nil {
			v := group.SortOrder
			beforeOrder = &v
		}
	}
	if afterUUID != "" {
		var group database.NodeGroup
		if err := database.DB.Where("uuid = ?", afterUUID).First(&group).Error; err == nil {
			v := group.SortOrder
			afterOrder = &v
		}
	}

	newOrder := database.SortInsert(beforeOrder, afterOrder)

	if err := database.DB.Model(&database.NodeGroup{}).Where("uuid = ?", uuid).Update("sort_order", newOrder).Error; err != nil {
		return 0, fmt.Errorf("failed to reorder: %w", err)
	}

	return newOrder, nil
}
