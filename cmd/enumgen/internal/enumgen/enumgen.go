package enumgen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"log/slog"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/thinkgos/proc-extra/cmd/enumgen/internal/enumerate"
	"golang.org/x/tools/go/packages"
)

type Gen struct {
	Pattern   []string
	OutputDir string
	Type      []string
	Tags      []string
	Version   string
	pkg       *enumerate.Package
}

func (g *Gen) Generate() error {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedImports |
			packages.NeedSyntax,
		Tests:      false,
		BuildFlags: []string{fmt.Sprintf("-tags=%s", strings.Join(g.Tags, " "))},
		Logf: func(format string, args ...interface{}) {
			slog.Debug(fmt.Sprintf(format, args...))
		},
	}
	pkgs, err := packages.Load(cfg, g.Pattern...)
	if err != nil {
		return err
	}
	if len(pkgs) != 1 {
		return fmt.Errorf("%d packages matching %v", len(pkgs), g.Pattern)
	}
	pkg := pkgs[0]
	g.pkg = &enumerate.Package{
		Name:  pkg.Name,
		Defs:  pkg.TypesInfo.Defs,
		Files: make([]*enumerate.File, len(pkg.Syntax)),
	}
	for i, file := range pkg.Syntax {
		g.pkg.Files[i] = &enumerate.File{
			File: file,
			Pkg:  g.pkg,
		}
	}
	for _, typeName := range g.Type {
		err = g.generateEnum(typeName)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *Gen) generateEnum(typeName string) error {
	typeComment := ""
	typeType := ""
	values := make([]*enumerate.Value, 0, 128)
	for _, file := range g.pkg.Files {
		// Set the state for this run of the walker.
		file.TypeName = typeName
		file.Values = nil
		file.TypeComment = ""
		file.Type = ""
		if file.File != nil {
			ast.Inspect(file.File, file.GenDecl)
			values = append(values, file.Values...)
			if file.TypeComment != "" {
				typeComment = file.TypeComment
			}
			if file.Type != "" {
				typeType = file.Type
			}
		}
	}
	if len(values) == 0 {
		return fmt.Errorf("no values defined for type %s", typeName)
	}
	sort.Stable(enumerate.SortValue(values))
	explain := enumerate.SortValue(values).ArrayString()
	if typeComment != "" {
		explain = typeComment + ": " + explain
	}
	f := &File{
		Version:      g.Version,
		IsDeprecated: false,
		Package:      g.pkg.Name,
		Enums: []*Enumerate{
			{
				Type:     typeType,
				TypeName: typeName,
				Explain:  explain,
				Values:   values,
			},
		},
	}
	buf := &bytes.Buffer{}
	err := f.execute(buf)
	if err != nil {
		return err
	}
	data, err := format.Source(buf.Bytes())
	if err != nil {
		data = buf.Bytes()
	}
	filename := path.Join(g.OutputDir, strings.ToLower(typeName)+".enum.gen.go")
	return os.WriteFile(filename, data, 0644)
}
