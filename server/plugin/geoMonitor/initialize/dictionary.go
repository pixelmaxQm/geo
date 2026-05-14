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
	entities = append(entities, model.SysDictionary{
		Name:   "geo_monitor_topic_type",
		Type:   "geo_monitor_topic_type",
		Status: &statusTrue,
		Desc:   "监控话题类型",
		SysDictionaryDetails: []model.SysDictionaryDetail{
			{Label: "事实核查", Value: "fact_check", Status: &statusTrue, Sort: 1},
			{Label: "实时资讯", Value: "news", Status: &statusTrue, Sort: 2},
			{Label: "知识问答", Value: "knowledge", Status: &statusTrue, Sort: 3},
			{Label: "政策解读", Value: "policy", Status: &statusTrue, Sort: 4},
			{Label: "技术评测", Value: "tech", Status: &statusTrue, Sort: 5},
			{Label: "其他", Value: "other", Status: &statusTrue, Sort: 6},
		},
	})
	utils.RegisterDictionaries(entities...)
}
