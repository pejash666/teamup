package handler

import (
	"teamup/constant"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/util"
)

type MatchDetail struct {
	RoundInfo []*RoundInfo `json:"round_info"`
	Settings  *Settings    `json:"settings"`
}

type RoundInfo struct {
	Home     []*model.Player `json:"home"`      // 主队球员
	HomeAvg  float32         `json:"home_avg"`  // 主队平均分
	AwayAvg  float32         `json:"away_avg"`  // 客队平均分
	Away     []*model.Player `json:"away"`      // 客队球员
	CourtNum int32           `json:"court_num"` // 场地号
	RoundNum int32           `json:"round_num"` // 轮次数
}

type Settings struct {
	ValidScorers []*model.Player `json:"valid_scorers"`
}

// StartScoring 用户点击开始记分
// 对于pedal运动：
// 1. Americano记分模式，有N个用户参加，则生成N-1场对局，俩俩搭配；
//  每场比赛后，每名玩家累计自己的比分，按照总分加和排名，这种case下，服务端一次性下发所有的轮次信息给到用户。
// 2. Mexicano记分模式，有N个用户参加，首先随机生成一轮对局，结束后，根据结果需要安排首轮第一与首轮第四一组 对抗 首轮第三与首轮第二
//    这种case下，服务端第一次只会下发首轮的对局。
// 3. Tennis记分模式，N个用户参加，按照网球规则三局两胜，每一局到6分就赢
//    这种case下，1号玩家与2号玩家 对抗 3号玩家与4号玩家即可，一次性下发三轮的信息

// 对于pickelball运动
// 1. ？？？

func StartScoring(c *model.TeamUpContext) (interface{}, error) {
	type Body struct {
		ScoreRule   string `json:"score_rule"`
		RoundTarget int    `json:"round_target"`
		FieldNum    int    `json:"field_num"` // 现在都是1，多块场地需要多个活动
		EventID     int    `json:"event_id"`
	}
	body := &Body{}
	err := c.BindJSON(body)
	if err != nil {
		util.Logger.Printf("[StartScoring] bindJson failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.ParamsError, "invalid record")
	}
	if body.ScoreRule == "" || body.EventID < 1 || body.RoundTarget < 1 {
		util.Logger.Printf("[StartScoring] invalid params")
		return nil, iface.NewBackEndError(iface.ParamsError, "invalid params")
	}

	res := &MatchDetail{}

	event := &mysql.EventMeta{}
	err = util.DB().Where("id = ?", body.EventID).Take(event).Error
	if err != nil {
		util.Logger.Printf("[StartScoring] query event failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.ParamsError, "query record failed")
	}
	event.ScoreRule = body.ScoreRule
	// 获取全部用户
	currentPlayers := make([]string, 0)
	players := make([]*model.Player, 0)
	if err != nil {
		util.Logger.Printf("[StartScoring] unmarshal currentPlayer failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.InternalError, "unmarshal failed")
	}
	if len(currentPlayers) < 1 {
		return nil, iface.NewBackEndError(iface.ParamsError, "invalid player number")
	}
	// 从mysql获取参赛人员数据
	var users []mysql.WechatUserInfo
	err = util.DB().Where("sport_type = ? AND open_id IN ?", event.SportType, currentPlayers).Find(users).Error
	if err != nil {
		util.Logger.Printf("[StartScoring] query user failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, "query record failed")
	}
	for _, user := range users {
		player := &model.Player{
			NickName:     user.Nickname,
			Avatar:       user.Avatar,
			OpenID:       user.OpenId,
			IsCalibrated: user.IsCalibrated == 1,
			Level:        float32(user.Level / 100),
		}
		players = append(players, player)
	}

	// 如果scorer为空，则所有人都可以记分
	if event.Scorers == "" {
		res.Settings = &Settings{
			ValidScorers: players,
		}
	}

	// 进行分组
	switch event.SportType {
	case constant.SportTypePedal:
		// 进行人数校验，网球规则必须为偶数
		if event.ScoreRule == constant.PedalScoreRuleTennis && len(players)%2 != 0 {
			util.Logger.Printf("[StartScoring] unmatched pedal game and player")
			return nil, iface.NewBackEndError(iface.ParamsError, "invalid gametype & player")
		}
		res.RoundInfo = dividePedal(event, players)
	case constant.SportTypePickelBall:
		// 进行人数校验
		if (event.GameType == constant.EventGameTypeDuo && len(players)%2 != 0) || (event.GameType == constant.EventGameTypeSolo && len(players) != 2) {
			util.Logger.Printf("[StartScoring] unmatched pickleball game and player")
			return nil, iface.NewBackEndError(iface.ParamsError, "invalid gametype & player")
		}
		res.RoundInfo = dividePickleBall(c, event, players)
	case constant.SportTypeTennis:
		// 进行人数校验
		if len(players)%2 != 0 {
			util.Logger.Printf("[StartScoring] unmatched tennis game and player")
			return nil, iface.NewBackEndError(iface.ParamsError, "invalid gametype & player")
		}
		res.RoundInfo = divideTennis(c, event, players)
	}

}

// dividePedal pedal分组
func dividePedal(c *model.TeamUpContext, event *mysql.EventMeta, players []*model.Player) []*RoundInfo {
	switch event.ScoreRule {
	case constant.PedalScoreRuleAmericano:
		return divideAmericano(event, players)
	case constant.PedalScoreRuleMexicano:
		return divideMexicano(event, players)
	case constant.PedalScoreRuleTennis:
		return divideTennis(c, event, players)
	default:
		return nil
	}
}

// dividePickleBall 匹克球分组（单打就是两个人，双打必须偶数人）
// 发球得分制：生成三局对阵，三局两胜
// 每球得分制：一局定胜负
func dividePickleBall(c *model.TeamUpContext, event *mysql.EventMeta, players []*model.Player) []*RoundInfo {
	// 打散用户的顺序
	ShufflePlayers(c, players)
	switch event.ScoreRule {
	case constant.PickleBallScoreRuleEvery:
		// 单打
		if event.GameType == constant.EventGameTypeSolo {
			return []*RoundInfo{
				{
					Home:     players[:1],
					Away:     players[1:2],
					HomeAvg:  players[0].Level,
					AwayAvg:  players[1].Level,
					CourtNum: 1,
					RoundNum: 1,
				},
			}
		}

		res := make([]*RoundInfo, 0)
		// 计算有几组运动员
		// 比如有6人参与，3组 a,b,c
		// a vs b; b vs c; a vs c
		groupNum := len(players) / 2
		for i := 1; i <= groupNum; i++ {
			// 计算对局的group
			vsGroup := i + 1
			if vsGroup > groupNum {
				vsGroup = 1
			}
			res = append(res, &RoundInfo{
				Home:     players[(i-1)*2 : i*2],
				Away:     players[(vsGroup-1)*2 : vsGroup*2],
				HomeAvg:  GetPlayersAvgLevel(players[(i-1)*2 : i*2]),
				AwayAvg:  GetPlayersAvgLevel(players[(vsGroup-1)*2 : vsGroup*2]),
				CourtNum: 1,
				RoundNum: int32(i),
			})
		}

	case constant.PickleBallScoreRuleServe:
		// 单打
		if event.GameType == constant.EventGameTypeSolo {
			return []*RoundInfo{
				{
					Home:     players[:1],
					Away:     players[1:2],
					HomeAvg:  players[0].Level,
					AwayAvg:  players[1].Level,
					CourtNum: 1,
					RoundNum: 1,
				},
				{
					Home:     players[:1],
					Away:     players[1:2],
					HomeAvg:  players[0].Level,
					AwayAvg:  players[1].Level,
					CourtNum: 1,
					RoundNum: 1,
				},
				{
					Home:     players[:1],
					Away:     players[1:2],
					HomeAvg:  players[0].Level,
					AwayAvg:  players[1].Level,
					CourtNum: 1,
					RoundNum: 1,
				},
			}
		}
		// 双打, 三局两胜制需要一轮生成3个一样的对局
		// 总共生成 对战组数 x 3 个对局
		res := make([]*RoundInfo, 0)
		groupNum := len(players) / 2
		for i := 1; i <= groupNum; i++ {
			// 计算对局的group
			vsGroup := i + 1
			if vsGroup > groupNum {
				vsGroup = 1
			}
			roundInfos := []*RoundInfo{
				{
					Home:     players[(i-1)*2 : i*2],
					Away:     players[(vsGroup-1)*2 : vsGroup*2],
					HomeAvg:  GetPlayersAvgLevel(players[(i-1)*2 : i*2]),
					AwayAvg:  GetPlayersAvgLevel(players[(vsGroup-1)*2 : vsGroup*2]),
					CourtNum: 1,
					RoundNum: int32(i),
				},
				{
					Home:     players[(i-1)*2 : i*2],
					Away:     players[(vsGroup-1)*2 : vsGroup*2],
					HomeAvg:  GetPlayersAvgLevel(players[(i-1)*2 : i*2]),
					AwayAvg:  GetPlayersAvgLevel(players[(vsGroup-1)*2 : vsGroup*2]),
					CourtNum: 1,
					RoundNum: int32(i),
				},
				{
					Home:     players[(i-1)*2 : i*2],
					Away:     players[(vsGroup-1)*2 : vsGroup*2],
					HomeAvg:  GetPlayersAvgLevel(players[(i-1)*2 : i*2]),
					AwayAvg:  GetPlayersAvgLevel(players[(vsGroup-1)*2 : vsGroup*2]),
					CourtNum: 1,
					RoundNum: int32(i),
				},
			}

			res = append(res, roundInfos...)
		}
	}
}

// divideTennis tennis规则，适用与pedal的tennis记分规则，和tennis sport_type 分组
// tennis规则下搭档也是固定的，所以如果是pedal运动，人数必须为偶数
func divideTennis(c *model.TeamUpContext, event *mysql.EventMeta, players []*model.Player) []*RoundInfo {
	return dividePickleBall(c, event, players)
}

// divideAmericano pedal的Americano规则
func divideAmericano(event *mysql.EventMeta, players []*model.Player) []*RoundInfo {
	// adaptive算法
	// 比如有abcd 4个人， 生成3场比赛
	// Game1: ab vs cd    a:1, b:1, c:1, d:1
	// Game2: ac vs bd    a:2, b:2, c:2, d:2
	// Game3: ad vs bc    a:3, b:3, c:3, d:3
	// 比如有abcde 5个人, 需要生成4场比赛
	// Game1 : ab vs cd/ce/de  随机选择一个锚点用户 ab vs cd , a: 1, b: 1, c: 1, d:1, e:0
	// Game2 : ac vs bd/bd/be  ac vs be , a:2, b:2, c:2, d:1, e:1
	// Game3 : ad vs bc/be/ce  ad vs ce , a:3, b:2, c:3, d:2, e:2
	// Game4 : ae vs bc/bd/cd  ae vs bd , a:4, b:3, c:3, d:3, e:3

	// 比如有abcdef 6个人 需要生成5场比赛
	// Game1: ab vs cd/ce/cf/de/df/ef 随机选择一个锚点用户 ab vs cd, a:1, b:1, c:1, d:1, e:0, f:0
	// Game2: ac vs bd/be/bf/de/df/ef  ac vs ef      a:2, b:1, c:2, d:1, e:1, f:1
	//                                 bf vs de      a:2, b:2, c:2, d:2, e:2, f:2
	//                                 ad vs bc      a:3, b:3, c:3, d:3, e:2, f:2
	//                                 ae vs bf      a:4, b:4, c:3, d:3, e:3, f:3
	//                                 ce vs df      a:4, b:4, c:4, d:4, e:4, f:4

	// Game3: ad vs bc/be/bf/ce/cf/ef  ad vs be      a:3, b:2, c:2, d:2, e:2, f:1
	// Game4: ae vs bc/bd/bf/cd/cf/df  ae vs bf      a:4, b:3, c:2, d:2, e:3, f:2
	// Game5: af vs bc/bd/be/cd/ce/de  af vs de      a:5, b:3, c:3, d:3, e:4, f:3

	// 记录每个用户参与的次数
	countMap := make(map[string]int)
	// 记录每个组合是否出现过，必须和不同的人搭档
	groupDedupMap := make(map[string]int)

	// 初始化，就选在前4个用户
	initGroup1 := players[0].OpenID + players[1].OpenID
	initGroup2 := players[2].OpenID + players[3].OpenID
	countMap[players[0].OpenID] += 1
	countMap[players[1].OpenID] += 1
	countMap[players[2].OpenID] += 1
	countMap[players[3].OpenID] += 1

	groupDedupMap[initGroup1] = 1
	groupDedupMap[initGroup2] = 1

	// 建立一个数学期望，分配完场次后，一个人参与的场次最多不会超过总人数数量
	expectation := len(players)
	// n个人 生成n-1场比赛，第一轮默认初始化生成，从第二轮开始
	for i := 2; i <= len(players)-1; i++ {
		// 遍历countMap，记录参与次数和expectation差距最大的用户，这些用户必须存在于下一轮的比赛中
		// 如果不足4人，则回溯寻找差距第二大的用户，用来填满一轮需要的四位玩家，填满为止

		for openID, times := range countMap {

		}

		// 找到满足要求的4个玩家后，进行随机俩俩组合，遍历groupDedupMap，一组的用户必须是首次组队
		// 记录找到的两个组合，构建roundInfo
	}

}

// divideMexicano pedal的Mexicano规则
func divideMexicano(event *mysql.EventMeta, players []*model.Player) []*RoundInfo {

}

// ShufflePlayers 对运动员进行随机打散顺序
func ShufflePlayers(c *model.TeamUpContext, players []*model.Player) {
	c.Rand.Shuffle(len(players), func(i, j int) {
		players[i], players[j] = players[j], players[i]
	})
}

func GetPlayersAvgLevel(players []*model.Player) float32 {
	res := float32(0)
	for _, player := range players {
		res += player.Level
	}
	return res / float32(len(players))
}
