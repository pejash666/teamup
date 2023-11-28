package handler

import (
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/service/login"
	"teamup/util"
)

type UpdateUserInfoReq struct {
	*model.WechatUserInfo
	RawData       string `json:"rawData"`       // 不包括敏感信息的原始数据字符串，用于计算签名
	Signature     string `json:"signature"`     // 使用 sha1( rawData + sessionkey ) 得到字符串，用于校验用户信息
	EncryptedData string `json:"encryptedData"` // 包括敏感数据在内的完整用户信息的加密数据
	Iv            string `json:"iv"`            // 加密算法的初始向量
	AvatarUrl     string `json:"avatar_url"`
	NickName      string `json:"nick_name"`
}

// UpdateUserInfo 前端调用getUserProfile后调用，更新后端用户信息表
func UpdateUserInfo(c *model.TeamUpContext) (interface{}, error) {
	body := &UpdateUserInfoReq{}
	err := c.BindJSON(body)
	if err != nil {
		util.Logger.Printf("[UpdateUserInfo] BindJSON failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.InternalError, "bindJson failed")
	}
	// 验签
	isPass, err := login.CheckFrontEndSignature(c, body.Signature, c.BasicUser.SessionKey, body.RawData)
	if err != nil {
		util.Logger.Printf("[UpdateUserInfo] CheckFrontEndSignature failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.InternalError, "CheckFrontEndSignature failed")
	}
	if !isPass {
		util.Logger.Printf("[UpdateUserInfo] signature not match")
		return nil, iface.NewBackEndError(iface.ParamsError, "invalid signature")
	}
	// 从DB获取用户信息
	user := &mysql.WechatUserInfo{}
	err = util.DB().Where("open_id = ?", c.BasicUser.OpenID).Take(user).Error
	if err != nil {
		util.Logger.Printf("[UpdateUserInfo] get record from db failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, "no record found")
	}
	//decryptedData, err := login.GetEncryptedData(c, c.BasicUser.SessionKey, body.EncryptedData, body.Iv)
	//if err != nil {
	//	util.Logger.Printf("[UpdateUserInfo] login.GetEncryptedData failed, err:%v", err)
	//	return nil, iface.NewBackEndError(iface.InternalError, "GetEncryptedData failed")
	//}
	// todo：不确定decryptedData这里面都有啥
	user.Avatar = body.AvatarUrl
	user.Nickname = body.NickName
	err = util.DB().Save(user).Error
	if err != nil {
		util.Logger.Printf("[UpdateUserInfo] save user record failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
	}
	return nil, nil
}
