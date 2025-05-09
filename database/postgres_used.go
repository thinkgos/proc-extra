//go:build postgres
// +build postgres

package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgres(c postgres.Config) gorm.Dialector {
	return postgres.New(c)
}
