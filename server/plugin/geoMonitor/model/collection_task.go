package model

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

type CollectionTask struct {
	global.GVA_MODEL
	TopicID         uint       `json:"topicId" gorm:"index;comment:监控话题ID"`
	TopicName       string     `json:"topicName" gorm:"comment:监控话题名称"`
	Prompt          string     `json:"prompt" gorm:"type:text;comment:本次执行提示词"`
	Mode            string     `json:"mode" gorm:"index;comment:采集模式"`
	Status          string     `json:"status" gorm:"index;comment:任务状态"`
	PlatformIDs     string     `json:"platformIds" gorm:"type:text;comment:平台ID列表JSON"`
	RequestedBy     uint       `json:"requestedBy" gorm:"index;comment:发起人ID"`
	RequestedByName string     `json:"requestedByName" gorm:"comment:发起人用户名"`
	StartedAt       *time.Time `json:"startedAt" gorm:"comment:开始时间"`
	FinishedAt      *time.Time `json:"finishedAt" gorm:"comment:结束时间"`
	ErrorMsg        string     `json:"errorMsg" gorm:"type:text;comment:任务级错误信息"`
}

func (CollectionTask) TableName() string {
	return "gva_geo_monitor_collection_tasks"
}
