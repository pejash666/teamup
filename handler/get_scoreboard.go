package handler

import (
	"github.com/bytedance/sonic"
	"teamup/constant"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/util"
)

type ScoreBoard struct {
	EventInfo *model.EventInfo    `json:"event_info"`
	Players   []*model.Player     `json:"players"`
	Options   *model.ScoreOptions `json:"options"`
}

var SportTypeScoreOptions = map[string]*model.ScoreOptions{
	constant.SportTypePedal: {
		AvailableScoreRule: []string{constant.PedalScoreRuleAmericano, constant.PedalScoreRuleTennis},
		AvailableRoundTarget: map[string][]int{
			constant.PedalScoreRuleAmericano: {8, 16, 24, 32},
			constant.PedalScoreRuleTennis:    {6},
		},
		FieldNum: 1,
	},
	constant.SportTypePickelBall: {
		AvailableScoreRule: []string{constant.PickleBallScoreRuleServe, constant.PickleBallScoreRuleEvery},
		AvailableRoundTarget: map[string][]int{
			constant.SportTypePickelBall: {11, 21},
		},
		FieldNum: 1,
	},
	constant.SportTypeTennis: {
		AvailableScoreRule: []string{constant.PedalScoreRuleTennis},
		AvailableRoundTarget: map[string][]int{
			constant.SportTypeTennis: {6},
		},
		FieldNum: 1,
	},
}

type GetScoreboardBody struct {
	EventID int64 `json:"event_id"`
}

type GetScoreboardResp struct {
	ErrNo   int32       `json:"err_no"`
	ErrTips string      `json:"err_tips"`
	Data    *ScoreBoard `json:"data"`
}

// GetScoreboard godoc
// @Summary      记分板页面
// @Description  记分板页面，包含用户可选规则项
// @Tags         /teamup/user
// @Accept       json
// @Produce      json
// @Param        get_scoreboard  body  {object} GetScoreboardBody  true  "获取记分板页面入参"
// @Success      200  {object}  GetScoreboardResp
// @Router       /teamup/user/get_scoreboard [post]
func GetScoreboard(c *model.TeamUpContext) (interface{}, error) {
	type Body struct {
		EventID int `json:"event_id"`
	}
	body := &Body{}
	err := c.BindJSON(body)
	if err != nil {
		util.Logger.Printf("[GetScoreBoard] bindJson failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.InternalError, err.Error())
	}
	// 获取活动的信息, 根据单/双打 和 运动类型，能够开始记分的标准不一样
	// 匹克球单打：2人
	// 匹克球双打：4人
	// pedal双打：4人
	event := &mysql.EventMeta{}
	err = util.DB().Where("id = ? AND status IN ?", body.EventID, []string{constant.EventStatusCreated, constant.EventStatusFull}).Take(event).Error
	if err != nil {
		util.Logger.Printf("[GetScoreBoard] query record failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
	}
	// 如果已经满员的，肯定OK
	if event.Status != constant.EventStatusFull {
		switch event.SportType {
		case constant.SportTypePedal:
			// pedal 最少有4个人
			if event.CurrentPlayerNum < 4 {
				util.Logger.Printf("[GetScoreBoard] currentNum is less than 4, cant start")
				return nil, iface.NewBackEndError(iface.ParamsError, "invalid player num")
			}
		case constant.SportTypePickelBall:
			// 双打必须是>4的偶数就是4个人
			if event.GameType == constant.EventGameTypeDuo {
				if event.CurrentPlayerNum < 4 || event.CurrentPlayerNum%2 != 0 {
					util.Logger.Printf("[GetScoreBoard] game is duo, and playerNum is invalid")
					return nil, iface.NewBackEndError(iface.ParamsError, "invalid player num")
				}
			} else {
				if event.CurrentPlayerNum != 2 {
					util.Logger.Printf("[GetScoreBoard] game is solo, invalid player number")
					return nil, iface.NewBackEndError(iface.ParamsError, "invalid player number")
				}
			}
		}
	}
	// 获取创建者信息
	creator := &mysql.WechatUserInfo{}
	err = util.DB().Where("open_id = ? AND sport_type = ?", event.Creator, event.SportType).Error
	if err != nil {
		return nil, iface.NewBackEndError(iface.MysqlError, "creator not found")
	}
	scoreBoard := &ScoreBoard{}
	// 组装EventInfo
	eventInfo := &model.EventInfo{
		Id:              int64(event.ID),
		Status:          event.Status,
		Date:            event.Date,
		City:            event.City,
		Name:            event.Name,
		CreatorNickname: creator.Nickname,
		CreatorAvatar:   creator.Avatar,
		Price:           event.Price,
		Desc:            event.Desc,
		SportType:       event.SportType,
		StartTime:       event.StartTime,
		StartTimeStr:    event.StartTimeStr,
		EndTime:         event.EndTime,
		EndTimeStr:      event.EndTimeStr,
		IsBooked:        event.IsBooked == 1,
		IsPublic:        event.IsPublic == 1,
		FieldName:       event.FieldName,
		IsCompetitive:   event.MatchType == constant.EventMatchTypeCompetitive,
		GameType:        event.GameType,
		LowestLevel:     float32(event.LowestLevel / 1000),
		HighestLevel:    float32(event.HighestLevel / 1000),
		Weekday:         event.Weekday,
		MaxPeopleNum:    event.MaxPlayerNum,
		CurrentPeople:   event.CurrentPlayerNum,
	}
	scoreBoard.EventInfo = eventInfo

	// 组装players
	currentPeople := make([]string, 0)
	err = sonic.UnmarshalString(event.CurrentPlayer, &currentPeople)
	if err != nil {
		util.Logger.Printf("[GetScoreBoard] Unmarshal currentPeople from DB failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.InternalError, "unmarshall error")
	}
	var users []mysql.WechatUserInfo
	result := util.DB().Where("open_id IN ? AND sport_type = ?", currentPeople, event.SportType).Find(&users)
	if result.Error != nil {
		util.Logger.Printf("[GetScoreBoard] find player in DB failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, "get record failed")
	}
	if len(users) < 1 || len(users) != int(event.CurrentPlayerNum) {
		util.Logger.Printf("[GetScoreBoard] unmatched currentplayer info")
		return nil, iface.NewBackEndError(iface.InternalError, "unmatched player info")
	}
	players := make([]*model.Player, 0)
	for _, user := range users {
		player := &model.Player{
			NickName:     user.Nickname,
			Avatar:       user.Avatar,
			IsCalibrated: user.IsCalibrated == 1,
			Level:        float32(user.Level) / 1000,
		}
		players = append(players, player)
	}
	scoreBoard.Players = players

	// 组装Option
	scoreBoard.Options = SportTypeScoreOptions[event.SportType]

	util.Logger.Printf("[GetScoreBoard] success")
	return scoreBoard, nil
}
