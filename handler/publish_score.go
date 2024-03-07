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

type PublishScoreBody struct {
	EventID       int                 `json:"event_id"`
	PlayersDetail []*PlayerAfterMatch `json:"players_detail"`
}

// PublishScore godoc
//
//	@Summary		发布比赛结果
//	@Description	根据计算出的等级变化，服务端更新场次，用户的信息
//	@Tags			/teamup/user
//	@Accept			json
//	@Produce		json
//	@Param			publish_score	body		{object}	PublishScoreBody	true	"比赛结果"
//	@Success		200				{object}	model.BackEndResp
//	@Router			/teamup/user/publish_score [post]
func PublishScore(c *model.TeamUpContext) (interface{}, error) {
	body := &PublishScoreBody{}
	err := c.BindJSON(body)
	if err != nil {
		util.Logger.Printf("[PublishScore] bindJSON failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.ParamsError, "invalid params")
	}
	// 获取sport_type
	event := &mysql.EventMeta{}
	err = util.DB().Where("id = ?", body.EventID).Take(event).Error
	if err != nil {
		util.Logger.Printf("[PublishScore] query event failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, "query record failed")
	}
	// 更新每个用户的分数记录，增加一条分数变化记录，更新活动的状态
	util.DB().Transaction(func(tx *gorm.DB) error {
		for _, player := range body.PlayersDetail {
			lc, err := strconv.ParseFloat(player.LevelChangeStr, 32)
			if err != nil {
				return err
			}
			user := &mysql.WechatUserInfo{}
			err = tx.Where("open_id = ? AND sport_type = ?", player.OpenID, event.SportType).Take(user).Error
			if err != nil {
				return err
			}
			user.Level += int(lc * 1000)
			err = tx.Save(user).Error
			if err != nil {
				return err
			}
			userEvent := &mysql.UserEvent{}
			err = tx.Where("open_id = ? AND sport_type = ? AND event_id = ? AND is_quit = 0", player.OpenID, event.SportType, event.ID).Take(userEvent).Error
			if err != nil {
				return err
			}
			userEvent.IsIncrease = uint(util.BoolToDB(player.LevelChangeStr > "0"))
			userEvent.LevelChange = int(lc * 1000)
			err = util.DB().Save(userEvent).Error
			if err != nil {
				return err
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

	util.Logger.Printf("[PublishScore] success")
	return nil, nil
}
