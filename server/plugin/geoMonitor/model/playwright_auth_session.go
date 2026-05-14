package model

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

type PlaywrightAuthSession struct {
	global.GVA_MODEL
	PlatformID     uint       `json:"platformId" gorm:"index;comment:平台ID"`
	PlatformCode   string     `json:"platformCode" gorm:"index;comment:平台编码"`
	Status         string     `json:"status" gorm:"index;comment:授权状态"`
	LoginURL       string     `json:"loginUrl" gorm:"type:text;comment:登录地址"`
	QrImagePath    string     `json:"qrImagePath" gorm:"comment:二维码图片路径"`
	ScreenshotPath string     `json:"screenshotPath" gorm:"comment:登录页截图路径"`
	StatePath      string     `json:"statePath" gorm:"comment:Playwright storageState路径"`
	ErrorMsg       string     `json:"errorMsg" gorm:"type:text;comment:错误信息"`
	ExpiresAt      *time.Time `json:"expiresAt" gorm:"comment:过期时间"`
	LastUsedAt     *time.Time `json:"lastUsedAt" gorm:"comment:最后使用时间"`
	CreatedBy      uint       `json:"createdBy" gorm:"index;comment:创建人ID"`
	CreatedByName  string     `json:"createdByName" gorm:"comment:创建人用户名"`
}

func (PlaywrightAuthSession) TableName() string {
	return "gva_geo_monitor_playwright_auth_sessions"
}
