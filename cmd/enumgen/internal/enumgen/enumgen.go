package enumgen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"log/slog"
	"os"
	"path"
	"slices"
	"sort"
	"strings"

	"github.com/thinkgos/proc-extra/cmd/enumgen/internal/enumerate"
	"golang.org/x/tools/go/packages"
)

const DefaultDictTypeTpl = "INSERT INTO `sys_dict_type` (`type`, `name`, `remark`, `status`) VALUES ('%s', '%s', '%s', 1);"
const DefaultDictItemTpl = "INSERT INTO `sys_dict_item` (`dict_type`, `label`, `value`, `sort`, `remark`, `status`) VALUES ('%s', '%s', '%s', %d, '%s', 1);"

type Gen struct {
	Pattern     []string           // 匹配路径
	OutputDir   string             // 输出路径
	Type        []string           // 相关类型
	Tags        []string           // 编译标签
	Version     string             // 版本
	Merge       bool               // 合并到一个文件
	Filename    string             // 合并文件名
	OmitZero    bool               // 忽略零值
	SqlDictType string             // 字典类型模板
	SqlDictItem string             // 字典项模板
	pkg         *enumerate.Package // 包
	enums       []*Enumerate       // 枚举列表
}

func (g *Gen) Init() error {
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
			File:     file,
			Pkg:      g.pkg,
			OmitZero: g.OmitZero,
		}
	}
	for _, typeName := range g.Type {
		enums := g.inspectEnum(typeName)
		if enums == nil {
			slog.Error("code gen enum", slog.String("err", fmt.Sprintf("no find type defined: %s", typeName)))
			continue
		}
		g.enums = append(g.enums, enums)
	}
	sort.Stable(SorEnumerate(g.enums))
	return nil
}

func (g *Gen) GenEnum() error {
	genEnumFile := func(es ...*Enumerate) ([]byte, error) {
		f := &File{
			Version:      g.Version,
			IsDeprecated: false,
			Package:      g.pkg.Name,
			HasInteger:   false,
			Enums:        es,
		}
		f.HasInteger = slices.ContainsFunc(f.Enums, func(v *Enumerate) bool { return !v.IsString })
		buf := &bytes.Buffer{}
		err := f.execute(buf)
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
		data, err := genEnumFile(g.enums...)
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
		for _, e := range g.enums {
			data, err := genEnumFile(e)
			if err != nil {
				slog.Error("code gen enum", slog.Any("err", err))
				continue
			}
			filename := path.Join(g.OutputDir, strings.ToLower(e.TypeName)+".enum.gen.go")
			err = os.WriteFile(filename, data, 0644)
			if err != nil {
				slog.Error("code gen enum", slog.Any("err", err))
				continue
			}
		}
		return nil
	}
}

func (g *Gen) GenSql() error {
	if g.SqlDictType == "" {
		g.SqlDictType = DefaultDictTypeTpl
	}
	if g.SqlDictItem == "" {
		g.SqlDictItem = DefaultDictItemTpl
	}

	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}
	for _, v := range g.enums {
		fmt.Fprintf(buf1, DefaultDictTypeTpl, v.TypeName, v.TypeComment, v.Explain)
		fmt.Fprintln(buf1)
		sort := 1
		for _, vv := range v.Values {
			fmt.Fprintf(buf2, DefaultDictItemTpl, v.TypeName, vv.Label, vv.RawValue, sort, vv.Label)
			fmt.Fprintln(buf2)
			sort++
		}
	}
	buf1.WriteTo(os.Stdout)
	fmt.Fprintln(os.Stdout)
	buf2.WriteTo(os.Stdout)
	return nil
}

func (g *Gen) inspectEnum(typeName string) *Enumerate {
	typeComment := ""
	typeType := ""
	typeIsString := false
	values := make([]*enumerate.Value, 0, 128)
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
				typeIsString = file.IsString
				typeComment = file.TypeComment
			}
			values = append(values, file.Values...)
		}
	}
	if typeType == "" {
		return nil
	}
	// sort.Stable(enumerate.SortValue(values))
	explain := enumerate.SortValues(values).ArrayString()
	if typeComment != "" {
		explain = typeComment + ": " + explain
	}
	return &Enumerate{
		Type:        typeType,
		TypeName:    typeName,
		TypeComment: typeComment,
		IsString:    typeIsString,
		Explain:     explain,
		Values:      values,
	}
}
