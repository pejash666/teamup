package handler

import (
	"github.com/bytedance/sonic"
	"teamup/constant"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/util"
	"time"
)

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
	// 如果自己也加入，则直接在创建时添加进去
	if event.SelfJoin {
		meta.CurrentPlayerNum += 1
		currentPeople := make([]string, 0)
		err = sonic.UnmarshalString(meta.CurrentPlayer, &currentPeople)
		if err != nil {
			util.Logger.Printf("UnmarshalString failed")
			return nil, err
		}
		currentPeople = append(currentPeople, c.BasicUser.OpenID)
		currentPeopleStr, err := sonic.MarshalString(currentPeople)
		if err != nil {
			util.Logger.Printf("MarshalString failed")
			return nil, err
		}
		meta.CurrentPlayer = currentPeopleStr
	}
	err = util.DB().Save(meta).Error
	if err != nil {
		util.Logger.Printf("[UpdateEvent] save updated event failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
	}

	util.Logger.Printf("[UpdateEvent] success, event_id:%d", event.Id)

	return map[string]int64{"event_id": event.Id}, nil
}
