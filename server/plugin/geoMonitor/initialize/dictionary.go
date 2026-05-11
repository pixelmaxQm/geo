package initialize

import (
	"context"
	model "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
)

func Dictionary(ctx context.Context) {
	statusTrue := true
	entities := []model.SysDictionary{
		{
			Name:   "geo_monitor_platform_status",
			Type:   "geo_monitor_platform_status",
			Status: &statusTrue,
			Desc:   "平台状态",
			SysDictionaryDetails: []model.SysDictionaryDetail{
				{Label: "启用", Value: "1", Status: &statusTrue, Sort: 1},
				{Label: "停用", Value: "0", Status: &statusTrue, Sort: 2},
			},
		},
	}
	utils.RegisterDictionaries(entities...)
}
