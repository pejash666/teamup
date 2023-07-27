package util

import (
	"log"
	"sync"
)

var (
	Logger     *log.Logger
	onceLogger sync.Once
)

func InitLogger() {
	if Logger != nil {
		return
	}
	onceLogger.Do(func() {
		Logger = log.Default()
		Logger.SetFlags(log.Llongfile | log.Ldate | log.Lmicroseconds)
	})
	Logger.Println("InitLoggerSuccess")
}
