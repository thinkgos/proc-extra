package driver

import (
	"github.com/thinkgos/proc/confuse"
)

type TempGrantGenerator struct{}

func (TempGrantGenerator) Name() string { return "temp-grant-generator" }
func (TempGrantGenerator) GenerateUniqueId() string {
	return confuse.Symbol(32)
}
