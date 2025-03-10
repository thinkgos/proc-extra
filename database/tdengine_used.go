//go:build taosSql
// +build taosSql

package database

import (
	tdengine_gorm "github.com/thinkgos/tdengine-gorm"
	"gorm.io/gorm"
)

func NewTaosSql(dsn string) gorm.Dialector {
	return &tdengine_gorm.Dialect{
		DriverName: tdengine_gorm.DefaultDriverName,
		DSN:        dsn,
		Conn:       nil,
	}
}
