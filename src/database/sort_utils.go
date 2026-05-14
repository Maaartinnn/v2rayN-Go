package database

import (
	"log"
	"reflect"
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
	return maxOrder + SortStep
}

// SortBetween 拖拽插入：插入到两条记录之间
func SortBetween(before, after int) int {
	return (before + after) / 2
}

// SortSequence 批量生成步长序列 [10, 20, 30, ...]
func SortSequence(count int) []int {
	seq := make([]int, count)
	for i := 0; i < count; i++ {
		seq[i] = (i + 1) * SortStep
	}
	return seq
}

// ========== 全表/范围重排 ==========

// Rebalance 全表重排：检查当前 sort_order 是否已符合 (i+1)*SortStep，
// 全部一致则不动数据库，返回 false；有不一致则批量 UPDATE，返回 true。
func Rebalance(model interface{}) bool {
	return RebalanceScoped(model, "1 = 1")
}

// RebalanceScoped 限定范围重排
func RebalanceScoped(model interface{}, query string, args ...interface{}) bool {
	type row struct {
		ID        uint
		SortOrder int
	}
	var rows []row
	tableName := getTableName(model)
	q := DB.Model(model).Table(tableName).Select("id, sort_order").Order("sort_order ASC, id ASC")
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

	// 执行重排
	tx := DB.Begin()
	for i, r := range rows {
		newOrder := (i + 1) * SortStep
		if r.SortOrder != newOrder {
			if err := tx.Table(tableName).Where("id = ?", r.ID).Update("sort_order", newOrder).Error; err != nil {
				tx.Rollback()
				log.Printf("[WARN] Rebalance failed for %s id=%d: %v", tableName, r.ID, err)
				return false
			}
		}
	}
	if err := tx.Commit().Error; err != nil {
		log.Printf("[WARN] Rebalance commit failed for %s: %v", tableName, err)
		return false
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
		seq[i] = maxOrder + (i+1)*SortStep
	}
	return seq
}
