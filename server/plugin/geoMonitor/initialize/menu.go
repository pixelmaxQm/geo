package initialize

import (
	"context"
	model "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
)

func Menu(ctx context.Context) {
	entities := []model.SysBaseMenu{
		{
			ParentId:  0,
			Path:      "geoMonitor",
			Name:      "geoMonitor",
			Hidden:    false,
			Component: "view/index",
			Sort:      1,
			Meta:      model.Meta{Title: "AI监控分析", Icon: "monitor"},
		},
		{
			ParentId:  0,
			Path:      "geoMonitorPlatform",
			Name:      "geoMonitorPlatform",
			Hidden:    false,
			Component: "plugin/geoMonitor/view/platform.vue",
			Sort:      1,
			Meta:      model.Meta{Title: "平台配置", Icon: "platform"},
		},
	}
	utils.RegisterMenus(entities...)
}
