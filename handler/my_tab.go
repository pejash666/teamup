package handler

import "teamup/model"

func GetMyTab(c *model.TeamUpContext) (interface{}, error) {
	// 此时已经登录
	c.BindJSON()
}
