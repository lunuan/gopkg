package controller

import (
	"github.com/gin-gonic/gin"
)

type BaseController struct {
	GinContext *gin.Context
}

func NewBaseController(ctx *gin.Context) *BaseController {
	return &BaseController{GinContext: ctx}
}

func (c *BaseController) GetInt(k string, def int) int {
	if val, ok := c.GinContext.Get(k); ok && val != nil {
		i, _ := val.(int)
		return i
	}
	return def
}

func (c *BaseController) GetInt32(k string, def int32) int32 {
	if val, ok := c.GinContext.Get(k); ok && val != nil {
		if i, ok := val.(int32); ok {
			return i
		}
	}
	return def
}

func (c *BaseController) GetInt64(k string, def int64) int64 {
	if val, ok := c.GinContext.Get(k); ok && val != nil {
		if i, ok := val.(int64); ok {
			return i
		}
	}
	return def
}

func (c *BaseController) GetString(k string, def string) string {
	if val, ok := c.GinContext.Get(k); ok && val != nil {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return def
}

func (c *BaseController) GetFloat64(k string, def float64) float64 {
	if val, ok := c.GinContext.Get(k); ok && val != nil {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return def
}

func (c *BaseController) GetBool(k string, def bool) bool {
	if val, ok := c.GinContext.Get(k); ok && val != nil {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return def
}
