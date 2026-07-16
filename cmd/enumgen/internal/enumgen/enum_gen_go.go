package enumgen

import (
	"bytes"
	"go/format"
	"io"
	"log/slog"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/thinkgos/proc-extra/cmd/enumgen/internal/astutil"
	"github.com/thinkgos/proc/enum_spec"
)

type GenGo struct {
	AstInspect astutil.AstInspect
	OutputDir  string // 输出路径
	Version    string // 版本
	Merge      bool   // 合并到一个文件
	Filename   string // 合并文件名/ts文件名
}

func (g *GenGo) Init() error {
	return g.AstInspect.Init()
}
func (g *GenGo) Gen() error {
	genEnumFile := func(es *enum_spec.Enums) ([]byte, error) {
		f := &GenGoFile{
			Version:           g.Version,
			IsDeprecated:      false,
			Package:           g.AstInspect.PackageName(),
			HasContainInteger: false,
			Enums:             es.Maps(),
		}
		f.HasContainInteger = ContainAny(es.Values(), func(v *enum_spec.Enumerate) bool {
			return v.Type == enum_spec.TypeInteger
		})
		buf := &bytes.Buffer{}
		err := f.execute(buf, goEnumTemplate)
		if err != nil {
			return nil, err
		}
		data, err := format.Source(buf.Bytes())
		if err != nil {
			data = buf.Bytes()
		}
		return data, nil
	}

	if g.Merge {
		data, err := genEnumFile(g.AstInspect.Enums())
		if err != nil {
			return err
		}
		filename := g.Filename
		if filename == "" {
			filename = "enum"
		}
		filename = path.Join(g.OutputDir, strings.ToLower(filename)+".enum.gen.go")
		return os.WriteFile(filename, data, 0644)
	} else {
		for typeName, e := range g.AstInspect.Enums().All() {
			data, err := genEnumFile(enum_spec.NewEnums().Set(typeName, e))
			if err != nil {
				slog.Error("code gen enum", slog.Any("err", err))
				continue
			}
			filename := path.Join(g.OutputDir, strings.ToLower(typeName)+".enum.gen.go")
			err = os.WriteFile(filename, data, 0644)
			if err != nil {
				slog.Error("code gen enum", slog.Any("err", err))
				continue
			}
		}
		return nil
	}
}

type GenGoFile struct {
	Version           string
	IsDeprecated      bool
	Package           string
	HasContainInteger bool
	Enums             map[string]*enum_spec.Enumerate
}

func (e *GenGoFile) execute(w io.Writer, tpl *template.Template) error {
	return tpl.Execute(w, e)
}
