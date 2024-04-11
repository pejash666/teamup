package handler

import (
	"teamup/constant"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/util"
)

type GetHostInfoResp struct {
	ErrNo   int32            `json:"err_no"`
	ErrTips string           `json:"err_tips"`
	Data    map[string]*Info `json:"data"`
}

type Info struct {
	IsHost         bool  `json:"is_host"`
	OrganizationID int64 `json:"organization_id"`
}

// GetHostInfo godoc
//
//	@Summary		用户组织信息
//	@Description	用户在不同运动类型下是否为“组织”身份
//	@Tags			/team_up/user
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	GetHostInfoResp
//	@Router			/team_up/user/get_host_info [get]
func GetHostInfo(c *model.TeamUpContext) (interface{}, error) {
	var users []*mysql.WechatUserInfo
	result := util.DB().Where("open_id = ? AND sport_type IN ?", c.BasicUser.OpenID, []string{constant.SportTypePadel, constant.SportTypeTennis, constant.SportTypePickelBall}).Find(&users)
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
