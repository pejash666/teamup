package handler

import (
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/jmoiron/sqlx"
	rand2 "math/rand"
	"sort"
	"teamup/constant"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/util"
	"time"
)

type MatchDetail struct {
	RoundInfo        []*RoundInfo `json:"round_info"`
	Settings         *Settings    `json:"settings"`
	RoundTargetScore int32        `json:"round_target_score"` // 目标得分
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

type Sortable struct {
	*model.Player
	Count int
}

type StartScoringBody struct {
	ScoreRule        string `json:"score_rule"`
	RoundTargetScore int32  `json:"round_target_score"`
	FieldNum         int    `json:"field_num"` // 现在都是1，多块场地需要多个活动
	EventID          int    `json:"event_id"`
}

type StartScoringResp struct {
	ErrNo   int32        `json:"err_no"`
	ErrTips string       `json:"err_tips"`
	Data    *MatchDetail `json:"data"`
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

// StartScoring godoc
// @Summary      记分详细规则
// @Description  根据用户选择的规则下发对应的详细对局信息
// @Tags         /teamup/user
// @Accept       json
// @Produce      json
// @Param        start_scoring  body  {object} StartScoringBody  true  "用户选择的记分规则"
// @Success      200  {object}  StartScoringResp
// @Router       /teamup/user/start_scoring [post]
func StartScoring(c *model.TeamUpContext) (interface{}, error) {
	body := &StartScoringBody{}
	err := c.BindJSON(body)
	if err != nil {
		util.Logger.Printf("[StartScoring] bindJson failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.ParamsError, "invalid record")
	}
	if body.ScoreRule == "" || body.EventID < 1 || body.RoundTargetScore < 1 {
		util.Logger.Printf("[StartScoring] invalid params")
		return nil, iface.NewBackEndError(iface.ParamsError, "invalid params")
	}

	res := &MatchDetail{}
	res.RoundTargetScore = body.RoundTargetScore

	event := &mysql.EventMeta{}
	err = util.DB().Where("id = ?", body.EventID).Take(event).Error
	if err != nil {
		util.Logger.Printf("[StartScoring] query event failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.ParamsError, "query record failed")
	}
	event.ScoreRule = body.ScoreRule
	// 获取全部用户
	var currentPlayers []string
	err = sonic.UnmarshalString(event.CurrentPlayer, &currentPlayers)
	if err != nil {
		util.Logger.Printf("[StartScoring] unmarshal currentPlayer failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.InternalError, "unmarshal failed")
	}
	if len(currentPlayers) < 1 {
		return nil, iface.NewBackEndError(iface.ParamsError, "invalid player number")
	}
	querySql := "SELECT * FROM wechat_user_info WHERE sport_type = ? AND open_id IN (?)"
	params := []interface{}{event.SportType, currentPlayers}
	querySql, params, err = sqlx.In(querySql, params...)
	util.Logger.Printf(fmt.Sprintf("StartScoring - querySql: %+v, params: %+v\n", querySql, params))
	if err != nil {
		return nil, iface.NewBackEndError(iface.InternalError, err.Error())
	}
	players := make([]*model.Player, 0)
	// 从mysql获取参赛人员数据
	var users []mysql.WechatUserInfo
	// 对于切片类型的IN条件，需要用sqlx进行打散传入
	err = util.DB().Where("sport_type = ? AND open_id IN ?", params[0], params[1:]).Find(&users).Error
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
			Level:        float32(user.Level) / 1000,
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
		res.RoundInfo = dividePedal(c, event, players)
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

	return res, nil
}

// dividePedal pedal分组
func dividePedal(c *model.TeamUpContext, event *mysql.EventMeta, players []*model.Player) []*RoundInfo {
	switch event.ScoreRule {
	case constant.PedalScoreRuleAmericano:
		return divideAmericano(c, event, players)
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
	ShufflePlayers(players)
	res := make([]*RoundInfo, 0)
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

		// 计算有几组运动员
		// 比如有6人参与，3组 a,b,c
		// a vs b; b vs c; a vs c

		groupNum := len(players) / 2

		for i := 1; i <= groupNum; i++ {
			// 计算对局的group
			vsGroup := i + 1
			if vsGroup > groupNum {
				if groupNum > 2 {
					vsGroup = 1
				} else {
					break
				}

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

	case constant.PickleBallScoreRuleServe, constant.PedalScoreRuleTennis:
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
	default:
		util.Logger.Printf("[dividePickleBall] invalid scoreRule")
		return nil
	}
	util.Logger.Printf("[dividePickball] success")
	return res
}

// divideTennis tennis规则，适用与pedal的tennis记分规则，和tennis sport_type 分组
// tennis规则下搭档也是固定的，所以如果是pedal运动，人数必须为偶数
func divideTennis(c *model.TeamUpContext, event *mysql.EventMeta, players []*model.Player) []*RoundInfo {
	return dividePickleBall(c, event, players)
}

// divideAmericano pedal的Americano规则
func divideAmericano(c *model.TeamUpContext, event *mysql.EventMeta, players []*model.Player) []*RoundInfo {
	// adaptive算法
	// 比如有abcd 4个人， 生成3场比赛
	// Game1: ab vs cd    a:1, b:1, c:1, d:1
	// Game2: ac vs bd    a:2, b:2, c:2, d:2
	// Game3: ad vs bc    a:3, b:3, c:3, d:3
	// 比如有abcde 5个人, 需要生成5场比赛
	// Game1 : ab vs cd/ce/de  随机选择一个锚点用户 ab vs cd , a: 1, b: 1, c: 1, d:1, e:0
	// Game2 : ac vs bd/bd/be  ac vs be , a:2, b:2, c:2, d:1, e:1
	// Game3 : ad vs bc/be/ce  ad vs ce , a:3, b:2, c:3, d:2, e:2
	// Game4 : ae vs bc/bd/cd  ae vs bd , a:4, b:3, c:3, d:3, e:3
	// Game5 :                 bc vs de , a:4, b:4, c:4, d:4, e:4

	// 比如有abcdef 6个人 需要生成6场比赛
	// Game1: ab vs cd/ce/cf/de/df/ef 随机选择一个锚点用户 ab vs cd, a:1, b:1, c:1, d:1, e:0, f:0
	// Game2: ac vs bd/be/bf/de/df/ef  ac vs ef      a:2, b:1, c:2, d:1, e:1, f:1
	//                                 bf vs de      a:2, b:2, c:2, d:2, e:2, f:2
	//                                 ad vs bc      a:3, b:3, c:3, d:3, e:2, f:2
	//                                 ae vs bf      a:4, b:4, c:3, d:3, e:3, f:3
	//                                 ce vs df      a:4, b:4, c:4, d:4, e:4, f:4

	// 比如有abcdefg 7个人
	// ab vs cd  a:1, b:1, c:1, d:1, e:0, f:0, g:0
	// fg vs ea  a:2, b:1, c:1, d:1, e:1, f:1, g:1
	// eg vs fd  a:2, b:1, c:1, d:2, e:2, f:2, g:2
	// bc vs ad  a:3, b:2, c:2, d:3, e:2, f:2, g:2
	// bg vs cf  a:3, b:3, c:3, d:3, e:2, f:3, g:3
	// ac vs eb  a:4, b:4, c:4, d:3, e:3, f:3, g:3
	// dg vs ef  a:4, b:4, c:4, d:4, e:4, f:4, g:4

	//ShufflePlayers(c, players)

	// 记录每个用户参与的次数
	countMap := make(map[string]int)
	// 记录每个组合是否出现过，必须和不同的人搭档
	groupDedupMap := make(map[string]int)

	// 初始化，就选在前4个用户
	countMap[players[0].OpenID] += 1
	countMap[players[1].OpenID] += 1
	countMap[players[2].OpenID] += 1
	countMap[players[3].OpenID] += 1

	// ab 和 ba是一样的
	groupDedupMap[players[0].OpenID+players[1].OpenID] = 1
	groupDedupMap[players[1].OpenID+players[0].OpenID] = 1
	groupDedupMap[players[2].OpenID+players[3].OpenID] = 1
	groupDedupMap[players[3].OpenID+players[2].OpenID] = 1

	// 初始化对局信息, 并将初始化的轮次放进去
	roundInfos := make([]*RoundInfo, 0)
	roundNum := int32(1)
	roundInfos = append(roundInfos, &RoundInfo{
		Home:     []*model.Player{players[0], players[1]},
		HomeAvg:  GetPlayersAvgLevel([]*model.Player{players[0], players[1]}),
		AwayAvg:  GetPlayersAvgLevel([]*model.Player{players[2], players[3]}),
		Away:     []*model.Player{players[2], players[3]},
		CourtNum: 1,
		RoundNum: roundNum,
	})

	// 建立一个数学期望，分配完场次后，一个人参与的场次最多不会超过总人数数量
	//expectation := len(players)
	// n个人 生成n场比赛，第一轮默认初始化生成，从第二轮开始
	expectedRoundNum := len(players)
	if len(players) == 4 {
		expectedRoundNum = 3
	}
	for i := 2; i <= expectedRoundNum; i++ {
		util.Logger.Printf("i:%d", i)
		// sort一下map，按照count
		sortable := make([]*Sortable, 0)
		for _, player := range players {
			sortable = append(sortable, &Sortable{
				Player: player,
				Count:  countMap[player.OpenID],
			})
		}

		// 根据用户的count降序排列
		sort.Slice(sortable, func(i, j int) bool {
			return sortable[i].Count < sortable[j].Count
		})

		// 对于count一样的sortable，需要进行随机取
		// 获取最低的count，最低的count一定要放进去
		lowest := 100
		for _, s := range sortable {
			if s.Count < lowest {
				lowest = s.Count
			}
		}
		startingIdx := 0
		// 优先将lowest的放进来
		candidates := make([]*Sortable, 0)
		for idx, s := range sortable {
			if s.Count == lowest {
				startingIdx = idx + 1
				candidates = append(candidates, s)
			}
		}

	retry:
		// 如果凑不够，再从后面取
		if len(candidates) < 4 {
			potential := make([]*Sortable, 0)
			more := 4 - len(candidates) // 还差几个人
			// 获取第二低的
			secondLowest := 100
			for _, s := range sortable[startingIdx:] {
				if s.Count < secondLowest {
					secondLowest = s.Count
				}
			}
			for _, s := range sortable[startingIdx:] {
				if s.Count == secondLowest {
					potential = append(potential, s)
				}
			}
			rand := rand2.New(rand2.NewSource(time.Now().UnixNano()))
			startIdx := rand.Intn(len(potential))
			// 超出范围
			if startIdx+more > len(potential) {
				candidates = append(candidates, potential[startIdx:]...)
				candidates = append(candidates, potential[startIdx-(4-len(candidates)):startIdx]...)
			} else {
				candidates = append(candidates, potential[startIdx:startingIdx+more]...)
			}
			// 8个人的case，如果超过4个人了，那么需要进行随机
		} else {
			ShuffleCandidates(candidates)
			candidates = candidates[:4]
		}

		//// 取出count最少的4个
		//candidates := sortable[:4]

		stopFlag := 0
		for {
			if stopFlag == 1 {
				break
			}
			possibleGroups := getTwoGroupsFrom4People(candidates)
			util.Logger.Printf("%d", len(possibleGroups))
			// 如果组合存在
			times := 1
			for _, groups := range possibleGroups {
				// 两组组合都没出现过
				group1 := groups[0].Player.OpenID + groups[1].Player.OpenID
				group2 := groups[2].Player.OpenID + groups[3].Player.OpenID
				util.Logger.Printf("group1:%v, group2:%v", group1, group2)
				// todo： 如果三次循环后还是不行，需要重新获取count最少的四个，否则会陷入无限循环
				if checkGroupDedup(groups[0].Player, groups[1].Player, groupDedupMap) && checkGroupDedup(groups[2].Player, groups[3].Player, groupDedupMap) {

					// 记录次数
					countMap[groups[0].Player.OpenID] += 1
					countMap[groups[1].Player.OpenID] += 1
					countMap[groups[2].Player.OpenID] += 1
					countMap[groups[3].Player.OpenID] += 1

					groupDedupMap[groups[0].Player.OpenID+groups[1].Player.OpenID] = 1
					groupDedupMap[groups[1].Player.OpenID+groups[0].Player.OpenID] = 1
					groupDedupMap[groups[2].Player.OpenID+groups[3].Player.OpenID] = 1
					groupDedupMap[groups[3].Player.OpenID+groups[2].Player.OpenID] = 1

					// 生成对局信息
					roundInfo := &RoundInfo{
						Home: []*model.Player{
							groups[0].Player,
							groups[1].Player,
						},
						HomeAvg: GetPlayersAvgLevel([]*model.Player{
							groups[0].Player,
							groups[1].Player,
						}),
						Away: []*model.Player{
							groups[2].Player,
							groups[3].Player,
						},
						AwayAvg: GetPlayersAvgLevel([]*model.Player{
							groups[2].Player,
							groups[3].Player,
						}),
						CourtNum: 1,
						RoundNum: roundNum,
					}

					// append进场次信息
					roundInfos = append(roundInfos, roundInfo)
					stopFlag = 1
					break
				}
				// 如果3个组合都不满足要求，需要回归到
				if times >= 3 {
					goto retry
				}
				times += 1
				util.Logger.Printf("cycle, times:%d", times)
			}
		}
	}
	return roundInfos

}

// 遍历countMap，记录参与次数和expectation差距最大的用户，这些用户必须存在于下一轮的比赛中
// 如果不足4人，则回溯寻找差距第二大的用户，用来填满一轮需要的四位玩家，填满为止

//tmp := make([]string, 0)
//for openID, times := range countMap {
//	if len(tmp) == 4 {
//		// 检查当前4个选手的组合，如果不行，需要重新从countMap获取
//		for _, t := range tmp {
//
//		}
//	}
//	tmp = append(tmp, openID)
//}

// 找到满足要求的4个玩家后，进行随机俩俩组合，遍历groupDedupMap，一组的用户必须是首次组队
// 记录找到的两个组合，构建roundInfo
//}

//}

// divideMexicano pedal的Mexicano规则
func divideMexicano(event *mysql.EventMeta, players []*model.Player) []*RoundInfo {
	return nil
}

// ShufflePlayers 对运动员进行随机打散顺序
func ShufflePlayers(players []*model.Player) {
	ran := rand2.New(rand2.NewSource(time.Now().UnixNano()))
	ran.Shuffle(len(players), func(i, j int) {
		players[i], players[j] = players[j], players[i]
	})
}

func ShuffleCandidates(candidates []*Sortable) {
	ran := rand2.New(rand2.NewSource(time.Now().UnixNano()))
	ran.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})
}

func GetPlayersAvgLevel(players []*model.Player) float32 {
	res := float32(0)
	for _, player := range players {
		res += player.Level
	}
	return res / float32(len(players))
}

// idx 0 & idx 1 VS idx 2 & idx 3
func getTwoGroupsFrom4People(sortable []*Sortable) [][]*Sortable {
	res := make([][]*Sortable, 3)
	res[0] = []*Sortable{
		{
			Player: sortable[0].Player,
			Count:  sortable[0].Count,
		},
		{
			Player: sortable[1].Player,
			Count:  sortable[1].Count,
		},
		{
			Player: sortable[2].Player,
			Count:  sortable[2].Count,
		},
		{
			Player: sortable[3].Player,
			Count:  sortable[3].Count,
		},
	}
	res[1] = []*Sortable{
		{
			Player: sortable[0].Player,
			Count:  sortable[0].Count,
		},
		{
			Player: sortable[2].Player,
			Count:  sortable[2].Count,
		},
		{
			Player: sortable[1].Player,
			Count:  sortable[1].Count,
		},
		{
			Player: sortable[3].Player,
			Count:  sortable[3].Count,
		},
	}

	res[2] = []*Sortable{
		{
			Player: sortable[0].Player,
			Count:  sortable[0].Count,
		},
		{
			Player: sortable[3].Player,
			Count:  sortable[3].Count,
		},
		{
			Player: sortable[1].Player,
			Count:  sortable[1].Count,
		},
		{
			Player: sortable[2].Player,
			Count:  sortable[2].Count,
		},
	}

	return res
}

func checkGroupDedup(player1, player2 *model.Player, dedupMap map[string]int) bool {
	if dedupMap[player1.OpenID+player2.OpenID] == 1 || dedupMap[player2.OpenID+player1.OpenID] == 1 {
		return false
	}
	return true
}
