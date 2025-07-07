package enumgen

import (
	"embed"
	"io"
	"strings"
	"text/template"

	"github.com/thinkgos/proc-extra/cmd/enumgen/internal/enumerate"
	"github.com/thinkgos/proc/infra"
)

//go:embed enum.tpl ts.enum.tpl
var Static embed.FS

var TemplateFuncs = template.FuncMap{
	"snakeCase":      func(s string) string { return infra.SnakeCase(s) },
	"kebabCase":      func(s string) string { return infra.Kebab(s) },
	"camelCase":      func(s string) string { return infra.PascalCase(s) },
	"smallCamelCase": func(s string) string { return infra.SmallCamelCase(s) },
	"styleName":      StyleName,
	"trimPrefix":     strings.TrimPrefix,
	"formatTsEnumValue": func(v, t string) string {
		if v == "" {
			return v
		}
		s := strings.TrimPrefix(strings.TrimPrefix(v, t), "_")
		if s == "" {
			return s
		}
		if s[0] >= '0' && s[0] <= '9' {
			return "X" + s
		}
		return s
	},
}

var (
	tpl = template.Must(template.New("components").
		Funcs(TemplateFuncs).
		ParseFS(Static, "enum.tpl", "ts.enum.tpl"))
	enumTemplate   = tpl.Lookup("enum.tpl")
	tsEnumTemplate = tpl.Lookup("ts.enum.tpl")
)

type File struct {
	Version      string
	IsDeprecated bool
	Package      string
	HasInteger   bool
	TypeStyle    string // 字典类型type风格
	Enums        []*Enumerate
}

type Enumerate struct {
	Type        string
	TypeName    string
	TypeComment string
	IsString    bool
	Explain     string
	Values      []*enumerate.Value
}

// SortEnumerates 按TypeName排序
type SortEnumerates []*Enumerate

func (b SortEnumerates) Len() int      { return len(b) }
func (b SortEnumerates) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b SortEnumerates) Less(i, j int) bool {
	return b[i].TypeName < b[j].TypeName
}

func (e *File) execute(w io.Writer, tpl *template.Template) error {
	return tpl.Execute(w, e)
}
