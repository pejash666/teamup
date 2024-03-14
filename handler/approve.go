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

type OrganizationApprovalBody struct {
	OrganizationID int64 `json:"organization_id"`
}

type CalibrationProofApprovalBody struct {
	OpenID    string `json:"open_id"`
	SportType string `json:"sport_type"`
}

// Approve godoc
//
//	@Summary		管理员审批
//	@Description	管理员审批：包含创建组织的申请与Pro级别的认证事件
//	@Tags			/team_up/admin
//	@Accept			json
//	@Produce		json
//	@Param			approve_type	query		string	true	"审批事件类型：organization 或者 calibration_proof"
//	@Param			organization_id	body		int		false	"组织ID"
//	@Param			open_id			body		string	false	"用户open_id"
//	@Param			sport_type		body		string	false	"用户运动类型"
//	@Success		200				{object}	model.BackEndResp
//	@Router			/team_up/admin/get_approval_items [get]
func Approve(c *model.TeamUpContext) (interface{}, error) {
	//  todo: 这里也需要增加前端页面
	approveType := c.Query("approve_type")
	switch approveType {
	case ApprovalTypeOrganization:
		body := &OrganizationApprovalBody{}
		err := c.BindJSON(body)
		if err != nil {
			util.Logger.Printf("[Approve] bindJson failed, err:%v", err)
			return nil, iface.NewBackEndError(iface.ParamsError, "invalid body")
		}
		// 开启事务，更新组织状态 + 更新用户host信息
		err = util.DB().Transaction(func(tx *gorm.DB) error {
			orga := &mysql.Organization{}
			err = tx.Where("id = ?", body.OrganizationID).Take(orga).Error
			if err != nil {
				return err
			}
			orga.IsApproved = 1
			orga.Reviewer = c.BasicUser.OpenID
			err = tx.Save(orga).Error
			if err != nil {
				return err
			}
			user := &mysql.WechatUserInfo{}
			err = tx.Where("open_id = ? AND sport_type = ?", orga.HostOpenID, orga.SportType).Take(user).Error
			if err != nil {
				return err
			}
			user.OrganizationID = int(orga.ID)
			user.IsHost = 1
			err = tx.Save(user).Error
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
		} else {
			util.Logger.Printf("[Approve] organization approve finished")
		}
	case ApproveTypeCalibrationProof:
		body := &CalibrationProofApprovalBody{}
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
	default:
		return nil, iface.NewBackEndError(iface.ParamsError, "invalid approve type")
	}
	return nil, nil
}
