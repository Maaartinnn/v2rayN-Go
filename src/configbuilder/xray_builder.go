package configbuilder

// xrayBuilder Xray 配置构建器
type xrayBuilder struct{}

func init() {
	Register("xray", &xrayBuilder{})
}

// Build 根据统一参数生成 Xray 配置文件
func (b *xrayBuilder) Build(p *BuildConfigParams) (string, error) {
	return SaveXrayConfig(p.Profile, p.Rules, p.ConfigDir, p.SocksPort, p.HTTPPort)
}
