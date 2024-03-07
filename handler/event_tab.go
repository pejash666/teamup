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
	EventInfo *model.EventInfo `json:"event_info"`
	Players   []*model.Player  `json:"players"`
}

type EventPageBody struct {
	EventID int64 `json:"event_id"`
}

// EventPage godoc
//
//	@Summary		获取活动详情页
//	@Description	活动详情页，包含活动元信息和参与的用户信息
//	@Tags			/teamup/event
//	@Accept			json
//	@Produce		json
//	@Param			get_event_tab	body		{object}	EventPageBody	true	"详情页入参"
//	@Success		200				{object}	EventTab
//	@Router			/teamup/event/page [post]
func EventPage(c *model.TeamUpContext) (interface{}, error) {
	body := &EventPageBody{}
	err := c.BindJSON(body)
	if err != nil {
		util.Logger.Printf("[EventPage] BindJSON failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.ParamsError, "invalid body")
	}
	// 获取任务元信息, 能看到的都是发布后的，不能是草稿状态
	eventMeta := &mysql.EventMeta{}
	result := util.DB().Where("id = ? AND status <> ?", body.EventID, constant.EventStatusDraft).Take(eventMeta)
	if result.Error != nil {
		util.Logger.Printf("[EventPage] get event meta from DB failed, err:%v", result.Error)
		return nil, iface.NewBackEndError(iface.MysqlError, "get record failed")
	}
	eventTab := &EventTab{}
	eventInfo := &model.EventInfo{}
	eventInfo.IsHost = eventMeta.IsHost == 1
	if eventInfo.IsHost {
		organization := &mysql.Organization{}
		result := util.DB().Where("id = ?", eventMeta.OrganizationID).Take(organization)
		if result.Error != nil {
			util.Logger.Printf("[EventPage] get organization record from DB failed, err:%v", result.Error)
			return nil, iface.NewBackEndError(iface.MysqlError, "get record failed")
		}
		eventInfo.OrganizationID = int64(organization.ID)
		eventInfo.OrganizationLogo = organization.Logo
	}
	eventInfo.Id = int64(eventMeta.ID)
	eventInfo.Desc = eventMeta.Desc
	eventInfo.Name = eventMeta.Name
	eventInfo.SportType = eventMeta.SportType
	eventInfo.Status = eventMeta.Status
	eventInfo.Weekday = eventMeta.Weekday
	eventInfo.City = eventMeta.City
	eventInfo.StartTime = eventMeta.StartTime
	eventInfo.StartTimeStr = eventMeta.StartTimeStr
	eventInfo.EndTime = eventMeta.EndTime
	eventInfo.EndTimeStr = eventMeta.EndTimeStr
	eventInfo.Date = eventMeta.Date
	eventInfo.GameType = eventMeta.GameType
	eventInfo.IsCompetitive = eventMeta.MatchType == constant.EventMatchTypeCompetitive
	eventInfo.IsPublic = eventMeta.IsPublic == 1
	eventInfo.IsBooked = eventMeta.IsBooked == 1
	if eventInfo.IsBooked {
		eventInfo.FieldName = eventMeta.FieldName
		eventInfo.FieldType = eventMeta.FieldType
	}
	eventInfo.LowestLevel = float32(eventMeta.LowestLevel) / 1000
	eventInfo.HighestLevel = float32(eventMeta.HighestLevel) / 1000
	eventInfo.Price = eventMeta.Price
	eventInfo.MaxPeopleNum = eventMeta.MaxPlayerNum
	// 获取creator的头像，名字
	creator := &mysql.WechatUserInfo{}
	err = util.DB().Where("open_id = ? AND sport_type = ?", eventMeta.Creator, eventMeta.SportType).Take(creator).Error
	if err != nil {
		util.Logger.Printf("[EventPage] query creator failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
	}
	eventInfo.CreatorNickname = creator.Nickname
	eventInfo.CreatorAvatar = creator.Avatar
	// 给结果赋值
	eventTab.EventInfo = eventInfo

	// 已经有人加入，则需要展示已加入用户信息
	if eventMeta.CurrentPlayerNum > 0 {
		currentPeople := make([]string, 0)
		err = sonic.UnmarshalString(eventMeta.CurrentPlayer, &currentPeople)
		if err != nil {
			util.Logger.Printf("[EventPage] Unmarshal currentPeople from DB failed, err:%v", err)
			return nil, iface.NewBackEndError(iface.InternalError, "unmarshall error")
		}
		var users []mysql.WechatUserInfo
		result := util.DB().Where("open_id IN ? AND sport_type = ?", currentPeople, eventMeta.SportType).Find(&users)
		if result.Error != nil {
			util.Logger.Printf("[EventPage] find player in DB failed, err:%v", err)
			return nil, iface.NewBackEndError(iface.MysqlError, "get record failed")
		}
		if len(users) != int(eventMeta.CurrentPlayerNum) {
			util.Logger.Printf("[EventPage] unmatched currentplayer info")
			return nil, iface.NewBackEndError(iface.InternalError, "unmatched player info")
		}
		players := make([]*model.Player, 0)
		for _, user := range users {
			player := &model.Player{
				NickName:     user.Nickname,
				Avatar:       user.Avatar,
				IsCalibrated: user.IsCalibrated == 1,
				Level:        float32(user.Level) / 1000,
			}
			players = append(players, player)
		}
		eventTab.Players = players
	} else {
		eventTab.Players = make([]*model.Player, 0)
	}
	util.Logger.Printf("[EventPage] success, res:%+v", eventInfo)
	return eventTab, nil
}
