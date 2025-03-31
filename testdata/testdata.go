package testdata

import (
	_ "embed"
)

//go:embed test.key
var PriveKey string

//go:embed test.pub
var PubKey string
