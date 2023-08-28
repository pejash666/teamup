package model

type EventInfo struct {
	Status        string `gorm:"status" json:"status"` // full, can_join, finished
	Creator       string `gorm:"creator" json:"creator"`
	SportType     string `gorm:"sport_type" json:"sport_type"` // 运动类型
	Date          string `gorm:"date" json:"date"`             // 日期 20060102
	City          string `gorm:"city" json:"city"`
	Title         string `gorm:"title" json:"title"`
	Desc          string `gorm:"desc" json:"desc"`
	StartTime     int64  `gorm:"start_time" json:"start_time"`
	EndTime       int64  `gorm:"end_time" json:"end_time"`
	FieldName     string `gorm:"field_name" json:"filed_name"`
	MaxPeople     uint   `gorm:"max_people" json:"max_people"`
	CurrentPeople uint   `gorm:"current_people" json:"current_people"`
	Price         uint   `gorm:"price" json:"price"`

	IsDraft bool `json:"is_draft"` // 是否是草稿请求
}
