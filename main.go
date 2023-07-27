package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	// 定时任务获取access_token
	// todo:
	// 加载
	// 商家侧使用的接口
	merchantGroup := r.Group("/team_up/core/merchant_user")
	merchantGroup.POST("/create_event")
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
