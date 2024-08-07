package handler

import (
	"errors"
	"github.com/bytedance/sonic"
	"gorm.io/gorm"
	"teamup/constant"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/util"
	"time"
)

type JoinEventBody struct {
	EventID    uint `json:"event_id"`    // 事件元信息ID
	IsInviting bool `json:"is_inviting"` // 是否通过邀请链接加入
}

// JoinEvent godoc
//
//	@Summary		加入活动场次
//	@Description	加入活动场次
//	@Tags			/team_up/user
//	@Accept			json
//	@Produce		json
//	@Param			event_id	body		int		true	"加入活动场次入参"
//	@Param			is_inviting	body		bool	true	"是否通过邀请链接加入"
//	@Success		200			{object}	model.BackEndResp
//	@Router			/team_up/user/join_event [post]
func JoinEvent(c *model.TeamUpContext) (interface{}, error) {
	util.Logger.Printf("[JoinEvent] starts")
	body := &JoinEventBody{}
	err := c.BindJSON(body)
	if err != nil {
		util.Logger.Printf("[JoinEvent] BindJSON failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.ParamsError, err.Error())
	}
	if body.EventID < 1 {
		util.Logger.Printf("[JoinEvent] invalid event_id:%d", body.EventID)
		return nil, iface.NewBackEndError(iface.ParamsError, "invalid event_id")
	}
	// 先检查下是否已经参与了这个活动 && 活动是否可以参加
	haveUserEvent := false
	userEvent := &mysql.UserEvent{}
	err = util.DB().Where("open_id = ? AND event_id = ?", c.BasicUser.OpenID, body.EventID).Take(userEvent).Error
	if err != nil {
		// 有记录
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			util.Logger.Printf("[JoinEvent] find user_event record in DB failed, err:%v", err)
			return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
		}
	}
	eventMeta := &mysql.EventMeta{}
	err = util.DB().Where("id = ?", body.EventID).Take(eventMeta).Error
	if err != nil {
		return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
	}
	// 看下用户是否已经参与了
	joinedPlayers := make([]string, 0)
	err = sonic.UnmarshalString(eventMeta.CurrentPlayer, &joinedPlayers)
	if err != nil {
		return nil, iface.NewBackEndError(iface.InternalError, err.Error())
	}
	for _, player := range joinedPlayers {
		if player == c.BasicUser.OpenID {
			return nil, iface.NewBackEndError(iface.ParamsError, "user already joined this event")
		}
	}
	if userEvent.IsQuit == 1 {
		haveUserEvent = true
	}
	util.Logger.Printf("[JoinEvent] openID:%s has not joined eventID:%d yet, check success", c.BasicUser.OpenID, body.EventID)
	event := &mysql.EventMeta{}
	if err = util.DB().Where("id = ? AND status = ? AND end_time > ?", body.EventID, constant.EventStatusCreated, time.Now().Unix()).Take(event).Error; err != nil {
		util.Logger.Printf("[JoinEvent] get event from DB failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
	}
	// 检查用户自己的level是否符合局的要求
	user := &mysql.WechatUserInfo{}
	err = util.DB().Where("open_id = ? AND sport_type = ?", c.BasicUser.OpenID, event.SportType).Take(user).Error
	if err != nil {
		return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
	}
	// 没有calibrated的用户不允许加入竞技类的比赛
	if user.IsCalibrated == 0 && event.MatchType == "competitive" {
		return nil, iface.NewBackEndError(iface.ParamsError, "user is not calibrated, cant join competitive game")
	}
	if user.Level < event.LowestLevel || user.Level > event.HighestLevel {
		return nil, iface.NewBackEndError(iface.ParamsError, "level is not in valid range of the event")
	}
	// 非公开活动只能通过邀请链接加入
	//// todo: 需要测试
	//if event.IsPublic == 0 && !body.IsInviting {
	//	util.Logger.Printf("[JoinEvent] event:%d is private, can only join via invitinglink", body.EventID)
	//	return nil, iface.NewBackEndError(iface.PrivateEventError, "private event")
	//}
	util.Logger.Printf("[JoinEvent] eventID:%d can be joined", body.EventID)
	// 开启事务
	err = util.DB().Transaction(func(tx *gorm.DB) error {
		// 增加一条user_event记录
		// 更新用户表（参与次数）
		// 更新事件表（参与人数）
		if haveUserEvent {
			userEvent.IsQuit = 0
			if err = tx.Save(userEvent).Error; err != nil {
				util.Logger.Printf("[JoinEvent] Create user event failed, err:%v", err)
				return err
			}
		} else {
			newUserEvent := &mysql.UserEvent{
				EventID:   body.EventID,
				SportType: event.SportType,
				OpenID:    c.BasicUser.OpenID,
				IsQuit:    0,
			}
			if err = tx.Save(newUserEvent).Error; err != nil {
				util.Logger.Printf("[JoinEvent] Create user event failed, err:%v", err)
				return err
			}
		}
		//wechatUser := &mysql.WechatUserInfo{}
		//if err = tx.Where("open_id = ? AND sport_type = ?", c.BasicUser.OpenID, event.SportType).Take(user).Error; err != nil {
		//	util.Logger.Printf("[JoinEvent] query user failed, err:%v", err)
		//	return err
		//}
		user.JoinedTimes += 1
		joinedEvent := make([]uint, 0)
		err = sonic.UnmarshalString(user.JoinedEvent, &joinedEvent)
		if err != nil {
			return err
		}
		joinedEvent = append(joinedEvent, body.EventID)
		joinedEventStr, Terr := sonic.MarshalString(joinedEvent)
		if Terr != nil {
			return Terr
		}
		user.JoinedEvent = joinedEventStr
		if err = tx.Save(user).Error; err != nil {
			util.Logger.Printf("[JoinEvent] save user failed, err:%v", err)
			return err
		}

		//meta := &mysql.EventMeta{}
		//if err = tx.Where("id = ?", body.EventID).Take(meta).Error; err != nil {
		//	util.Logger.Printf("[JoinEvent] query user meta failed, err:%v", err)
		//	return err
		//}
		event.CurrentPlayerNum += 1
		if event.CurrentPlayerNum == event.MaxPlayerNum {
			util.Logger.Printf("[JoinEvent] event:%d is full now", body.EventID)
			event.Status = constant.EventStatusFull
		}
		currentPeople := make([]string, 0)
		err = sonic.UnmarshalString(event.CurrentPlayer, &currentPeople)
		if err != nil {
			return err
		}
		currentPeople = append(currentPeople, c.BasicUser.OpenID)
		currentPeopleStr, Terr := sonic.MarshalString(currentPeople)
		if Terr != nil {
			return Terr
		}
		event.CurrentPlayer = currentPeopleStr
		if err = tx.Save(event).Error; err != nil {
			util.Logger.Printf("[JoinEvent] save event meta failed, err:%v", err)
			return err
		}

		return nil
	})
	if err != nil {
		return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
	}

	util.Logger.Printf("[JoinEvent] user:%s join event:%d success", c.BasicUser.OpenID, body.EventID)
	return nil, nil
}
