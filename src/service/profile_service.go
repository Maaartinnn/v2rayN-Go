package service

import (
	"fmt"
	"strings"

	"v2rayn-go/database"
	"v2rayn-go/parser"
)

// ProfileService 节点业务逻辑层
type ProfileService struct{}

// NewProfileService 创建节点服务
func NewProfileService() *ProfileService {
	return &ProfileService{}
}

// List 获取所有节点，按 sort_order 排序
func (s *ProfileService) List() ([]database.Profile, error) {
	var profiles []database.Profile
	if err := database.DB.Order("sort_order ASC").Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("failed to list profiles: %w", err)
	}
	return profiles, nil
}

// Get 根据 UUID 获取单个节点
func (s *ProfileService) Get(uuid string) (*database.Profile, error) {
	var profile database.Profile
	if err := database.DB.Where("uuid = ?", uuid).First(&profile).Error; err != nil {
		return nil, NewNotFound("profile not found", err)
	}
	return &profile, nil
}

// Create 创建节点（含分组校验、UUID 生成、排序）
func (s *ProfileService) Create(profile *database.Profile) error {
	if profile.GroupUUID == "" {
		return NewValidation("group_uuid is required", nil)
	}
	var group database.NodeGroup
	if err := database.DB.Where("uuid = ?", profile.GroupUUID).First(&group).Error; err != nil {
		return NewNotFound("group not found", err)
	}
	profile.SortOrder = database.SortNewScoped(&database.Profile{}, "group_uuid = ?", profile.GroupUUID)
	profile.UUID = database.GenerateUUID()
	if err := database.DB.Create(profile).Error; err != nil {
		return fmt.Errorf("failed to create profile: %w", err)
	}
	return nil
}

// ImportLinks 解析链接文本并批量导入到指定分组
func (s *ProfileService) ImportLinks(linksText string, groupUUID string) (int, error) {
	if groupUUID == "" {
		return 0, NewValidation("group_uuid is required", nil)
	}
	var group database.NodeGroup
	if err := database.DB.Where("uuid = ?", groupUUID).First(&group).Error; err != nil {
		return 0, NewNotFound("group not found", err)
	}

	profiles, err := parser.ParseLinks(strings.Split(linksText, "\n"))
	if err != nil {
		return 0, fmt.Errorf("failed to parse links: %w", err)
	}

	return s.importParsedProfiles(profiles, groupUUID)
}

// ImportParsedLinks 将已解析的链接列表导入到指定分组（供图片导入等复用）
func (s *ProfileService) ImportParsedLinks(links []string, groupUUID string) (int, error) {
	if groupUUID == "" {
		return 0, NewValidation("group_uuid is required", nil)
	}
	var group database.NodeGroup
	if err := database.DB.Where("uuid = ?", groupUUID).First(&group).Error; err != nil {
		return 0, NewNotFound("group not found", err)
	}

	profiles, err := parser.ParseLinks(links)
	if err != nil {
		return 0, fmt.Errorf("failed to parse links: %w", err)
	}

	return s.importParsedProfiles(profiles, groupUUID)
}

// importParsedProfiles 内部方法：将已解析的 Profile 列表批量写入数据库
func (s *ProfileService) importParsedProfiles(profiles []*database.Profile, groupUUID string) (int, error) {
	seq := database.SortNewBatch(&database.Profile{}, "group_uuid = ?", len(profiles), groupUUID)

	for i, profile := range profiles {
		profile.SortOrder = seq[i]
		profile.GroupUUID = groupUUID
		if err := database.DB.Create(profile).Error; err != nil {
			return 0, fmt.Errorf("failed to create profile %d: %w", i, err)
		}
	}

	return len(profiles), nil
}

// Select 选择指定节点为活跃节点（先取消全部，再激活目标）
func (s *ProfileService) Select(uuid string) error {
	var profile database.Profile
	if err := database.DB.Where("uuid = ?", uuid).First(&profile).Error; err != nil {
		return NewNotFound("profile not found", err)
	}
	if err := database.DB.Model(&database.Profile{}).Where("is_active = ?", true).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to deactivate profiles: %w", err)
	}
	if err := database.DB.Model(&profile).Update("is_active", true).Error; err != nil {
		return fmt.Errorf("failed to activate profile: %w", err)
	}
	return nil
}

// Update 更新节点
func (s *ProfileService) Update(uuid string, updates map[string]interface{}) (*database.Profile, error) {
	var profile database.Profile
	if err := database.DB.Where("uuid = ?", uuid).First(&profile).Error; err != nil {
		return nil, NewNotFound("profile not found", err)
	}

	// 验证分组存在
	groupUUID, _ := updates["group_uuid"].(string)
	if groupUUID == "" {
		return nil, NewValidation("group_uuid is required", nil)
	}
	var group database.NodeGroup
	if err := database.DB.Where("uuid = ?", groupUUID).First(&group).Error; err != nil {
		return nil, NewNotFound("group not found", err)
	}

	// 删除不可修改字段
	delete(updates, "uuid")
	delete(updates, "sort_order")
	delete(updates, "ID")
	delete(updates, "id")

	if err := database.DB.Model(&profile).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}
	// 重新加载以获取完整数据
	if err := database.DB.First(&profile, profile.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload profile: %w", err)
	}
	return &profile, nil
}

// Delete 删除指定节点
func (s *ProfileService) Delete(uuid string) error {
	var profile database.Profile
	if err := database.DB.Where("uuid = ?", uuid).First(&profile).Error; err != nil {
		return NewNotFound("profile not found", err)
	}
	if err := database.DB.Delete(&profile).Error; err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}
	return nil
}

// DedupResult 去重结果
type DedupResult struct {
	Removed int `json:"removed"`
	Total   int `json:"total"`
}

// Dedup 去重节点（可按分组过滤）
func (s *ProfileService) Dedup(groupUUID string) (*DedupResult, error) {
	var profiles []database.Profile
	query := database.DB.Order("sort_order ASC")
	if groupUUID != "" {
		query = query.Where("group_uuid = ?", groupUUID)
	}
	if err := query.Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("failed to list profiles: %w", err)
	}

	seen := make(map[string]bool)
	var duplicates []uint

	for _, p := range profiles {
		key := p.RawLink
		if idx := strings.LastIndex(key, "#"); idx != -1 {
			key = key[:idx]
		}
		if key == "" {
			key = fmt.Sprintf("%s:%d:%s", p.ProxyAddress, p.ProxyPort, p.ProxyProtocol)
			if p.ProxyCredential != "" {
				key += ":" + p.ProxyCredential
			}
		}
		if seen[key] {
			duplicates = append(duplicates, p.ID)
		} else {
			seen[key] = true
		}
	}

	if len(duplicates) > 0 {
		if err := database.DB.Delete(&database.Profile{}, duplicates).Error; err != nil {
			return nil, fmt.Errorf("failed to delete duplicates: %w", err)
		}
	}

	return &DedupResult{
		Removed: len(duplicates),
		Total:   len(profiles),
	}, nil
}
