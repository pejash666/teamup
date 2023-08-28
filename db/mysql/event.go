package mysql

import "gorm.io/gorm"

type EventMeta struct {
	gorm.Model
	Status        string `gorm:"status" json:"status"` // full, can_join, finished
	Creator       string `gorm:"creator" json:"creator"`
	SportType     string `gorm:"sport_type" json:"sport_type"` // 运动类型
	Date          string `gorm:"date" json:"date"`             // 日期 20060102
	City          string `gorm:"city" json:"city"`
	Title         string `gorm:"title" json:"title"`
	Desc          string `gorm:"desc" json:"desc"`
	StartTime     int64  `gorm:"start_time" json:"start_time"`
	StartTimeStr  string `gorm:"start_time_str" json:"start_time_str"`
	EndTime       int64  `gorm:"end_time" json:"end_time"`
	EndTimeStr    string `gorm:"end_time_str" json:"end_time_str"`
	FieldName     string `gorm:"field_name" json:"filed_name"`
	MaxPeople     uint   `gorm:"max_people" json:"max_people"`
	CurrentPeople uint   `gorm:"current_people" json:"current_people"`
	Price         uint   `gorm:"price" json:"price"`
}

func (e EventMeta) TableName() string {
	return "event_meta"
}
