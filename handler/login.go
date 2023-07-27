package handler

import (
	"errors"
	"teamup/model"
	"teamup/service/login"
	"teamup/util"
)

func UserLogin(c *model.TeamUpContext) (interface{}, error) {
	util.Logger.Println("UserLogin started")
	body := &model.UserLoginBody{}
	err := c.BindJSON(body)
	if err != nil {
		util.Logger.Printf("[UserLogin] BindJSON failed,err:%v", err)
		return nil, err
	}

	c2s, err := login.Code2Session(c, body.Code)
	if err != nil {
		return nil, err
	}
	// 判断错误码
	if c2s.ErrCode != 0 {
		// todo: 塞进err表？
		return nil, errors.New("errcode not 0")
	}

	// todo: upsert用户表

	login.CreateToken(c2s.OpenID, c2s.SessionKey)

}
