package handler

import (
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/util"
)

func HackCreateOrganization(c *model.TeamUpContext) (interface{}, error) {
	body := &CreateOrganizationBody{}
	err := c.BindJSON(body)
	if err != nil {
		return nil, iface.NewBackEndError(iface.ParamsError, err.Error())
	}
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
	return nil, nil
}
