package handler

import (
	"errors"
	"gorm.io/gorm"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/util"
)

type GetApprovalItemsResp struct {
	ErrNo   int32                   `json:"err_no"`
	ErrTips string                  `json:"err_tips"`
	Data    *GetApprovalItemsResult `json:"data"`
}

type GetApprovalItemsResult struct {
	OrganizationItems []mysql.Organization   `json:"organization_items"`
	CalibrationItems  []mysql.WechatUserInfo `json:"calibration_items"`
}

// GetApprovalItems godoc
// @Summary      获取待审批的事件信息
// @Description  包含创建组织的申请与Pro级别的认证事件
// @Tags         /teamup/admin
// @Accept       json
// @Produce      json
// @Success      200  {object}  GetApprovalItemsResp
// @Router       /teamup/admin/get_approval_items [get]
func GetApprovalItems(c *model.TeamUpContext) (interface{}, error) {
	// 获取待审批事件
	res := &GetApprovalItemsResult{}
	var organizations []mysql.Organization
	// 获取未经审批的组织申请
	err := util.DB().Where("is_approved = 0").Find(&organizations).Error
	if err != nil {
		// 没查到话，res对应为空
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res.OrganizationItems = make([]mysql.Organization, 0)
		} else {
			util.Logger.Printf("[GetApprovalItems] find record failed, err:%v", err)
			return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
		}
	} else {
		res.OrganizationItems = organizations
	}
	// 获取自我认定为7.0职业水平，且未经校准的人
	var users []mysql.WechatUserInfo
	err = util.DB().Where("level >= 700 AND is_calibrated = 0").Find(&users).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res.CalibrationItems = make([]mysql.WechatUserInfo, 0)
		} else {
			util.Logger.Printf("[GetApprovalItems] find record failed, err:%v", err)
			return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
		}
	} else {
		res.CalibrationItems = users
	}

	util.Logger.Printf("[GetApprovalItems] success")
	return res, nil
}
