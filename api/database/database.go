package database

import (
	"database/sql"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"sync"
	"time"
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
	connString := viper.GetString("database")
	if connString == "" {
		connString = "discord:discord@/discord?charset=utf8mb4&parseTime=True"
	}

	db, err = gorm.Open(mysql.Open(connString), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
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
