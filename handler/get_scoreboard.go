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
		AvailableScoreRule:   []string{constant.PedalScoreRuleAmericano, constant.PedalScoreRuleTennis},
		AvailableRoundTarget: []int{8, 16, 24, 32},
		FieldNum:             1,
	},
	constant.SportTypePickelBall: {
		AvailableScoreRule:   []string{constant.PickleBallScoreRuleServe, constant.PickleBallScoreRuleEvery},
		AvailableRoundTarget: []int{11, 21},
		FieldNum:             1,
	},
	constant.SportTypeTennis: {
		AvailableScoreRule:   []string{constant.PedalScoreRuleTennis},
		AvailableRoundTarget: []int{6},
		FieldNum:             1,
	},
}

// GetScoreboard 记分板页面
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

	scoreBoard := &ScoreBoard{}
	// 组装EventInfo
	eventInfo := &model.EventInfo{
		Name:          event.Name,
		SportType:     event.SportType,
		StartTimeStr:  event.StartTimeStr,
		EndTimeStr:    event.EndTimeStr,
		IsBooked:      event.IsBooked == 1,
		FieldName:     event.FieldName,
		IsCompetitive: event.MatchType == constant.EventMatchTypeCompetitive,
		LowestLevel:   float32(event.LowestLevel / 100),
		HighestLevel:  float32(event.HighestLevel / 100),
		Weekday:       event.Weekday,
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
			Level:        float32(user.Level / 100),
		}
		players = append(players, player)
	}
	scoreBoard.Players = players

	// 组装Option
	scoreBoard.Options = SportTypeScoreOptions[event.SportType]

	util.Logger.Printf("[GetScoreBoard] success")
	return scoreBoard, nil
}
