package handler

import (
	"fmt"
	"math"
	"sort"
	"teamup/constant"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/util"
)

type PlayerAfterMatch struct {
	*model.Player
	LevelChange    float64 `json:"level_change,omitempty"` // 等级的变化,不对外返回
	LevelChangeStr string  `json:"level_change_str"`       // 对外返回的等级变化
	TotalScore     int32   `json:"total_score"`            // 总得分
	Rank           int32   `json:"rank"`                   // 排名
	WinRound       int32   `json:"win_round"`
	TieRound       int32   `json:"tie_round"`
	LoseRound      int32   `json:"lose_round"`
}

type MatchResult struct {
	PlayersDetail []*PlayerAfterMatch `json:"players_detail"`
	RoundDetail   []*UploadRoundInfo  `json:"round_detail"`
}

// UploadRoundInfos model info
// @Description 上传轮次信息
type UploadRoundInfos struct {
	EventID   int                `json:"event_id"` // 活动ID
	RoundInfo []*UploadRoundInfo `json:"upload_round_info"`
}

type UploadRoundInfo struct {
	Home            []*model.Player `json:"home"`       // 主队球员
	HomeAvg         float32         `json:"home_avg"`   // 主队平均分
	AwayAvg         float32         `json:"away_avg"`   // 客队平均分
	Away            []*model.Player `json:"away"`       // 客队球员
	CourtNum        int32           `json:"court_num"`  // 场地号
	RoundNum        int32           `json:"round_num"`  // 轮次数
	HomeScore       int32           `json:"home_score"` // 主队本局分数
	AwayScore       int32           `json:"away_score"` // 客队本局分数
	Winner          string          `json:"winner"`     // 获胜者
	HomeLevelChange float64         `json:"home_level_change"`
	AwayLevelChange float64         `json:"away_level_change"`
}

type GetScoreResultResp struct {
	ErrNo   int32        `json:"err_no"`
	ErrTips string       `json:"err_tips"`
	Data    *MatchResult `json:"data"`
}

// GetScoreResult godoc
//
//	@Summary		获取记分结果
//	@Description	用户上传分数信息，服务端计算用户等级变化
//	@Tags			/team_up/user
//	@Accept			json
//	@Produce		json
//	@Param			get_score_result	body		string	true	"参考UploadRoundInfos"
//	@Success		200					{object}	GetScoreResultResp
//	@Router			/team_up/user/get_score_result [post]
func GetScoreResult(c *model.TeamUpContext) (interface{}, error) {
	body := &UploadRoundInfos{}
	err := c.BindJSON(body)
	if err != nil {
		util.Logger.Printf("[GetScoreResult] bindJSON failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.ParamsError, "bindJSON failed")
	}
	roundInfos := body.RoundInfo
	if len(roundInfos) < 1 {
		return nil, iface.NewBackEndError(iface.ParamsError, "params error, please check")
	}
	event := &mysql.EventMeta{}
	// 获取活动信息
	err = util.DB().Where("id = ?", body.EventID).Take(event).Error
	if err != nil {
		util.Logger.Printf("[GetScoreResult] query event from DB failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, "query record failed")
	}
	res := &MatchResult{
		PlayersDetail: make([]*PlayerAfterMatch, 0),
		RoundDetail:   make([]*UploadRoundInfo, 0),
	}

	playerMap := make(map[string]*PlayerAfterMatch)
	playerSlice := make([]*PlayerAfterMatch, 0)
	roundSlice := make([]*UploadRoundInfo, 0)
	isCompetitive := event.MatchType == "competitive"

	// 第一次遍历首先将player都填充好
	for _, round := range roundInfos {
		roundTmp := *round
		for _, player := range roundTmp.Home {
			// map中不存在 则放入map
			if _, ok := playerMap[player.OpenID]; !ok {
				tmp := &PlayerAfterMatch{
					Player: player,
				}
				playerMap[player.OpenID] = tmp
			}
			if roundTmp.Winner == "home" {
				playerMap[player.OpenID].WinRound += 1
			} else if roundTmp.Winner == "away" {
				playerMap[player.OpenID].LoseRound += 1
			}
			playerMap[player.OpenID].TotalScore += roundTmp.HomeScore
			// 如果是竞技类比赛需要计算等级的变化
			if isCompetitive {
				// 如果是pedal americano 或者 pickleball的单打，levelChange是每个人都不一样的
				// 其他方式，levelChange是每个队伍之间不一样，队内是一样的
				// 对于用户的前5场比赛，factor需要大一些，能够更快的校准
				factor := 0.1
				needTeamLevelChange := true
				if (event.SportType == constant.SportTypePedal && event.ScoreRule == constant.PedalScoreRuleAmericano) || (event.SportType == constant.SportTypePickelBall && event.GameType == constant.EventGameTypeSolo) {
					needTeamLevelChange = false
					ppl := &mysql.WechatUserInfo{}
					err = util.DB().Where("open_id = ? AND sport_type = ?", player.OpenID, event.SportType).Take(ppl).Error
					if err != nil {
						util.Logger.Printf("[GetScoreResult] query player from DB failed, err:%v", err)
						return nil, iface.NewBackEndError(iface.MysqlError, "query player failed")
					}
					if ppl.JoinedTimes < 5 {
						factor = 0.5
					}
				}
				levelChange := GetLevelChange("home", roundTmp.Winner, roundTmp.HomeAvg, roundTmp.AwayAvg, factor)
				playerMap[player.OpenID].LevelChange += levelChange
				if needTeamLevelChange {
					roundTmp.HomeLevelChange = levelChange
				}
			}
		}
		for _, player := range roundTmp.Away {
			if _, ok := playerMap[player.OpenID]; !ok {
				tmp := &PlayerAfterMatch{
					Player: player,
				}
				playerMap[player.OpenID] = tmp
			}
			if roundTmp.Winner == "away" {
				playerMap[player.OpenID].WinRound += 1
			} else if roundTmp.Winner == "home" {
				playerMap[player.OpenID].LoseRound += 1
			}
			playerMap[player.OpenID].TotalScore += roundTmp.AwayScore
			// 如果是竞技类比赛需要计算等级的变化
			if isCompetitive {
				factor := 0.1
				needTeamLevelChange := true
				if (event.SportType == constant.SportTypePedal && event.ScoreRule == constant.PedalScoreRuleAmericano) || (event.SportType == constant.SportTypePickelBall && event.GameType == constant.EventGameTypeSolo) {
					needTeamLevelChange = false
					ppl := &mysql.WechatUserInfo{}
					err = util.DB().Where("open_id = ? AND sport_type = ?", player.OpenID, event.SportType).Take(ppl).Error
					if err != nil {
						util.Logger.Printf("[GetScoreResult] query player from DB failed, err:%v", err)
						return nil, iface.NewBackEndError(iface.MysqlError, "query player failed")
					}
					if ppl.JoinedTimes < 5 {
						factor = 0.5
					}
				}
				levelChange := GetLevelChange("away", roundTmp.Winner, roundTmp.HomeAvg, roundTmp.AwayAvg, factor)
				playerMap[player.OpenID].LevelChange += levelChange
				if needTeamLevelChange {
					roundTmp.AwayLevelChange = levelChange
				}
			}
		}
		roundSlice = append(roundSlice, &roundTmp)
	}
	for _, player := range playerMap {
		player.LevelChangeStr = fmt.Sprintf("%.3f", player.LevelChange)
		player.LevelChange = 0 // 精度问题不返回
		playerSlice = append(playerSlice, player)
	}
	// 对player进行排序, 排名越高的越往上
	sort.Slice(playerSlice, func(i, j int) bool {
		return playerSlice[i].TotalScore > playerSlice[j].TotalScore
	})
	rank := int32(1)
	for _, player := range playerSlice {
		player.Rank = rank
		rank += 1
	}

	res.RoundDetail = roundSlice
	res.PlayersDetail = playerSlice

	util.Logger.Printf("[GetScoreResult] success, res:%v", res)
	return res, nil
}

func GetLevelChange(team string, winner string, homeAvg, awayAvg float32, factor float64) float64 {
	tmp := float64(0)
	if winner == "home" {
		tmp = math.Pow(float64(10), float64(1*(awayAvg-homeAvg)))
	} else {
		tmp = math.Pow(float64(10), float64(1*(homeAvg-awayAvg)))
	}
	possibility := 1 / (1 + tmp)
	res := math.Trunc((1-possibility)*factor*math.Pow10(3)) / math.Pow10(3)
	if team == winner {
		return res
	} else {
		return -res
	}
}
