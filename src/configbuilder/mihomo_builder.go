// mihomo_builder.go — Mihomo ConfigBuilder 接口实现
//
// 实现 configbuilder.ConfigBuilder 接口，将 Mihomo 配置构建逻辑
// 注册到 builderRegistry 中，使 CoreService 可以通过 GetBuilder("mihomo")
// 统一调用，与 Xray / Sing-box 保持完全一致的调用模式。
//
// 注意事项：
//   - Mihomo 使用 YAML 格式（而非 JSON），BuildBytes() 返回 YAML 字节
//   - Build() 保存为 .yaml 文件
//   - stdin 模式（mihomo -f -）接受 YAML 输入，无需特殊处理
//   - 通过 init() 自动注册，导入即生效

package configbuilder

import (
	"gopkg.in/yaml.v3"
)

// mihomoBuilder Mihomo 配置构建器
//
// 实现 ConfigBuilder 接口的两个方法：
//   - Build():      生成配置文件并写入磁盘，返回文件路径
//   - BuildBytes(): 生成配置数据（YAML 字节），不写入文件，用于 stdin 无文件落地模式
type mihomoBuilder struct{}

// init 在包导入时自动注册 Mihomo 配置构建器。
// 注册后可通过 GetBuilder("mihomo") 获取实例。
func init() {
	Register("mihomo", &mihomoBuilder{})
}

// Build 根据统一参数生成 Mihomo 配置文件（写入磁盘）
//
// 输出路径: {ConfigDir}/binConfig/mihomo_config.yaml
// 与 Xray/Singbox 的 Build 方法行为一致，区别在于使用 YAML 格式。
func (b *mihomoBuilder) Build(p *BuildConfigParams) (string, error) {
	return SaveMihomoConfig(p.Profile, p.Rules, p.ConfigDir, p.SocksPort)
}

// BuildBytes 根据统一参数生成 Mihomo 配置数据（仅返回 YAML 字节，不写入文件）
//
// 用于 stdin 无文件落地模式（mihomo -f -），配置数据通过 cmd.Stdin 管道注入。
// 返回的字节是 YAML 格式（Mihomo 原生格式），而非 JSON。
func (b *mihomoBuilder) BuildBytes(p *BuildConfigParams) ([]byte, error) {
	cfg, err := BuildMihomoConfig(p.Profile, p.Rules, p.ConfigDir, p.SocksPort)
	if err != nil {
		return nil, err
	}
	return yaml.Marshal(cfg)
}
