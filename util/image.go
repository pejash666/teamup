package util

import (
	"teamup/model"
)

const (
	// todo: 替换为线上的网址
	NetPrefix = "localhost:8080/team_up/static_image"
)

// GetImageUrl 获取存储在服务器的静态资源链接
func GetImageUrl(c *model.TeamUpContext, loc string) string {
	// 将域名组 与 图片路径进行拼接
	return NetPrefix + loc
}
