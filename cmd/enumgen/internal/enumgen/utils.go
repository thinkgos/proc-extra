package enumgen

import "github.com/thinkgos/proc/infra"

const (
	StyleSmallCamelCase = "smallCamelCase"
	StylePascalCase     = "pascalCase"
	StyleSnakeCase      = "snakeCase"
	StyleKebab          = "kebab"
)

func StyleName(kind, name string) string {
	vv := name
	switch kind {
	case StyleSmallCamelCase:
		vv = infra.SmallCamelCase(name)
	case StylePascalCase:
		vv = infra.PascalCase(name)
	case StyleSnakeCase:
		vv = infra.SnakeCase(name)
	case StyleKebab:
		vv = infra.Kebab(name)
	}
	return vv
}
