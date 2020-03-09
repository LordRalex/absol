package database

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/spf13/viper"
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

	if err == nil && databaseConn != nil {

	}

	return databaseConn, err
}

func Close() {
	if databaseConn != nil {
		_ = databaseConn.Close()
	}
}

func load() (db *gorm.DB, err error) {
	connString := viper.GetString("database")
	if connString == "" {
		connString = "discord:discord@/discord"
	}

	db, err = gorm.Open("mysql", connString)
	if db != nil {
		db.LogMode(true)
		db.DB().SetConnMaxLifetime(time.Second * 10)
		db.DB().SetMaxIdleConns(0)
		db.DB().SetMaxOpenConns(10)
	}
	return
}
