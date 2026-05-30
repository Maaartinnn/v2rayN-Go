package configbuilder

import "encoding/json"

// singboxBuilder Sing-box 配置构建器
type singboxBuilder struct{}

func init() {
	Register("sing-box", &singboxBuilder{})
}

// Build 根据统一参数生成 Sing-box 配置文件（写入磁盘）
// Sing-box 使用 mixedPort，这里复用 SocksPort 作为混合代理端口
func (b *singboxBuilder) Build(p *BuildConfigParams) (string, error) {
	return SaveSingboxConfig(p.Profile, p.Rules, p.ConfigDir, p.SocksPort)
}

// BuildBytes 根据统一参数生成 Sing-box 配置数据（仅返回 JSON 字节，不写入文件）
func (b *singboxBuilder) BuildBytes(p *BuildConfigParams) ([]byte, error) {
	cfg, err := BuildSingboxConfig(p.Profile, p.Rules, p.ConfigDir, p.SocksPort)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(cfg, "", "  ")
}
