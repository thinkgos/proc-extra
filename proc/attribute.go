package proc

// NameValue like `#[ident(name=value]`
type NameValue struct {
	// name
	Name string `parser:"@Ident '='"`
	// one of follow
	// String, Integer, Float, Bool,
	// StringList, IntegerList, FloatList, BoolList,
	Value ValueType `parser:"@@"`
}

// NameValueMap name:value map.
// NOTE: The same key pairs, later ones will override earlier ones.
func NameValueMap(vs []*NameValue) map[string]ValueType {
	attrs := make(map[string]ValueType, len(vs))
	for _, attr := range vs {
		attrs[attr.Name] = attr.Value
	}
	return attrs
}

type ValueType interface {
	Type() string
}

type String struct {
	Value string `parser:"@String"`
}

func (String) Type() string { return "string" }

type Integer struct {
	Value int64 `parser:"@Int"`
}

func (Integer) Type() string { return "integer" }

type Float struct {
	Value float64 `parser:"@Float"`
}

func (Float) Type() string { return "float" }

type Bool struct {
	Value Boolean `parser:"@('true' | 'false')"`
}

func (Bool) Type() string { return "bool" }

type Boolean bool

func (b *Boolean) Capture(values []string) error {
	*b = values[0] == "true"
	return nil
}

type StringList struct {
	Value []string `parser:"'[' (@String (',' @String)*)? ']'"`
}

func (StringList) Type() string { return "slice<string>" }

type IntegerList struct {
	Value []int64 `parser:"'[' (@Int (',' @Int)*)? ']'"`
}

func (IntegerList) Type() string { return "slice<integer>" }

// NOTE: FloatList float list. must be first is float.
type FloatList struct {
	Value []float64 `parser:"'[' (@Float (',' (@Float | @Int))*)? ']'"`
}

func (FloatList) Type() string { return "slice<float>" }

type BoolList struct {
	Value []Boolean `parser:"'[' (@('true' | 'false') (',' @('true' | 'false'))*)? ']'"`
}

func (BoolList) Type() string { return "slice<bool>" }

type Map struct {
	Entries []*NameValue `parser:"'{' (@@ (',' @@)*)? '}'"`
}

func (Map) Type() string { return "map[string]T" }

// EntryMap entry map
// NOTE: The same key pairs, later ones will override earlier ones.
func (d *Map) EntryMap() map[string]ValueType {
	return NameValueMap(d.Entries)
}
