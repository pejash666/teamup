package mysql

import "gorm.io/gorm"

// UserEvent: 每一个用户参与的每一次活动会有一个记录
type UserEvent struct {
	gorm.Model
	EventID uint `gorm:"column:event_id" json:"event_id"` // 关联上事件元信息
	UserID  uint `gorm:"column:open_id" json:"open_id"`   // 参与的用户
}
