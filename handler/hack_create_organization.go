package handler

import (
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/util"
)

func HackCreateOrganization(c *model.TeamUpContext) (interface{}, error) {
	util.Logger.Printf("[HackCreateOrganization] starts")
	body := &CreateOrganizationBody{}
	err := c.BindJSON(body)
	if err != nil {
		return nil, iface.NewBackEndError(iface.ParamsError, err.Error())
	}
	util.Logger.Printf("[HackCreateOrganization] body:%v", util.ToReadable(body))
	orga := &mysql.Organization{
		SportType:  body.SportType,
		Name:       body.Name,
		City:       body.City,
		Address:    body.Address,
		Longitude:  body.Longitude,
		Latitude:   body.Latitude,
		IsApproved: 1,
	}
	err = util.DB().Create(orga).Error
	if err != nil {
		return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
	}
	util.Logger.Printf("[HackCreateOrganization] success")
	return nil, nil
}
