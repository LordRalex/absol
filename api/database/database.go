package database

import (
	"database/sql"
	"errors"
	"github.com/lordralex/absol/api/env"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"sync"
	"time"
)

var databaseConn *gorm.DB
var locker sync.Mutex

var dialects = map[string]Dialect{}

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
	var dialect Dialect
	dialectName := env.Get("database.dialect")
	if dialectName == "" {
		if len(dialects) == 1 {
			for _, v := range dialects {
				dialect = v
				break
			}
		} else {
			return nil, errors.New("no database dialects or more than 1 dialect available with none selected using database.dialect")
		}
	} else {
		dialect = dialects[dialectName]
	}

	if dialect == nil {
		return nil, errors.New("unknown database dialect " + dialectName)
	}

	db, err = gorm.Open(dialect.Load(), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
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

type Dialect interface {
	Load() gorm.Dialector
}
