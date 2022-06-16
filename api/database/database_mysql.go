//go:build databases.mysql || databases.all
// +build databases.mysql databases.all

package database

import (
	"github.com/lordralex/absol/api/env"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MySql struct {
	Dialect
}

func (*MySql) Load() gorm.Dialector {
	connString := env.Get("database.url")

	if connString == "" {
		connString = "discord:discord@/discord?charset=utf8mb4&parseTime=True"
	}

	return mysql.Open(connString)
}

func init() {
	dialects["mysql"] = &MySql{}
}
