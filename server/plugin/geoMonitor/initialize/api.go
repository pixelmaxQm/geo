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
	}
	utils.RegisterApis(entities...)
}
