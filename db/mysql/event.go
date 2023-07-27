package mysql

import "gorm.io/gorm"

// todo: 都需要展示那些信息
type Event struct {
	gorm.Model
	Status    string `gorm:"status" json:"status"`
	UserID    uint   `gorm:"user_id" json:"user_id"`       // user表主键
	EventType string `gorm:"event_type" json:"event_type"` // 运动类型

}
