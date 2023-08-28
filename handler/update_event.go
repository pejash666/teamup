package handler

import (
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
	event := &mysql.EventMeta{}
	err := c.BindJSON(event)
	if err != nil {
		util.Logger.Printf("[UpdateEvent] bindJson failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.InternalError, err.Error())
	}
	meta := &mysql.EventMeta{}
	err = util.DB().Where("id = ?", event.ID).Take(meta).Error
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
	meta.MaxPeople = event.MaxPeople
	meta.FieldName = event.FieldName
	meta.SportType = event.SportType
	meta.Status = event.Status
	meta.Date = time.Unix(event.StartTime, 0).Format("20060102")
	meta.StartTime = event.StartTime
	meta.StartTimeStr = time.Unix(event.StartTime, 0).Format("20060102 15:04")
	meta.EndTime = event.EndTime
	meta.EndTimeStr = time.Unix(event.EndTime, 0).Format("20060102 15:04")
	meta.Desc = event.Desc
	meta.Title = event.Title
	meta.City = event.City

	err = util.DB().Save(meta).Error
	if err != nil {
		util.Logger.Printf("[UpdateEvent] save updated event failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
	}

	util.Logger.Printf("[UpdateEvent] success, event_id:%d", event.ID)

	return map[string]uint{"event_id": event.ID}, nil
}
