package model

import "github.com/flipped-aurora/gin-vue-admin/server/global"

// Platform 平台定义
type Platform struct {
	global.GVA_MODEL
	Code    string `json:"code" gorm:"uniqueIndex;comment:平台唯一标识"`
	Name    string `json:"name" gorm:"comment:平台显示名称"`
	ApiBase string `json:"apiBase" gorm:"comment:API端点地址"`
	ApiKey  string `json:"apiKey" gorm:"comment:API Key（加密存储）"`
	Status  int    `json:"status" gorm:"default:1;comment:状态 1启用 0停用"`
	Sort    int    `json:"sort" gorm:"default:0;comment:排序"`
	Remark  string `json:"remark" gorm:"comment:备注"`
}

func (Platform) TableName() string {
	return "gva_geo_monitor_platforms"
}
