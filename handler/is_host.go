package handler

import (
	"teamup/constant"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/util"
)

func GetUserHostInfo(c *model.TeamUpContext) (interface{}, error) {
	var users []*mysql.WechatUserInfo
	result := util.DB().Where("open_id = ? AND sport_type IN ?", c.BasicUser.OpenID, []string{constant.SportTypePedal, constant.SportTypeTennis, constant.SportTypePickelBall}).Find(&users)
	if result.Error != nil {
		util.Logger.Printf("[IsUserHostBySportType] DB select failed, err:%v", result.Error)
		return nil, iface.NewBackEndError(iface.MysqlError, result.Error.Error())
	}
	res := make(map[string]interface{})
	type info struct {
		IsHost         bool  `json:"is_host"`
		OrganizationID int64 `json:"organization_id"`
	}
	for _, user := range users {
		tmp := user
		res[tmp.SportType] = &info{
			IsHost:         user.IsHost == 1,
			OrganizationID: int64(user.OrganizationID),
		}
	}
	util.Logger.Printf("[IsUserHostBySportType] success, res:%+v", res)
	return res, nil
}
