package handler

import (
	"strconv"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/service/info"
	"teamup/util"
)

// UpdateUserPhoneNumber godoc
//
//	@Summary		获取用户手机号
//	@Description	前端获取加密的用户手机号，服务端进行解码，存储
//	@Tags			/teamup/user
//	@Accept			json
//	@Produce		json
//	@Param			code	body		string	true	"微信Code"
//	@Success		200		{object}	model.BackEndResp
//	@Router			/teamup/user/update_phone_number [post]
func UpdateUserPhoneNumber(c *model.TeamUpContext) (interface{}, error) {
	util.Logger.Println("UpdateUserPhoneNumber started")
	body := &model.GeneralCodeBody{}
	err := c.BindJSON(body)
	if err != nil {
		util.Logger.Printf("[UpdateUserPhoneNumber] BindJSON failed,err:%v", err)
		return nil, iface.NewBackEndError(iface.InternalError, err.Error())
	}
	phoneInfo, err := info.GetUserPhoneNumber(c, body)
	if err != nil {
		return nil, iface.NewBackEndError(iface.InternalError, err.Error())
	}
	if phoneInfo.WechatBase != nil && phoneInfo.WechatBase.ErrCode != 0 {
		util.Logger.Printf("[UpdateUserPhoneNumber] wechat resp code is not 0")
		return nil, iface.NewBackEndError(phoneInfo.ErrCode, phoneInfo.ErrMsg)
	}
	// 更新user表的手机号码
	user := &mysql.WechatUserInfo{}
	err = util.DB().Where("open_id = ? AND is_primary = 1", c.BasicUser.OpenID).Take(user).Error
	if err != nil {
		util.Logger.Printf("[UpdateUserPhoneNumber] get user from DB failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
	}
	num, err := strconv.ParseUint(phoneInfo.PhoneInfo.PhoneNumber, 10, 64)
	if err != nil {
		util.Logger.Printf("[UpdateUserPhoneNumber] ParseUint for phoneNumber failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.InternalError, err.Error())
	}
	user.PhoneNumber = uint(num)
	util.DB().Save(user)
	util.Logger.Printf("[UpdateUserPhoneNumber] success")
	return nil, nil
}
