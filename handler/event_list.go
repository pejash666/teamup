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
)

type GetEventListBody struct {
	SportType string `json:"sport_type"`
	City      string `json:"city"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
	Num       int    `json:"num"`
	Offset    int    `json:"offset"`
}

type GetEventListResp struct {
	ErrNo   int32             `json:"err_no"`
	ErrTips string            `json:"err_tips"`
	Data    *GetEventListData `json:"data"`
}

type GetEventListData struct {
	EventList []*Event `json:"event_list"`
	HasMore   bool     `json:"has_more"`
	Offset    int      `json:"offset"`
}

// EventList godoc
//
//	@Summary		获取活动列表
//	@Description	根据筛选条件获取活动列表
//	@Tags			/team_up/event
//	@Accept			json
//	@Produce		json
//	@Param			sport_type	body		string	true	"运动类型：pedal, tennis, pickleball"
//	@Param			city		body		string	true	"城市"
//	@Param			start_time	body		int		true	"开始时间，秒级时间戳"
//	@Param			num			body		int		true	"获取的数量"
//	@Param			offset		body		int		true	"偏移量"
//	@Success		200			{object}	GetEventListResp
//	@Router			/team_up/event/list [post]
func EventList(c *model.TeamUpContext) (interface{}, error) {
	// 从query获取sport_type
	body := &GetEventListBody{}
	err := c.BindJSON(body)
	if err != nil {
		util.Logger.Printf("[GetEventList] invalid body:%v", body)
		return nil, iface.NewBackEndError(iface.ParamsError, "invalid query")
	}
	// 从DB根据查询条件获取对应的event
	if body.SportType == "" || body.City == "" || body.StartTime == 0 || body.Num == 0 {
		util.Logger.Printf("[GetEventList] invalid body:%v", body)
		return nil, iface.NewBackEndError(iface.ParamsError, "invalid query")
	}
	var eventMetaList []mysql.EventMeta
	err = util.DB().Where("sport_type = ? AND city = ? AND status IN ? AND start_time > ?", body.SportType, body.City, []string{constant.EventStatusCreated, constant.EventStatusFull}, body.StartTime).Offset(body.Offset).Limit(body.Num).Find(&eventMetaList).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			util.Logger.Printf("[GetEventList] query with current options return no result")
			return &GetEventListData{
				HasMore: false,
				Offset:  body.Offset,
			}, nil
		}
		util.Logger.Printf("[GetEventList] query events failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
	}
	hasMore := false
	offset := 0
	if len(eventMetaList) == body.Num {
		hasMore = true
		offset = body.Offset + body.Num
	}
	res := &GetEventListData{
		HasMore: hasMore,
		Offset:  offset,
	}
	eventList := make([]*Event, 0)
	// 转成eventShow
	for _, event := range eventMetaList {
		eventInfo, err := EventMetaToEventInfo(&event)
		if err != nil {
			util.Logger.Printf("[GetEventList] EventMetaToEventInfo failed, err:%v", err)
			// 偶尔的失败进行continue
			continue
		}
		eventList = append(eventList, eventInfo)
	}
	res.EventList = eventList

	util.Logger.Printf("[GetEventList] success, res:%+v", res)
	return res, nil
}

func EventMetaToEventInfo(event *mysql.EventMeta) (*Event, error) {
	eventShow := &Event{
		StartTime:        event.StartTime,
		EndTime:          event.EndTime,
		IsBooked:         event.IsBooked == 1,
		FieldName:        event.FieldName,
		CurrentPlayerNum: event.CurrentPlayerNum,
		MaxPlayerNum:     event.MaxPlayerNum,
		GameType:         event.GameType,
		MatchType:        event.MatchType,
		LowestLevel:      float32(event.LowestLevel) / 100,
		HighestLevel:     float32(event.HighestLevel) / 100,
	}
	// 获取参与这个活动的用户信息
	currentPeople := make([]string, 0)
	err := sonic.UnmarshalString(event.CurrentPlayer, &currentPeople)
	if err != nil {
		util.Logger.Printf("[GetMyTab] unmarshal failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, "unmarshal failed")
	}
	var joiners []mysql.WechatUserInfo
	err = util.DB().Where("open_id IN ? AND is_primary = 1", currentPeople).Find(&joiners).Error
	if err != nil {
		util.Logger.Printf("[GetMyTab] query joiners info failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, "query failed")
	}
	currentPLayers := make([]*UserInfo, 0)
	for _, joiner := range joiners {
		currentPLayers = append(currentPLayers, &UserInfo{
			NickName:  joiner.Nickname,
			AvatarUrl: joiner.Avatar,
		})
	}
	eventShow.CurrentPlayer = currentPLayers
	return eventShow, nil
}
