package mysql

import "gorm.io/gorm"

type WechatUserInfo struct {
	gorm.Model         // embedded the basics
	UnionId     string `gorm:"column:union_id" json:"union_id"`         // 微信用户union_id
	OpenId      string `gorm:"column:open_id" json:"open_id"`           // 微信用户open_id
	Nickname    string `gorm:"column:nickname" json:"nickname"`         // 微信用户昵称
	Avatar      string `gorm:"column:avatar" json:"avatar"`             // 微信用户头像
	Gender      int8   `gorm:"column:gender" json:"gender"`             // 性别
	PhoneNumber uint   `gorm:"column:phone_number" json:"phone_number"` // 用户手机号
}
