package database

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/spf13/viper"
	"sync"
)

var databaseConn *gorm.DB
var locker sync.Locker

func Get() (*gorm.DB, error) {
	var err error

	locker.Lock()
	defer locker.Unlock()
	if databaseConn == nil {
		databaseConn, err = load()
	}

	return databaseConn, err
}

func load() (db *gorm.DB, err error) {
	connString := viper.GetString("database")
	if connString == "" {
		connString = "discord:discord@/discord"
	}

	return gorm.Open("mysql", connString)
}
