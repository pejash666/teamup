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
	"strconv"
	"strings"
	"teamup/constant"
	"teamup/iface"
	"teamup/model"
	"time"
)

const (
	Success = 0
)

var AdminList = []string{
	"oLI--62QB5L0bVJrFTFMTQj-lcGw", // 沙昊
	"oLI--69NwbFabrcwx7tDFgyEJkqc", // 李宇宙
	"oLI--635r8g0hwFquInKJQCatnik", // Flynn
	"oLI--60yUDyagDYRZf-tgolEU7W8", //  brd                          // Brd
}

func API(handler iface.HandlerFunc, opt model.APIOption) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {

			if err := recover(); err != nil {
				Logger.Printf("panic recover, err:%v, stack:%v", err, string(debug.Stack()))
			}
		}()
		// 获取token

		ctx, err := NewTeamUpContext(c, opt)
		if err != nil {
			Logger.Printf("API.NewTeamUpContext failed, err:%v", err)
			respData := makeUpRespData(nil, err)
			resp, _ := sonic.Marshal(respData)
			c.Data(http.StatusOK, "application/json; charset=utf-8", resp)
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
	// 跳过校验
	if opt.HackLogic {
		return ctx, nil
	}
	// 防止接口重放
	timeStampStr := c.GetHeader("timestamp")
	ts, err := strconv.ParseInt(timeStampStr, 10, 64)
	if err != nil {
		Logger.Printf("NewTeamUpContext get ts from header failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.ParamsError, "missing timestamp")
	}
	// （测试先注释掉）
	//randomStr := c.GetHeader("nonce")
	//if randomStr == "" {
	//	Logger.Printf("NewTeamUpContext get nonce from header failed, err:%v", err)
	//	return nil, errors.New("nonce missing")
	//}
	//antiReplayKey := fmt.Sprintf("server_antispam_timestamp_%d_nonce_%s", ts, randomStr)
	//res, err := RedisGet(antiReplayKey)
	//if err != nil && res != "" {
	//	Logger.Printf("NewTeamUpContext req too frequent")
	//	return nil, errors.New("too frequent")
	//}
	//_ = RedisSet(antiReplayKey, 1, time.Second*15)

	ctx.Timestamp = ts
	// 获取语言
	lang := c.GetHeader("lang")
	if lang == "" {
		lang = "zh_CN"
	}
	ctx.Language = lang

	// 生成随机种子
	ran := rand.New(rand.NewSource(time.Now().UnixNano()))
	ctx.Rand = ran
	ctx.ID = ran.Int()
	// 获取access_token
	accessToken, err := GetAccessToken(constant.AppID, constant.AppSecret)
	if err != nil {
		Logger.Printf("NewTeamUpContext GetAccessToken failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.WechatError, err.Error())
	}
	ctx.AccessToken = accessToken
	if opt.NeedLoginStatus {
		wechatToken := c.GetHeader("wechat_token")
		if wechatToken == "" {
			Logger.Printf("NewTeamUpContext c.GetHeader failed, err:%v", err)
			return nil, iface.NewBackEndError(iface.ParamsError, "invalid wechat_token")
		}
		jwt, errA := ParseJWTToken(wechatToken)
		if errA != nil {
			Logger.Printf("NewTeamUpContext ParseJWTToken failed, err:%v", err)
			return nil, iface.NewBackEndError(iface.InternalError, err.Error())
		}
		//// 还要去DB获取到自己维护的UserID
		//user := &mysql.WechatUserInfo{}
		//if err = DB().Where("open_id = ?", jwt.OpenID).Take(user).Error; err != nil {
		//	Logger.Printf("NewTeamUpContext get user info from DB failed, err:%v", err)
		//	return nil, err
		//}
		ctx.BasicUser = &model.BasicUser{
			OpenID:     jwt.OpenID,
			UnionID:    "",
			SessionKey: jwt.SessionKey,
		}
	}
	// 检查管理员权限
	if opt.NeedAdminClearance {
		if !opt.NeedLoginStatus {
			return nil, iface.NewBackEndError(iface.ParamsError, "admin need to login")
		}
		if !IsAdmin(ctx.BasicUser.OpenID) {
			return nil, iface.NewBackEndError(iface.ParamsError, "user not admin")
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

func IsAdmin(openID string) bool {
	// 只有几个人的openid有这个权限
	for _, admin := range AdminList {
		if admin == openID {
			return true
		}
	}
	return false
}
