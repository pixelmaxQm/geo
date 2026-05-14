package playwright

// CollectDeepSeek 通过 Playwright 访问 DeepSeek 网页版并采集回答
func CollectDeepSeek(webUrl string, prompt string, screenshotPath string) (*PageCollectResult, error) {
	return collectWithPage(webUrl, prompt, screenshotPath, "#chat-input, textarea[placeholder], .chat-input")
}

// TestDeepSeek 通过 Playwright 访问 DeepSeek 网页版
func TestDeepSeek(webUrl string) (bool, error) {
	_, err := CollectDeepSeek(webUrl, "ping", "")
	return err == nil, err
}
