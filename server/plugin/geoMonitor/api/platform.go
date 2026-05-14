package api

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/model"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/model/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var Platform = new(platform)

type platform struct{}

// GetPlatformList 分页获取平台列表
// @Tags      GeoMonitorPlatform
// @Summary   分页获取平台列表
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  query     request.PlatformSearch                               true  "分页和筛选参数"
// @Success   200   {object}  response.Response{data=response.PageResult,msg=string}  "获取成功"
// @Router    /geoMonitor/platform/list [get]
func (a *platform) GetPlatformList(c *gin.Context) {
	var pageInfo request.PlatformSearch
	if err := c.ShouldBindQuery(&pageInfo); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := platformService.GetPlatformList(pageInfo)
	if err != nil {
		global.GVA_LOG.Error("获取平台列表失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功", c)
}

// GetPlatform 获取平台详情
// @Tags      GeoMonitorPlatform
// @Summary   获取平台详情
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     id   path      int  true  "平台ID"
// @Success   200  {object}  response.Response{data=model.Platform,msg=string}  "获取成功"
// @Router    /geoMonitor/platform/{id} [get]
func (a *platform) GetPlatform(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}
	info, err := platformService.GetPlatform(uint(id))
	if err != nil {
		global.GVA_LOG.Error("获取平台详情失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithData(info, c)
}

// CreatePlatform 新增平台
// @Tags      GeoMonitorPlatform
// @Summary   新增平台
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      model.Platform  true  "平台信息"
// @Success   200   {object}  response.Response{msg=string}  "创建成功"
// @Router    /geoMonitor/platform [post]
func (a *platform) CreatePlatform(c *gin.Context) {
	var info model.Platform
	if err := c.ShouldBindJSON(&info); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := platformService.CreatePlatform(&info); err != nil {
		global.GVA_LOG.Error("创建平台失败!", zap.Error(err))
		response.FailWithMessage("创建失败", c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// UpdatePlatform 编辑平台
// @Tags      GeoMonitorPlatform
// @Summary   编辑平台
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     id    path      int             true  "平台ID"
// @Param     data  body      model.Platform  true  "平台信息"
// @Success   200   {object}  response.Response{msg=string}  "更新成功"
// @Router    /geoMonitor/platform/{id} [put]
func (a *platform) UpdatePlatform(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}
	var info model.Platform
	if err := c.ShouldBindJSON(&info); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	info.ID = uint(id)
	if err := platformService.UpdatePlatform(&info); err != nil {
		global.GVA_LOG.Error("更新平台失败!", zap.Error(err))
		response.FailWithMessage("更新失败", c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeletePlatform 删除平台
// @Tags      GeoMonitorPlatform
// @Summary   删除平台
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     id   path      int  true  "平台ID"
// @Success   200  {object}  response.Response{msg=string}  "删除成功"
// @Router    /geoMonitor/platform/{id} [delete]
func (a *platform) DeletePlatform(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}
	if err := platformService.DeletePlatform(uint(id)); err != nil {
		global.GVA_LOG.Error("删除平台失败!", zap.Error(err))
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// TestPlatform 连通性测试
// @Tags      GeoMonitorPlatform
// @Summary   连通性测试
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     id   path      int  true  "平台ID"
// @Success   200  {object}  response.Response{data=service.PlatformTestResult,msg=string}  "测试完成"
// @Router    /geoMonitor/platform/test/{id} [post]
func (a *platform) TestPlatform(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}
	result, err := platformService.TestConnectivity(uint(id))
	if err != nil {
		global.GVA_LOG.Error("连通性测试失败!", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(result, result.Message, c)
}

// TestAllPlatforms 一键测试所有平台连通性
// @Tags      GeoMonitorPlatform
// @Summary   一键测试所有平台连通性
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Success   200  {object}  response.Response{data=[]service.PlatformTestResult,msg=string}  "测试完成"
// @Router    /geoMonitor/platform/testAll [post]
func (a *platform) TestAllPlatforms(c *gin.Context) {
	results, err := platformService.TestAllConnectivity()
	if err != nil {
		global.GVA_LOG.Error("一键测试失败!", zap.Error(err))
		response.FailWithMessage("测试失败", c)
		return
	}
	response.OkWithData(results, c)
}
