package coredef

import "runtime"

// CoreType 内核类型（全局唯一定义）
type CoreType string

const (
	TypeXray    CoreType = "xray"
	TypeSingBox CoreType = "sing-box"
	TypeMihomo  CoreType = "mihomo"
)

// CoreMeta 内核静态元数据
type CoreMeta struct {
	Type        CoreType
	DisplayName string
	SubDir      string
	Repo        string
}

// Registry 所有支持内核的元数据注册表（唯一事实来源）
var Registry = map[CoreType]CoreMeta{
	TypeXray: {
		Type:        TypeXray,
		DisplayName: "Xray-core",
		SubDir:      "xray",
		Repo:        "XTLS/Xray-core",
	},
	TypeSingBox: {
		Type:        TypeSingBox,
		DisplayName: "Sing-box",
		SubDir:      "sing_box",
		Repo:        "SagerNet/sing-box",
	},
	TypeMihomo: {
		Type:        TypeMihomo,
		DisplayName: "Mihomo",
		SubDir:      "mihomo",
		Repo:        "MetaCubeX/mihomo",
	},
}

// BinaryName 根据当前操作系统生成可执行文件名
func (m CoreMeta) BinaryName() string {
	name := string(m.Type)
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

// GetSupportedCores 返回所有支持内核的元数据列表
func GetSupportedCores() []CoreMeta {
	cores := make([]CoreMeta, 0, len(Registry))
	for _, meta := range Registry {
		cores = append(cores, meta)
	}
	return cores
}
