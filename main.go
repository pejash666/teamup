package main

import (
	"fmt"
	"github.com/fvbock/endless"
	"log"
	"net/http"
	"syscall"
	"teamup/handler"
	"teamup/model"
	"teamup/util"

	"github.com/gin-gonic/gin"
)

var (
	API = util.API
)

func init() {
	Init()
}

func main() {
	//r := gin.Default()
	//// 定时任务获取access_token
	//// todo:
	//// 加载
	//// 商家侧使用的接口
	//userGroup := r.Group("/team_up/user")
	//userGroup.POST("/login", API(handler.UserLogin, model.APIOption{
	//	RequireMerchantUser: false}))
	//userGroup.GET("/phone_number", API(handler.GetPhoneNumber, model.APIOption{
	//	RequireMerchantUser: false,
	//	NeedLoginStatus:     true,
	//}))
	//r.GET("/ping", func(c *gin.Context) {
	//	c.JSON(http.StatusOK, gin.H{
	//		"message": "pong",
	//	})
	//})
	//r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	// 优雅退出，后续服务需要更新时候
	// 1. 修改代码并上传到实例
	// 2. 执行 go build ./ （确保在main.go级别的目录下）
	// 3. 使用 lsof -i tcp:8080 检查占用8080的进程pid
	// 4. 执行 kill -1 <pid> ，使得fork一个新进程监听8080端口即可
	// 这样即可保证在服务更新的过程中，所有的处理中请求不会被干扰，老进程在处理完存量请求后就会结束，后续由新进程执行请求
	endPoint := fmt.Sprintf(":%d", 8080)
	server := endless.NewServer(endPoint, HttpHandler())
	server.BeforeBegin = func(add string) {
		log.Printf("Actual pid is %d", syscall.Getpid())
	}
	server.ListenAndServe()
}

func HttpHandler() *gin.Engine {
	r := gin.Default()
	// todo: 区分 r.Group()
	userGroup := r.Group("/team_up/user")
	userGroup.POST("/login", API(handler.UserLogin, model.APIOption{
		RequireMerchantUser: false}))
	userGroup.GET("/phone_number", API(handler.GetPhoneNumber, model.APIOption{
		RequireMerchantUser: false,
		NeedLoginStatus:     true,
	}))
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	return r
}
