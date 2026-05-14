package api

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/model"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/model/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var Topic = new(topic)

type topic struct{}

// GetTopicList 分页获取话题列表
// @Tags      GeoMonitorTopic
// @Summary   分页获取话题列表
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  query     request.TopicSearch                               true  "分页和筛选参数"
// @Success   200   {object}  response.Response{data=response.PageResult,msg=string}  "获取成功"
// @Router    /geoMonitor/topic/list [get]
func (a *topic) GetTopicList(c *gin.Context) {
	var pageInfo request.TopicSearch
	if err := c.ShouldBindQuery(&pageInfo); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := topicService.GetTopicList(pageInfo, utils.GetUserID(c), utils.GetUserAuthorityId(c))
	if err != nil {
		global.GVA_LOG.Error("获取话题列表失败!", zap.Error(err))
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

// GetTopic 获取话题详情
// @Tags      GeoMonitorTopic
// @Summary   获取话题详情
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     id   path      int  true  "话题ID"
// @Success   200  {object}  response.Response{data=model.MonitorTopic,msg=string}  "获取成功"
// @Router    /geoMonitor/topic/{id} [get]
func (a *topic) GetTopic(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}
	info, err := topicService.GetTopic(uint(id), utils.GetUserID(c), utils.GetUserAuthorityId(c))
	if err != nil {
		global.GVA_LOG.Error("获取话题详情失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithData(info, c)
}

// CreateTopic 新增话题
// @Tags      GeoMonitorTopic
// @Summary   新增话题
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      model.MonitorTopic  true  "话题信息"
// @Success   200   {object}  response.Response{msg=string}  "创建成功"
// @Router    /geoMonitor/topic [post]
func (a *topic) CreateTopic(c *gin.Context) {
	var info model.MonitorTopic
	if err := c.ShouldBindJSON(&info); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	info.UserID = utils.GetUserID(c)
	info.UserName = utils.GetUserName(c)
	if err := topicService.CreateTopic(&info); err != nil {
		global.GVA_LOG.Error("创建话题失败!", zap.Error(err))
		response.FailWithMessage("创建失败", c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// UpdateTopic 编辑话题
// @Tags      GeoMonitorTopic
// @Summary   编辑话题
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     id    path      int                 true  "话题ID"
// @Param     data  body      model.MonitorTopic  true  "话题信息"
// @Success   200   {object}  response.Response{msg=string}  "更新成功"
// @Router    /geoMonitor/topic/{id} [put]
func (a *topic) UpdateTopic(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}
	var info model.MonitorTopic
	if err := c.ShouldBindJSON(&info); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	info.ID = uint(id)
	if err := topicService.UpdateTopic(&info, utils.GetUserID(c), utils.GetUserAuthorityId(c)); err != nil {
		global.GVA_LOG.Error("更新话题失败!", zap.Error(err))
		response.FailWithMessage("更新失败", c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteTopic 删除话题
// @Tags      GeoMonitorTopic
// @Summary   删除话题
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     id   path      int  true  "话题ID"
// @Success   200  {object}  response.Response{msg=string}  "删除成功"
// @Router    /geoMonitor/topic/{id} [delete]
func (a *topic) DeleteTopic(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}
	if err := topicService.DeleteTopic(uint(id), utils.GetUserID(c), utils.GetUserAuthorityId(c)); err != nil {
		global.GVA_LOG.Error("删除话题失败!", zap.Error(err))
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}
