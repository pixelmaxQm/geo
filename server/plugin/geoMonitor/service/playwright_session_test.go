package service

import (
	"strings"
	"testing"
)

func TestIsPlaywrightAuthExpired_DeepSeekSignInURL(t *testing.T) {
	raw := `{"url":"https://chat.deepseek.com/sign_in?from=wechat","message":"页面结构异常，可能需要登录授权"}`
	if !IsPlaywrightAuthExpired("deepseek", raw) {
		t.Fatalf("expected deepseek sign_in raw response to be treated as expired auth")
	}
}

func TestIsPlaywrightAuthExpired_DeepSeekChatURL(t *testing.T) {
	raw := `{"url":"https://chat.deepseek.com/chat/s/abc123","message":"ok"}`
	if IsPlaywrightAuthExpired("deepseek", raw) {
		t.Fatalf("expected deepseek chat raw response not to be treated as expired auth")
	}
}

func TestDeepSeekAuthorizedURLPattern(t *testing.T) {
	authorizedURLs := []string{
		"https://chat.deepseek.com/",
		"https://chat.deepseek.com/chat/s/abc123",
		"https://chat.deepseek.com/?utm=1",
	}
	for _, rawURL := range authorizedURLs {
		if strings.Contains(rawURL, "/sign_in") {
			t.Fatalf("test setup invalid, authorized URL contains sign_in: %s", rawURL)
		}
	}
}

func TestBuildSessionDiagnosticsMessage(t *testing.T) {
	message := buildSessionDiagnosticsMessage("waiting_scan", "https://chat.deepseek.com/sign_in?from=wechat", "DeepSeek", false)
	checks := []string{
		"status=waiting_scan",
		"url=https://chat.deepseek.com/sign_in?from=wechat",
		"title=DeepSeek",
		"inputVisible=false",
	}
	for _, check := range checks {
		if !strings.Contains(message, check) {
			t.Fatalf("expected diagnostics message to contain %q, got %q", check, message)
		}
	}
}
