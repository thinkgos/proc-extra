package astutil

import (
	"fmt"
	"go/ast"
	"log/slog"
	"strings"

	"github.com/thinkgos/proc/enum_spec"
	"golang.org/x/tools/go/packages"
)

type AstInspect struct {
	Pattern []string         // 匹配路径
	Type    []string         // 枚举类型
	Tags    []string         // 编译标签
	pkg     *Package         // 包
	enums   *enum_spec.Enums // 枚举
}

func (g *AstInspect) Enums() *enum_spec.Enums { return g.enums }

func (g *AstInspect) PackageName() string { return g.pkg.Name }

func (g *AstInspect) Init() error {
	pkgs, err := packages.Load(
		&packages.Config{
			Mode: packages.NeedName |
				packages.NeedTypes |
				packages.NeedTypesInfo |
				packages.NeedImports |
				packages.NeedSyntax,
			Tests:      false,
			BuildFlags: []string{fmt.Sprintf("-tags=%s", strings.Join(g.Tags, " "))},
			Logf: func(format string, args ...any) {
				slog.Debug(fmt.Sprintf(format, args...))
			},
		},
		g.Pattern...,
	)
	if err != nil {
		return err
	}
	if len(pkgs) != 1 {
		return fmt.Errorf("%d packages matching %v", len(pkgs), g.Pattern)
	}
	pkg := pkgs[0]
	g.pkg = &Package{
		Name:  pkg.Name,
		Defs:  pkg.TypesInfo.Defs,
		Files: make([]*File, len(pkg.Syntax)),
	}
	for i, file := range pkg.Syntax {
		g.pkg.Files[i] = &File{
			File: file,
			Pkg:  g.pkg,
		}
	}
	if g.enums == nil {
		g.enums = enum_spec.NewEnums()
	}
	for _, typeName := range g.Type {
		enums := g.findEnumerateViaTypeName(typeName)
		if enums == nil {
			slog.Error("code gen enum", slog.String("err", fmt.Sprintf("no find type defined: %s", typeName)))
			continue
		}
		g.enums.Set(typeName, enums)
	}
	return nil
}

func (g *AstInspect) findEnumerateViaTypeName(typeName string) *enum_spec.Enumerate {
	typeComment := ""
	typeType := ""
	values := make([]*Value, 0, 128)
	for _, file := range g.pkg.Files {
		// Set the state for this run of the walker.
		file.TypeName = typeName
		file.TypeComment = ""
		file.Type = ""
		file.Values = nil
		if file.File != nil {
			ast.Inspect(file.File, file.GenDecl)
			if file.Type != "" {
				typeType = file.Type
				typeComment = file.TypeComment
			}
			values = append(values, file.Values...)
		}
	}
	if typeType == "" {
		return nil
	}
	oneof := make([]*enum_spec.EnumerateValue, 0, len(values))
	for _, v := range values {
		oneof = append(oneof, &enum_spec.EnumerateValue{
			GoTypeName: v.OriginalName,
			Name:       strings.TrimPrefix(strings.TrimPrefix(v.OriginalName, typeName), "_"),
			Const:      v.Const,
			Label:      v.Label,
			RawValue:   v.RawValue,
		})
	}
	realType := typeType
	if typeType != enum_spec.TypeString {
		realType = enum_spec.TypeInteger
	}
	e := &enum_spec.Enumerate{
		Type:        realType,
		Format:      typeType,
		Description: typeComment,
		Explain:     enum_spec.EnumerateValueSlices(oneof).Explain(),
		Oneof:       oneof,
	}
	return e
}
