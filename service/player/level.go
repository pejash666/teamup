package player

import (
	"teamup/db/mysql"
	"teamup/model"
)

type PlayerLevel struct {
	SportType   string `json:"sport_type"`
	Initialized bool   `json:"level_status"` // 是否已经定级
	Score       int64  `json:"score"`        // 分数
}

func GetPlayerLevelModule(c *model.TeamUpContext, user *mysql.WechatUserInfo) (*PlayerLevel, error) {
	res := &PlayerLevel{
		SportType:   user.SportType,
		Initialized: false,
		Score:       0,
	}
}
