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
		dsn := fmt.Sprintf(constant.MySQLDSN, "root", "sdqd960410", "tcp(10.0.16.9:3306)", "teamup")
		MySQLDB, err = gorm.Open(mysql.Open(dsn))
		if err != nil {
			Logger.Panicf("InitMySQL failed, err:%v", err)
		}
	})
	Logger.Println("mysql init success")
}

func DB() *gorm.DB {
	return MySQLDB
}

// InsertRecord 插入一条数据到数据库，value必须为指针！
func InsertRecord(value interface{}) error {
	Logger.Printf("inserting :%v", value)
	if err := DB().Create(value).Error; err != nil {
		Logger.Printf("Insert %v into DB failed, err:%v", value, err)
		return err
	}
	return nil
}
