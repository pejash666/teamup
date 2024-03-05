package model

type ScoreOptions struct {
	AvailableScoreRule   []string         `json:"available_score_rule"`   // 可选的记分赛制
	AvailableRoundTarget map[string][]int `json:"available_round_target"` // 可选的每轮目标分
	FieldNum             int              `json:"field_num"`              // 比赛场地数
}
