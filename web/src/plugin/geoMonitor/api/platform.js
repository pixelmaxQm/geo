import service from '@/utils/request'

// @Tags GeoMonitorPlatform
// @Summary 分页获取平台列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query request.PlatformSearch true "分页和筛选参数"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /geoMonitor/platform/list [get]
export const getPlatformList = (params) => {
  return service({
    url: '/geoMonitor/platform/list',
    method: 'get',
    params
  })
}

// @Tags GeoMonitorPlatform
// @Summary 获取平台详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "平台ID"
// @Success 200 {object} response.Response{data=model.Platform,msg=string} "获取成功"
// @Router /geoMonitor/platform/:id [get]
export const getPlatform = (id) => {
  return service({
    url: `/geoMonitor/platform/${id}`,
    method: 'get'
  })
}

// @Tags GeoMonitorPlatform
// @Summary 新增平台
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body model.Platform true "平台信息"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /geoMonitor/platform [post]
export const createPlatform = (data) => {
  return service({
    url: '/geoMonitor/platform',
    method: 'post',
    data
  })
}

// @Tags GeoMonitorPlatform
// @Summary 编辑平台
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "平台ID"
// @Param data body model.Platform true "平台信息"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /geoMonitor/platform/:id [put]
export const updatePlatform = (id, data) => {
  return service({
    url: `/geoMonitor/platform/${id}`,
    method: 'put',
    data
  })
}

// @Tags GeoMonitorPlatform
// @Summary 删除平台
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "平台ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /geoMonitor/platform/:id [delete]
export const deletePlatform = (id) => {
  return service({
    url: `/geoMonitor/platform/${id}`,
    method: 'delete'
  })
}

// @Tags GeoMonitorPlatform
// @Summary 连通性测试
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "平台ID"
// @Success 200 {object} response.Response{msg=string} "连接成功"
// @Router /geoMonitor/platform/test/:id [post]
export const testPlatform = (id) => {
  return service({
    url: `/geoMonitor/platform/test/${id}`,
    method: 'post'
  })
}

// @Tags GeoMonitorPlatform
// @Summary 一键测试所有平台连通性
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=[]object,msg=string} "测试完成"
// @Router /geoMonitor/platform/testAll [post]
export const testAllPlatforms = () => {
  return service({
    url: '/geoMonitor/platform/testAll',
    method: 'post'
  })
}

export const getEnabledPlatformsByMode = (mode) => {
  return service({
    url: '/geoMonitor/platform/list',
    method: 'get',
    params: { page: 1, pageSize: 1000, mode, status: 1 }
  })
}
