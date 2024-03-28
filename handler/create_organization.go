package handler

import (
	"errors"
	"gorm.io/gorm"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/util"
)

type CreateOrganizationBody struct {
	Name      string `json:"name"`       // 组织名称
	SportType string `json:"sport_type"` // 运动类型
	City      string `json:"city"`       // 城市
	Address   string `json:"address"`    // 地址
	Contact   string `json:"contact"`    // 联系方式
	Logo      string `json:"logo"`       // 组织logo
}

// CreateOrganization godoc
//
//	@Summary		创建组织
//	@Description	用户上传组织信息
//	@Tags			/team_up/organization
//	@Accept			json
//	@Produce		json
//	@Param			name		body		string	true	"组织名称"
//	@Param			sport_type	body		string	true	"运动类型"
//	@Param			city		body		string	true	"城市"
//	@Param			address		body		string	true	"地址"
//	@Param			contact		body		string	true	"联系方式"
//	@Param			logo		body		string	true	"logo图"
//	@Success		200			{object}	model.BackEndResp
//	@Router			/team_up/organization/create [post]
func CreateOrganization(c *model.TeamUpContext) (interface{}, error) {
	util.Logger.Printf("[CreateOrganization] starts")
	// 一个open_id在一个sport_type下只能有一个organization
	organization := &mysql.Organization{}
	err := util.DB().Where("host_open_id = ? AND sport_type = ?", c.BasicUser.OpenID, c.PostForm("sport_type")).Take(organization).Error
	if err != nil {
		// 没有的话创建一个新的, 先将基础信息写入mysql
		if errors.Is(err, gorm.ErrRecordNotFound) {
			body := &CreateOrganizationBody{}
			err = c.BindJSON(body)
			if err != nil {
				util.Logger.Printf("[CreateOrganization] BindJSON failed, err:%v", err)
				return nil, iface.NewBackEndError(iface.ParamsError, "invalid params")
			}
			organization.Name = body.Name
			organization.HostOpenID = c.BasicUser.OpenID
			organization.SportType = body.SportType
			organization.City = body.City
			organization.Address = body.Address
			organization.Contact = body.Contact
			organization.Logo = body.Logo
			//// 将logo资源存在服务器内
			//file, err := c.FormFile("logo")
			//if err != nil {
			//	util.Logger.Printf("[CreateOrganization] FormFile failed, err:%v", err)
			//	return nil, iface.NewBackEndError(iface.ParamsError, "invalid logo")
			//}
			//// 不能大于1mb
			//if file.Size > 1<<20 {
			//	util.Logger.Printf("[CreateOrganization] file size is too big")
			//	return nil, iface.NewBackEndError(iface.ParamsError, "file too big")
			//}
			//fileName := strings.Split(file.Filename, ".")
			//if fileName[len(fileName)-1] != "png" && fileName[len(fileName)-1] != "jpeg" {
			//	util.Logger.Printf("[CreateOrganization] invalid file, should either png or jpeg")
			//	return nil, iface.NewBackEndError(iface.ParamsError, "invalid filename")
			//}
			//dst := path.Join("./organization_logos", strconv.FormatInt(int64(organization.ID), 10)+"_logo."+fileName[len(fileName)-1])
			//err = c.SaveUploadedFile(file, dst)
			//if err != nil {
			//	util.Logger.Printf("[CreateOrganization] iSaveUploadedFile failed, err:%v", err)
			//	return nil, iface.NewBackEndError(iface.ParamsError, "save file failed")
			//}
			//organization.Name = c.PostForm("name")
			//organization.HostOpenID = c.BasicUser.OpenID
			//organization.SportType = c.PostForm("sport_type")
			//organization.City = c.PostForm("city")
			//organization.Address = c.PostForm("address")
			//organization.Contact = c.PostForm("contact")
			//imagePath := "/organization_logos/" + strconv.FormatInt(int64(organization.ID), 10) + "_logo." + fileName[len(fileName)-1]
			//organization.Logo = util.GetImageUrl(c, imagePath)
			result := util.DB().Create(organization)
			if result.Error != nil {
				util.Logger.Printf("[CreateOrganization] DB().Create failed, err:%v", result.Error)
				return nil, iface.NewBackEndError(iface.MysqlError, "create record failed")
			}
			util.Logger.Printf("[CreateOrganization] success")
			return nil, nil
		} else {
			util.Logger.Printf("[CreateOrganization] get record from DB failed, err:%v", err)
			return nil, iface.NewBackEndError(iface.MysqlError, "get record failed")
		}
	}
	return nil, iface.NewBackEndError(iface.ParamsError, "user already have organization in this sport type")
}
