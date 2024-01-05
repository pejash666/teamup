package constant

const (
	EventStatusDraft    = "draft"    // 草稿模式
	EventStatusCreated  = "created"  // 等待开始 （创建后的初始状态）
	EventStatusFull     = "full"     // 已经满员
	EventStatusFinished = "finished" // 已结束
	EventStatusDeleted  = "deleted"

	EventMatchTypeEntertainment = "entertainment" // 休闲活动
	EventMatchTypeCompetitive   = "competitive"

	EventGameTypeDuo  = "duo"  // 双打
	EventGameTypeSolo = "solo" // 单打
)
