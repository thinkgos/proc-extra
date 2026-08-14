# proc

`proc` annotation syntax.

basic syntax:

- `#[ident]`: headless syntax
- `#[ident(name1=value1,name2=value2)]`: no headless syntax

`ident`is a identity, `name=value` slice in the `()`.
`value` support the following syntax

- `string`: `"hello"`
- `integer`: `123`
- `float`: `1.0`
- `bool`: `true`,`false`
- `Object`: `{k1="v1",k2="v2"}`, in the object is `name=value` slice too.
- `string slice`: `["hello","world"]`
- `integer slice`: `[123,12,1]`
- `float slice`: `[1.0,1.1,1.2]`, ***NOTE***: the first value in slice must be a float type, like `[1,1.1,1.2]` will parsed as integer slice, then failure.
- `bool slice`: `[true,false,true]`
- `map[string]Value`: `{k1="v1", k2="v2"}`, the value can be the defined.

example:

- `#[ident]`
- `#[ident(k1=1,k2="2")]`
- `#[ident(k1=[1,2,3],k2=["1","2","3"])]`
- `#[ident(k1="hello",k2=["1","2","3"])]`
- `#[ident(k1={k2="v2",k3="v3"})]`
