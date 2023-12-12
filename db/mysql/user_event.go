package mysql

import "gorm.io/gorm"

// UserEvent: 每一个用户参与的每一次活动会有一个
type UserEvent struct {
	gorm.Model
	EventID     uint   `gorm:"column:event_id" json:"event_id"`     // 关联上事件元信息
	OpenID      string `gorm:"column:open_id" json:"open_id"`       // 用户OpenID
	SportType   string `gorm:"column:sport_type" json:"sport_type"` // 运动类型
	IsQuit      uint   `gorm:"column:is_quit" json:"is_quit"`       // 是否已经退出
	LevelChange string `gorm:"level_change" json:"level_change"`    // 这场活动的level变化
}
