package service

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/model"
	pwutils "github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/utils/playwright"
)

func (s *collector) InitPlaywright() error {
	if err := pwutils.Install(); err != nil {
		return err
	}
	return pwutils.Launch()
}

func (s *collector) collectWithPlaywright(platform model.Platform, prompt string, taskID uint) (CollectOutput, error) {
	started := time.Now()
	screenshotPath := filepath.ToSlash(filepath.Join("uploads", "file", fmt.Sprintf("gm-task-%d-platform-%d.png", taskID, platform.ID)))
	storageStatePath := latestAuthorizedStorageStatePath(platform.ID)

	result, err := pwutils.CollectByCode(platform.Code, platform.ApiBase, prompt, screenshotPath, storageStatePath)
	if err != nil {
		if IsPlaywrightAuthExpired(platform.Code, result.RawResponse) {
			markSessionExpiredByStatePath(storageStatePath)
		}
		output := CollectOutput{
			Citations:      "[]",
			ScreenshotPath: screenshotPath,
			DurationMs:     int(time.Since(started).Milliseconds()),
			ErrorMsg:       err.Error(),
		}
		if result != nil {
			output.Answer = result.Answer
			output.ScreenshotPath = result.ScreenshotPath
			output.RawResponse = result.RawResponse
			output.RunLog = result.RunLog
		}
		if output.RunLog == "" {
			runLog := NewRunLog()
			runLog.Add("playwright_collect", "failed", err.Error(), int64(output.DurationMs))
			output.RunLog = runLog.JSON()
		}
		return output, err
	}

	return CollectOutput{
		Answer:         result.Answer,
		Citations:      "[]",
		ScreenshotPath: result.ScreenshotPath,
		DurationMs:     int(time.Since(started).Milliseconds()),
		RawResponse:    result.RawResponse,
		RunLog:         result.RunLog,
	}, nil
}

func latestAuthorizedStorageStatePath(platformID uint) string {
	var session model.PlaywrightAuthSession
	if global.GVA_DB == nil {
		return ""
	}
	err := global.GVA_DB.Where("platform_id = ? AND status = ? AND state_path <> '' AND (expires_at IS NULL OR expires_at > ?)", platformID, PlaywrightSessionStatusAuthorized, time.Now()).Order("updated_at desc, id desc").First(&session).Error
	if err != nil {
		return ""
	}
	if _, err := os.Stat(session.StatePath); err != nil {
		_ = global.GVA_DB.Model(&model.PlaywrightAuthSession{}).Where("id = ?", session.ID).Update("status", PlaywrightSessionStatusExpired).Error
		return ""
	}
	now := time.Now()
	_ = global.GVA_DB.Model(&model.PlaywrightAuthSession{}).Where("id = ?", session.ID).Update("last_used_at", &now).Error
	return session.StatePath
}
