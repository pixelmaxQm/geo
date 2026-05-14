package model

import "github.com/flipped-aurora/gin-vue-admin/server/global"

// MonitorTopic 监控话题
type MonitorTopic struct {
	global.GVA_MODEL
	Type     string `json:"type" gorm:"comment:监控类型"`
	Name     string `json:"name" gorm:"comment:话题名称"`
	Prompt   string `json:"prompt" gorm:"type:text;comment:AI搜索提示词"`
	Status   int    `json:"status" gorm:"default:1;comment:1启用 0停用"`
	UserID   uint   `json:"userID" gorm:"index;comment:创建者ID"`
	UserName string `json:"userName" gorm:"comment:创建者用户名"`
	Remark   string `json:"remark" gorm:"comment:备注"`
}

func (MonitorTopic) TableName() string {
	return "gva_geo_monitor_topics"
}
