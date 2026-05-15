package model

import "github.com/flipped-aurora/gin-vue-admin/server/global"

type CollectionCitation struct {
	global.GVA_MODEL
	CollectionResultID uint   `json:"collectionResultId" gorm:"index;comment:采集结果ID"`
	Title              string `json:"title" gorm:"type:text;comment:引用标题"`
	URL                string `json:"url" gorm:"type:text;comment:引用网址"`
	Icon               string `json:"icon" gorm:"type:text;comment:引用图标URL"`
	Summary            string `json:"summary" gorm:"type:longtext;comment:引用摘要"`
	Sort               int    `json:"sort" gorm:"comment:排序"`
}

func (CollectionCitation) TableName() string {
	return "gva_geo_monitor_collection_citations"
}
