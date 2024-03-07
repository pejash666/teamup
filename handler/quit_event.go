package handler

import (
	"github.com/bytedance/sonic"
	"gorm.io/gorm"
	"teamup/constant"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/util"
	"time"
)

const (
	ThreeHoursSeconds = 10800
)

type QuitEventBody struct {
	EventID int64 `json:"event_id"`
}

// QuitEvent godoc
//
//	@Summary		退出活动场次
//	@Description	退出活动场次
//	@Tags			/teamup/user
//	@Accept			json
//	@Produce		json
//	@Param			event_id	body		int	true	"活动ID"
//	@Success		200			{object}	model.BackEndResp
//	@Router			/teamup/user/quit_event [post]
func QuitEvent(c *model.TeamUpContext) (interface{}, error) {
	// 获取当前活动信息
	body := &QuitEventBody{}
	err := c.BindJSON(body)
	if err != nil {
		util.Logger.Printf("[QuitEvent] bindJSON failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.ParamsError, "invalid param")
	}
	event := &mysql.EventMeta{}
	// 只有被创建 和 full状态的活动才可以退出
	err = util.DB().Where("id = ? AND status IN ?", body.EventID, []string{constant.EventStatusFull, constant.EventStatusCreated}).Take(event).Error
	if err != nil {
		util.Logger.Printf("[[QuitEvent] query eventMeta failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, "query failed")
	}
	// 判断当前时间距离开始时间，只有大于三小时才能退出
	timeNow := time.Now().Unix()
	eventStartTime := event.StartTime
	if timeNow >= eventStartTime || (eventStartTime-timeNow) < ThreeHoursSeconds {
		util.Logger.Printf("[QuitEvent] time is not invalid, eventStartTime:%v, timeNow:%v", eventStartTime, timeNow)
		return nil, iface.NewBackEndError(iface.ParamsError, "invalid quit time")
	}
	// 开启事务
	util.DB().Transaction(func(tx *gorm.DB) error {
		// 更新event表
		event.CurrentPlayerNum -= 1
		currentPlayer := make([]string, 0)
		err = sonic.UnmarshalString(event.CurrentPlayer, &currentPlayer)
		if err != nil {
			util.Logger.Printf("[QuitEvent] unmarshalString failed, err:%v", err)
			return err
		}
		playerAfterQuit := make([]string, 0)
		for _, player := range currentPlayer {
			// 新的名单去掉当前要退出的用户
			if player != c.BasicUser.OpenID {
				playerAfterQuit = append(playerAfterQuit, player)
			}
		}
		playerAfterQuitStr, err := sonic.MarshalString(playerAfterQuit)
		if err != nil {
			util.Logger.Printf("[QuitEvent] marshalString failed, err:%v", err)
			return err
		}
		event.CurrentPlayer = playerAfterQuitStr
		// 如果退出后，人不满了，需要修改状态
		if event.CurrentPlayerNum < event.MaxPlayerNum {
			event.Status = constant.EventStatusCreated
		}
		err = tx.Save(event).Error
		if err != nil {
			util.Logger.Printf("[QuitEvent] Save failed, err:%v", err)
			return err
		}
		// 更新用户信息
		user := &mysql.WechatUserInfo{}
		err = tx.Where("open_id = ? AND sport_type = ?", c.BasicUser.OpenID, event.SportType).Take(user).Error
		if err != nil {
			util.Logger.Printf("[QuitEvent] query user failed, err:%v", err)
			return err
		}
		user.JoinedTimes -= 1
		joinedEvent := make([]uint, 0)
		err = sonic.UnmarshalString(user.JoinedEvent, &joinedEvent)
		if err != nil {
			util.Logger.Printf("[QuitEvent] unmarshalString failed, err:%v", err)
			return err
		}
		newJoinedEvent := make([]uint, 0)
		for _, joined := range joinedEvent {
			// 参与的活动要去掉这次
			if joined != event.ID {
				newJoinedEvent = append(newJoinedEvent, joined)
			}
		}
		newJoinedEventStr, err := sonic.MarshalString(newJoinedEvent)
		if err != nil {
			util.Logger.Printf("[QuitEvent] marshalString failed, err:%v", err)
			return err
		}
		user.JoinedEvent = newJoinedEventStr
		err = tx.Save(user).Error
		if err != nil {
			util.Logger.Printf("[QuitEvent] Save failed, err:%v", err)
			return err
		}
		// 更新user_event表
		userEvent := &mysql.UserEvent{}
		err = tx.Where("open_id = ? AND sport_type = ? AND event_id = ?", c.BasicUser.OpenID, event.SportType, event.ID).Take(userEvent).Error
		if err != nil {
			util.Logger.Printf("[QuitEvent] get record failed, err:%v", err)
			return err
		}
		userEvent.IsQuit = 1
		err = tx.Save(userEvent).Error
		if err != nil {
			util.Logger.Printf("[QuitEvent] save user_event failed, err:%v", err)
			return err
		}
		util.Logger.Printf("[QuitEvent] transaction success")
		return nil
	})

	return nil, nil
}
