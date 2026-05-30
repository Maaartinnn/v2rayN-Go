package configbuilder

import "encoding/json"

// xrayBuilder Xray 配置构建器
type xrayBuilder struct{}

func init() {
	Register("xray", &xrayBuilder{})
}

// Build 根据统一参数生成 Xray 配置文件（写入磁盘）
func (b *xrayBuilder) Build(p *BuildConfigParams) (string, error) {
	return SaveXrayConfig(p.Profile, p.Rules, p.ConfigDir, p.SocksPort, p.HTTPPort)
}

// BuildBytes 根据统一参数生成 Xray 配置数据（仅返回 JSON 字节，不写入文件）
func (b *xrayBuilder) BuildBytes(p *BuildConfigParams) ([]byte, error) {
	cfg, err := BuildXrayConfig(p.Profile, p.Rules, p.ConfigDir, p.SocksPort, p.HTTPPort)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(cfg, "", "  ")
}
