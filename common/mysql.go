package common

import (
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type MysqlDBPool struct {
	dbs map[string]*gorm.DB
}

func NewMysqlDBPool() *MysqlDBPool {
	return &MysqlDBPool{
		dbs: make(map[string]*gorm.DB),
	}
}

// AddDB add a mysql instance into poll
// note that, all instance should be added in init status of Application
//
// example:
// AddDB("user", "root:root@tcp(127.0.0.1:3306)/user?charset=utf8mb4&parseTime=True&loc=Local")
// then, you can use code `GetDB("user")` to get it
func (p MysqlDBPool) AddDB(name string, sqlInfo string, maxConn int, idleConn int, maxLeftTime time.Duration) error {
	db, err := gorm.Open("mysql", sqlInfo)
	if err != nil {
		return err
	}

	db.DB().SetMaxOpenConns(maxConn)
	db.DB().SetMaxIdleConns(idleConn)

	// Mysql usually close a conn if it doesn't use in eight hours,
	// So maxLeftTime best less than eight hours,
	// We close it active if don't use in maxLeftTime. avoid it turn into a broke pipe
	db.DB().SetConnMaxLifetime(maxLeftTime)

	if err = db.DB().Ping(); err != nil {
		return err
	}

	p.dbs[name] = db

	return nil
}

func (p MysqlDBPool) GetDB(name string) *gorm.DB {
	if db, ok := p.dbs[name]; ok {
		return db
	}

	return nil
}

func (p MysqlDBPool) ReleasePool() {
	for _, db := range p.dbs {
		_ = db.Close()
	}
}

// --------------------------------------------------------------------

var defaultPool *MysqlDBPool

func init() {
	defaultPool = NewMysqlDBPool()
}

func AddDB(name string, sqlInfo string, maxConn int, idleConn int, maxLeftTime time.Duration) error {
	return defaultPool.AddDB(name, sqlInfo, maxConn, idleConn, maxLeftTime)
}

func GetDB(name string) *gorm.DB {
	return defaultPool.GetDB(name)
}

func MustGetDB(name string) *gorm.DB {
	db := GetDB(name)
	if db == nil {
		panic("DB " + name + " not exist")
	}
	return db
}

func ReleaseMysqlDBPool() {
	defaultPool.ReleasePool()
}
