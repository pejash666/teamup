package handler

import (
	"teamup/iface"
	"teamup/model"
	"teamup/util"
)

type PlayerAfterMatch struct {
	*model.Player
	LevelChange float32 `json:"level_change"` // 等级的变化
	TotalScore  int32   `json:"total_score"`  // 总得分
	Rank        int32   `json:"rank"`         // 排名
	WinRound    int32   `json:"win_round"`
	TieRound    int32   `json:"tie_round"`
	LoseRound   int32   `json:"lose_round"`
}

type MatchResult struct {
	PlayersDetail []*PlayerAfterMatch `json:"players_detail"`
	RoundDetail   []*UploadRoundInfo  `json:"round_detail"`
}

type UploadRoundInfos struct {
	RoundInfo []*UploadRoundInfo `json:"upload_round_info"`
}

type UploadRoundInfo struct {
	Home            []*model.Player `json:"home"`              // 主队球员
	HomeAvg         float32         `json:"home_avg"`          // 主队平均分
	AwayAvg         float32         `json:"away_avg"`          // 客队平均分
	Away            []*model.Player `json:"away"`              // 客队球员
	CourtNum        int32           `json:"court_num"`         // 场地号
	RoundNum        int32           `json:"round_num"`         // 轮次数
	HomeScore       int32           `json:"home_score"`        // 主队本局分数
	AwayScore       int32           `json:"away_score"`        // 客队本局分数
	Winner          string          `json:"winner"`            // 获胜者
	HomeLevelChange float32         `json:"home_level_change"` //主队等级变化
	AwayLevelChange float32         `json:"away_level_change"` // 客队等级变化
}

func UploadScore(c *model.TeamUpContext) (interface{}, error) {
	body := &UploadRoundInfos{}
	err := c.BindJSON(body)
	if err != nil {
		util.Logger.Printf("[UploadScore] bindJSON failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.ParamsError, "bindJSON failed")
	}
	roundInfos := body.RoundInfo

	res := &MatchResult{
		PlayersDetail: make([]*PlayerAfterMatch, 0),
		RoundDetail:   make([]*UploadRoundInfo, 0),
	}

	playerMap := make(map[string]*PlayerAfterMatch)

	// 第一次遍历首先将player都填充好
	for _, round := range roundInfos {
		for _, player := range round.Home {
			if playerMap[player.OpenID] == nil {
				tmp := &PlayerAfterMatch{
					Player: player,
				}
				playerMap[player.OpenID] = tmp
			} else {
				if round.Winner == "home" {
					playerMap[player.OpenID].WinRound += 1
				} else if round.Winner == "away" {
					playerMap[player.OpenID].LoseRound += 1
				}
				playerMap[player.OpenID].TotalScore += round.HomeScore
			}
		}
		for _, player := range round.Away {
			if playerMap[player.OpenID] == nil {
				tmp := &PlayerAfterMatch{
					Player: player,
				}
				playerMap[player.OpenID] = tmp
			} else {
				if round.Winner == "away" {
					playerMap[player.OpenID].WinRound += 1
				} else if round.Winner == "home" {
					playerMap[player.OpenID].LoseRound += 1
				}
				playerMap[player.OpenID].TotalScore += round.AwayScore
			}
		}
		// 计算level的变化

	}

	// 第二次遍历计算level的变化

}

func GetLevelChange(homeAvg, awayAvg, factor float32)
