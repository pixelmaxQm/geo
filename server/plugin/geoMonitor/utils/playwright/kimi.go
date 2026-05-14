package playwright

// CollectKimi 通过 Playwright 访问 Kimi 网页版并采集回答
func CollectKimi(webUrl string, prompt string, screenshotPath string) (*PageCollectResult, error) {
	return collectWithPage(webUrl, prompt, screenshotPath, "textarea, [contenteditable], .chat-input")
}

// TestKimi 通过 Playwright 访问 Kimi 网页版
func TestKimi(webUrl string) (bool, error) {
	_, err := CollectKimi(webUrl, "ping", "")
	return err == nil, err
}
