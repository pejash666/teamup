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

type CreateEventResp struct {
	ErrNo   int32         `json:"err_no"`
	ErrTips string        `json:"err_tips"`
	Data    CreateEventID `json:"data"`
}

type CreateEventID struct {
	EventID int64 `json:"event_id"`
}

// CreateEvent godoc
//
//	@Summary		创建活动
//	@Description	个人或者组织创建活动
//	@Tags			/team_up/event
//	@Accept			json
//	@Produce		json
//	@Param			create_event	body		string	true	"参考EventInfo Model"
//	@Success		200				{object}	CreateEventResp
//	@Router			/team_up/event/create [post]
func CreateEvent(c *model.TeamUpContext) (interface{}, error) {
	util.Logger.Println("[CreateEvent] starts")
	event := &model.EventInfo{}
	err := c.BindJSON(event)
	if err != nil {
		util.Logger.Printf("[CreateEvent] bindJson failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.InternalError, err.Error())
	}
	util.Logger.Printf("[CreateEvent] req:%+v", event)

	isPass, reason := paramsCheck(event)
	if !isPass {
		util.Logger.Printf("[CreateEvent] paramCheck failed")
		return nil, iface.NewBackEndError(iface.ParamsError, reason)
	}
	// 通过organization_id获取组织信息
	// 查询orga表，获取经纬度和fieldname
	orga := &mysql.Organization{}
	err = util.DB().Where("id = ?", event.OrganizationID).Take(orga).Error
	if err != nil {
		return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
	}
	// 无论个人还是组织创建的活动，都需要传organization_id,所以场地相关的信息都是从组织信息中获取
	event.Longitude = orga.Longitude
	event.Latitude = orga.Latitude
	event.FieldType = "outdoor" // 默认outdoor
	event.FieldName = orga.Name // event的场地名字 就是组织名字
	// 如果是host创建的，需要额外check这个用户在这个sport_type下是不是host
	if event.IsHost {
		user := &mysql.WechatUserInfo{}
		err = util.DB().Where("open_id = ? AND sport_type = ?", c.BasicUser.OpenID, event.SportType).Take(user).Error
		if err != nil {
			return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
		}
		if user.IsHost == 0 {
			util.Logger.Printf("[CreateEvent] user not host")
			return nil, iface.NewBackEndError(iface.NotHostError, "not host")
		}
		// 如果是以场馆身份创建的，则需要检查organization_id是不是当前用户的organization_id
		if event.OrganizationID != int64(user.OrganizationID) {
			util.Logger.Printf("[CreateEvent] organization id not match")
			return nil, iface.NewBackEndError(iface.ParamsError, "organization id not match")
		}
		event.IsBooked = true // 场地创建的活动标记为已订场地
	}

	meta, err := EventMeta(c, event)
	if err != nil {
		util.Logger.Printf("[CreateEvent] EventMeta failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.InternalError, "EventMeta failed")
	}

	// todo: 开启DB事务, 要更新organization表对应组织的活动次数 和 用户自身参与次数
	err = util.DB().Transaction(func(tx *gorm.DB) error {
		// 创建活动
		err = tx.Create(meta).Error
		if err != nil {
			util.Logger.Printf("[CreateEvent] DB Create failed, err:%v", err)
			return err
		}
		// 如果是以场馆身份创建的，则需要给场馆活动次数+1
		if event.IsHost {
			organization := &mysql.Organization{}
			err = tx.Where("id = ?", meta.OrganizationID).Take(organization).Error
			if err != nil {
				util.Logger.Printf("[CreateEvent] query organization failed, err:%v", err)
				return err
			}
			organization.TotalEventNum += 1
			err = tx.Save(organization).Error
			if err != nil {
				return err
			}
		}
		// 如果自己也加入，则需要给user参与次数，参与活动ID进行变更
		if event.SelfJoin {
			util.Logger.Printf("[CreateEvent] self join detected")
			userEvent := &mysql.UserEvent{
				EventID:   meta.ID,
				SportType: event.SportType,
				OpenID:    c.BasicUser.OpenID,
				IsQuit:    0,
			}
			if err = tx.Save(userEvent).Error; err != nil {
				util.Logger.Printf("[JoinEvent] Create user event failed, err:%v", err)
				return err
			}
			user := &mysql.WechatUserInfo{}
			err = tx.Where("open_id = ? AND sport_type = ?", c.BasicUser.OpenID, event.SportType).Take(user).Error
			if err != nil {
				util.Logger.Printf("[CreateEvent] query user failed, err:%v", err)
				return err
			}
			util.Logger.Printf("[CreateEvent] user:%+v", user)
			user.JoinedTimes += 1
			joinedEvent := make([]uint, 0)
			err = sonic.UnmarshalString(user.JoinedEvent, &joinedEvent)
			if err != nil {
				return err
			}
			joinedEvent = append(joinedEvent, meta.ID)
			joinedEventStr, err := sonic.MarshalString(joinedEvent)
			if err != nil {
				return err
			}
			user.JoinedEvent = joinedEventStr
			if err := tx.Save(user).Error; err != nil {
				util.Logger.Printf("[CreateEvent] save failed, err:%v", err.Error())
				return err
			}
		}
		util.Logger.Printf("[CreateEvent] success")
		return nil
	})

	if err != nil {
		util.Logger.Printf("[CreateEvent] save record to DB failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
	} else {
		util.Logger.Printf("[CreateEvent] save record to DB success, eventID:%d", meta.ID)
	}
	return &CreateEventID{EventID: int64(meta.ID)}, nil
}

func paramsCheck(event *model.EventInfo) (bool, string) {
	if event == nil {
		return false, "empty params"
	}
	if event.Name == "" || event.City == "" ||
		event.StartTime == 0 ||
		event.EndTime == 0 ||
		event.MaxPeopleNum == 0 ||
		event.SportType == "" ||
		event.GameType == "" || event.OrganizationID < 1 {
		return false, "invalid param"
	}
	// 判断时间是否符合预期
	timeNow := time.Now()
	startTime := time.Unix(event.StartTime, 0)
	endTime := time.Unix(event.EndTime, 0)
	if timeNow.After(endTime) {
		return false, "invalid time"
	}
	if startTime.After(endTime) {
		return false, "invalid time"
	}
	// 个人已经预定场地的需要检查场地名称和场地类型
	//if !event.IsHost && event.IsBooked && (event.FieldName == "" || event.FieldType == "") && (event.Longitude == "" || event.Latitude == "") {
	//	return false, "invalid field"
	//}
	if event.IsBooked && event.Price == 0 {
		return false, "booked filed must have price"
	}
	// padel只允许双打 人数 >= 4 <= 8
	if event.SportType == constant.SportTypePadel {
		if event.GameType == constant.EventGameTypeSolo {
			return false, "padel only can duo"
		}
		if event.MaxPeopleNum < 4 || event.MaxPeopleNum > 8 {
			return false, "invalid max_people_number"
		}
	}
	// 匹克球的人数必须为偶数 且小于8人
	if event.SportType == constant.SportTypePickelBall {
		if event.MaxPeopleNum > 8 || event.MaxPeopleNum%2 != 0 {
			return false, "invalid max_people_number"
		}
	}
	// 只有组织创建的才能上传eventImage
	if !event.IsHost && event.EventImage != "" {
		return false, "only host can upload event image"
	}
	return true, ""
}

func EventMeta(c *model.TeamUpContext, event *model.EventInfo) (*mysql.EventMeta, error) {
	meta := &mysql.EventMeta{
		Creator:        c.BasicUser.OpenID,
		SportType:      event.SportType,
		GameType:       event.GameType,
		IsBooked:       util.BoolToDB(event.IsBooked),
		IsPublic:       util.BoolToDB(event.IsPublic),
		IsHost:         util.BoolToDB(event.IsHost),
		LowestLevel:    int(event.LowestLevel * 1000),
		HighestLevel:   int(event.HighestLevel * 1000),
		Date:           time.Unix(event.StartTime, 0).Format("2006-01-02"),
		Weekday:        time.Unix(event.StartTime, 0).Weekday().String(),
		City:           event.City,
		Name:           event.Name,
		Desc:           event.Desc,
		StartTime:      event.StartTime,
		StartTimeStr:   time.Unix(event.StartTime, 0).Format("15:04"), // 分钟级别
		EndTime:        event.EndTime,
		EndTimeStr:     time.Unix(event.EndTime, 0).Format("15:04"),
		FieldName:      event.FieldName,
		FieldType:      event.FieldType,
		MaxPlayerNum:   event.MaxPeopleNum,
		Price:          event.Price,
		Latitude:       event.Latitude,
		Longitude:      event.Longitude,
		OrganizationID: event.OrganizationID,
	}

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
	// 如果自己也加入，则直接在创建时添加进去
	currentPeople := make([]string, 0)
	if event.SelfJoin {
		meta.CurrentPlayerNum = 1
		currentPeople = append(currentPeople, c.BasicUser.OpenID)
	}
	currentPeopleStr, err := sonic.MarshalString(currentPeople)
	if err != nil {
		util.Logger.Printf("[EventMeta] self join failed")
		return nil, err
	}
	meta.CurrentPlayer = currentPeopleStr

	if event.IsHost {
		meta.EventImage = event.EventImage
	}
	util.Logger.Printf("[CreateEvent] eventMeta:%+v", meta)
	return meta, nil
}
