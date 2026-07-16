package enumgen

import (
	"embed"
	"strings"
	"text/template"

	"github.com/thinkgos/proc/infra"
)

//go:embed enum.go.tpl enum.ts.tpl
var Static embed.FS

//go:embed enum.def.ts
var enumDefTs []byte

var TemplateFuncs = template.FuncMap{
	"snakeCase":  func(s string) string { return infra.SnakeCase(s) },
	"kebabCase":  func(s string) string { return infra.Kebab(s) },
	"pascalCase": func(s string) string { return infra.PascalCase(s) },
	"camelCase":  func(s string) string { return infra.SmallCamelCase(s) },
	"styleName":  StyleName,
	"trimPrefix": strings.TrimPrefix,
	"formatName": func(v string) string {
		if v == "" {
			return v
		}
		if v[0] >= '0' && v[0] <= '9' {
			return "X" + v
		}
		return v
	},
}

var (
	tpl = template.Must(template.New("components").
		Funcs(TemplateFuncs).
		ParseFS(Static, "enum.go.tpl", "enum.ts.tpl"))
	goEnumTemplate = tpl.Lookup("enum.go.tpl")
	tsEnumTemplate = tpl.Lookup("enum.ts.tpl")
)
