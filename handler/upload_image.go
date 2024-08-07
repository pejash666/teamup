package handler

import (
	"fmt"
	"teamup/iface"
	"teamup/model"
	"teamup/util"
	"time"
)

const (
	ImageTypeOrganizationLogo = "organization_logo"
	ImageTypeCalibrationProof = "calibration_proof"
	ImageTypeEventImage       = "event_image"
	ImageTypeAvatar           = "avatar_image"
)

// UploadImage godoc
//
//	@Summary		前端上传文件流给服务端
//	@Description	前端获取加密的用户手机号，服务端进行解码，存储
//	@Tags			/team_up/user
//	@Accept			json
//	@Produce		json
//	@Param			image_type	formData	string	true	"图片类型:organization_logo, calibration_proof, event_image, avatar_image"
//	@Param			file		formData	file	true	"文件流"
//	@Success		200			{object}	model.BackEndResp
//	@Router			/team_up/user/upload_image [post]
func UploadImage(c *model.TeamUpContext) (interface{}, error) {
	imageType := c.Query("image_type")
	if imageType != ImageTypeOrganizationLogo && imageType != ImageTypeCalibrationProof && imageType != ImageTypeEventImage && imageType != ImageTypeAvatar {
		util.Logger.Printf("[UploadImage] invalid image type")
		return nil, iface.NewBackEndError(iface.ParamsError, "UploadImage不合法图片类型")
	}
	// 使用分布式锁，保证无并发问题
	get := util.GetLock(imageType, time.Second)
	defer util.DelLock(imageType)
	if !get {
		util.Logger.Printf("[UploadImage] concurrent request for upload file, please try again")
		return nil, iface.NewBackEndError(iface.ParamsError, "UploadImage请重试")
	}

	path := fmt.Sprintf("./%s", imageType)
	//// 执行系统命令获取当前的数量
	//command := cmd.NewCmd("bash", "-c", fmt.Sprintf("ls -l %s | grep \"^-\" | grep -c \"png$\"", path))
	//<-command.Start()
	//util.Logger.Printf("[UploadImage] command:%d", command.Status().Stdout)
	//currentNumStr := command.Status().Stdout[0]
	//currentNum, err := strconv.ParseInt(currentNumStr, 10, 64)
	//if err != nil {
	//	util.Logger.Printf("[UploadImage] ParseInt failed, err:%v", err)
	//	return nil, iface.NewBackEndError(iface.InternalError, "get image count failed")
	//}

	// 获取当前的图片名
	imageName := imageType + "_" + fmt.Sprintf("%d.png", time.Now().UnixNano()) // organization_logo_时间戳.png

	// 获取用户上传的图片
	file, err := c.FormFile("file")
	if err != nil {
		util.Logger.Printf("[UploadImage] FormFile failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.ParamsError, "UploadImage不合法的图片")
	}
	// 不能大于3mb
	if file.Size > 3<<20 {
		util.Logger.Printf("[UploadImage] file size is too big")
		return nil, iface.NewBackEndError(iface.ParamsError, "UploadImage图片过大")
	}
	filePath := fmt.Sprintf("%s/%s", path, imageName)

	err = c.SaveUploadedFile(file, filePath)
	if err != nil {
		util.Logger.Printf("[UploadImage] iSaveUploadedFile failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.ParamsError, "UploadImage保存图片失败")
	}
	// 返回给前端一个图片的url
	url := util.GetImageUrl(c, imageType, imageName)

	util.Logger.Printf("[UploadImage] success, image_url:%s", url)

	return map[string]interface{}{"data": url}, nil
}
