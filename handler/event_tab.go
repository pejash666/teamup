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
	IsUserIn  bool             `json:"is_user_in"` // 表示当前用户是否已加入此活动
}

type EventPageBody struct {
	EventID int64 `json:"event_id"`
}

// EventPage godoc
//
//	@Summary		获取活动详情页
//	@Description	活动详情页，包含活动元信息和参与的用户信息
//	@Tags			/team_up/event
//	@Accept			json
//	@Produce		json
//	@Param			event_id	body int	true	"活动ID"
//	@Success		200				{object}	EventTab
//	@Router			/team_up/event/page [post]
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
	// event都会有organization_id
	organization := &mysql.Organization{}
	err = util.DB().Where("id = ?", eventMeta.OrganizationID).Take(organization).Error
	if err != nil {
		util.Logger.Printf("[EventPage] get organization record from DB failed, err:%v", result.Error)
		return nil, iface.NewBackEndError(iface.MysqlError, "get record failed")
	}
	eventInfo.OrganizationID = int64(organization.ID)
	eventInfo.OrganizationLogo = organization.Logo
	eventInfo.OrganizationAddress = organization.Address
	//if eventInfo.IsHost {
	//	organization := &mysql.Organization{}
	//	result := util.DB().Where("id = ?", eventMeta.OrganizationID).Take(organization)
	//	if result.Error != nil {
	//		util.Logger.Printf("[EventPage] get organization record from DB failed, err:%v", result.Error)
	//		return nil, iface.NewBackEndError(iface.MysqlError, "get record failed")
	//	}
	//	eventInfo.OrganizationID = int64(organization.ID)
	//	eventInfo.OrganizationLogo = organization.Logo
	//	eventInfo.OrganizationAddress = organization.Address
	//}
	eventInfo.Id = int64(eventMeta.ID)
	eventInfo.Desc = eventMeta.Desc
	eventInfo.Name = eventMeta.Name
	eventInfo.SportType = eventMeta.SportType
	eventInfo.Status = eventMeta.Status
	eventInfo.Weekday = eventMeta.Weekday
	eventInfo.City = eventMeta.City
	eventInfo.Longitude = eventMeta.Longitude
	eventInfo.Latitude = eventMeta.Latitude
	eventInfo.FieldName = eventMeta.FieldName
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
	eventInfo.CurrentPeople = eventMeta.CurrentPlayerNum
	eventInfo.EventImage = eventMeta.EventImage
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
		err = util.DB().Where("open_id IN ? AND sport_type = ?", currentPeople, eventMeta.SportType).Find(&users).Error
		if err != nil {
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
				OpenID:       user.OpenId,
			}
			players = append(players, player)
			// 当前用户已经加入了这个活动的标识，前端用来展示退出button
			if player.OpenID == c.BasicUser.OpenID {
				eventTab.IsUserIn = true
			}
		}
		eventTab.Players = players
	} else {
		eventTab.Players = make([]*model.Player, 0)
	}

	util.Logger.Printf("[EventPage] success, res:%+v", *eventTab)
	return eventTab, nil
}
