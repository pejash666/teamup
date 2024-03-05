package model

type EventInfo struct {
	Id               int64   `json:"id"`                        // 数据库主键
	IsHost           bool    `json:"is_host"`                   // 是否是组织创建
	OrganizationID   int64   `json:"organization_id,omitempty"` // 组织ID
	Status           string  `json:"status"`                    // full, can_join, finished
	CreatorNickname  string  `json:"creator_nickname"`          // 创建者昵称
	CreatorAvatar    string  `json:"creator_avatar"`            // 创建者头像
	SportType        string  `json:"sport_type"`                // 运动类型
	IsCompetitive    bool    `json:"is_competitive"`            // 是否是竞赛类型
	GameType         string  `json:"game_type"`                 // 对局类型，solo/duo
	IsPublic         bool    `json:"is_public"`                 // 是否是公开比赛
	IsBooked         bool    `json:"is_booked"`                 // 是否已定场
	SelfJoin         bool    `json:"self_join,omitempty"`       // 自己是否加入
	FieldType        string  `json:"field_type,omitempty"`      // 场地类型
	LowestLevel      float32 `json:"lowest_level"`              // 适合的最低级别
	HighestLevel     float32 `json:"highest_level"`             // 适合的最高级别
	Date             string  `json:"date"`                      // 日期 20060102
	Weekday          string  `json:"weekday"`                   // 星期几
	City             string  `json:"city"`
	Name             string  `json:"name"`
	Desc             string  `json:"desc"`
	StartTime        int64   `json:"start_time"`
	StartTimeStr     string  `json:"start_time_str"`
	EndTime          int64   `json:"end_time"`
	EndTimeStr       string  `json:"end_time_str"`
	FieldName        string  `json:"field_name,omitempty"`
	MaxPeopleNum     uint    `json:"max_people_num"`
	CurrentPeople    uint    `json:"current_people"`
	Price            uint    `json:"price"`
	OrganizationLogo string  `json:"organization_logo,omitempty"`

	IsDraft bool `json:"is_draft,omitempty"` // 是否是草稿请求
}
