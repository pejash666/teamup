package handler

import (
	"github.com/bytedance/sonic"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/util"
)

// UpdateScorerRequest model info
// @Description UpdateScorerRequest model info
type UpdateScorerRequest struct {
	EventID int64    `json:"event_id"`
	OpenIDs []string `json:"open_ids"`
}

// UpdateScorer 用户可以配置可记分的用户
//
//	@Summary		更新活动的记分员
//	@Description	更新活动的记分员
//	@Tags			/team_up/event
//	@Accept			json
//	@Produce		json
//	@Param			code	body		string	true	"详见UpdateScorerRequest定义"
//	@Success		200		{object}	model.BackEndResp
//	@Router			/team_up/event/update_scorer [post]
func UpdateScorer(c *model.TeamUpContext) (interface{}, error) {
	body := &UpdateScorerRequest{}
	err := c.BindJSON(body)
	if err != nil {
		util.Logger.Printf("[UpdateScorer] bindJSON failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.ParamsError, "invalid request")
	}
	if len(body.OpenIDs) < 1 {
		util.Logger.Printf("[UpdateScorer] openIDs length < 1")
		return nil, iface.NewBackEndError(iface.ParamsError, "invalid openIDs")
	}
	// 更新event 中的scorer
	event := &mysql.EventMeta{}
	err = util.DB().Where("id = ?", body.EventID).Take(event).Error
	if err != nil {
		util.Logger.Printf("[UpdateScorer] query record failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, "query record failed")
	}
	openIDs, err := sonic.MarshalString(body.OpenIDs)
	if err != nil {
		util.Logger.Printf("[UpdateScorer] marshalString failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.InternalError, "marshal failed")
	}
	event.Scorers = openIDs
	err = util.DB().Save(event).Error
	if err != nil {
		util.Logger.Printf("[UpdateScorer] save failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, "save record failed")
	}

	util.Logger.Printf("[UpdateScorer] success")
	return nil, nil
}
