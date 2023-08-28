package util

import (
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"
	"teamup/constant"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"time"
)

const (
	Success = 0
)

func API(handler iface.HandlerFunc, opt model.APIOption) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				Logger.Printf("panic recover, err:%v, stack:%v", err, string(debug.Stack()))
			}
		}()
		ctx, err := NewTeamUpContext(c, opt)
		if err != nil {
			Logger.Printf("API.NewTeamUpContext failed, err:%v", err)
			return
		}
		data, err := handler(ctx)
		respData := makeUpRespData(data, err)
		resp, err := sonic.Marshal(respData)
		if err != nil {
			Logger.Printf("API MarshalString failed, err:%v", err)
			return
		}
		c.Data(http.StatusOK, "application/json; charset=utf-8", resp)
	}
}

func getLogInfo(handler iface.HandlerFunc) model.LogInfo {
	fPointer := reflect.ValueOf(handler).Pointer()
	f := runtime.FuncForPC(fPointer)

	fullFuncName := f.Name()
	funcStrs := strings.Split(fullFuncName, ".")
	funcName := funcStrs[len(funcStrs)-1]

	fullFileName, line := f.FileLine(fPointer)
	fileStrs := strings.Split(fullFileName, "/")
	fileName := fileStrs[len(fileStrs)-1]

	location := fmt.Sprintf("%s:%d", fileName, line) // e.g. task_page.go:12
	return model.LogInfo{FuncName: funcName, Location: location}
}

func NewTeamUpContext(c *gin.Context, opt model.APIOption) (*model.TeamUpContext, error) {
	ctx := &model.TeamUpContext{
		Context: c,
		AppInfo: &model.AppInfo{
			AppID:  constant.AppID,
			Secret: constant.AppSecret,
		},
	}
	// 生成随机种子
	ran := rand.New(rand.NewSource(time.Now().UnixNano()))
	ctx.ID = ran.Int63()
	// 获取access_token
	accessToken, err := GetAccessToken(constant.AppID, constant.AppSecret)
	if err != nil {
		Logger.Printf("NewTeamUpContext GetAccessToken failed, err:%v", err)
		return nil, err
	}
	ctx.AccessToken = accessToken
	if opt.NeedLoginStatus {
		// 解密JWT，放入ctx
		body := struct {
			WechatToken string `json:"wechat_token"`
		}{}
		err := c.BindJSON(&body)
		if err != nil {
			Logger.Printf("NewTeamUpContext c.BindJSON failed, err:%v", err)
			return nil, err
		}
		jwt, err := ParseJWTToken(body.WechatToken)
		if err != nil {
			Logger.Printf("NewTeamUpContext ParseJWTToken failed, err:%v", err)
			return nil, err
		}
		// 还要去DB获取到自己维护的UserID
		user := &mysql.WechatUserInfo{}
		if err = DB().Where("open_id = ?", jwt.OpenID).Take(user).Error; err != nil {
			Logger.Printf("NewTeamUpContext get user info from DB failed, err:%v", err)
			return nil, err
		}
		ctx.BasicUser = &model.BasicUser{
			OpenID:     jwt.OpenID,
			UserID:     user.ID,
			UnionID:    "",
			SessionKey: jwt.SessionKey,
		}
	}

	return ctx, nil
}

func makeUpRespData(data interface{}, err error) *model.BackEndResp {
	errNo := int32(Success)
	errTips := "success"
	if err != nil {
		beError := err.(*iface.BackEndError)
		errNo = beError.ErrNumber()
		errTips = beError.Error()
	}
	return &model.BackEndResp{
		ErrNo:   errNo,
		ErrTips: errTips,
		Data:    data,
	}
}
