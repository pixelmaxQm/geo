package service

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/model"
	apiutils "github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/utils/api"
	pwutils "github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/utils/playwright"
)

func TestManualAPICollectResult(t *testing.T) {
	apiBase := os.Getenv("GEO_MONITOR_API_BASE")
	apiKey := os.Getenv("GEO_MONITOR_API_KEY")
	platformCode := os.Getenv("GEO_MONITOR_API_PLATFORM")
	prompt := os.Getenv("GEO_MONITOR_PROMPT")
	if platformCode == "" {
		platformCode = "deepseek"
	}
	if prompt == "" {
		prompt = "请联网搜索并列出今天一个科技新闻，回答里带上来源。"
	}
	if apiBase == "" || apiKey == "" {
		t.Skip("set GEO_MONITOR_API_BASE and GEO_MONITOR_API_KEY to run this manual API collect test")
	}

	collector := &collector{}
	output, err := collector.collectWithAPI(model.Platform{Code: platformCode, ApiBase: apiBase, ApiKey: apiKey}, prompt)
	if err != nil {
		t.Fatalf("collectWithAPI failed: %v\nrawResponse=%s\nrunLog=%s", err, output.RawResponse, output.RunLog)
	}

	logCollectOutput(t, output)
}

func TestManualPlaywrightCollectResult(t *testing.T) {
	webURL := os.Getenv("GEO_MONITOR_PLAYWRIGHT_URL")
	platformCode := os.Getenv("GEO_MONITOR_PLAYWRIGHT_PLATFORM")
	statePath := os.Getenv("GEO_MONITOR_PLAYWRIGHT_STATE")
	prompt := os.Getenv("GEO_MONITOR_PROMPT")
	if platformCode == "" {
		platformCode = "deepseek"
	}
	if prompt == "" {
		prompt = "请搜索并总结一个最新科技新闻，回答里带上来源。"
	}
	if webURL == "" {
		t.Skip("set GEO_MONITOR_PLAYWRIGHT_URL to run this manual Playwright collect test")
	}

	if err := pwutils.Launch(); err != nil {
		t.Fatalf("launch playwright: %v", err)
	}
	defer pwutils.Close()

	result, err := pwutils.CollectByCode(platformCode, webURL, prompt, "uploads/file/gm-manual-"+platformCode+".png", statePath)
	output := CollectOutput{Citations: "[]"}
	if result != nil {
		output.Answer = result.Answer
		output.ScreenshotPath = result.ScreenshotPath
		output.RawResponse = result.RawResponse
		output.RunLog = result.RunLog
	}
	if err != nil {
		t.Fatalf("CollectByCode failed: %v\nrawResponse=%s\nrunLog=%s\nscreenshot=%s", err, output.RawResponse, output.RunLog, output.ScreenshotPath)
	}

	logCollectOutput(t, output)
}

func TestManualAPIUtilityDirect(t *testing.T) {
	apiBase := os.Getenv("GEO_MONITOR_API_BASE")
	apiKey := os.Getenv("GEO_MONITOR_API_KEY")
	prompt := os.Getenv("GEO_MONITOR_PROMPT")
	if prompt == "" {
		prompt = "请联网搜索并列出今天一个科技新闻，回答里带上来源。"
	}
	if apiBase == "" || apiKey == "" {
		t.Skip("set GEO_MONITOR_API_BASE and GEO_MONITOR_API_KEY to run this direct API utility test")
	}

	result, err := apiutils.CollectDeepSeek(apiBase, apiKey, prompt)
	if err != nil {
		t.Fatalf("CollectDeepSeek failed: %v\nrawResponse=%s", err, result.RawResponse)
	}
	t.Logf("answer:\n%s", result.Answer)
	t.Logf("rawResponse:\n%s", result.RawResponse)
}

func logCollectOutput(t *testing.T, output CollectOutput) {
	t.Helper()
	t.Logf("answer:\n%s", output.Answer)
	t.Logf("citations JSON:\n%s", prettyJSON(output.Citations))
	t.Logf("runLog JSON:\n%s", prettyJSON(output.RunLog))
	t.Logf("rawResponse:\n%s", output.RawResponse)
	t.Logf("screenshotPath: %s", output.ScreenshotPath)
}

func prettyJSON(raw string) string {
	if raw == "" {
		return ""
	}
	var data any
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return raw
	}
	formatted, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return raw
	}
	return string(formatted)
}
