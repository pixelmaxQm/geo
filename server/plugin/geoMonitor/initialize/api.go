package initialize

import (
	"context"
	model "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
)

func Api(ctx context.Context) {
	entities := []model.SysApi{
		{
			Path:        "/geoMonitor/platform/list",
			Description: "平台列表",
			ApiGroup:    "geoMonitor平台",
			Method:      "GET",
		},
		{
			Path:        "/geoMonitor/platform/:id",
			Description: "平台详情",
			ApiGroup:    "geoMonitor平台",
			Method:      "GET",
		},
		{
			Path:        "/geoMonitor/platform",
			Description: "新增平台",
			ApiGroup:    "geoMonitor平台",
			Method:      "POST",
		},
		{
			Path:        "/geoMonitor/platform/:id",
			Description: "编辑平台",
			ApiGroup:    "geoMonitor平台",
			Method:      "PUT",
		},
		{
			Path:        "/geoMonitor/platform/:id",
			Description: "删除平台",
			ApiGroup:    "geoMonitor平台",
			Method:      "DELETE",
		},
		{
			Path:        "/geoMonitor/platform/test/:id",
			Description: "连通性测试",
			ApiGroup:    "geoMonitor平台",
			Method:      "POST",
		},
		{
			Path:        "/geoMonitor/platform/testAll",
			Description: "一键测试所有平台",
			ApiGroup:    "geoMonitor平台",
			Method:      "POST",
		},
		{
			Path:        "/geoMonitor/collect/run",
			Description: "执行采集",
			ApiGroup:    "geoMonitor采集",
			Method:      "POST",
		},
		{
			Path:        "/geoMonitor/collect/task/:id",
			Description: "采集任务详情",
			ApiGroup:    "geoMonitor采集",
			Method:      "GET",
		},
		{
			Path:        "/geoMonitor/collect/result/list",
			Description: "采集结果列表",
			ApiGroup:    "geoMonitor采集",
			Method:      "GET",
		},
		{
			Path:        "/geoMonitor/playwright/session/start",
			Description: "创建Playwright登录会话",
			ApiGroup:    "geoMonitor采集",
			Method:      "POST",
		},
		{
			Path:        "/geoMonitor/playwright/session/:id",
			Description: "Playwright登录会话详情",
			ApiGroup:    "geoMonitor采集",
			Method:      "GET",
		},
		{
			Path:        "/geoMonitor/playwright/session/:id/refresh",
			Description: "刷新Playwright登录会话",
			ApiGroup:    "geoMonitor采集",
			Method:      "POST",
		},
		{
			Path:        "/geoMonitor/playwright/session/:id",
			Description: "删除Playwright登录会话",
			ApiGroup:    "geoMonitor采集",
			Method:      "DELETE",
		},
		{
			Path:        "/geoMonitor/topic/list",
			Description: "话题列表",
			ApiGroup:    "geoMonitor话题",
			Method:      "GET",
		},
		{
			Path:        "/geoMonitor/topic/:id",
			Description: "话题详情",
			ApiGroup:    "geoMonitor话题",
			Method:      "GET",
		},
		{
			Path:        "/geoMonitor/topic",
			Description: "新增话题",
			ApiGroup:    "geoMonitor话题",
			Method:      "POST",
		},
		{
			Path:        "/geoMonitor/topic/:id",
			Description: "编辑话题",
			ApiGroup:    "geoMonitor话题",
			Method:      "PUT",
		},
		{
			Path:        "/geoMonitor/topic/:id",
			Description: "删除话题",
			ApiGroup:    "geoMonitor话题",
			Method:      "DELETE",
		},
	}
	utils.RegisterApis(entities...)
}
