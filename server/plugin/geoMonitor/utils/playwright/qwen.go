package playwright

// CollectQwen 通过 Playwright 访问通义千问网页版并采集回答
func CollectQwen(webUrl string, prompt string, screenshotPath string) (*PageCollectResult, error) {
	return collectWithPage(webUrl, prompt, screenshotPath, "textarea, [contenteditable], .editor")
}

// TestQwen 通过 Playwright 访问通义千问网页版
func TestQwen(webUrl string) (bool, error) {
	_, err := CollectQwen(webUrl, "ping", "")
	return err == nil, err
}
