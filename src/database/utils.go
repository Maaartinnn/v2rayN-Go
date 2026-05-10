package database

import "github.com/google/uuid"

// GenerateUUID 生成一个新的 UUID 字符串
func GenerateUUID() string {
	return uuid.New().String()
}
