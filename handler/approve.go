package handler

import (
	"gorm.io/gorm"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/util"
)

const (
	ApproveTypeCalibrationProof = "calibration_proof"
	ApprovalTypeOrganization    = "organization"
)

func Approve(c *model.TeamUpContext) (interface{}, error) {
	//  todo: 这里也需要增加前端页面
	approveType := c.Query("approve_type")
	switch approveType {
	case ApprovalTypeOrganization:
		type Body struct {
			OrganizationID int `json:"organization_id"`
		}
		body := &Body{}
		err := c.BindJSON(body)
		if err != nil {
			util.Logger.Printf("[Approve] bindJson failed, err:%v", err)
			return nil, iface.NewBackEndError(iface.ParamsError, "invalid body")
		}
		// 开启事务，更新组织状态 + 更新用户host信息
		util.DB().Transaction(func(tx *gorm.DB) error {
			orga := &mysql.Organization{}
			err = util.DB().Where("id = ?", body.OrganizationID).Take(orga).Error
			if err != nil {
				return err
			}
			orga.IsApproved = 1
			orga.Reviewer = c.BasicUser.OpenID
			err = util.DB().Save(orga).Error
			if err != nil {
				return err
			}
			user := &mysql.WechatUserInfo{}
			err = util.DB().Where("open_id = ? AND sport_type = ?", orga.HostOpenID, orga.SportType).Take(user).Error
			if err != nil {
				return err
			}
			user.OrganizationID = int(orga.ID)
			user.IsHost = 1
			err = util.DB().Save(user).Error
			if err != nil {
				return err
			}
			return nil
		})
		util.Logger.Printf("[Approve] organization approve success")
	case ApproveTypeCalibrationProof:
		type Body struct {
			OpenID    string `json:"open_id"`
			SportType string `json:"sport_type"`
		}
		body := &Body{}
		err := c.BindJSON(body)
		if err != nil {
			util.Logger.Printf("[Approve] bindJson failed, err:%v", err)
			return nil, iface.NewBackEndError(iface.ParamsError, "invalid body")
		}
		// 更新用户信息
		user := &mysql.WechatUserInfo{}
		err = util.DB().Where("open_id = ? AND sport_type = ?", body.OpenID, body.SportType).Take(user).Error
		if err != nil {
			return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
		}
		user.IsCalibrated = 1
		user.Reviewer = c.BasicUser.OpenID
		err = util.DB().Save(user).Error
		if err != nil {
			return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
		}
		util.Logger.Printf("[Approve] calibration_proof approve success")
	}
	return nil, nil
}
