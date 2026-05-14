package model

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// Platform 平台定义
type Platform struct {
	global.GVA_MODEL
	Code                     string                         `json:"code" gorm:"uniqueIndex:idx_code_mode;comment:平台唯一标识"`
	Name                     string                         `json:"name" gorm:"comment:平台显示名称"`
	Mode                     string                         `json:"mode" gorm:"default:api;uniqueIndex:idx_code_mode;comment:采集模式 api/playwright"`
	ApiBase                  string                         `json:"apiBase" gorm:"comment:API端点地址或网页URL"`
	ApiKey                   string                         `json:"apiKey" gorm:"comment:API Key（API模式使用）"`
	Status                   int                            `json:"status" gorm:"default:1;comment:状态 1启用 0停用"`
	Sort                     int                            `json:"sort" gorm:"default:0;comment:排序"`
	Remark                   string                         `json:"remark" gorm:"comment:备注"`
	CurrentAuthorizedSession *PlatformSessionSummary        `json:"currentAuthorizedSession" gorm:"-"`
}

type PlatformSessionSummary struct {
	ID             uint       `json:"id"`
	Status         string     `json:"status"`
	ScreenshotPath string     `json:"screenshotPath"`
	QrImagePath    string     `json:"qrImagePath"`
	StatePath      string     `json:"statePath"`
	ExpiresAt      *time.Time `json:"expiresAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

func (Platform) TableName() string {
	return "gva_geo_monitor_platforms"
}
