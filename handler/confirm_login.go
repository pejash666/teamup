package handler

import (
	"errors"
	"github.com/bytedance/sonic"
	"gorm.io/gorm"
	"strconv"
	"teamup/constant"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/service/info"
	"teamup/service/login"
	"teamup/util"
)

type ConfirmLoginBody struct {
	SilentCode string `json:"silent_code"` // 静默登录的code（前段自动拿到的静默code）
	PhoneCode  string `json:"phone_code"`  // 获取电话号的code（需要用户显式授权）
}

type ConfirmLoginResp struct {
	ErrNo   int32             `json:"err_no"`
	ErrTips string            `json:"err_tips"`
	Data    map[string]string `json:"data"`
}

// ConfirmLogin godoc
//
//	@Summary		用户登录+获取手机号
//	@Description	前端使用微信code+获取手机号的code请求服务端登录
//	@Tags			/team_up/user
//	@Accept			json
//	@Produce		json
//	@Param			silent_code	body		string	true	"静默登录的code"
//	@Param			phone_code	body		string	true	"获取电话号的code"
//	@Success		200			{object}	ConfirmLoginResp
//	@Router			/team_up/user/confirm_login [post]
func ConfirmLogin(c *model.TeamUpContext) (interface{}, error) {
	util.Logger.Println("ConfirmLogin started")
	body := &ConfirmLoginBody{}
	err := c.BindJSON(body)
	if err != nil {
		util.Logger.Printf("[ConfirmLogin] bindJSON failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.ParamsError, err.Error())
	}
	if body.PhoneCode == "" || body.SilentCode == "" {
		util.Logger.Printf("[ConfirmLogin] missing code, body:%+v", body)
		return nil, iface.NewBackEndError(iface.ParamsError, "missing code")
	}
	// 获取用户openID和sessionKey
	c2s, err := login.Code2Session(c, body.SilentCode)
	if err != nil {
		return nil, iface.NewBackEndError(iface.InternalError, err.Error())
	}
	// 判断错误码
	if c2s.WechatBase != nil && c2s.WechatBase.ErrCode != 0 {
		// todo: 塞进err表？
		return nil, iface.NewBackEndError(c2s.ErrCode, c2s.ErrMsg)
	}
	// 获取用户手机号码
	phoneInfo, err := info.GetUserPhoneNumber(c, &model.GeneralCodeBody{Code: body.PhoneCode})
	if err != nil {
		return nil, iface.NewBackEndError(iface.InternalError, err.Error())
	}
	if phoneInfo.WechatBase != nil && phoneInfo.WechatBase.ErrCode != 0 {
		util.Logger.Printf("[ConfirmLogin] wechat resp code is not 0")
		return nil, iface.NewBackEndError(phoneInfo.ErrCode, phoneInfo.ErrMsg)
	}
	phoneNum, err := strconv.ParseInt(phoneInfo.PhoneInfo.PhoneNumber, 10, 64)
	if err != nil {
		util.Logger.Printf("[ConfirmLogin] parse phoneInfo failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.InternalError, err.Error())
	}
	// 创造三条新的纪录
	user := &mysql.WechatUserInfo{}
	err = util.DB().Where("open_id = ? AND is_primary = ?", c2s.OpenID, 1).Take(user).Error
	if err != nil {
		joinedEvent := make([]uint, 0)
		jes, _ := sonic.MarshalString(joinedEvent)
		joinedGroup := make([]string, 0)
		jgs, _ := sonic.MarshalString(joinedGroup)
		preference := make([]string, 0)
		ps, _ := sonic.MarshalString(preference)
		tags := make([]string, 0)
		ts, _ := sonic.MarshalString(tags)
		// 如果没找到主要记录则需要插入3条新的
		if errors.Is(err, gorm.ErrRecordNotFound) {
			var users = []mysql.WechatUserInfo{
				{
					OpenId:      c2s.OpenID,
					SessionKey:  c2s.SessionKey,
					UnionId:     c2s.UnionID,
					IsPrimary:   1,
					SportType:   constant.SportTypePadel,
					JoinedEvent: jes,
					JoinedGroup: jgs,
					Preference:  ps,
					Tags:        ts,
					PhoneNumber: uint(phoneNum),
				},
				{
					OpenId:      c2s.OpenID,
					IsPrimary:   0,
					SportType:   constant.SportTypePickelBall,
					JoinedEvent: jes,
					JoinedGroup: jgs,
					Preference:  ps,
					Tags:        ts,
				},
				{
					OpenId:      c2s.OpenID,
					IsPrimary:   0,
					SportType:   constant.SportTypeTennis,
					JoinedEvent: jes,
					JoinedGroup: jgs,
					Preference:  ps,
					Tags:        ts,
				},
			}

			err = util.DB().Create(&users).Error
			if err != nil {
				util.Logger.Printf("[ConfirmLogin] create user failed, err:%v", err)
				return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
			}
			util.Logger.Printf("[ConfirmLogin] save to DB success")
			goto jwtCreate
		}
		util.Logger.Printf("[ConfirmLogin] query user record failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
	}
	// 如果前端没有缓存的jwt token，但是服务端有数据，只需要更新session_key和电话号码
	user.SessionKey = c2s.SessionKey
	user.PhoneNumber = uint(phoneNum)
	err = util.DB().Save(user).Error
	if err != nil {
		util.Logger.Printf("[UserLogin] save user record failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
	}

jwtCreate:
	// 要将这个jwt返回给前端，前端缓存到local_storage
	jwt, err := util.CreateJWTToken(c, c2s.OpenID, c2s.SessionKey)
	if err != nil {
		util.Logger.Printf("login.CreateJWTToken failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.InternalError, err.Error())
	}

	return map[string]string{"user_token": jwt}, nil
}
