//go:build databases.mysql || databases.all
// +build databases.mysql databases.all

package database

import (
	"fmt"
	"github.com/lordralex/absol/api/env"
	"github.com/lordralex/absol/api/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/url"
)

type MySql struct {
	Dialect
}

func (*MySql) Load() gorm.Dialector {
	user := env.GetOr("database.user", "discord")
	pass := url.PathEscape(env.GetOr("database.pass", "discord"))
	host := env.Get("database.host")
	dbName := env.GetOr("database.db", "discord")

	logger.Debug().Printf("Connecting to DB: %s - %s", host, dbName)

	connString := fmt.Sprintf("%s:%s@%s/%s?charset=utf8mb4&parseTime=True", user, pass, host, dbName)
	return mysql.Open(connString)
}

func init() {
	dialects["mysql"] = &MySql{}
}
