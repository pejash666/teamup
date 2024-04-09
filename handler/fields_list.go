package handler

import (
	"errors"
	"gorm.io/gorm"
	"strconv"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/util"
)

type GetFieldsListBody struct {
	SportType string `json:"sport_type"`
	City      string `json:"city"`
}

type GetFieldsListResp struct {
	ErrNo   int32             `json:"err_no"`
	ErrTips string            `json:"err_tips"`
	Data    *GetFieldsListRes `json:"data"`
}

type GetFieldsListRes struct {
	FieldsList []Organization `json:"fields_list"`
}

// GetFieldsList godoc
//
//	@Summary		获取场地列表
//	@Description	根据运动类型和城市获取场地列表
//	@Tags			/team_up/event
//	@Accept			json
//	@Produce		json
//	@Param			city		body		string	true	"城市"
//	@Param			sport_type	body		string	true	"运动类型"
//	@Success		200			{object}	GetFieldsListResp
//	@Router			/team_up/event/fields_list [post]
func GetFieldsList(c *model.TeamUpContext) (interface{}, error) {
	body := &GetFieldsListBody{}
	err := c.BindJSON(body)
	if err != nil {
		return nil, iface.NewBackEndError(iface.ParamsError, err.Error())
	}
	if body.SportType == "" || body.City == "" {
		return nil, iface.NewBackEndError(iface.ParamsError, "invalid params")
	}
	var organizations []mysql.Organization
	// 只展示通过审核的
	err = util.DB().Where("sport_type = ? AND city = ? AND is_approved = 1", body.SportType, body.City).Find(&organizations).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return make([]Organization, 0), nil
		} else {
			return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
		}

	}
	res := make([]Organization, 0)
	for _, orga := range organizations {
		tmp := Organization{
			ID:       orga.ID,
			Name:     orga.Name,
			Logo:     orga.Logo,
			Address:  orga.Address,
			EventNum: orga.TotalEventNum,
			Status:   "approved",
		}
		latitude, err := strconv.ParseFloat(orga.Latitude, 64)
		if err != nil {
			return nil, iface.NewBackEndError(iface.InternalError, err.Error())
		}
		longitude, err := strconv.ParseFloat(orga.Longitude, 64)
		if err != nil {
			return nil, iface.NewBackEndError(iface.InternalError, err.Error())
		}
		tmp.Longitude = longitude
		tmp.Latitude = latitude
		res = append(res, tmp)
	}

	return res, nil
}
