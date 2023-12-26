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
	//userGroup.GET("/phone_number", API(handler.UpdateUserPhoneNumber, model.APIOption{
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
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	// 用户类接口
	userGroup := r.Group("/team_up/user")
	// 登录接口，用户进入小程序就要请求（通了）
	userGroup.POST("/login", API(handler.UserLogin, model.APIOption{
		NeedLoginStatus: false,
	}))
	// 前端获取密文手机号，服务端解码并存储
	userGroup.POST("/update_phone_number", API(handler.UpdateUserPhoneNumber, model.APIOption{
		NeedLoginStatus: true,
	}))
	// 更新用户信息（头像 & 昵称）
	userGroup.POST("/update_user_info", API(handler.UpdateUserInfo, model.APIOption{
		NeedLoginStatus: true,
	}))
	// 获取个人主页
	userGroup.GET("/my_tab", API(handler.GetMyTab, model.APIOption{
		NeedLoginStatus: false,
	}))
	// 获取定级问题
	userGroup.POST("/get_calibration_questions", API(handler.GetCalibrationQuestions, model.APIOption{
		NeedLoginStatus: true,
	}))
	// 定级
	userGroup.POST("/calibrate", API(handler.Calibrate, model.APIOption{
		NeedLoginStatus: true,
	}))
	// 获取用户组织相关信息
	userGroup.GET("/host_info", API(handler.GetUserHostInfo, model.APIOption{
		NeedLoginStatus: true,
	}))
	// 加入活动
	userGroup.POST("/join_event", API(handler.JoinEvent, model.APIOption{
		NeedLoginStatus: true,
	}))
	// 退出活动
	userGroup.POST("/quit_event", API(handler.QuitEvent, model.APIOption{
		NeedLoginStatus: true,
	}))
	// 上传比赛结果
	userGroup.POST("/upload_event_result", API(handler.UploadEventResult, model.APIOption{
		NeedLoginStatus: true,
	}))

	// 组织类接口
	organizationGroup := r.Group("/team_up/organization")
	// 创建组织
	organizationGroup.POST("/create", API(handler.CreateOrganization, model.APIOption{
		NeedLoginStatus: true,
	}))

	// 活动类接口
	eventGroup := r.Group("/team_up/event")
	// 活动页面
	eventGroup.POST("/page", API(handler.GetEventTab, model.APIOption{
		NeedLoginStatus: true,
	}))
	// 创建活动
	eventGroup.POST("/create", API(handler.CreateEvent, model.APIOption{
		NeedLoginStatus: true,
	}))
	// 更新活动信息
	eventGroup.POST("/update", API(handler.UpdateEvent, model.APIOption{
		NeedLoginStatus: true,
	}))
	// 获取活动结果
	eventGroup.POST("/get_result", API(handler.GetEventResult, model.APIOption{
		NeedLoginStatus: true,
	}))

	// 静态资源Group
	imageGroup := r.Group("/team_up/static_image")
	// 用户定级职业的证明
	imageGroup.Static("/calibration_proof", "./user_calibration_proof")
	// 用户创建组织的logo
	imageGroup.Static("/organization_logo", "./organization_logos")

	return r
}
