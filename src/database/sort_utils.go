package database

import (
	"errors"
	"fmt"
	"log"
	"math"
	"reflect"

	"gorm.io/gorm"
)

// SortStep 排序步长，集中定义方便修改
const SortStep = 10

// ========== 新增记录 ==========

// SortNew 全局追加：max(sort_order) + SortStep
func SortNew(model interface{}) int {
	return SortNewScoped(model, "1 = 1")
}

// SortNewScoped 限定范围追加：max(sort_order) + SortStep
// query 可以是 "group_uuid = ?" 之类的条件
func SortNewScoped(model interface{}, query string, args ...interface{}) int {
	var maxOrder int
	tableName := getTableName(model)
	databaseQuery := DB.Model(model).Table(tableName).Select("COALESCE(MAX(sort_order), 0)")
	if query != "1 = 1" {
		databaseQuery = databaseQuery.Where(query, args...)
	}
	databaseQuery.Scan(&maxOrder)
	return mustAdd(maxOrder, SortStep)
}

// SortBetween 拖拽插入：插入到两条记录之间
// 使用 before + (after-before)/2 避免 before+after 整数溢出
func SortBetween(before, after int) int {
	return before + (after-before)/2
}

// SortSequence 批量生成步长序列 [10, 20, 30, ...]
func SortSequence(count int) []int {
	seq := make([]int, count)
	for i := 0; i < count; i++ {
		seq[i] = mustAdd(0, (i+1)*SortStep)
	}
	return seq
}

// ========== 全表/范围重排 ==========

// Rebalance 全表重排：检查当前 sort_order 是否已符合 (i+1)*SortStep，
// 全部一致则不动数据库，返回 false；有不一致则批量 UPDATE，返回 true。
func Rebalance(model interface{}) bool {
	return RebalanceScoped(model, "1 = 1")
}

// RebalanceScoped 限定范围重排（使用全局 DB）
func RebalanceScoped(model interface{}, query string, args ...interface{}) bool {
	return RebalanceScopedTx(DB, model, query, args...)
}

// RebalanceScopedTx 限定范围重排（使用指定事务/连接）
// 调用方可传入 DB 或事务内的 tx
func RebalanceScopedTx(tx *gorm.DB, model interface{}, query string, args ...interface{}) bool {
	type row struct {
		ID        uint
		SortOrder int
	}
	var rows []row
	tableName := getTableName(model)
	q := tx.Model(model).Table(tableName).Select("id, sort_order").Order("sort_order ASC, id ASC")
	if query != "1 = 1" {
		q = q.Where(query, args...)
	}
	q.Find(&rows)

	if len(rows) == 0 {
		return false
	}

	// 检查是否已经符合要求
	needRebalance := false
	for i, r := range rows {
		expected := (i + 1) * SortStep
		if r.SortOrder != expected {
			needRebalance = true
			break
		}
	}
	if !needRebalance {
		return false
	}

	// 执行重排（在同一个 tx 内，不开子事务）
	for i, r := range rows {
		newOrder := (i + 1) * SortStep
		if r.SortOrder != newOrder {
			if err := tx.Table(tableName).Where("id = ?", r.ID).Update("sort_order", newOrder).Error; err != nil {
				log.Printf("[WARN] Rebalance failed for %s id=%d: %v", tableName, r.ID, err)
				return false
			}
		}
	}

	log.Printf("[INFO] Rebalanced sort_order for %s (%d records)", tableName, len(rows))
	return true
}

// RebalanceAll 在程序启动时对所有有序表执行重排检查
func RebalanceAll() {
	Rebalance(&NodeGroup{})
	Rebalance(&RoutingRule{})
	Rebalance(&StrategyGroup{})

	// Profile 按每个分组单独重排
	var groups []NodeGroup
	DB.Find(&groups)
	for _, g := range groups {
		RebalanceScoped(&Profile{}, "group_uuid = ?", g.UUID)
	}
}

// ========== 内部辅助 ==========

// getTableName 通过反射或 GORM 获取表名
func getTableName(model interface{}) string {
	// 优先使用 GORM Tabler 接口
	if t, ok := model.(interface{ TableName() string }); ok {
		return t.TableName()
	}
	// 通过反射取类型名并转 snake_case
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return toSnakeCase(t.Name()) + "s"
}

// toSnakeCase CamelCase -> snake_case
func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, r+32) // to lower
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// SortInsert 根据前后元素的排序值计算插入位置
// before == nil 表示拖到第一个位置：after - SortStep
// after == nil 表示拖到最后一个位置：before + SortStep
// 都非 nil 表示插入中间：SortBetween(before, after)
// 溢出时返回 SortStep（由调用方触发 rebalance 重建间距）
func SortInsert(before, after *int) int {
	if before == nil && after == nil {
		return SortStep // 只有一个元素，给默认值
	} else if before == nil {
		return mustSub(*after, SortStep)
	} else if after == nil {
		return mustAdd(*before, SortStep)
	}
	return SortBetween(*before, *after)
}

// SortInsertSafe 带冲突检测的插入排序
// 返回 (计算结果, 是否冲突)。冲突 = 结果与邻居值相同（整数除法坍缩）
// 调用方检测到冲突时应触发 Rebalance 后重新查询邻居值再重算
func SortInsertSafe(before, after *int) (int, bool) {
	result := SortInsert(before, after)
	if before != nil && result == *before {
		return result, true
	}
	if after != nil && result == *after {
		return result, true
	}
	return result, false
}

// SortNewBatch 批量追加：为 count 个新记录生成排序值，起始值 = max + SortStep
func SortNewBatch(model interface{}, query string, count int, args ...interface{}) []int {
	var maxOrder int
	tableName := getTableName(model)
	databaseQuery := DB.Model(model).Table(tableName).Select("COALESCE(MAX(sort_order), 0)")
	if query != "1 = 1" {
		databaseQuery = databaseQuery.Where(query, args...)
	}
	databaseQuery.Scan(&maxOrder)

	seq := make([]int, count)
	for i := 0; i < count; i++ {
		seq[i] = mustAdd(maxOrder, (i+1)*SortStep)
	}
	return seq
}

// ========== 整数安全运算 ==========

// mustAdd 安全加法，溢出时返回 0（由调用方触发 rebalance）
func mustAdd(a, b int) int {
	if result, ok := safeAdd(a, b); ok {
		return result
	}
	return 0
}

// mustSub 安全减法，溢出时返回 0（由调用方触发 rebalance）
func mustSub(a, b int) int {
	if result, ok := safeSub(a, b); ok {
		return result
	}
	return 0
}

// safeAdd 安全加法，返回 (结果, 是否合法)
func safeAdd(a, b int) (int, bool) {
	if b > 0 && a > math.MaxInt-b {
		return 0, false
	}
	if b < 0 && a < math.MinInt-b {
		return 0, false
	}
	return a + b, true
}

// safeSub 安全减法，返回 (结果, 是否合法)
func safeSub(a, b int) (int, bool) {
	if b < 0 && a > math.MaxInt+b {
		return 0, false
	}
	if b > 0 && a < math.MinInt+b {
		return 0, false
	}
	return a - b, true
}

// ========== 通用拖拽重排序 ==========

// ReorderEntity 通用拖拽重排序，自动检测整数除法坍缩并触发 rebalance
// 整个流程在一个事务中完成，保证原子性
//   - model: 表模型（如 &NodeGroup{}）
//   - uuid: 被拖拽记录的 UUID
//   - beforeUUID, afterUUID: 前后邻居的 UUID（可为空串表示首/尾）
//   - query, args: 限定范围条件，全表传 "1=1"，限定范围传 "group_uuid = ?", groupUUID
func ReorderEntity(model interface{}, uuid, beforeUUID, afterUUID string, query string, args ...any) (int, error) {
	if uuid == "" {
		return 0, fmt.Errorf("reorder: uuid is required")
	}

	tableName := getTableName(model)

	var newOrder int
	err := DB.Transaction(func(tx *gorm.DB) error {
		// 1. 查询被拖拽记录是否存在
		var target struct{ ID uint }
		if err := tx.Table(tableName).Where("uuid = ?", uuid).Select("id").First(&target).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("reorder: record %s not found", uuid)
			}
			return fmt.Errorf("reorder: failed to query target: %w", err)
		}

		// 2. 查询前后邻居的 sort_order
		var beforeOrder, afterOrder *int
		if beforeUUID != "" {
			var v int
			if err := tx.Table(tableName).Where("uuid = ?", beforeUUID).Select("sort_order").Scan(&v).Error; err == nil {
				beforeOrder = &v
			}
		}
		if afterUUID != "" {
			var v int
			if err := tx.Table(tableName).Where("uuid = ?", afterUUID).Select("sort_order").Scan(&v).Error; err == nil {
				afterOrder = &v
			}
		}

		// 3. 计算插入位置，检测冲突
		result, conflict := SortInsertSafe(beforeOrder, afterOrder)
		if conflict {
			// 整数除法坍缩 → 在事务内 rebalance → 重新查询 → 重新计算
			RebalanceScopedTx(tx, model, query, args...)

			// 重新查询邻居
			if beforeUUID != "" {
				var v int
				if err := tx.Table(tableName).Where("uuid = ?", beforeUUID).Select("sort_order").Scan(&v).Error; err == nil {
					beforeOrder = &v
				}
			}
			if afterUUID != "" {
				var v int
				if err := tx.Table(tableName).Where("uuid = ?", afterUUID).Select("sort_order").Scan(&v).Error; err == nil {
					afterOrder = &v
				}
			}
			result, _ = SortInsertSafe(beforeOrder, afterOrder) // rebalance 后不会再冲突
		}

		// 4. 更新目标记录
		if err := tx.Table(tableName).Where("id = ?", target.ID).Update("sort_order", result).Error; err != nil {
			return fmt.Errorf("reorder: failed to update sort_order: %w", err)
		}

		newOrder = result
		return nil
	})

	if err != nil {
		return 0, err
	}
	return newOrder, nil
}
