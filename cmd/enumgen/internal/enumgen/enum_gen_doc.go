package enumgen

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/thinkgos/proc-extra/cmd/enumgen/internal/astutil"
	"github.com/thinkgos/proc/enum_spec"
)

type GenDoc struct {
	AstInspect astutil.AstInspect
	Version    string // 版本
	Filename   string // 文件名
}

func (g *GenDoc) Init() error {
	return g.AstInspect.Init()
}
func (g *GenDoc) Gen() error {
	t := enum_spec.T{
		Version: enum_spec.Version,
		Info:    nil,
		Enums:   g.AstInspect.Enums(),
	}
	data, err := json.MarshalIndent(t, " ", "  ")
	if err != nil {
		return err
	}
	filename := g.Filename
	if !strings.HasSuffix(filename, ".json") {
		filename = g.Filename + ".json"
	}
	return os.WriteFile(filename, data, 0644)
}
