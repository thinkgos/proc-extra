package enumgen

import (
	"embed"
	"text/template"
)

//go:embed enum.go.tpl
var Static embed.FS

var (
	tpl            = template.Must(template.New("components").ParseFS(Static, "enum.go.tpl"))
	goEnumTemplate = tpl.Lookup("enum.go.tpl")
)
