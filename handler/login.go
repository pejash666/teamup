package handler

import (
	"errors"
	"github.com/bytedance/sonic"
	"gorm.io/gorm"
	"teamup/constant"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/service/login"
	"teamup/util"
)

type UserLoginResp struct {
	ErrNo   int32             `json:"err_no"`
	ErrTips string            `json:"err_tips"`
	Data    map[string]string `json:"data"`
}

// UserLogin godoc
//
//	@Summary		用户登录(废弃，请使用confirm_login)
//	@Description	前端使用微信code请求服务端登录
//	@Tags			/team_up/user
//	@Accept			json
//	@Produce		json
//	@Param			code	body		string	true	"微信Code"
//	@Success		200		{object}	UserLoginResp
//	@Router			/team_up/user/login [post]
func UserLogin(c *model.TeamUpContext) (interface{}, error) {
	util.Logger.Println("UserLogin started")
	body := &model.GeneralCodeBody{}
	err := c.BindJSON(body)
	if err != nil {
		util.Logger.Printf("[UserLogin] BindJSON failed,err:%v", err)
		return nil, iface.NewBackEndError(iface.InternalError, err.Error())
	}

	c2s, err := login.Code2Session(c, body.Code)
	if err != nil {
		return nil, iface.NewBackEndError(iface.InternalError, err.Error())
	}
	// 判断错误码
	if c2s.WechatBase != nil && c2s.WechatBase.ErrCode != 0 {
		// todo: 塞进err表？
		return nil, iface.NewBackEndError(c2s.ErrCode, c2s.ErrMsg)
	}
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
				util.Logger.Printf("[UserLogin] create user failed, err:%v", err)
				return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
			}
			util.Logger.Printf("[UserLogin] save to DB success")
			goto jwtCreate
		}
		util.Logger.Printf("[UserLogin] query user record failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
	}
	// 如果前端没有缓存的jwt token，但是服务端有数据，只需要更新session_key
	user.SessionKey = c2s.SessionKey
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
