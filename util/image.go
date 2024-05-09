package util

import (
	"fmt"
	"teamup/model"
)

const (
	// todo: 替换为线上的网址
	LocalTestSchema = "localhost:8080/team_up/static_image/%s/%s"
	ImageUrlSchema  = "https://www.teamupup.cn/team_up/user/image/%s/%s"
)

// GetImageUrl 获取存储在服务器的静态资源链接
func GetImageUrl(c *model.TeamUpContext, imageType, imageName string) string {
	// 将域名组 与 图片路径进行拼接
	return fmt.Sprintf(ImageUrlSchema, imageType, imageName)
}
