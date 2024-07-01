package handler

import (
	"errors"
	"github.com/bytedance/sonic"
	"gorm.io/gorm"
	"strconv"
	"teamup/constant"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/util"
	"time"
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
	Status       string         `json:"status"` // 定级状态：need_to_calibrate;wait_for_approve;approved(7.0以下自动审批)
	CurrentLevel float32        `json:"current_level"`
	LevelDetail  []*LevelChange `json:"level_detail"`
}

type LevelChange struct {
	Change float32 `json:"change"`
	Date   string  `json:"date"`
}

type Event struct {
	ID                  uint        `json:"id"`
	EventName           string      `json:"even_name"`
	StartTime           int64       `json:"start_time"`
	EndTime             int64       `json:"end_time"`
	Weekday             string      `json:"weekday"`
	Date                string      `json:"date"`
	StartTimeStr        string      `json:"start_time_str"`
	EndTimeStr          string      `json:"end_time_str"`
	IsBooked            bool        `json:"is_booked"`
	FieldName           string      `json:"field_name"`
	Longitude           float64     `json:"longitude"`
	Latitude            float64     `json:"latitude"`
	CurrentPlayer       []*UserInfo `json:"current_player"`
	CurrentPlayerNum    uint        `json:"current_player_num"`
	MaxPlayerNum        uint        `json:"max_player_num"`
	GameType            string      `json:"game_type"`
	MatchType           string      `json:"match_type"`
	LowestLevel         float32     `json:"lowest_level"`
	HighestLevel        float32     `json:"highest_level"`
	Status              string      `json:"status"` // event状态：created;full;finished
	EventImage          string      `json:"event_image"`
	IsHost              bool        `json:"is_host"`
	OrganizationLogo    string      `json:"organization_logo"` // 组织图片
	OrganizationAddress string      `json:"organization_address"`
}

type Organization struct {
	ID        uint    `json:"id"`
	Name      string  `json:"name"`
	Logo      string  `json:"logo"`
	Address   string  `json:"address"`
	Longitude float64 `json:"longitude,omitempty"`
	Latitude  float64 `json:"latitude,omitempty"`
	EventNum  int     `json:"event_num"`
	Status    string  `json:"status"` // 组织状态：no_organization;wait_for_approve;approved
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
		levelInfo := &LevelInfo{}
		levelInfo.Status = GetCalibrationStatus(user.IsCalibrated == 1, user.Level)
		// 只有定级过且审批通过的的人，才返回信息
		if levelInfo.Status == "approved" {
			levelInfo.CurrentLevel = float32(user.Level / 1000)
			levelInfo.LevelDetail = GetRecentLevelChanges(user.OpenId, user.SportType, 10)
		}
		sportTypeInfo.LevelInfo = levelInfo

		// 用sport_type + open_id 查找organization表
		organization := &mysql.Organization{}
		myOrganization := &Organization{}
		err = util.DB().Where("sport_type = ? AND host_open_id = ?", user.SportType, user.OpenId).Take(organization).Error
		if err != nil {
			// 名下无组织
			if errors.Is(err, gorm.ErrRecordNotFound) {
				util.Logger.Printf("[GetMyTab] user:%s has no organization in this sport_type:%s", user.Nickname, user.SportType)
				myOrganization.Status = "no_organization"
			} else {
				util.Logger.Printf("[GetMyTab] query organization using soprt_type and openID failed, err:%v", err)
				return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
			}
			// 查到记录了, 检查状态
		} else {
			myOrganization.ID = organization.ID
			myOrganization.Name = organization.Name
			myOrganization.Address = organization.Address
			myOrganization.Logo = organization.Logo
			myOrganization.EventNum = organization.TotalEventNum
			// 已经审批通过了
			if organization.IsApproved == 1 {
				myOrganization.Status = "approved"
				// 还没审批通过呢
			} else {
				myOrganization.Status = "wait_for_approve"
			}
		}
		sportTypeInfo.MyOrganization = myOrganization

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
					ID:               event.ID,
					EventName:        event.Name,
					StartTime:        event.StartTime,
					EndTime:          event.EndTime,
					Date:             time.Unix(event.StartTime, 0).Format("2006-01-02"),
					Weekday:          time.Unix(event.StartTime, 0).Weekday().String(),
					StartTimeStr:     event.StartTimeStr,
					EndTimeStr:       event.EndTimeStr,
					IsBooked:         event.IsBooked == 1,
					FieldName:        event.FieldName,
					CurrentPlayerNum: event.CurrentPlayerNum,
					MaxPlayerNum:     event.MaxPlayerNum,
					GameType:         event.GameType,
					MatchType:        event.MatchType,
					LowestLevel:      float32(event.LowestLevel) / 1000,
					HighestLevel:     float32(event.HighestLevel) / 1000,
					EventImage:       event.EventImage,
					IsHost:           event.IsHost == 1,
				}
				// 获取组织的信息
				if eventShow.IsHost {
					orga := &mysql.Organization{}
					err = util.DB().Where("id = ?", event.OrganizationID).Take(orga).Error
					if err != nil {
						util.Logger.Printf("[GetMyTab] query organization failed for id:%d", event.OrganizationID)
						return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
					}
					util.Logger.Printf("[GetMyTab] event_id:%d, organization_id:%d, logo:%v, address:%v", event.ID, event.OrganizationID, orga.Logo, orga.Address)
					eventShow.OrganizationLogo = orga.Logo
					eventShow.OrganizationAddress = orga.Address
				}
				util.Logger.Printf("[GetMyTab] event_id:%d, start_time:%d, end_time:%d, time_now:%v", event.ID, event.StartTime, event.EndTime, time.Now().Unix())
				// 获取status
				if time.Now().Unix() > event.StartTime && time.Now().Unix() < event.EndTime {
					eventShow.Status = constant.EventStatusInProgress
				} else {
					eventShow.Status = event.Status
				}
				// 订场地只能从已合作的场地里面选择
				if event.FieldName != "" && event.Latitude != "" && event.Longitude != "" {
					longitude, errP := strconv.ParseFloat(event.Longitude, 64)
					if errP != nil {
						return nil, iface.NewBackEndError(iface.InternalError, err.Error())
					}
					latitude, errP := strconv.ParseFloat(event.Latitude, 64)
					if errP != nil {
						return nil, iface.NewBackEndError(iface.InternalError, err.Error())
					}
					eventShow.Latitude = latitude
					eventShow.Longitude = longitude
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
	resStr, _ := sonic.MarshalString(res)
	util.Logger.Printf("[GetMytTab] success, res is %v", resStr)
	return res, nil
}

// GetRecentLevelChanges 获取最近 limit 场次内，分数的变化情况
func GetRecentLevelChanges(openID, sportType string, limit int) []*LevelChange {
	var records []mysql.UserEvent
	err := util.DB().Where("open_id = ? AND sport_type = ?", openID, sportType).Order("created_at desc").Find(&records).Limit(limit).Error
	if err != nil {
		// 没有变化信息，返回空
		if errors.Is(err, gorm.ErrRecordNotFound) {
			util.Logger.Printf("[GetRecentLevelChanges] no record for level_change")
			return make([]*LevelChange, 0)
		}
	}
	res := make([]*LevelChange, 0)
	for _, record := range records {
		change := float32(record.LevelChange) / 1000
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

// GetCalibrationStatus 获取用户定级状态
func GetCalibrationStatus(isCalibrated bool, level int) string {
	if isCalibrated {
		return "approved"
	} else {
		if level == 7000 {
			return "wait_for_approve"
		} else {
			return "need_to_calibrate"
		}
	}
}
