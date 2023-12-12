package handler

import (
	"github.com/bytedance/sonic"
	"teamup/constant"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/util"
)

type EventTab struct {
	IsHost           bool     `json:"is_host"`
	SportType        string   `json:"sport_type"`
	StartTime        int64    `json:"start_time"`
	EndTime          int64    `json:"end_time"`
	GameType         string   `json:"game_type"`
	MatchType        string   `json:"match_type"`
	IsPublic         bool     `json:"is_public"`
	IsBooked         bool     `json:"is_booked"`
	FieldName        string   `json:"field_name"`
	FieldType        string   `json:"field_type"`
	LowestLevel      int      `json:"lowest_level"`
	HighestLevel     int      `json:"highest_level"`
	Price            uint     `json:"price"`
	MaxPlayerNum     uint     `json:"max_player_num"`
	CurrentPlayerNum uint     `json:"current_player_num"`
	Players          []Player `json:"players"`

	OrganizationImage string `json:"organization_image"` // 只有is_host时才有
}

type Player struct {
	NickName     string `json:"nick_name"`
	Avatar       string `json:"avatar"`
	IsCalibrated bool   `json:"is_calibrated"`
	Level        int    `json:"level"`
}

func GetEventTab(c *model.TeamUpContext) (interface{}, error) {
	type Req struct {
		EventID int64 `json:"event_id"`
	}
	req := &Req{}
	err := c.BindJSON(req)
	if err != nil {
		util.Logger.Printf("[GetEventTab] BindJSON failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.ParamsError, "invalid req")
	}
	// 获取任务元信息, 能看到的都是发布后的，不能是草稿状态
	eventMeta := &mysql.EventMeta{}
	result := util.DB().Where("id = ? AND status <> ?", req.EventID, constant.EventStatusDraft).Take(eventMeta)
	if result.Error != nil {
		util.Logger.Printf("[GetEventTab] get event meta from DB failed, err:%v", result.Error)
		return nil, iface.NewBackEndError(iface.MysqlError, "get record failed")
	}
	eventTab := &EventTab{}
	eventTab.IsHost = eventMeta.IsHost == 1
	if eventTab.IsHost {
		organization := &mysql.Organization{}
		result := util.DB().Where("id = ?", eventMeta.OrganizationID).Take(organization)
		if result.Error != nil {
			util.Logger.Printf("[GetEventTab] get organization record from DB failed, err:%v", result.Error)
			return nil, iface.NewBackEndError(iface.MysqlError, "get record failed")
		}
		eventTab.OrganizationImage = organization.Logo
	}
	eventTab.SportType = eventMeta.SportType
	eventTab.StartTime = eventMeta.StartTime
	eventTab.EndTime = eventMeta.EndTime
	eventTab.GameType = eventMeta.GameType
	eventTab.MatchType = eventMeta.MatchType
	eventTab.IsPublic = eventMeta.IsPublic == 1
	eventTab.IsBooked = eventMeta.IsBooked == 1
	if eventTab.IsBooked {
		eventTab.FieldName = eventMeta.FieldName
		eventTab.FieldType = eventMeta.FieldType
	}
	eventTab.LowestLevel = eventMeta.LowestLevel
	eventTab.HighestLevel = eventMeta.HighestLevel
	eventTab.Price = eventMeta.Price
	eventTab.MaxPlayerNum = eventMeta.MaxPlayerNum
	// 已经有人加入，则需要展示已加入用户信息
	if eventMeta.CurrentPlayerNum > 0 {
		currentPeople := make([]string, 0)
		err = sonic.UnmarshalString(eventMeta.CurrentPlayer, &currentPeople)
		if err != nil {
			util.Logger.Printf("[GetEventTab] Unmarshal currentPeople from DB failed, err:%v", err)
			return nil, iface.NewBackEndError(iface.InternalError, "unmarshall error")
		}
		var users []mysql.WechatUserInfo
		result := util.DB().Where("open_id IN ? AND sport_type = ?", currentPeople, eventMeta.SportType).Find(&users)
		if result.Error != nil {
			util.Logger.Printf("[GetEventTab] find player in DB failed, err:%v", err)
			return nil, iface.NewBackEndError(iface.MysqlError, "get record failed")
		}
		if len(users) != int(eventMeta.CurrentPlayerNum) {
			util.Logger.Printf("[GetEventTab] unmatched currentplayer info")
			return nil, iface.NewBackEndError(iface.InternalError, "unmatched player info")
		}
		players := make([]Player, 0)
		for _, user := range users {
			player := Player{
				NickName:     user.Nickname,
				Avatar:       user.Avatar,
				IsCalibrated: user.IsCalibrated == 1,
				Level:        user.Level,
			}
			players = append(players, player)
		}
		eventTab.Players = players
	} else {
		eventTab.Players = make([]Player, 0)
	}
	util.Logger.Printf("[GetEventTab] success, res:%+v", eventTab)
	return eventTab, nil
}
