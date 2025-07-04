package enumerate

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"log/slog"
	"os"
	"slices"
	"strings"
)

type Package struct {
	Name  string
	Defs  map[*ast.Ident]types.Object
	Files []*File
}

type File struct {
	Pkg         *Package
	File        *ast.File
	TypeName    string
	TypeComment string
	Type        string
	IsString    bool
	Values      []*Value
	OmitZero    bool
}

// Value represents a declared constant.
type Value struct {
	OriginalName string // 常量定义的名称
	Label        string // 注释, 如果没有, 则同常量名称
	// value相关
	Value    uint64 // 需要时转为`int64`(integer有效).
	Signed   bool   // `constant`是否是有符号类型(integer有效),
	RawValue string // 纯值, 字符串不包含引号, 整型直接格式化成字符串
	IsString bool   // 是否是string, 否则为integer
	Val      string // `constant`的字符串值,由"go/constant"包提供.
}

func (v *Value) String() string { return v.Val }

// SortValues 使我们可以将`constants`进行排序
// 字符串不排序, 整型谨慎地按照有符号或无符号的顺序进行恰当的排序
type SortValues []*Value

func (b SortValues) Len() int      { return len(b) }
func (b SortValues) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b SortValues) Less(i, j int) bool {
	if !b[i].IsString {
		if b[i].Signed {
			return int64(b[i].Value) < int64(b[j].Value)
		} else {
			return b[i].Value < b[j].Value
		}
	}
	return false
}

func (vs SortValues) Clone() SortValues {
	sortValues := make([]*Value, 0, len(vs))
	for _, v := range vs {
		tv := *v
		sortValues = append(sortValues, &tv)
	}
	return sortValues
}

// ArrayString convert to array string format [0:aaa,1:bbb,3:ccc]
func (vs SortValues) ArrayString() string {
	if len(vs) == 0 {
		return "[]"
	}
	b := strings.Builder{}
	b.WriteString("[")
	for i, k := range vs {
		if i != 0 {
			b.WriteString(",")
		}
		b.WriteString(k.RawValue)
		b.WriteString(":")
		b.WriteString(k.Label)
	}
	b.WriteString("]")
	return b.String()
}

// GenDecl processes one declaration clause.
func (f *File) GenDecl(node ast.Node) bool {
	decl, ok := node.(*ast.GenDecl)
	if !ok || !slices.Contains([]token.Token{token.CONST, token.TYPE}, decl.Tok) {
		// 只关心 const, Type 声明
		return true
	}

	// The name of the type of the constants we are declaring.
	// Can change if this is a multi-element declaration.
	typ := ""
	// Loop over the elements of the declaration. Each element is a ValueSpec:
	// a list of names possibly followed by a type, possibly followed by values.
	// If the type and value are both missing, we carry down the type (and value,
	// but the "go/types" package takes care of that).
	for _, spec := range decl.Specs {
		if decl.Tok == token.TYPE { // type spec
			tsepc := spec.(*ast.TypeSpec)          // 必定是 TYPE
			if tsepc.Name.String() == f.TypeName { // 找到这个类型
				obj, ok := f.Pkg.Defs[tsepc.Name]
				if !ok {
					slog.Error(fmt.Sprintf("no value for type %s", f.TypeName))
					os.Exit(1)
				}
				basic := obj.Type().Underlying().(*types.Basic)
				if (basic.Info() & (types.IsInteger | types.IsString)) == 0 {
					slog.Error(fmt.Sprintf("can't handle non-integer or non-string constant type %s", typ))
					os.Exit(1)
				}
				f.IsString = (basic.Info() & types.IsString) != 0
				f.Type = basic.Name() // 拿到类型
				if c := tsepc.Comment.Text(); c != "" {
					f.TypeComment += strings.TrimSuffix(strings.TrimSpace(c), "\n")
				} else {
					f.TypeComment = f.TypeName
				}
			}
		} else { // const spec
			vspec := spec.(*ast.ValueSpec) // 必定是 CONST.
			if vspec.Type == nil && len(vspec.Values) > 0 {
				// "X = 1". With no type but a value. If the constant is untyped,
				// skip this vspec and reset the remembered type.
				typ = ""

				// If this is a simple type conversion, remember the type.
				// We don't mind if this is actually a call; a qualified call won't
				// be matched (that will be SelectorExpr, not Ident), and only unusual
				// situations will result in a function call that appears to be
				// a type conversion.
				ce, ok := vspec.Values[0].(*ast.CallExpr)
				if !ok {
					continue
				}
				id, ok := ce.Fun.(*ast.Ident)
				if !ok {
					continue
				}
				typ = id.Name
			}
			if vspec.Type != nil {
				// "X T". We have a type. Remember it.
				ident, ok := vspec.Type.(*ast.Ident)
				if !ok {
					continue
				}
				typ = ident.Name
			}
			if typ != f.TypeName {
				// This is not the type we're looking for.
				continue
			}
			// We now have a list of names (from one line of source code) all being
			// declared with the desired type.
			// Grab their names and actual values and store them in f.values.
			for _, name := range vspec.Names {
				if name.Name == "_" {
					continue
				}
				// This dance lets the type checker find the values for us. It's a
				// bit tricky: look up the object declared by the name, find its
				// types.Const, and extract its value.
				obj, ok := f.Pkg.Defs[name]
				if !ok {
					slog.Error(fmt.Sprintf("no value for constant %s", name))
					os.Exit(1)
				}
				info := obj.Type().Underlying().(*types.Basic).Info()
				if (info & (types.IsInteger | types.IsString)) == 0 {
					slog.Error(fmt.Sprintf("can't handle non-integer or non-string constant type %s", typ))
					os.Exit(1)
				}
				value := obj.(*types.Const).Val() // Guaranteed to succeed as this is CONST.
				if value.Kind() != constant.Int && value.Kind() != constant.String {
					slog.Error(fmt.Sprintf("can't happen: constant is not an integer or a string %s", name))
					os.Exit(1)
				}

				v := &Value{
					OriginalName: name.Name,
					Label:        "",
					Value:        0,
					Signed:       false,
					RawValue:     value.String(),
					IsString:     (info & types.IsString) != 0,
					Val:          value.String(),
				}
				if c := vspec.Comment; c != nil && len(c.List) == 1 {
					v.Label = strings.TrimSpace(c.Text())
				} else {
					v.Label = v.OriginalName
				}

				if v.IsString {
					v.RawValue = constant.StringVal(value)
					if f.OmitZero && v.RawValue == "" {
						continue
					}
				} else {
					i64, isInt := constant.Int64Val(value)
					u64, isUint := constant.Uint64Val(value)
					if !isInt && !isUint {
						slog.Error(fmt.Sprintf("internal error: value of %s is not an integer: %s", name, value.String()))
						os.Exit(1)
					}
					if !isInt {
						u64 = uint64(i64)
					}
					if f.OmitZero && u64 == 0 {
						continue
					}
					v.Value = u64
					v.Signed = info&types.IsUnsigned == 0
				}
				f.Values = append(f.Values, v)
			}
		}
	}
	return false
}
