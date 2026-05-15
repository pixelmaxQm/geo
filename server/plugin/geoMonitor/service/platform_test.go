package service

import (
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/model"
	pwutils "github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/utils/playwright"
)

func TestPlatformTestReusesAuthorizedStorageState(t *testing.T) {
	originalCollect := collectByCodeForPlatformTest
	originalLookup := latestAuthorizedStorageStatePathForPlatformTest
	defer func() {
		collectByCodeForPlatformTest = originalCollect
		latestAuthorizedStorageStatePathForPlatformTest = originalLookup
	}()

	latestAuthorizedStorageStatePathForPlatformTest = func(platformID uint) string {
		if platformID != 11 {
			t.Fatalf("expected platform id 11, got %d", platformID)
		}
		return "uploads/file/gm-platform-11-state.json"
	}

	var gotStatePath string
	collectByCodeForPlatformTest = func(code string, webURL string, prompt string, screenshotPath string, storageStatePath string) (*pwutils.PageCollectResult, error) {
		gotStatePath = storageStatePath
		return &pwutils.PageCollectResult{ScreenshotPath: screenshotPath, RawResponse: "ok"}, nil
	}

	result := (&platform{}).testPlaywrightWithSnapshot(model.Platform{
		GVA_MODEL: global.GVA_MODEL{ID: 11},
		Code:      "deepseek",
		Name:      "DeepSeek",
		ApiBase:   "https://chat.deepseek.com",
	})

	if gotStatePath != "uploads/file/gm-platform-11-state.json" {
		t.Fatalf("expected authorized storage state to be reused, got %q", gotStatePath)
	}
	if !result.Ok {
		t.Fatalf("expected platform test to succeed, got status=%s message=%s", result.Status, result.Message)
	}
}
