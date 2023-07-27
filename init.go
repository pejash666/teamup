package main

import "teamup/util"

func Init() {
	// 初始化Logger
	util.InitLogger()
	// 初始化数据库
	util.InitMySQL()
	// 初始化Redis
	util.InitRedis()
}
