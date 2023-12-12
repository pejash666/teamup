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

func CreateEvent(c *model.TeamUpContext) (interface{}, error) {
	util.Logger.Println("[CreateEvent] starts")
	event := &model.EventInfo{}
	err := c.BindJSON(event)
	if err != nil {
		util.Logger.Printf("[CreateEvent] bindJson failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.InternalError, err.Error())
	}
	util.Logger.Printf("[CreateEvent] req:%+v", event)
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
		event.OrganizationID = int64(user.OrganizationID)
	}
	if !paramsCheck(event) {
		util.Logger.Printf("[CreateEvent] paramCheck failed")
		return nil, iface.NewBackEndError(iface.ParamsError, "invalid params")
	}
	meta, err := EventMeta(c, event)
	if err != nil {
		util.Logger.Printf("[CreateEvent] EventMeta failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.InternalError, "EventMeta failed")
	}

	// todo: 开启DB事务, 要更新organization表对应组织的活动次数 和 用户自身参与次数
	util.DB().Transaction(func(tx *gorm.DB) error {
		// 创建活动
		err = util.DB().Create(meta).Error
		if err != nil {
			util.Logger.Printf("[CreateEvent] DB Create failed, err:%v", err)
			return err
		}
		// 如果是以场馆身份创建的，则需要给场馆活动次数+1
		if event.IsHost {
			organization := &mysql.Organization{}
			err = util.DB().Where("id = ?", meta.OrganizationID).Take(organization).Error
			if err != nil {
				util.Logger.Printf("[CreateEvent] query organization failed, err:%v", err)
				return err
			}
			organization.TotalEventNum += 1
			err = util.DB().Save(organization).Error
			if err != nil {
				return err
			}
		}
		// 如果自己也加入，则需要给user参与次数，参与活动ID进行变更
		if event.SelfJoin {
			user := &mysql.WechatUserInfo{}
			err = util.DB().Where("open_id = ? AND sport_type = ?", c.BasicUser.OpenID, event.SportType).Take(user).Error
			if err != nil {
				util.Logger.Printf("[CreateEvent] query user failed, err:%v", err)
				return err
			}
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
			err = util.DB().Save(user).Error
			if err != nil {
				return err
			}
		}
		util.Logger.Printf("[CreateEvent] success")
		return nil
	})

	util.Logger.Printf("[CreateEvent] save record to DB success, eventID:%d", meta.ID)

	return map[string]uint{"event_id": meta.ID}, nil
}

func paramsCheck(event *model.EventInfo) bool {
	if event == nil {
		return false
	}
	if event.Name == "" || event.City == "" ||
		event.StartTime == 0 ||
		event.EndTime == 0 ||
		event.MaxPeople == 0 || event.SportType == "" ||
		event.Price == 0 || event.GameType == "" {
		return false
	}
	// 判断时间是否符合预期
	timeNow := time.Now()
	startTime := time.Unix(event.StartTime, 0)
	endTime := time.Unix(event.EndTime, 0)
	if timeNow.After(endTime) {
		return false
	}
	if startTime.After(endTime) {
		return false
	}
	// 已经预定场地的需要检查场地名称和场地类型
	if event.IsBooked && (event.FieldName == "" || event.FieldType == "") {
		return false
	}
	return true
}

func EventMeta(c *model.TeamUpContext, event *model.EventInfo) (*mysql.EventMeta, error) {
	meta := &mysql.EventMeta{
		Creator:        c.BasicUser.OpenID,
		SportType:      event.SportType,
		GameType:       event.GameType,
		IsBooked:       util.BoolToDB(event.IsBooked),
		IsPublic:       util.BoolToDB(event.IsPublic),
		IsHost:         util.BoolToDB(event.IsHost),
		LowestLevel:    event.LowestLevel,
		HighestLevel:   event.HighestLevel,
		Date:           time.Unix(event.StartTime, 0).Format("20060102"),
		City:           event.City,
		Name:           event.Name,
		Desc:           event.Desc,
		StartTime:      event.StartTime,
		StartTimeStr:   time.Unix(event.StartTime, 0).Format("20060102 15:04"), // 分钟级别
		EndTime:        event.EndTime,
		EndTimeStr:     time.Unix(event.EndTime, 0).Format("20060102 15:04"),
		FieldName:      event.FieldName,
		FieldType:      event.FieldType,
		MaxPlayerNum:   event.MaxPeople,
		Price:          event.Price,
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
	// 如果自己也加入，则直接在创建时添加进去（只有个人创建的才可以）
	if !event.IsHost && event.SelfJoin {
		meta.CurrentPlayerNum = 1
		currentPeople := make([]string, 0)
		currentPeople = append(currentPeople, c.BasicUser.OpenID)
		currentPeopleStr, err := sonic.MarshalString(currentPeople)
		if err != nil {
			util.Logger.Printf("[EventMeta] self join failed")
			return nil, err
		}
		meta.CurrentPlayer = currentPeopleStr
	}
	util.Logger.Printf("[CreateEvent] eventMeta:%+v", meta)
	return meta, nil
}
