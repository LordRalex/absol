package database

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/lordralex/absol/api/env"
	"github.com/lordralex/absol/api/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var databaseConn *gorm.DB
var locker sync.Mutex

func Get() (*gorm.DB, error) {
	var err error

	locker.Lock()
	defer locker.Unlock()
	if databaseConn == nil {
		databaseConn, err = load()
	}
	if databaseConn == nil {
		databaseConn, err = load()
	}

	return databaseConn, err
}

func load() (db *gorm.DB, err error) {
	user := env.GetOr("database.user", "discord")
	pass := env.GetOr("database.pass", "discord")
	host := env.GetOr("database.host", "localhost:3306")
	dbName := env.GetOr("database.db", "discord")

	logger.Debug().Printf("Connecting to DB: %s - %s", host, dbName)

	connString := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True", user, pass, host, dbName)

	db, err = gorm.Open(mysql.Open(connString), &gorm.Config{Logger: gormLogger.Default.LogMode(gormLogger.Info)})
	if db != nil {
		sqlDb, _ := db.DB()
		sqlDb.SetConnMaxLifetime(time.Second * 10)
		sqlDb.SetMaxIdleConns(0)
		sqlDb.SetMaxOpenConns(10)
	}

	return
}

func Execute(stmt *sql.Stmt, args ...interface{}) error {
	defer stmt.Close()
	_, err := stmt.Exec(args...)
	return err
}
