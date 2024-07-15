package handler

import (
	"gorm.io/gorm"
	"strconv"
	"teamup/constant"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/util"
)

// PublishScoreBody model info
//
//	@Description	发布分数
type PublishScoreBody struct {
	EventID         int                 `json:"event_id"`
	PlayersDetail   []*PlayerAfterMatch `json:"players_detail"`
	EventResultJson string              `json:"event_result_json"`
}

// PublishScore godoc
//
//	@Summary		发布比赛结果
//	@Description	根据计算出的等级变化，服务端更新场次，用户的信息
//	@Tags			/team_up/user
//	@Accept			json
//	@Produce		json
//	@Param			event_id		body		int		true	"活动ID"
//	@Param			player_detail	body		string	true	"用户详情, 参考PublishScoreBody"
//	@Success		200				{object}	model.BackEndResp
//	@Router			/team_up/user/publish_score [post]
func PublishScore(c *model.TeamUpContext) (interface{}, error) {
	body := &PublishScoreBody{}
	err := c.BindJSON(body)
	if err != nil {
		util.Logger.Printf("[PublishScore] bindJSON failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.ParamsError, "invalid params")
	}
	util.Logger.Printf("[PublishScore] starts, body:%v", util.ToReadable(body))
	// 获取sport_type
	event := &mysql.EventMeta{}
	err = util.DB().Where("id = ?", body.EventID).Take(event).Error
	if err != nil {
		util.Logger.Printf("[PublishScore] query event failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, "query record failed")
	}
	// 更新每个用户的分数记录，增加一条分数变化记录，更新活动的状态
	err = util.DB().Transaction(func(tx *gorm.DB) error {
		for _, player := range body.PlayersDetail {
			lc, errT := strconv.ParseFloat(player.LevelChangeStr, 64)
			if errT != nil {
				return errT
			}
			user := &mysql.WechatUserInfo{}
			errT = tx.Where("open_id = ? AND sport_type = ?", player.OpenID, event.SportType).Take(user).Error
			if errT != nil {
				return errT
			}
			user.Level += int(lc * 1000)
			errT = tx.Save(user).Error
			if errT != nil {
				return errT
			}
			userEvent := &mysql.UserEvent{}
			errT = tx.Where("open_id = ? AND sport_type = ? AND event_id = ? AND is_quit = 0", player.OpenID, event.SportType, event.ID).Take(userEvent).Error
			if errT != nil {
				return errT
			}
			userEvent.IsIncrease = uint(util.BoolToDB(lc > 0))
			userEvent.LevelSnapshot = user.Level
			userEvent.LevelChange = int(lc * 1000)
			userEvent.IsPublished = 1
			errT = util.DB().Save(userEvent).Error
			if errT != nil {
				return errT
			}
		}
		// 更新活动的状态
		event.Status = constant.EventStatusFinished
		err = tx.Save(event).Error
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		util.Logger.Printf("[PublishScore] update user failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
	}

	util.Logger.Printf("[PublishScore] success")
	return nil, nil
}
