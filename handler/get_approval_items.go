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
	OrganizationItems []mysql.OrganizationWithoutGorm   `json:"organization_items"` // 待审批的组织列表
	CalibrationItems  []mysql.WechatUserInfoWithoutGorm `json:"calibration_items"`  // 待审批的定位为职业的用户列表
}

// GetApprovalItems godoc
//
//	@Summary		待审批事项清单
//	@Description	管理员审批：获取待审批的事项（包括定级别为pro级别的用户或者用户创建的组织信息）
//	@Tags			/team_up/admin
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	GetApprovalItemsResp
//	@Router			/team_up/admin/get_approval_items [get]
func GetApprovalItems(c *model.TeamUpContext) (interface{}, error) {
	// 获取待审批事件
	res := &GetApprovalItemsResult{}
	var organizations []mysql.Organization
	// 获取未经审批的组织申请
	err := util.DB().Where("is_approved = 0").Find(&organizations).Error
	if err != nil {
		// 没查到话，res对应为空
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res.OrganizationItems = make([]mysql.OrganizationWithoutGorm, 0)
		} else {
			util.Logger.Printf("[GetApprovalItems] find record failed, err:%v", err)
			return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
		}
	} else {
		res.OrganizationItems = organizationWithoutGorm(organizations)
	}
	// 获取自我认定为7.0职业水平，且未经校准的人
	var users []mysql.WechatUserInfo
	err = util.DB().Where("level >= 700 AND is_calibrated = 0").Find(&users).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res.CalibrationItems = make([]mysql.WechatUserInfoWithoutGorm, 0)
		} else {
			util.Logger.Printf("[GetApprovalItems] find record failed, err:%v", err)
			return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
		}
	} else {
		res.CalibrationItems = userWithoutGorm(users)
	}

	util.Logger.Printf("[GetApprovalItems] success")
	return res, nil
}

func userWithoutGorm(ogs []mysql.WechatUserInfo) []mysql.WechatUserInfoWithoutGorm {
	res := make([]mysql.WechatUserInfoWithoutGorm, 0)
	for _, og := range ogs {
		res = append(res, mysql.WechatUserInfoWithoutGorm{
			SportType:        og.SportType,
			IsCalibrated:     og.IsCalibrated,
			Level:            og.Level,
			Reviewer:         og.Reviewer,
			UnionId:          og.UnionId,
			OpenId:           og.OpenId,
			SessionKey:       og.SessionKey,
			IsPrimary:        og.IsPrimary,
			Nickname:         og.Nickname,
			IsHost:           og.IsHost,
			OrganizationID:   og.OrganizationID,
			Avatar:           og.Avatar,
			Gender:           og.Gender,
			PhoneNumber:      og.PhoneNumber,
			JoinedTimes:      og.JoinedTimes,
			JoinedEvent:      og.JoinedEvent,
			Preference:       og.Preference,
			Tags:             og.Tags,
			JoinedGroup:      og.JoinedGroup,
			CalibrationProof: og.CalibrationProof,
		})
	}
	return res
}

func organizationWithoutGorm(ogs []mysql.Organization) []mysql.OrganizationWithoutGorm {
	res := make([]mysql.OrganizationWithoutGorm, 0)
	for _, og := range ogs {
		res = append(res, mysql.OrganizationWithoutGorm{
			ID:            og.ID,
			SportType:     og.SportType,
			Name:          og.Name,
			City:          og.City,
			Address:       og.Address,
			HostOpenID:    og.HostOpenID,
			Contact:       og.Contact,
			Logo:          og.Logo,
			TotalEventNum: og.TotalEventNum,
			IsApproved:    og.IsApproved,
			Reviewer:      og.Reviewer,
		})
	}
	return res
}
