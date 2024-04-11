package mysql

import "gorm.io/gorm"

type EventMeta struct {
	gorm.Model
	Status           string `gorm:"column:status" json:"status"` // full, can_join, finished, field_booked
	Creator          string `gorm:"column:creator" json:"creator"`
	IsHost           int    `gorm:"column:is_host" json:"is_host"`                 // 是否为组织发布
	OrganizationID   int64  `gorm:"column:organization_id" json:"organization_id"` // 组织ID
	SportType        string `gorm:"column:sport_type" json:"sport_type"`           // 运动类型，padel，tennis
	MatchType        string `gorm:"column:match_type" json:"match_type"`           // 比赛类型，competitive，entertainment，这两种类型都可以记分，但是competition会影响个人的level
	GameType         string `gorm:"column:game_type" json:"game_type"`             // 对局类型，单打/双打(padel只能双打)
	ScoreRule        string `gorm:"column:score_rule" json:"score_rule"`           // 记分规则，只有用户上传分数后才会记录
	Scorers          string `gorm:"column:scorers" json:"scorers"`                 // 记分员，记分时需要用户指定，如果为空，则所有参与用户均可都可以
	LowestLevel      int    `gorm:"column:lowest_level" json:"lowest_level"`       // 适合的最低级别
	HighestLevel     int    `gorm:"column:highest_level" json:"highest_level"`     // 适合的最高级别
	IsPublic         int    `gorm:"column:is_public" json:"is_public"`             // 是否是公开的
	IsBooked         int    `gorm:"column:is_booked" json:"is_booked"`             // 是否已定场地
	FieldType        string `gorm:"column:field_type" json:"field_type"`           // 场地类型（室外，室内）
	Date             string `gorm:"column:date" json:"date"`                       // 日期 20060102
	Weekday          string `gorm:"weekday" json:"weekday"`                        // Monday...
	City             string `gorm:"column:city" json:"city"`
	Name             string `gorm:"column:name" json:"name"`
	Desc             string `gorm:"column:desc" json:"desc"`
	StartTime        int64  `gorm:"column:start_time" json:"start_time"`
	StartTimeStr     string `gorm:"column:start_time_str" json:"start_time_str"`
	EndTime          int64  `gorm:"column:end_time" json:"end_time"`
	EndTimeStr       string `gorm:"column:end_time_str" json:"end_time_str"`
	FieldName        string `gorm:"column:field_name" json:"filed_name"`
	Longitude        string `gorm:"column:longitude" json:"longitude"`
	Latitude         string `gorm:"column:latitude" json:"latitude"`
	MaxPlayerNum     uint   `gorm:"column:max_player_num" json:"max_player_num"` // 匹克球只能是偶数
	CurrentPlayerNum uint   `gorm:"column:current_player_num" json:"current_player_num"`
	CurrentPlayer    string `gorm:"column:current_player" json:"current_player"`
	Price            uint   `gorm:"column:price" json:"price"`
}

func (e EventMeta) TableName() string {
	return "event_info"
}
