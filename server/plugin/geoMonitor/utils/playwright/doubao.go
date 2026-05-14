package playwright

// CollectDoubao 通过 Playwright 访问豆包网页版并采集回答
func CollectDoubao(webUrl string, prompt string, screenshotPath string) (*PageCollectResult, error) {
	return collectWithPage(webUrl, prompt, screenshotPath, "textarea, [contenteditable], .chat-input-box")
}

// TestDoubao 通过 Playwright 访问豆包网页版
func TestDoubao(webUrl string) (bool, error) {
	_, err := CollectDoubao(webUrl, "ping", "")
	return err == nil, err
}
