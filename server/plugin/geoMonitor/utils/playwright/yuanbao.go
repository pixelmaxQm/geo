package playwright

// CollectYuanbao 通过 Playwright 访问元宝网页版并采集回答
func CollectYuanbao(webUrl string, prompt string, screenshotPath string) (*PageCollectResult, error) {
	return collectWithPage(webUrl, prompt, screenshotPath, "textarea, [contenteditable], .input-box")
}

// TestYuanbao 通过 Playwright 访问元宝网页版
func TestYuanbao(webUrl string) (bool, error) {
	_, err := CollectYuanbao(webUrl, "ping", "")
	return err == nil, err
}
