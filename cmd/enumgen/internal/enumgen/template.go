package enumgen

import (
	"embed"
	"html/template"
	"io"

	"github.com/thinkgos/proc-extra/cmd/enumgen/internal/enumerate"
)

//go:embed enum.tpl
var Static embed.FS

var enumTemplate = template.Must(template.New("components").
	ParseFS(Static, "enum.tpl")).
	Lookup("enum.tpl")

type File struct {
	Version      string
	IsDeprecated bool
	Package      string
	Enums        []*Enumerate
}

type Enumerate struct {
	Type        string
	TypeName    string
	TypeComment string
	Explain     string
	Values      []*enumerate.Value
}

func (e *File) execute(w io.Writer) error {
	return enumTemplate.Execute(w, e)
}
