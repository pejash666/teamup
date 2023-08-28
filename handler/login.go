package handler

import (
	"errors"
	"gorm.io/gorm"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/service/login"
	"teamup/util"
)

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
	if c2s.ErrCode != 0 {
		// todo: 塞进err表？
		return nil, iface.NewBackEndError(c2s.ErrCode, c2s.ErrMsg)
	}
	user := &mysql.WechatUserInfo{}
	err = util.DB().Where("open_id = ?", c2s.OpenID).Take(user).Error
	if err != nil {
		// 如果没找到则需要插入一条新的
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user.OpenId = c2s.OpenID
			user.SessionKey = c2s.SessionKey
			user.UnionId = c2s.UnionID
			err = util.DB().Create(user).Error
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
	// 应该只要更新session_key?
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
