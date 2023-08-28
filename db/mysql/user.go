package mysql

import (
	"gorm.io/gorm"
)

// StringSlice 使用gorm默认的关键字 serializer 进行序列化
type StringSlice []string

type WechatUserInfo struct {
	gorm.Model              // embedded the basics
	SportType   string      `gorm:"column:sport_type" json:"sport_type"`                 // 运动类型
	UnionId     string      `gorm:"column:union_id" json:"union_id"`                     // 微信用户union_id
	OpenId      string      `gorm:"column:open_id" json:"open_id"`                       // 微信用户open_id
	SessionKey  string      `gorm:"session_key" json:"session_key"`                      // 微信session_key 用于解密
	Nickname    string      `gorm:"column:nickname" json:"nickname"`                     // 微信用户昵称
	Avatar      string      `gorm:"column:avatar" json:"avatar"`                         // 微信用户头像
	Gender      int8        `gorm:"column:gender" json:"gender"`                         // 性别
	PhoneNumber uint        `gorm:"column:phone_number" json:"phone_number"`             // 用户手机号
	JoinedTimes uint        `gorm:"joined_times" json:"joined_times"`                    // 参与次数
	Preference  StringSlice `gorm:"serializer:json;type:varchar(255)" json:"preference"` // 偏好
	Tags        StringSlice `gorm:"serializer:json;type:varchar(255)" json:"tags"`       // 用户标签
	JoinedGroup string      `gorm:"joined_group" json:"joined_group"`                    // 参加的组织
}

func (u WechatUserInfo) TableName() string {
	return "wechat_user_info"
}
