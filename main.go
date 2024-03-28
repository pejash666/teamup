package main

import (
	"fmt"
	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"io"
	"os"

	"log"
	"net/http"
	"syscall"
	_ "teamup/docs"
	"teamup/handler"
	"teamup/model"
	"teamup/util"
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
	// 登录与获取手机号融合登录
	userGroup.POST("/confirm_login", API(handler.ConfirmLogin, model.APIOption{
		NeedLoginStatus: false,
	}))
	// 前端获取密文手机号，服务端解码并存储(通了)
	userGroup.POST("/update_phone_number", API(handler.UpdateUserPhoneNumber, model.APIOption{
		NeedLoginStatus: true,
	}))
	// 更新用户信息（头像 & 昵称）(通了)
	userGroup.POST("/update_user_info", API(handler.UpdateUserInfo, model.APIOption{
		NeedLoginStatus: true,
	}))
	// 获取个人主页
	userGroup.GET("/my_tab", API(handler.GetMyTab, model.APIOption{
		NeedLoginStatus: true,
	}))
	// 获取定级问题(通了)
	userGroup.POST("/get_calibration_questions", API(handler.GetCalibrationQuestions, model.APIOption{
		NeedLoginStatus: true,
	}))
	// 定级（通了）
	userGroup.POST("/calibrate", API(handler.Calibrate, model.APIOption{
		NeedLoginStatus: true,
	}))
	// 获取用户组织相关信息（通了）
	userGroup.GET("/get_host_info", API(handler.GetHostInfo, model.APIOption{
		NeedLoginStatus: true,
	}))
	// 加入活动(通了)
	userGroup.POST("/join_event", API(handler.JoinEvent, model.APIOption{
		NeedLoginStatus: true,
	}))
	// 退出活动(通了)
	userGroup.POST("/quit_event", API(handler.QuitEvent, model.APIOption{
		NeedLoginStatus: true,
	}))
	// 下发记分配置(通了)
	userGroup.POST("/get_scoreboard", API(handler.GetScoreboard, model.APIOption{
		NeedLoginStatus: true,
	}))
	// 开始比赛（下发对局信息）(通了)
	userGroup.POST("/start_scoring", API(handler.StartScoring, model.APIOption{
		NeedLoginStatus: true,
	}))
	// 获取比赛的结果分数（通了）
	userGroup.POST("/get_score_result", API(handler.GetScoreResult, model.APIOption{
		NeedLoginStatus: true,
	}))
	// 发布分数(通了)
	userGroup.POST("/publish_score", API(handler.PublishScore, model.APIOption{
		NeedLoginStatus: true}))
	// 上传图片
	userGroup.POST("/upload_image", API(handler.UploadImage, model.APIOption{
		NeedLoginStatus: true,
	}))

	// 组织类接口
	organizationGroup := r.Group("/team_up/organization")
	// 创建组织(通了)
	organizationGroup.POST("/create", API(handler.CreateOrganization, model.APIOption{
		NeedLoginStatus: true,
	}))

	// 活动类接口
	eventGroup := r.Group("/team_up/event")
	// 活动页面(通了)
	eventGroup.POST("/page", API(handler.EventPage, model.APIOption{
		NeedLoginStatus: true,
	}))
	// 创建活动（通了）
	eventGroup.POST("/create", API(handler.CreateEvent, model.APIOption{
		NeedLoginStatus: true,
	}))
	// 更新活动信息（通了）
	eventGroup.POST("/update", API(handler.UpdateEvent, model.APIOption{
		NeedLoginStatus: true,
	}))
	//// 获取活动结果
	//eventGroup.POST("/get_result", API(handler.GetEventResult, model.APIOption{
	//	NeedLoginStatus: true,
	//}))

	// 管理员接口
	adminGroup := r.Group("/team_up/admin")
	// 获取待审批事件 (通了)
	adminGroup.GET("/get_approval_items", API(handler.GetApprovalItems, model.APIOption{
		NeedLoginStatus:    true,
		NeedAdminClearance: true,
	}))
	// 审批 （通了）
	adminGroup.POST("/approve", API(handler.Approve, model.APIOption{
		NeedLoginStatus:    true,
		NeedAdminClearance: true,
	}))

	// 静态资源Group
	imageGroup := r.Group("/team_up/static_image")
	// 用户定级职业的证明
	imageGroup.Static("/user_calibration_proof", "./calibration_proof")
	// 用户创建组织的logo
	imageGroup.Static("/organization_logo", "./organization_logo")

	// swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 处理请求图片响应的路由
	userGroup.GET("/image/:directoryname/:filename", func(c *gin.Context) {
		filename := c.Param("filename")
		directoryname := c.Param("directoryname")
		// 检查文件是否存在
		_, err := os.Stat(fmt.Sprintf("%s/%s", directoryname, filename))
		if os.IsNotExist(err) {
			c.String(http.StatusNotFound, "File not found")
			return
		}

		// 设置Content-Type头为image/jpeg
		c.Header("Content-Type", "image/jpeg")

		// 打开图片文件并将其写入响应
		file, err := os.Open(fmt.Sprintf("%s/%s", directoryname, filename))
		if err != nil {
			c.String(http.StatusInternalServerError, "Internal server error")
			return
		}
		defer file.Close()

		io.Copy(c.Writer, file)
	})

	return r
}
