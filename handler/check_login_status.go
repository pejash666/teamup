package handler

import (
	"teamup/iface"
	"teamup/model"
	"teamup/service/login"
	"teamup/util"
)

type CheckLoginStatusReq struct {
	RawData string `json:"raw_data"`
}

// CheckLoginStatus 校验用户登录态，前端传来rawData和signature
func CheckLoginStatus(c *model.TeamUpContext) (interface{}, error) {
	if c.BasicUser == nil {
		return nil, iface.NewBackEndError(iface.InternalError, "user not exist")
	}
	if c.BasicUser.SessionKey == "" {
		return nil, iface.NewBackEndError(iface.InternalError, "session key not exist")
	}
	body := &CheckLoginStatusReq{}
	err := c.BindJSON(body)
	if err != nil {
		util.Logger.Printf("[CheckLoginStatus] BindJSON failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.InvalidRequest, "invalid request")
	}
	// 校验服务端的session_key有没有过期
	isExpired, err := login.CheckDBSessionKeyExpired(c, c.BasicUser.SessionKey, body.RawData)
	if err != nil {
		util.Logger.Printf("[CheckLoginStatus] CheckDBSessionKeyExpired failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.InternalError, "CheckDBSessionKeyExpired")
	}
	res := make(map[string]interface{})
	res["is_expired"] = isExpired
	return res, nil
}
