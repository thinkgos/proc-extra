package enumgen

import (
	"embed"
	"io"
	"text/template"

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
	HasInteger   bool
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

// SorEnumerate 按TypeName排序
type SorEnumerate []*Enumerate

func (b SorEnumerate) Len() int      { return len(b) }
func (b SorEnumerate) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b SorEnumerate) Less(i, j int) bool {
	return b[i].TypeName < b[j].TypeName
}

func (e *File) execute(w io.Writer) error {
	return enumTemplate.Execute(w, e)
}
