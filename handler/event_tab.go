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
	EventInfo *model.EventInfo
	Players   []*model.Player
}

func GetEventTab(c *model.TeamUpContext) (interface{}, error) {
	type Body struct {
		EventID int64 `json:"event_id"`
	}
	body := &Body{}
	err := c.BindJSON(body)
	if err != nil {
		util.Logger.Printf("[GetEventTab] BindJSON failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.ParamsError, "invalid body")
	}
	// 获取任务元信息, 能看到的都是发布后的，不能是草稿状态
	eventMeta := &mysql.EventMeta{}
	result := util.DB().Where("id = ? AND status <> ?", body.EventID, constant.EventStatusDraft).Take(eventMeta)
	if result.Error != nil {
		util.Logger.Printf("[GetEventTab] get event meta from DB failed, err:%v", result.Error)
		return nil, iface.NewBackEndError(iface.MysqlError, "get record failed")
	}
	eventTab := &EventTab{}
	eventInfo := &model.EventInfo{}
	eventInfo.IsHost = eventMeta.IsHost == 1
	if eventInfo.IsHost {
		organization := &mysql.Organization{}
		result := util.DB().Where("id = ?", eventMeta.OrganizationID).Take(organization)
		if result.Error != nil {
			util.Logger.Printf("[GetEventTab] get organization record from DB failed, err:%v", result.Error)
			return nil, iface.NewBackEndError(iface.MysqlError, "get record failed")
		}
		eventInfo.OrganizationLogo = organization.Logo
	}
	eventInfo.SportType = eventMeta.SportType
	eventInfo.Weekday = eventMeta.Weekday
	eventInfo.StartTime = eventMeta.StartTime
	eventInfo.EndTime = eventMeta.EndTime
	eventInfo.GameType = eventMeta.GameType
	eventInfo.IsCompetitive = eventMeta.MatchType == constant.EventMatchTypeCompetitive
	eventInfo.IsPublic = eventMeta.IsPublic == 1
	eventInfo.IsBooked = eventMeta.IsBooked == 1
	if eventInfo.IsBooked {
		eventInfo.FieldName = eventMeta.FieldName
		eventInfo.FieldType = eventMeta.FieldType
	}
	eventInfo.LowestLevel = float32(eventMeta.LowestLevel / 100)
	eventInfo.HighestLevel = float32(eventMeta.HighestLevel / 100)
	eventInfo.Price = eventMeta.Price
	eventInfo.MaxPeopleNum = eventMeta.MaxPlayerNum
	// 给结果赋值
	eventTab.EventInfo = eventInfo

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
		players := make([]*model.Player, 0)
		for _, user := range users {
			player := &model.Player{
				NickName:     user.Nickname,
				Avatar:       user.Avatar,
				IsCalibrated: user.IsCalibrated == 1,
				Level:        float32(user.Level / 100),
			}
			players = append(players, player)
		}
		eventTab.Players = players
	} else {
		eventTab.Players = make([]*model.Player, 0)
	}
	util.Logger.Printf("[GetEventTab] success, res:%+v", eventInfo)
	return eventInfo, nil
}