package handler

import (
	"fmt"
	"github.com/bytedance/sonic"
	"gorm.io/gorm"
	"teamup/constant"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/util"
	"time"
)

// UpdateEvent godoc
//
//	@Summary		更新活动元信息
//	@Description	个人或者组织更新活动元信息
//	@Tags			/team_up/event
//	@Accept			json
//	@Produce		json
//	@Param			update_event	body		string	true	"更新活动入参,参考EventInfo"
//	@Success		200				{object}	CreateEventResp
//	@Router			/team_up/event/update [post]
func UpdateEvent(c *model.TeamUpContext) (interface{}, error) {
	util.Logger.Printf("[UpdateEvent] starts")
	// 从DB获取event
	event := &model.EventInfo{}
	err := c.BindJSON(event)
	if err != nil {
		util.Logger.Printf("[UpdateEvent] bindJson failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.InternalError, err.Error())
	}
	meta := &mysql.EventMeta{}
	err = util.DB().Where("id = ?", event.Id).Take(meta).Error
	if err != nil {
		util.Logger.Printf("[UpdateEvent] find record in DB failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
	}
	// 只有草稿状态下的事件才允许被编辑
	if meta.Status != constant.EventStatusDraft {
		util.Logger.Printf("[UpdateEvent] record is not in draft status, can't edit")
		return nil, iface.NewBackEndError(iface.ParamsError, "cant edit")
	}
	meta.Price = event.Price
	meta.MaxPlayerNum = event.MaxPeopleNum
	meta.FieldName = event.FieldName
	meta.FieldType = event.FieldType
	meta.SportType = event.SportType
	meta.GameType = event.GameType
	meta.IsPublic = util.BoolToDB(event.IsPublic)
	meta.IsBooked = util.BoolToDB(event.IsBooked)
	meta.SportType = event.SportType
	meta.Status = event.Status
	meta.Date = time.Unix(event.StartTime, 0).Format("20060102")
	meta.Weekday = time.Unix(event.StartTime, 0).Weekday().String()
	meta.StartTime = event.StartTime
	meta.StartTimeStr = time.Unix(event.StartTime, 0).Format("15:04")
	meta.EndTime = event.EndTime
	meta.EndTimeStr = time.Unix(event.EndTime, 0).Format("15:04")
	meta.Desc = event.Desc
	meta.Name = event.Name
	meta.City = event.City
	if event.IsCompetitive {
		meta.MatchType = constant.EventMatchTypeCompetitive
	} else {
		meta.MatchType = constant.EventMatchTypeEntertainment
	}
	if event.IsDraft {
		meta.Status = constant.EventStatusDraft
	} else {
		meta.Status = constant.EventStatusCreated
	}
	currentPlayers := make([]string, 0)
	err = sonic.UnmarshalString(meta.CurrentPlayer, &currentPlayers)
	if err != nil {
		util.Logger.Printf("UnmarshalString failed")
		return nil, err
	}
	selfIn := false
	for _, player := range currentPlayers {
		if player == c.BasicUser.OpenID {
			selfIn = true
			break
		}
	}
	// 当前数据库有，但是要改成自己不加入
	if selfIn && !event.SelfJoin {
		util.DB().Transaction(func(tx *gorm.DB) error {
			meta.CurrentPlayerNum -= 1
			newPlayers := make([]string, 0)
			for _, player := range currentPlayers {
				if player == c.BasicUser.OpenID {
					continue
				}
				newPlayers = append(newPlayers, player)
			}
			newPlayerStr, err := sonic.MarshalString(newPlayers)
			if err != nil {
				return err
			}
			meta.CurrentPlayer = newPlayerStr
			// 更改事件
			err = tx.Save(meta).Error
			if err != nil {
				return err
			}
			// 更改用户
			user := &mysql.WechatUserInfo{}
			err = tx.Where("open_id = ? AND sport_type = ?", c.BasicUser.OpenID, meta.SportType).Take(user).Error
			if err != nil {
				return err
			}
			joinedEvents := make([]uint, 0)
			err = sonic.UnmarshalString(user.JoinedEvent, &joinedEvents)
			if err != nil {
				return err
			}
			newJoinedEvent := make([]uint, 0)
			for _, joinedEvent := range joinedEvents {
				if joinedEvent == meta.ID {
					continue
				}
				newJoinedEvent = append(newJoinedEvent, joinedEvent)
			}
			newJoinedEventStr, err := sonic.MarshalString(newJoinedEvent)
			if err != nil {
				return err
			}
			user.JoinedEvent = newJoinedEventStr
			user.JoinedTimes -= 1
			err = tx.Save(user).Error
			if err != nil {
				return err
			}
			return nil
		})

		// 当前数据库没有，改成自己加入
	} else if !selfIn && event.SelfJoin {
		util.DB().Transaction(func(tx *gorm.DB) error {
			meta.CurrentPlayerNum += 1
			currentPlayers = append(currentPlayers, c.BasicUser.OpenID)
			currentPeopleStr, err := sonic.MarshalString(currentPlayers)
			if err != nil {
				util.Logger.Printf("MarshalString failed")
				return err
			}
			meta.CurrentPlayer = currentPeopleStr
			// 更改事件
			err = tx.Save(meta).Error
			if err != nil {
				return err
			}
			// 更改用户
			user := &mysql.WechatUserInfo{}
			err = tx.Where("open_id = ? AND sport_type = ?", c.BasicUser.OpenID, meta.SportType).Take(user).Error
			if err != nil {
				return err
			}
			joinedEvents := make([]uint, 0)
			err = sonic.UnmarshalString(user.JoinedEvent, &joinedEvents)
			if err != nil {
				return err
			}
			joinedEvents = append(joinedEvents, meta.ID)
			JoinedEventStr, err := sonic.MarshalString(joinedEvents)
			if err != nil {
				return err
			}
			user.JoinedEvent = JoinedEventStr
			user.JoinedTimes += 1
			err = tx.Save(user).Error
			if err != nil {
				return err
			}
			return nil
		})
	}
	fmt.Println(meta.IsHost)
	fmt.Println(event.IsHost)
	// 之前是组织创建的活动，现在改成个人创建
	if meta.IsHost == 1 && !event.IsHost {
		meta.IsHost = 0
		meta.OrganizationID = 0
		err = util.DB().Save(meta).Error
		if err != nil {
			return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
		}
	} else if meta.IsHost == 0 && event.IsHost {
		fmt.Println("here")
		// 检查此用户是不是此运动类型的host
		user := &mysql.WechatUserInfo{}
		err = util.DB().Where("open_id = ? AND sport_type = ?", c.BasicUser.OpenID, event.SportType).Take(user).Error
		if err != nil {
			return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
		}
		if user.IsHost == 0 {
			util.Logger.Printf("[CreateEvent] user not host")
			return nil, iface.NewBackEndError(iface.NotHostError, "not host")
		}
		meta.IsHost = 1
		meta.OrganizationID = int64(user.OrganizationID)
		err = util.DB().Save(meta).Error
		if err != nil {
			return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
		}
	}
	util.Logger.Printf("[UpdateEvent] success, event_id:%d", event.Id)

	return &CreateEventID{EventID: event.Id}, nil
}
