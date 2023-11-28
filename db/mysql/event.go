package mysql

import "gorm.io/gorm"

type EventMeta struct {
	gorm.Model
	Status           string `gorm:"column:status" json:"status"` // full, can_join, finished, field_booked
	Creator          string `gorm:"column:creator" json:"creator"`
	SportType        string `gorm:"column:sport_type" json:"sport_type"`       // 运动类型，pedal，tennis
	MatchType        string `gorm:"column:match_type" json:"match_type"`       // 比赛类型，competition，entertainment
	GameType         string `gorm:"column:game_type" json:"game_type"`         // 对局类型，单打/双打
	LowestLevel      int    `gorm:"column:lowest_level" json:"lowest_level"`   // 适合的最低级别
	HighestLevel     int    `gorm:"column:highest_level" json:"highest_level"` // 适合的最高级别
	IsPublic         int    `gorm:"column:is_public" json:"is_public"`         // 是否是公开的
	IsBooked         int    `gorm:"column:is_booked" json:"is_booked"`         // 是否已定场地
	FieldType        string `gorm:"column:field_type" json:"field_type"`       // 场地类型（室外，室内）
	Date             string `gorm:"column:date" json:"date"`                   // 日期 20060102
	City             string `gorm:"column:city" json:"city"`
	Name             string `gorm:"column:name" json:"name"`
	Desc             string `gorm:"column:desc" json:"desc"`
	StartTime        int64  `gorm:"column:start_time" json:"start_time"`
	StartTimeStr     string `gorm:"column:start_time_str" json:"start_time_str"`
	EndTime          int64  `gorm:"column:end_time" json:"end_time"`
	EndTimeStr       string `gorm:"column:end_time_str" json:"end_time_str"`
	FieldName        string `gorm:"column:field_name" json:"filed_name"`
	MaxPeopleNum     uint   `gorm:"column:max_people_num" json:"max_people_num"`
	CurrentPeopleNum uint   `gorm:"column:current_people_num" json:"current_people_num"`
	CurrentPeople    string `gorm:"column:current_people" json:"current_people"`
	Price            uint   `gorm:"column:price" json:"price"`
}

func (e EventMeta) TableName() string {
	return "event_meta"
}
