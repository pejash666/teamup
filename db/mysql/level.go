package mysql

import "gorm.io/gorm"

type UserLevel struct {
	gorm.Model
	UserID     uint   `gorm:"column:user_id" json:"user_id"` // 参与的用户
	OpenID     string `gorm:"column:open_id" json:"open_id"`
	SportType  string `gorm:"column:sport_type" json:"sport_type"` // 运动类型
	Calibrated bool   `gorm:"column:calibrated" json:"calibrated"` // 是否已经定级
	Level      int32  `gorm:"column:level" json:"level"`           // 级别分数
}

func (u UserLevel) TableName() string {
	return "user_level"
}
