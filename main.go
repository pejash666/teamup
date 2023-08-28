package main

import (
	"net/http"
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
	r := gin.Default()
	// 定时任务获取access_token
	// todo:
	// 加载
	// 商家侧使用的接口
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
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
