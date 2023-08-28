package handler

import (
	"teamup/constant"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/util"
	"time"
)

func CreateEvent(c *model.TeamUpContext) (interface{}, error) {
	util.Logger.Println("[CreateEvent] starts ")
	event := &model.EventInfo{}
	err := c.BindJSON(event)
	if err != nil {
		util.Logger.Printf("[CreateEvent] bindJson failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.InternalError, err.Error())
	}
	if !paramsCheck(event) {
		util.Logger.Printf("[CreateEvent] paramCheck failed")
		return nil, iface.NewBackEndError(iface.ParamsError, "invalid params")
	}

	meta := EventMeta(c, event)
	if err = util.DB().Create(meta).Error; err != nil {
		util.Logger.Printf("[CreateEvent] create record in DB failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
	}

	util.Logger.Printf("[CreateEvent] save record to DB success, eventID:%d", meta.ID)

	return map[string]uint{"event_id": meta.ID}, nil
}

func paramsCheck(event *model.EventInfo) bool {
	if event == nil {
		return false
	}
	if event.Title == "" || event.City == "" ||
		event.StartTime == 0 ||
		event.EndTime == 0 || event.FieldName == "" ||
		event.MaxPeople == 0 || event.SportType == "" ||
		event.Price == 0 {
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

	return true
}

func EventMeta(c *model.TeamUpContext, event *model.EventInfo) *mysql.EventMeta {
	meta := &mysql.EventMeta{
		Creator:       c.BasicUser.OpenID,
		SportType:     event.SportType,
		Date:          time.Unix(event.StartTime, 0).Format("20060102"),
		City:          event.City,
		Title:         event.Title,
		Desc:          event.Desc,
		StartTime:     event.StartTime,
		StartTimeStr:  time.Unix(event.StartTime, 0).Format("20060102 15:04"), // 分钟级别
		EndTime:       event.EndTime,
		EndTimeStr:    time.Unix(event.EndTime, 0).Format("20060102 15:04"),
		FieldName:     event.FieldName,
		MaxPeople:     event.MaxPeople,
		CurrentPeople: 0,
		Price:         event.Price,
	}

	if event.IsDraft {
		meta.Status = constant.EventStatusDraft
	} else {
		meta.Status = constant.EventStatusCreated
	}
	util.Logger.Printf("[CreateEvent] eventMeta:%+v", meta)
	return meta
}
