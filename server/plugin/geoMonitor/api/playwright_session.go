package api

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	geoReq "github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/model/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var PlaywrightSession = new(playwrightSession)

type playwrightSession struct{}

func (a *playwrightSession) Start(c *gin.Context) {
	var req geoReq.StartPlaywrightSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	data, err := playwrightSessionService.Start(req, utils.GetUserID(c), utils.GetUserName(c), utils.GetUserAuthorityId(c))
	if err != nil {
		global.GVA_LOG.Error("创建 Playwright 登录会话失败!", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(data, "创建成功", c)
}

func (a *playwrightSession) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}
	data, err := playwrightSessionService.Get(uint(id), utils.GetUserID(c), utils.GetUserAuthorityId(c))
	if err != nil {
		global.GVA_LOG.Error("获取 Playwright 登录会话失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *playwrightSession) Refresh(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}
	data, err := playwrightSessionService.Refresh(uint(id), utils.GetUserID(c), utils.GetUserAuthorityId(c))
	if err != nil {
		global.GVA_LOG.Error("刷新 Playwright 登录会话失败!", zap.Error(err))
		response.FailWithMessage("刷新失败", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *playwrightSession) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}
	if err := playwrightSessionService.Delete(uint(id), utils.GetUserID(c), utils.GetUserAuthorityId(c)); err != nil {
		global.GVA_LOG.Error("删除 Playwright 登录会话失败!", zap.Error(err))
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}
