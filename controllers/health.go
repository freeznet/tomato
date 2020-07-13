package controllers

import (
	"github.com/astaxie/beego"
	"github.com/freeznet/tomato/types"
)

// HealthController 检测服务器健康状态
type HealthController struct {
	beego.Controller
}

// Get 直接返回状态 200
// @router / [get]
func (h *HealthController) Get() {
	h.Data["json"] = types.M{
		"status": "ok",
	}
	h.ServeJSON()
}
