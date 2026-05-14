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

var Collect = new(collect)

type collect struct{}

func (a *collect) Run(c *gin.Context) {
	var req geoReq.RunCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	data, err := collectorService.Run(req, utils.GetUserID(c), utils.GetUserName(c), utils.GetUserAuthorityId(c))
	if err != nil {
		global.GVA_LOG.Error("执行采集失败!", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(data, "执行成功", c)
}

func (a *collect) GetTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}
	data, err := collectorService.GetTask(uint(id), utils.GetUserID(c), utils.GetUserAuthorityId(c))
	if err != nil {
		global.GVA_LOG.Error("获取任务失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *collect) GetResultList(c *gin.Context) {
	var req geoReq.CollectionResultSearch
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	list, total, err := collectorService.GetResultList(req.TaskID, req.Page, req.PageSize, utils.GetUserID(c), utils.GetUserAuthorityId(c))
	if err != nil {
		global.GVA_LOG.Error("获取结果列表失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, "获取成功", c)
}
