package astutil

import (
	"fmt"
	"go/ast"
	"log/slog"
	"strconv"
	"strings"

	"github.com/thinkgos/proc/enum_spec"
	"golang.org/x/tools/go/packages"
)

type AstInspect struct {
	Pattern []string                        // 匹配路径
	Type    []string                        // 枚举类型
	Tags    []string                        // 编译标签
	pkg     *Package                        // 包
	enums   map[string]*enum_spec.Enumerate // 枚举
}

func (g *AstInspect) Enums() map[string]*enum_spec.Enumerate { return g.enums }

func (g *AstInspect) PackageName() string { return g.pkg.Name }

func (g *AstInspect) Init() error {
	pkgs, err := packages.Load(
		&packages.Config{
			Mode: packages.NeedName |
				packages.NeedTypes |
				packages.NeedTypesInfo |
				packages.NeedImports |
				packages.NeedDeps |
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
		g.enums = make(map[string]*enum_spec.Enumerate)
	}
	for _, typeName := range g.Type {
		enums := g.findEnumerateViaTypeName(typeName)
		if enums == nil {
			slog.Error("code gen enum", slog.String("err", fmt.Sprintf("no find type defined: %s", typeName)))
			continue
		}
		g.enums[typeName] = enums
	}
	return nil
}

// findEnumerateViaTypeName 从 Go 源码提取枚举定义, 复用注册表模型 enum_spec.Enumerate:
//   - Title = 类型注释 (中文名称); 成员 Name = Go 常量名 (生成代码按此引用);
//   - Format = 取值宽度 (词表 int8..uint64); Go 默认 int 与 string 留空 (不填 = 语言默认整数类型), uint/uintptr 等 Go 特有宽度原样存入, 仅供生成;
//   - Const/Label = 取值/标签。
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
		var cv enum_spec.Valuer
		if v.isInteger {
			i, err := strconv.ParseInt(v.Const, 10, 64)
			if err != nil {
				slog.Error("code gen enum", slog.String("err", fmt.Sprintf("invalid integer value %q of %s", v.Const, v.OriginalName)))
				continue
			}
			cv = enum_spec.IntegerValue(i)
		} else {
			cv = enum_spec.StringValue(v.Const)
		}
		oneof = append(oneof, &enum_spec.EnumerateValue{
			Name:  v.OriginalName,
			Const: cv,
			Label: v.Label,
		})
	}
	realType := typeType
	if typeType != enum_spec.TypeString {
		realType = enum_spec.TypeInteger
	}
	// Format 仅存词表宽度; Go 默认 int 与 string 留空 (不填 = 语言默认整数类型), 生成签名由模板兜底。
	format := typeType
	if typeType == enum_spec.TypeString || typeType == "int" {
		format = ""
	}
	return &enum_spec.Enumerate{
		Type:   realType,
		Format: format,
		Title:  typeComment,
		OneOf:  oneof,
	}
}
