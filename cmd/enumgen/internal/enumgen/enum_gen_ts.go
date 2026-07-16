package enumgen

import (
	"bytes"
	"io"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/thinkgos/proc/enum_spec"
)

type GenTs struct {
	Uri       string // uri: file://enum.json 或 http://xxx.com/a.json
	Version   string // 版本
	Filename  string // 文件名
	TypeStyle string // 字典Key风格
}

func (g *GenTs) loadFromUri() (*enum_spec.T, error) {
	if filename, ok := strings.CutPrefix(g.Uri, "file://"); ok {
		return enum_spec.NewLoader().LoadFromFile(filename)
	} else {
		return enum_spec.NewLoader().LoadFromURL(g.Uri)
	}
}

func (g *GenTs) Gen() error {
	t, err := g.loadFromUri()
	if err != nil {
		return err
	}
	outputDir := path.Dir(g.Filename)
	err = os.WriteFile(path.Join(outputDir, "dictDef.ts"), enumDefTs, 0644)
	if err != nil {
		return err
	}
	f := &GenTsFile{
		Version:   g.Version,
		TypeStyle: g.TypeStyle,
		Enums:     t.Enums.Maps(),
	}
	buf := &bytes.Buffer{}
	err = f.execute(buf, tsEnumTemplate)
	if err != nil {
		return err
	}
	return os.WriteFile(g.Filename, buf.Bytes(), 0644)
}

type GenTsFile struct {
	Version   string
	TypeStyle string // 字典类型type风格
	Enums     map[string]*enum_spec.Enumerate
}

func (e *GenTsFile) execute(w io.Writer, tpl *template.Template) error {
	return tpl.Execute(w, e)
}
