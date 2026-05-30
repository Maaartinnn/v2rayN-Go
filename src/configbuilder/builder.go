package configbuilder

import "v2rayn-go/database"

// BuildConfigParams 统一的配置构建参数
type BuildConfigParams struct {
	Profile   *database.Profile
	Rules     []database.RoutingRule
	ConfigDir string
	SocksPort int
	HTTPPort  int
}

// ConfigBuilder 配置构建器接口
type ConfigBuilder interface {
	// Build 根据统一参数生成配置文件，返回配置文件路径
	Build(params *BuildConfigParams) (configPath string, err error)
	// BuildBytes 根据统一参数生成配置数据（JSON 字节），不写入文件。
	// 用于 stdin 无文件落地模式，避免不必要的磁盘 I/O。
	BuildBytes(params *BuildConfigParams) (data []byte, err error)
}

var builderRegistry = make(map[string]ConfigBuilder)

// Register 注册配置构建器。重复注册同名 coreType 会 panic（与 database/sql.Register 行为一致）。
func Register(coreType string, builder ConfigBuilder) {
	if _, exists := builderRegistry[coreType]; exists {
		panic("configbuilder: duplicate registration for core type: " + coreType)
	}
	builderRegistry[coreType] = builder
}

// GetBuilder 根据 coreType 获取已注册的配置构建器
func GetBuilder(coreType string) (ConfigBuilder, bool) {
	b, ok := builderRegistry[coreType]
	return b, ok
}
