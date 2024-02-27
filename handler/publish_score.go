package handler

import (
	"gorm.io/gorm"
	"teamup/constant"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/util"
)

func PublishScore(c *model.TeamUpContext) (interface{}, error) {
	type Body struct {
		EventID       int                 `json:"event_id"`
		PlayersDetail []*PlayerAfterMatch `json:"players_detail"`
	}
	body := &Body{}
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
			err = util.DB().Model(&mysql.WechatUserInfo{}).Where("open_id = ? AND sport_type = ?", player.OpenID, event.SportType).Update("level", int(player.LevelChange*100)).Error
			if err != nil {
				return err
			}
			userEvent := &mysql.UserEvent{
				EventID:     event.ID,
				OpenID:      player.OpenID,
				SportType:   event.SportType,
				IsQuit:      0,
				IsIncrease:  uint(util.BoolToDB(player.LevelChange > 0)),
				LevelChange: int(player.LevelChange * 100),
			}
			err = util.DB().Save(userEvent).Error
			if err != nil {
				return err
			}
		}
		// 更新活动的状态
		event.Status = constant.EventStatusFinished
		err = util.DB().Save(event).Error
		if err != nil {
			return err
		}
		return nil
	})

	util.Logger.Printf("[PublishScore] success")
	return nil, nil
}
