package model

type EventInfo struct {
	Id             int64  `json:"id"`                   // 数据库主键
	IsHost         bool   `json:"is_host"`              // 是否是组织创建
	OrganizationID int64  `json:"organization_id"`      // 组织ID
	Status         string `gorm:"status" json:"status"` // full, can_join, finished
	Creator        string `gorm:"creator" json:"creator"`
	SportType      string `gorm:"sport_type" json:"sport_type"` // 运动类型
	IsCompetitive  bool   `json:"is_competitive"`               // 是否是竞赛类型
	GameType       string `json:"game_type"`                    // 对局类型，solo/duo
	IsPublic       bool   `json:"is_public"`                    // 是否是公开比赛
	IsBooked       bool   `json:"is_booked"`                    // 是否已定场
	SelfJoin       bool   `json:"self_join"`                    // 自己是否加入
	FieldType      string `json:"field_type"`                   // 场地类型
	LowestLevel    int    `json:"lowest_level"`                 // 适合的最低级别
	HighestLevel   int    `json:"highest_level"`                // 适合的最高级别
	Date           string `gorm:"date" json:"date"`             // 日期 20060102
	City           string `gorm:"city" json:"city"`
	Name           string `gorm:"name" json:"name"`
	Desc           string `gorm:"desc" json:"desc"`
	StartTime      int64  `gorm:"start_time" json:"start_time"`
	EndTime        int64  `gorm:"end_time" json:"end_time"`
	FieldName      string `gorm:"field_name" json:"field_name"`
	MaxPeople      uint   `gorm:"max_people" json:"max_people"`
	CurrentPeople  uint   `gorm:"current_people" json:"current_people"`
	Price          uint   `gorm:"price" json:"price"`

	IsDraft bool `json:"is_draft"` // 是否是草稿请求
}
