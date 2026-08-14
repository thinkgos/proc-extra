package proc

import (
	"github.com/alecthomas/participle/v2"
)

var parser = participle.MustBuild[Proc](
	participle.Unquote(),
	participle.Union[ValueType](
		String{}, Integer{}, Float{}, Bool{},
		StringList{}, IntegerList{}, FloatList{}, BoolList{},
		Map{},
	),
)

type Proc struct {
	Derive Derive `parser:"'#' '[' @@ ']'"`
}

// Match 匹配注解
// `#[ident]`
// `#[ident(k1=1,k2="2")]`
// `#[ident(k1=[1,2,3],k2=["1","2","3"])]`
func Match(s string) (*Derive, error) {
	p, err := parser.ParseString("", s)
	if err != nil {
		return nil, err
	}
	return &p.Derive, nil
}
