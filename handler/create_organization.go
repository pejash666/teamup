package handler

import (
	"errors"
	"gorm.io/gorm"
	"path"
	"strconv"
	"strings"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/util"
)

func CreateOrganization(c *model.TeamUpContext) (interface{}, error) {
	util.Logger.Printf("[CreateOrganization] starts")
	// 一个open_id在一个sport_type下只能有一个organization
	organization := &mysql.Organization{}
	err := util.DB().Where("open_id = ? AND sport_type = ?", c.BasicUser.OpenID, c.PostForm("sport_type")).Take(organization).Error
	if err != nil {
		// 没有的话创建一个新的, 先将基础信息写入mysql
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 将logo资源存在服务器内
			file, err := c.FormFile("logo")
			if err != nil {
				util.Logger.Printf("[CreateOrganization] FormFile failed, err:%v", err)
				return nil, iface.NewBackEndError(iface.ParamsError, "invalid file")
			}
			// 不能大于1mb
			if file.Size > 1<<20 {
				util.Logger.Printf("[CreateOrganization] file size is too big")
				return nil, iface.NewBackEndError(iface.ParamsError, "file too big")
			}
			fileName := strings.Split(file.Filename, ".")
			if fileName[len(fileName)-1] != "png" || fileName[len(fileName)-1] != "jpeg" {
				util.Logger.Printf("[CreateOrganization] invalid file, should either png or jpeg")
				return nil, iface.NewBackEndError(iface.ParamsError, "invalid filename")
			}
			dst := path.Join("./organization_logos", strconv.FormatInt(int64(organization.ID), 10)+fileName[len(fileName)-1])
			err = c.SaveUploadedFile(file, dst)
			if err != nil {
				util.Logger.Printf("[CreateOrganization] iSaveUploadedFile failed, err:%v", err)
				return nil, iface.NewBackEndError(iface.ParamsError, "save file failed")
			}
			organization.Name = c.PostForm("name")
			organization.SportType = c.PostForm("sport_type")
			organization.City = c.PostForm("city")
			organization.Address = c.PostForm("address")
			organization.Contact = c.PostForm("contact")
			organization.Logo = dst
			result := util.DB().Create(organization)
			if result.Error != nil {
				util.Logger.Printf("[CreateOrganization] DB().Create failed, err:%v", result.Error)
				return nil, iface.NewBackEndError(iface.MysqlError, "create record failed")
			}
			util.Logger.Printf("[CreateOrganization] success")
		} else {
			util.Logger.Printf("[CreateOrganization] get record from DB failed, err:%v", err)
			return nil, iface.NewBackEndError(iface.MysqlError, "get record failed")
		}
	}
	return nil, iface.NewBackEndError(iface.InternalError, "create failed")
}
