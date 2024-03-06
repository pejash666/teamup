package handler

import (
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
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

// UpdateUserInfo godoc
// @Summary      前端调用getUserProfile后调用，更新后端用户信息表
// @Description  前端获取加密的用户手机号，服务端进行解码，存储
// @Tags         /teamup/user
// @Accept       json
// @Produce      json
// @Param        code  body     UpdateUserInfoReq  true  "请求"
// @Success      200  {object}  model.BackEndResp
// @Router       /teamup/user/update_user_info [post]
func UpdateUserInfo(c *model.TeamUpContext) (interface{}, error) {
	body := &UpdateUserInfoReq{}
	err := c.BindJSON(body)
	if err != nil {
		util.Logger.Printf("[UpdateUserInfo] BindJSON failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.InternalError, "bindJson failed")
	}
	// 从DB获取用户信息
	user := &mysql.WechatUserInfo{}
	err = util.DB().Where("open_id = ? AND is_primary = 1", c.BasicUser.OpenID).Take(user).Error
	if err != nil {
		util.Logger.Printf("[UpdateUserInfo] get record from db failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, "no record found")
	}
	user.Avatar = body.AvatarUrl
	user.Nickname = body.NickName
	err = util.DB().Save(user).Error
	if err != nil {
		util.Logger.Printf("[UpdateUserInfo] save user record failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
	}
	return nil, nil
}
