package handler

import (
	"strconv"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/service/info"
	"teamup/util"
)

func GetPhoneNumber(c *model.TeamUpContext) (interface{}, error) {
	util.Logger.Println("GetPhoneNumber started")
	body := &model.GeneralCodeBody{}
	err := c.BindJSON(body)
	if err != nil {
		util.Logger.Printf("[GetPhoneNumber] BindJSON failed,err:%v", err)
		return nil, iface.NewBackEndError(iface.InternalError, err.Error())
	}
	phoneInfo, err := info.GetUserPhoneNumber(c, body)
	if err != nil {
		return nil, iface.NewBackEndError(iface.InternalError, err.Error())
	}
	if phoneInfo.ErrCode != 0 {
		util.Logger.Printf("[GetPhoneNumber] wechat resp code is not 0")
		return nil, iface.NewBackEndError(phoneInfo.ErrCode, phoneInfo.ErrMsg)
	}
	// 更新user表的手机号码
	user := &mysql.WechatUserInfo{}
	err = util.DB().Where("open_id = ?", c.BasicUser.OpenID).Take(user).Error
	if err != nil {
		util.Logger.Printf("[GetPhoneNumber] get user from DB failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
	}
	num, err := strconv.ParseUint(phoneInfo.PhoneInfo.PhoneNumber, 10, 64)
	if err != nil {
		util.Logger.Printf("[GetPhoneNumber] ParseUint for phoneNumber failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.InternalError, err.Error())
	}
	user.PhoneNumber = uint(num)
	util.DB().Save(user)
	util.Logger.Printf("[GetPhoneNumber] success")
	return map[string]string{"code": body.Code}, nil
}
