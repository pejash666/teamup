package util

import (
	"github.com/gin-gonic/gin"
	"teamup/model"
)

type DefaultMiddlewareChecker func(c *gin.Context, opt model.APIOption) bool
