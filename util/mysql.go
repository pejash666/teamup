package util

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"sync"
	"teamup/constant"
)

var (
	MySQLDB   *gorm.DB
	onceMySQL sync.Once
)

func InitMySQL() {
	if MySQLDB != nil {
		return
	}
	var err error
	onceMySQL.Do(func() {
		dsn := fmt.Sprintf(constant.MySQLDSN, "team_up_db_w", "sdqd960410", "tcp(127.0.0.1:3306)", "team_up_db")
		MySQLDB, err = gorm.Open(mysql.Open(dsn))
		if err != nil {
			Logger.Panicf("InitMySQL failed, err:%v", err)
		}
	})
}

func DB() *gorm.DB {
	return MySQLDB
}
