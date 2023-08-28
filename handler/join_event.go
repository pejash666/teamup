package handler

import (
	"errors"
	"gorm.io/gorm"
	"teamup/constant"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/util"
	"time"
)

type JoinEventBody struct {
	EventID uint `json:"event_id"` // 事件元信息ID
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
	if err = util.DB().Where("user_id = ? AND event_id = ?", c.BasicUser.UserID, body.EventID).Take(userEvent).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		util.Logger.Printf("[JoinEvent] find user_event record in DB failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
	}
	util.Logger.Printf("[JoinEvent] userID:%d has not joined eventID:%d yet, check success", c.BasicUser.UserID, body.EventID)
	event := &mysql.EventMeta{}
	if err = util.DB().Where("id = ? AND status = ? AND end_time > ?", body.EventID, constant.EventStatusCreated, time.Now().Unix()).Take(event).Error; err != nil {
		util.Logger.Printf("[JoinEvent] get event from DB failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.InternalError, err.Error())
	}
	util.Logger.Printf("[JoinEvent] eventID:%d can be joined", body.EventID)
	// 开启事务
	util.DB().Transaction(func(tx *gorm.DB) error {
		// 增加一条user_event记录
		// 更新用户表（参与次数）
		// 更新事件表（参与人数）
		newUserEvent := &mysql.UserEvent{
			EventID: body.EventID,
			UserID:  c.BasicUser.UserID,
		}
		if err = tx.Create(newUserEvent).Error; err != nil {
			return err
		}
		user := &mysql.WechatUserInfo{}
		if err = tx.Where("id = ?", c.BasicUser.UserID).Take(user).Error; err != nil {
			return err
		}
		user.JoinedTimes += 1
		if err = tx.Save(user).Error; err != nil {
			return err
		}

	})
}
