package handler

import (
	"errors"
	"github.com/bytedance/sonic"
	"gorm.io/gorm"
	"math"
	"sort"
	"strconv"
	"teamup/constant"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/util"
	"time"
)

type GetEventListBody struct {
	SportType string `json:"sport_type"`
	City      string `json:"city"`
	StartTime int64  `json:"start_time"`
	//EndTime                int64                  `json:"end_time"`
	EventListFilterOptions *EventListFilterOption `json:"filter_option"`
	EventListOrderOption   *EventListOrderOption  `json:"order_option"`
	Num                    int                    `json:"num"`
	Offset                 int                    `json:"offset"`
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

type EventListFilterOption struct {
	PlayerGameOnly   bool `json:"player_game_only"`   // 只展示个人比赛
	JoinableGameOnly bool `json:"joinable_game_only"` // 只展示可加入的活动
	FieldGameOnly    bool `json:"field_game_only"`    // 只展示已定场比赛
}

type EventListOrderOption struct {
	OrderBy   string  `json:"order_by"`  // 排序方式(by_time;by_level;by_distance)
	Latitude  *string `json:"latitude"`  // 纬度
	Longitude *string `json:"longitude"` // 经度
}

// EventList godoc
//
//	@Summary		获取活动列表
//	@Description	根据筛选条件获取活动列表
//	@Tags			/team_up/event
//	@Accept			json
//	@Produce		json
//	@Param			sport_type	body		string	true	"运动类型：padel, tennis, pickleball"
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
	util.Logger.Printf("[GetEventList] req:%v", util.ToReadable(body))
	// 获取筛选条件
	hostOption, fieldOption, statusOption := GetEventListOptions(body)
	util.Logger.Printf("%v, %v, %v", hostOption, fieldOption, statusOption)
	var eventMetaList []mysql.EventMeta
	err = util.DB().Where("sport_type = ? AND city = ? AND is_public = 1 AND start_time > ? AND is_host IN ? AND is_booked IN ? AND status IN ?", body.SportType, body.City, body.StartTime, hostOption, fieldOption, statusOption).Offset(body.Offset).Limit(body.Num).Find(&eventMetaList).Error
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
		tmp := event
		eventInfo, err := EventMetaToEventInfo(&tmp)
		if err != nil {
			util.Logger.Printf("[GetEventList] EventMetaToEventInfo failed, err:%v", err)
			// 偶尔的失败进行continue
			continue
		}
		eventList = append(eventList, eventInfo)
	}

	// 根据用户选择的order_by进行定制化排序
	if body.EventListOrderOption == nil || body.EventListOrderOption.OrderBy == "by_time" {
		SortByTime(eventList)
	} else if body.EventListOrderOption.OrderBy == "by_distance" {
		if body.EventListOrderOption.Latitude == nil || body.EventListOrderOption.Longitude == nil {
			util.Logger.Printf("[GetEventList] invalid order option")
			return nil, iface.NewBackEndError(iface.ParamsError, "invalid order option")
		}
		err = SortByDistance(body.EventListOrderOption.Latitude, body.EventListOrderOption.Longitude, eventList)
		if err != nil {
			return nil, iface.NewBackEndError(iface.InternalError, err.Error())
		}
	} else if body.EventListOrderOption.OrderBy == "by_level" {
		if c.BasicUser == nil {
			return nil, iface.NewBackEndError(iface.ParamsError, "need login")
		}
		var user mysql.WechatUserInfo
		err = util.DB().Where(" open_id = ? AND sport_type = ?", c.BasicUser.OpenID, body.SportType).Take(&user).Error
		if err != nil {
			return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
		}
		SortByLevel(user.Level, eventList)
	}
	res.EventList = eventList

	eventListStr, _ := sonic.MarshalString(eventList)
	util.Logger.Printf("[GetEventList] success, res is %v", eventListStr)
	return res, nil
}

func EventMetaToEventInfo(event *mysql.EventMeta) (*Event, error) {
	eventShow := &Event{
		ID:               event.ID,
		EventName:        event.Name,
		StartTime:        event.StartTime,
		StartTimeStr:     event.StartTimeStr,
		EndTime:          event.EndTime,
		EndTimeStr:       event.EndTimeStr,
		Weekday:          time.Unix(event.StartTime, 0).Weekday().String(),
		Date:             time.Unix(event.StartTime, 0).Format("2006-01-02"),
		IsBooked:         event.IsBooked == 1,
		FieldName:        event.FieldName,
		CurrentPlayerNum: event.CurrentPlayerNum,
		MaxPlayerNum:     event.MaxPlayerNum,
		GameType:         event.GameType,
		MatchType:        event.MatchType,
		LowestLevel:      float32(event.LowestLevel) / 1000,
		HighestLevel:     float32(event.HighestLevel) / 1000,
		EventImage:       event.EventImage,
		IsHost:           event.IsHost == 1,
	}
	// 如果是组织创建的，获取组织的地址，logo
	if event.IsHost == 1 {
		organization := &mysql.Organization{}
		err := util.DB().Where("id = ?", event.OrganizationID).Take(organization).Error
		if err != nil {
			util.Logger.Printf("[EventMetaToEventInfo] query organization failed for id:%d", event.OrganizationID)
			return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
		}
		eventShow.OrganizationLogo = organization.Logo
		eventShow.OrganizationAddress = organization.Address
	}
	util.Logger.Printf("[EventMetaToEventInfo] event_id:%d, start_time:%d, end_time:%d, time_now:%v", event.ID, event.StartTime, event.EndTime, time.Now().Unix())
	// 获取status
	if time.Now().Unix() > event.StartTime && time.Now().Unix() < event.EndTime {
		eventShow.Status = constant.EventStatusInProgress
	} else {
		eventShow.Status = event.Status
	}
	// 订场地只能从已合作的场地里面选择
	if event.FieldName != "" && event.Latitude != "" && event.Longitude != "" {
		longitude, err := strconv.ParseFloat(event.Longitude, 64)
		if err != nil {
			return nil, iface.NewBackEndError(iface.InternalError, err.Error())
		}
		latitude, err := strconv.ParseFloat(event.Latitude, 64)
		if err != nil {
			return nil, iface.NewBackEndError(iface.InternalError, err.Error())
		}
		eventShow.Latitude = latitude
		eventShow.Longitude = longitude
	}
	// 获取参与这个活动的用户信息
	currentPeople := make([]string, 0)
	// 只有此活动有人参与，才需要解析
	if event.CurrentPlayer != "" {
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
	}
	//err := sonic.UnmarshalString(event.CurrentPlayer, &currentPeople)
	//if err != nil {
	//	util.Logger.Printf("[GetMyTab] unmarshal failed, err:%v", err)
	//	return nil, iface.NewBackEndError(iface.MysqlError, "unmarshal failed")
	//}
	//var joiners []mysql.WechatUserInfo
	//err := util.DB().Where("open_id IN ? AND is_primary = 1", currentPeople).Find(&joiners).Error
	//if err != nil {
	//	util.Logger.Printf("[GetMyTab] query joiners info failed, err:%v", err)
	//	return nil, iface.NewBackEndError(iface.MysqlError, "query failed")
	//}
	//currentPLayers := make([]*UserInfo, 0)
	//for _, joiner := range joiners {
	//	currentPLayers = append(currentPLayers, &UserInfo{
	//		NickName:  joiner.Nickname,
	//		AvatarUrl: joiner.Avatar,
	//	})
	//}
	//eventShow.CurrentPlayer = currentPLayers
	return eventShow, nil
}

func GetEventListOptions(param *GetEventListBody) (hostOption []int, fieldOption []int, statusOption []string) {
	hostOption = []int{0, 1}
	fieldOption = []int{0, 1}
	statusOption = []string{constant.EventStatusCreated, constant.EventStatusFull, constant.EventStatusFinished}
	if param.EventListFilterOptions != nil {
		if param.EventListFilterOptions.PlayerGameOnly {
			hostOption = []int{1}
		}
		if param.EventListFilterOptions.JoinableGameOnly {
			statusOption = []string{constant.EventStatusCreated}
		}
		if param.EventListFilterOptions.FieldGameOnly {
			fieldOption = []int{1}
		}
	}
	return hostOption, fieldOption, statusOption
}

func SortByDistance(latitude, longitude *string, events []*Event) error {
	const R = 6367000
	if latitude == nil || longitude == nil || *latitude == "" || *longitude == "" {
		return iface.NewBackEndError(iface.ParamsError, "invalid longitude/latitude")
	}
	lat, err := strconv.ParseFloat(*latitude, 64)
	if err != nil {
		return iface.NewBackEndError(iface.InternalError, err.Error())
	}
	longi, err := strconv.ParseFloat(*longitude, 64)
	if err != nil {
		return iface.NewBackEndError(iface.InternalError, err.Error())
	}
	sort.Slice(events, func(i, j int) bool {
		if events[i].IsBooked && events[j].IsBooked {
			// 如果都是已定场，那么按照距离排序
			c1 := math.Sin(lat)*math.Sin(events[i].Longitude)*math.Cos(longi-events[i].Longitude) + math.Cos(lat)*math.Cos(events[i].Latitude)
			d1 := R * math.Acos(c1) * math.Pi / 180

			c2 := math.Sin(lat)*math.Sin(events[j].Longitude)*math.Cos(longi-events[j].Longitude) + math.Cos(lat)*math.Cos(events[j].Latitude)
			d2 := R * math.Acos(c2) * math.Pi / 180
			return d1 < d2
			// 其他情况优先展示已定场的
		} else if events[i].IsBooked && !events[j].IsBooked {
			return true
		} else if !events[i].IsBooked && events[j].IsBooked {
			return false
		} else if !events[i].IsBooked && !events[j].IsBooked {
			return events[i].StartTime < events[j].StartTime
		}
		return events[i].StartTime < events[j].StartTime
	})
	return nil
}

func SortByTime(events []*Event) {
	sort.Slice(events, func(i, j int) bool {
		return events[i].StartTime > events[j].StartTime
	})
}

func SortByLevel(level int, events []*Event) {
	sort.Slice(events, func(i, j int) bool {
		// 未定级用户, 优先展示低等级的
		if level == 0 {
			return events[i].LowestLevel < events[j].LowestLevel
		}
		// todo: 这里需要细化，对于登陆的用户，展示更贴近自己区间的比赛
		return events[i].LowestLevel < events[j].LowestLevel
	})
}
