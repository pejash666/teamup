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
	userEvent := &mysql.UserEvent{}
	if err = util.DB().Where("open_id = ? AND event_id = ?", c.BasicUser.OpenID, body.EventID).Take(userEvent).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		util.Logger.Printf("[JoinEvent] find user_event record in DB failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
	}
	util.Logger.Printf("[JoinEvent] openID:%s has not joined eventID:%d yet, check success", c.BasicUser.OpenID, body.EventID)
	event := &mysql.EventMeta{}
	if err = util.DB().Where("id = ? AND status = ? AND end_time > ?", body.EventID, constant.EventStatusCreated, time.Now().Unix()).Take(event).Error; err != nil {
		util.Logger.Printf("[JoinEvent] get event from DB failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.InternalError, err.Error())
	}
	// 非公开活动只能通过邀请链接加入
	if event.IsPublic == 0 && !body.IsInviting {
		util.Logger.Printf("[JoinEvent] event:%d is private, can only join via invitinglink", body.EventID)
		return nil, iface.NewBackEndError(iface.PrivateEventError, "private event")
	}
	util.Logger.Printf("[JoinEvent] eventID:%d can be joined", body.EventID)
	// 开启事务
	util.DB().Transaction(func(tx *gorm.DB) error {
		// 增加一条user_event记录
		// 更新用户表（参与次数）
		// 更新事件表（参与人数）
		newUserEvent := &mysql.UserEvent{
			EventID:   body.EventID,
			SportType: event.SportType,
			OpenID:    c.BasicUser.OpenID,
		}
		if err = tx.Create(newUserEvent).Error; err != nil {
			util.Logger.Printf("[JoinEvent] Create user event failed, err:%v", err)
			return err
		}
		user := &mysql.WechatUserInfo{}
		if err = tx.Where("open_id = ? AND sport_type = ?", c.BasicUser.OpenID, event.SportType).Take(user).Error; err != nil {
			util.Logger.Printf("[JoinEvent] query user failed, err:%v", err)
			return err
		}
		user.JoinedTimes += 1
		joinedEvent := make([]uint, 0)
		err = sonic.UnmarshalString(user.JoinedEvent, &joinedEvent)
		if err != nil {
			return err
		}
		joinedEvent = append(joinedEvent, body.EventID)
		joinedEventStr, err := sonic.MarshalString(joinedEvent)
		if err != nil {
			return err
		}
		user.JoinedEvent = joinedEventStr
		if err = tx.Save(user).Error; err != nil {
			util.Logger.Printf("[JoinEvent] save user failed, err:%v", err)
			return err
		}

		meta := &mysql.EventMeta{}
		if err = tx.Where("id = ?", body.EventID).Take(meta).Error; err != nil {
			util.Logger.Printf("[JoinEvent] query user meta failed, err:%v", err)
			return err
		}
		meta.CurrentPlayerNum += 1
		if meta.CurrentPlayerNum == meta.MaxPlayerNum {
			util.Logger.Printf("[JoinEvent] event:%d is full now", body.EventID)
			meta.Status = constant.EventStatusFull
		}
		currentPeople := make([]string, 0)
		err = sonic.UnmarshalString(meta.CurrentPlayer, &currentPeople)
		if err != nil {
			return err
		}
		currentPeople = append(currentPeople, c.BasicUser.OpenID)
		currentPeopleStr, err := sonic.MarshalString(currentPeople)
		if err != nil {
			return err
		}
		meta.CurrentPlayer = currentPeopleStr
		if err = tx.Save(meta).Error; err != nil {
			util.Logger.Printf("[JoinEvent] save event meta failed, err:%v", err)
			return err
		}

		return nil
	})

	util.Logger.Printf("[JoinEvent] user:%s join event:%d success", c.BasicUser.OpenID, body.EventID)
	return nil, nil
}
