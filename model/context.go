package model

import (
	"github.com/gin-gonic/gin"
	"math/rand"
)

type TeamUpContext struct {
	*gin.Context // 默认Gin的上下文

	AppInfo     *AppInfo   `json:"app_info"`
	BasicUser   *BasicUser `json:"basic_user"`
	AccessToken string     `json:"access_token"`
	ID          int        `json:"id"`
	Rand        *rand.Rand `json:"rand"`
	Timestamp   int64      `json:"timestamp"`
	Language    string     `json:"language"`
}
