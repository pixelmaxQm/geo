package model

import "github.com/flipped-aurora/gin-vue-admin/server/global"

type CollectionResult struct {
	global.GVA_MODEL
	TaskID         uint   `json:"taskId" gorm:"index;comment:任务ID"`
	PlatformID     uint   `json:"platformId" gorm:"index;comment:平台ID"`
	PlatformName   string `json:"platformName" gorm:"comment:平台名称"`
	PlatformCode   string `json:"platformCode" gorm:"index;comment:平台编码"`
	Mode           string `json:"mode" gorm:"index;comment:采集模式"`
	Prompt         string `json:"prompt" gorm:"type:text;comment:本次执行提示词"`
	Answer         string `json:"answer" gorm:"type:longtext;comment:回答内容"`
	Status         string `json:"status" gorm:"index;comment:执行结果状态"`
	ErrorMsg       string `json:"errorMsg" gorm:"type:text;comment:错误信息"`
	Citations      string `json:"citations" gorm:"type:longtext;comment:引用信息JSON"`
	AnalysisJSON   string `json:"analysisJson" gorm:"type:longtext;comment:关键词分析JSON"`
	ScreenshotPath string `json:"screenshotPath" gorm:"comment:截图路径"`
	DurationMs     int    `json:"durationMs" gorm:"comment:耗时毫秒"`
	RawResponse    string `json:"rawResponse" gorm:"type:longtext;comment:原始响应"`
	RunLog         string `json:"runLog" gorm:"type:longtext;comment:运行日志JSON"`
}

func (CollectionResult) TableName() string {
	return "gva_geo_monitor_collection_results"
}
