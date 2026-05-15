package configbuilder

// singboxBuilder Sing-box 配置构建器
type singboxBuilder struct{}

func init() {
	Register("sing-box", &singboxBuilder{})
}

// Build 根据统一参数生成 Sing-box 配置文件
// Sing-box 使用 mixedPort，这里复用 SocksPort 作为混合代理端口
func (b *singboxBuilder) Build(p *BuildConfigParams) (string, error) {
	return SaveSingboxConfig(p.Profile, p.Rules, p.ConfigDir, p.SocksPort)
}
