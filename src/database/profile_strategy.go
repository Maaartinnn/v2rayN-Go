package database

import (
	"encoding/json"
	"fmt"

	"v2rayn-go/coredef"
)

// ParseMemberUUIDs 解析策略组的成员 UUID 列表
func ParseMemberUUIDs(sg *Profile) ([]string, error) {
	if sg.StrategyMemberUUIDs == "" {
		return nil, nil
	}
	var uuids []string
	if err := json.Unmarshal([]byte(sg.StrategyMemberUUIDs), &uuids); err != nil {
		return nil, fmt.Errorf("invalid strategy_member_uuids: %w", err)
	}
	return uuids, nil
}

// ResolveStrategyMembers 解析策略组成员，返回有效的 Profile 映射 + 孤儿 UUID 列表
// 这是数据库到配置生成之间的翻译层
func ResolveStrategyMembers(sg *Profile, profileMap map[string]*Profile) (members []*Profile, orphanUUIDs []string) {
	uuids, err := ParseMemberUUIDs(sg)
	if err != nil {
		return nil, nil
	}
	for _, uid := range uuids {
		if p, ok := profileMap[uid]; ok {
			members = append(members, p)
		} else {
			orphanUUIDs = append(orphanUUIDs, uid)
		}
	}
	return
}

// CheckStrategyCycle 检查策略组是否存在循环嵌套（DFS 环路检测）
// 返回 error 表示检测到循环，error 消息包含循环路径
func CheckStrategyCycle(startUUID string, profileMap map[string]*Profile) error {
	visited := make(map[string]bool)
	path := []string{}
	return dfsCycleCheck(startUUID, profileMap, visited, path)
}

// dfsCycleCheck 深度优先搜索环路检测
func dfsCycleCheck(currentUUID string, profileMap map[string]*Profile, visited map[string]bool, path []string) error {
	// 检查当前节点是否已在路径中（环路）
	if visited[currentUUID] {
		// 找到环路，构建错误信息
		cycleStart := -1
		for i, uid := range path {
			if uid == currentUUID {
				cycleStart = i
				break
			}
		}
		if cycleStart >= 0 {
			return fmt.Errorf("检测到策略组循环嵌套: %v", append(path[cycleStart:], currentUUID))
		}
		return fmt.Errorf("检测到策略组循环嵌套: %v", append(path, currentUUID))
	}

	// 查找当前 Profile
	p, ok := profileMap[currentUUID]
	if !ok {
		return nil // 节点不存在，不算错误
	}

	// 只对策略组类型进行递归检测
	if !coredef.IsStrategyProtocol(p.ProxyProtocol) {
		return nil
	}

	// 标记为已访问，加入路径
	visited[currentUUID] = true
	path = append(path, currentUUID)

	// 解析成员 UUID 列表并递归检查
	memberUUIDs, err := ParseMemberUUIDs(p)
	if err != nil {
		return nil // 解析失败，跳过
	}

	for _, memberUUID := range memberUUIDs {
		if err := dfsCycleCheck(memberUUID, profileMap, visited, path); err != nil {
			return err
		}
	}

	// 回溯：取消标记（允许其他路径访问）
	visited[currentUUID] = false
	return nil
}
