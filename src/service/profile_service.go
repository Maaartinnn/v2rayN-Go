package service

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"v2rayn-go/coredef"
	"v2rayn-go/database"
	"v2rayn-go/parser"

	"gorm.io/gorm"
)

// ProfileService 节点业务逻辑层
type ProfileService struct{}

// NewProfileService 创建节点服务
func NewProfileService() *ProfileService {
	return &ProfileService{}
}

// List 按条件查询节点列表，始终按 sort_order ASC 排序。
// 预留排序扩展点：未来可通过 sortBy 参数支持按名称、延迟等排序。
//   - groupUUID: 可选，按分组 UUID 精确过滤
//   - q: 可选，文本搜索，LIKE 匹配 name/proxy_address/proxy_protocol 及关联的 group alias
func (s *ProfileService) List(groupUUID, q string) ([]database.Profile, error) {
	var profiles []database.Profile
	tx := database.DB.Model(&database.Profile{})

	// 按分组 UUID 精确过滤
	if groupUUID != "" {
		tx = tx.Where("profiles.group_uuid = ?", groupUUID)
	}

	// 文本搜索：LIKE 匹配 name/proxy_address/proxy_protocol 及关联的 group alias
	if q != "" {
		like := "%" + q + "%"
		tx = tx.Joins("LEFT JOIN node_groups ON node_groups.uuid = profiles.group_uuid").
			Where("profiles.name LIKE ? OR profiles.proxy_address LIKE ? OR profiles.proxy_protocol LIKE ? OR node_groups.alias LIKE ?",
				like, like, like, like)
	}

	if err := tx.Order("profiles.sort_order ASC").Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("failed to list profiles: %w", err)
	}
	return profiles, nil
}

// ListSummary 返回精简的节点列表（仅查询展示所需字段 + 后端计算颜色）。
// 与 List 不同，此方法不返回 raw_link、proxy_credential 等大字段，
// 适合前端列表页面展示，减少传输数据量。
func (s *ProfileService) ListSummary(groupUUID, q string) ([]database.ProfileListItem, error) {
	var profiles []database.Profile
	tx := database.DB.Model(&database.Profile{})

	// 复用 List 的筛选逻辑：按分组 UUID 精确过滤
	if groupUUID != "" {
		tx = tx.Where("profiles.group_uuid = ?", groupUUID)
	}

	// 文本搜索：LIKE 匹配 name/proxy_address/proxy_protocol 及关联的 group alias
	if q != "" {
		like := "%" + q + "%"
		tx = tx.Joins("LEFT JOIN node_groups ON node_groups.uuid = profiles.group_uuid").
			Where("profiles.name LIKE ? OR profiles.proxy_address LIKE ? OR profiles.proxy_protocol LIKE ? OR node_groups.alias LIKE ?",
				like, like, like, like)
	}

	// 只查询列表展示所需的字段，避免传输 raw_link 等大字段
	if err := tx.Order("profiles.sort_order ASC").
		Select("profiles.uuid", "profiles.name", "profiles.proxy_protocol",
			"profiles.proxy_address", "profiles.proxy_port", "profiles.core_type",
			"profiles.test_result", "profiles.is_active", "profiles.group_uuid").
		Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("failed to list profiles: %w", err)
	}

	// 转换为精简 DTO 并计算颜色
	items := make([]database.ProfileListItem, len(profiles))
	for i, p := range profiles {
		items[i] = database.ProfileListItem{
			UUID:          p.UUID,
			Name:          p.Name,
			ProxyProtocol: p.ProxyProtocol,
			ProxyAddress:  p.ProxyAddress,
			ProxyPort:     p.ProxyPort,
			CoreType:      p.CoreType,
			TestResult:    p.TestResult,
			IsActive:      p.IsActive,
			GroupUUID:     p.GroupUUID,
			ProtocolColor: database.GetProtocolColor(p.ProxyProtocol),
			CoreColor:     database.GetCoreColor(p.CoreType),
			LatencyColor:  database.GetLatencyColor(p.TestResult),
		}
	}
	return items, nil
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
// 策略组节点（proxy_protocol 为策略类型）跳过代理字段验证，校验环路。
func (s *ProfileService) Create(profile *database.Profile) error {
	if profile.GroupUUID == "" {
		return NewValidation("group_uuid is required", nil)
	}
	var group database.NodeGroup
	if err := database.DB.Where("uuid = ?", profile.GroupUUID).First(&group).Error; err != nil {
		return NewNotFound("group not found", err)
	}

	// 策略组节点校验：检查环路
	if coredef.IsStrategyProtocol(profile.ProxyProtocol) {
		if err := s.validateStrategyCycle(profile); err != nil {
			return err
		}
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
// 使用事务保证 deactivate-all + activate-one 的原子性，防止并发 Select 导致多个节点同时 is_active=true
func (s *ProfileService) Select(uuid string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var profile database.Profile
		if err := tx.Where("uuid = ?", uuid).First(&profile).Error; err != nil {
			return NewNotFound("profile not found", err)
		}
		if err := tx.Model(&database.Profile{}).Where("is_active = ?", true).Update("is_active", false).Error; err != nil {
			return fmt.Errorf("failed to deactivate profiles: %w", err)
		}
		if err := tx.Model(&profile).Update("is_active", true).Error; err != nil {
			return fmt.Errorf("failed to activate profile: %w", err)
		}
		return nil
	})
}

// Update 更新节点
func (s *ProfileService) Update(uuid string, updates map[string]any) (*database.Profile, error) {
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

	// 策略组节点校验：检查环路
	if proto, ok := updates["proxy_protocol"].(string); ok && coredef.IsStrategyProtocol(proto) {
		// 构建临时 Profile 用于环路检测
		temp := profile
		if m, ok := updates["strategy_member_uuids"].(string); ok {
			temp.StrategyMemberUUIDs = m
		}
		if err := s.validateStrategyCycle(&temp); err != nil {
			return nil, err
		}
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

// Delete 删除指定节点（级联清理其他策略组中的引用）
func (s *ProfileService) Delete(uuid string) error {
	var profile database.Profile
	if err := database.DB.Where("uuid = ?", uuid).First(&profile).Error; err != nil {
		return NewNotFound("profile not found", err)
	}

	// 清理其他策略组中对被删节点的引用
	s.cleanStrategyMemberRefs(uuid)

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

// validateStrategyCycle 校验策略组是否存在循环嵌套
func (s *ProfileService) validateStrategyCycle(profile *database.Profile) error {
	// 构建 profileMap（查询所有 Profile）
	var allProfiles []database.Profile
	if err := database.DB.Find(&allProfiles).Error; err != nil {
		return fmt.Errorf("failed to load profiles for cycle check: %w", err)
	}
	profileMap := make(map[string]*database.Profile, len(allProfiles)+1)
	for i := range allProfiles {
		profileMap[allProfiles[i].UUID] = &allProfiles[i]
	}
	// 加入当前待创建/更新的 Profile（覆盖 DB 中的旧版本）
	profileMap[profile.UUID] = profile

	return database.CheckStrategyCycle(profile.UUID, profileMap)
}

// cleanStrategyMemberRefs 清理所有策略组中对被删节点 UUID 的引用
func (s *ProfileService) cleanStrategyMemberRefs(deletedUUID string) {
	var strategyProfiles []database.Profile
	database.DB.Where("proxy_protocol IN ?",
		[]string{coredef.ProtocolSelector, coredef.ProtocolURLTest, coredef.ProtocolFallback, coredef.ProtocolLoadBalance}).
		Find(&strategyProfiles)

	for _, sp := range strategyProfiles {
		if sp.StrategyMemberUUIDs == "" {
			continue
		}
		var uuids []string
		if err := json.Unmarshal([]byte(sp.StrategyMemberUUIDs), &uuids); err != nil {
			continue
		}
		var cleaned []string
		for _, uid := range uuids {
			if uid != deletedUUID {
				cleaned = append(cleaned, uid)
			}
		}
		if len(cleaned) != len(uuids) {
			newJSON, _ := json.Marshal(cleaned)
			if err := database.DB.Model(&sp).Update("strategy_member_uuids", string(newJSON)).Error; err != nil {
				slog.Warn("failed to clean strategy member refs", "profile", sp.UUID, "error", err)
			}
		}
	}
}
