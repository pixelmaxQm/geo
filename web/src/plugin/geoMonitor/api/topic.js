import service from '@/utils/request'

// @Tags GeoMonitorTopic
// @Summary 分页获取话题列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query request.TopicSearch true "分页和筛选参数"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /geoMonitor/topic/list [get]
export const getTopicList = (params) => {
  return service({
    url: '/geoMonitor/topic/list',
    method: 'get',
    params
  })
}

// @Tags GeoMonitorTopic
// @Summary 获取话题详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "话题ID"
// @Success 200 {object} response.Response{data=model.MonitorTopic,msg=string} "获取成功"
// @Router /geoMonitor/topic/:id [get]
export const getTopic = (id) => {
  return service({
    url: `/geoMonitor/topic/${id}`,
    method: 'get'
  })
}

// @Tags GeoMonitorTopic
// @Summary 新增话题
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body model.MonitorTopic true "话题信息"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /geoMonitor/topic [post]
export const createTopic = (data) => {
  return service({
    url: '/geoMonitor/topic',
    method: 'post',
    data
  })
}

// @Tags GeoMonitorTopic
// @Summary 编辑话题
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "话题ID"
// @Param data body model.MonitorTopic true "话题信息"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /geoMonitor/topic/:id [put]
export const updateTopic = (id, data) => {
  return service({
    url: `/geoMonitor/topic/${id}`,
    method: 'put',
    data
  })
}

// @Tags GeoMonitorTopic
// @Summary 删除话题
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "话题ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /geoMonitor/topic/:id [delete]
export const deleteTopic = (id) => {
  return service({
    url: `/geoMonitor/topic/${id}`,
    method: 'delete'
  })
}

export const getAllTopics = () => {
  return service({
    url: '/geoMonitor/topic/list',
    method: 'get',
    params: { page: 1, pageSize: 1000, status: 1 }
  })
}
