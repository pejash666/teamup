package handler

import (
	"errors"
	"github.com/bytedance/sonic"
	"gorm.io/gorm"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/util"
)

type MyTab struct {
	UserInfo       *UserInfo        `json:"user_info"`
	SportTypeInfos []*SportTypeInfo `json:"sport_type_infos"`
}

type UserInfo struct {
	NickName  string `json:"nick_name"`
	AvatarUrl string `json:"avatar_url"`
}

type SportTypeInfo struct {
	SportType      string        `json:"sport_type"`
	LevelInfo      *LevelInfo    `json:"level_info"` // 级别信息
	MyGames        []*Event      `json:"my_games"`
	MyOrganization *Organization `json:"my_organization"`
}

type LevelInfo struct {
	IsCalibrated bool           `json:"is_calibrated"`
	CurrentLevel float32        `json:"current_level"`
	LevelDetail  []*LevelChange `json:"level_detail"`
}

type LevelChange struct {
	Change float32 `json:"change"`
	Date   string  `json:"date"`
}

type Event struct {
	StartTime        int64       `json:"start_time"`
	EndTime          int64       `json:"end_time"`
	IsBooked         bool        `json:"is_booked"`
	FieldName        string      `json:"field_name"`
	CurrentPlayer    []*UserInfo `json:"current_player"`
	CurrentPlayerNum uint        `json:"current_player_num"`
	MaxPlayerNum     uint        `json:"max_player_num"`
	GameType         string      `json:"game_type"`
	MatchType        string      `json:"match_type"`
	LowestLevel      float32     `json:"lowest_level"`
	HighestLevel     float32     `json:"highest_level"`
}

type Organization struct {
	Name     string `json:"name"`
	EventNum int    `json:"event_num"`
}

type GetMyTabResp struct {
	ErrNo   int32  `json:"err_no"`
	ErrTips string `json:"err_tips"`
	Data    *MyTab `json:"data"`
}

// GetMyTab godoc
//
//	@Summary		我的页面
//	@Description	包含用户信息，级别信息和参与的活动
//	@Tags			/team_up/user
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	GetMyTabResp
//	@Router			/team_up/user/my_tab [get]
func GetMyTab(c *model.TeamUpContext) (interface{}, error) {
	res := &MyTab{}
	// 未登录返回空
	if c.BasicUser == nil {
		return res, nil
	}
	// 一次性返回所有支持的sport_type的信息
	var users []mysql.WechatUserInfo
	err := util.DB().Where("open_id = ?", c.BasicUser.OpenID).Find(&users).Error
	if err != nil {
		util.Logger.Printf("[GetMyTab] query users failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, "query users failed")
	}
	res.SportTypeInfos = make([]*SportTypeInfo, 0)
	for _, user := range users {
		// 从主记录获取用户信息
		if user.IsPrimary == 1 {
			res.UserInfo = &UserInfo{
				NickName:  user.Nickname,
				AvatarUrl: user.Avatar,
			}
		}
		sportTypeInfo := &SportTypeInfo{}
		sportTypeInfo.SportType = user.SportType
		sportTypeInfo.LevelInfo = &LevelInfo{
			IsCalibrated: user.IsCalibrated == 1,
			CurrentLevel: float32(user.Level) / 1000,
			LevelDetail:  GetRecentLevelChanges(user.OpenId, user.SportType, 10),
		}
		// 用户在这个sport_type里是host，获取organization信息
		if user.IsHost == 1 {
			organization := &mysql.Organization{}
			err = util.DB().Where("id = ?", user.OrganizationID).Take(organization).Error
			if err != nil {
				util.Logger.Printf("[GetMyTab] query organization failed, err:%v", err)
				return nil, iface.NewBackEndError(iface.MysqlError, "query organization failed")
			}
			sportTypeInfo.MyOrganization = &Organization{
				Name:     organization.Name,
				EventNum: organization.TotalEventNum,
			}
		}

		// 用户如果有参与的活动，则需要在这里展示
		if user.JoinedEvent != "" {
			// 获取用户参与的event
			joinedEvent := make([]uint, 0)
			err = sonic.UnmarshalString(user.JoinedEvent, &joinedEvent)
			if err != nil {
				util.Logger.Printf("[GetMyTab] unmarshalString failed, err:%v", err)
				return nil, iface.NewBackEndError(iface.InternalError, "unmarshal failed")
			}
			var events []mysql.EventMeta
			err = util.DB().Where("id IN ?", joinedEvent).Find(&events).Error
			if err != nil {
				util.Logger.Printf("[GetMyTab] query event failed, err:%v", err)
				return nil, iface.NewBackEndError(iface.MysqlError, "query failed")
			}
			sportTypeInfo.MyGames = make([]*Event, 0)
			for _, event := range events {
				eventShow := &Event{
					StartTime:        event.StartTime,
					EndTime:          event.EndTime,
					IsBooked:         event.IsBooked == 1,
					FieldName:        event.FieldName,
					CurrentPlayerNum: event.CurrentPlayerNum,
					MaxPlayerNum:     event.MaxPlayerNum,
					GameType:         event.GameType,
					MatchType:        event.MatchType,
					LowestLevel:      float32(event.LowestLevel) / 100,
					HighestLevel:     float32(event.HighestLevel) / 100,
				}
				// 获取参与这个活动的用户信息
				currentPeople := make([]string, 0)
				err = sonic.UnmarshalString(event.CurrentPlayer, &currentPeople)
				if err != nil {
					util.Logger.Printf("[GetMyTab] unmarshal failed, err:%v", err)
					return nil, iface.NewBackEndError(iface.MysqlError, "unmarshal failed")
				}
				var joiners []mysql.WechatUserInfo
				err = util.DB().Where("open_id IN ? AND is_primary = 1", currentPeople).Find(&joiners).Error
				if err != nil {
					util.Logger.Printf("[GetMyTab] query joiners info failed, err:%v", err)
					return nil, iface.NewBackEndError(iface.MysqlError, "query failed")
				}
				currentPLayers := make([]*UserInfo, 0)
				for _, joiner := range joiners {
					currentPLayers = append(currentPLayers, &UserInfo{
						NickName:  joiner.Nickname,
						AvatarUrl: joiner.Avatar,
					})
				}
				eventShow.CurrentPlayer = currentPLayers
				sportTypeInfo.MyGames = append(sportTypeInfo.MyGames, eventShow)
			}
		} else {
			util.Logger.Printf("[GetMytTab] user hasn't joined any event in this sport_type")
			sportTypeInfo.MyGames = make([]*Event, 0)
		}
		res.SportTypeInfos = append(res.SportTypeInfos, sportTypeInfo)
	}
	util.Logger.Printf("[GetMytTab] success, res:%+v", res)
	return res, nil
}

// GetRecentLevelChanges 获取最近 limit 场次内，分数的变化情况
func GetRecentLevelChanges(openID, sportType string, limit int) []*LevelChange {
	var records []mysql.UserEvent
	err := util.DB().Where("open_id = ? AND sport_type = ?", openID, sportType).Order("created_at desc").Find(&records).Limit(limit).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			util.Logger.Printf("[GetRecentLevelChanges] no record for level_change")
		}
	}
	res := make([]*LevelChange, 0)
	for _, record := range records {
		change := float32(record.LevelChange) / 100
		// 这里会出现一天多次记录的情况，前端需要额外关注
		date := record.CreatedAt.Format("20060102")
		if record.IsIncrease == 0 {
			change = change * (-1)
		}
		levelChange := &LevelChange{
			Change: change,
			Date:   date,
		}
		res = append(res, levelChange)
	}
	return res
}
