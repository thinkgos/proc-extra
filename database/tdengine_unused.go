//go:build !taosSql
// +build !taosSql

package database

import (
	"gorm.io/gorm"
)

func NewTaosSql(dsn string) gorm.Dialector {
	panic("please build tags with taosSql!")
}
