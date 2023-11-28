package model

import (
	"github.com/gin-gonic/gin"
)

type TeamUpContext struct {
	*gin.Context // 默认Gin的上下文

	AppInfo     *AppInfo   `json:"app_info"`
	BasicUser   *BasicUser `json:"basic_user"`
	AccessToken string     `json:"access_token"`
	ID          int64      `json:"id"`
	Timestamp   int64      `json:"timestamp"`
	Language    string     `json:"language"`
}
