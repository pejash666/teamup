package util

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"
	"teamup/constant"
	"teamup/iface"
	"teamup/model"
)

func API(handler iface.HandlerFunc, opt model.APIOption) gin.HandlerFunc {
	//logInfo := getLogInfo(handler)

	ginWrapper := func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				Logger.Printf("panic recover, err:%v, stack:%v", err, string(debug.Stack()))
			}
		}()
		ctx, err := NewTeamUpContext(c)
		if err != nil {
			Logger.Printf("API.NewTeamUpContext failed, err:%v", err)
			return
		}
		data, err := handler(ctx)

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

func NewTeamUpContext(c *gin.Context) (*model.TeamUpContext, error) {
	ctx := &model.TeamUpContext{
		Context: c,
		AppInfo: &model.AppInfo{
			AppID:  constant.AppID,
			Secret: constant.AppSecret,
		},
	}
	// 获取access_token
	token, err := GetAccessToken(constant.AppID, constant.AppSecret)
	if err != nil {
		Logger.Printf("NewTeamUpContext GetAccessToken failed, err:%v", err)
		return nil, err
	}
	ctx.AccessToken = token
	// todo: 从db获取用户信息

	return ctx, nil
}

func makeUpRespData(data interface{}, err error) *model.BackEndResp {
	if err != nil {
		err.Error()
	}
}