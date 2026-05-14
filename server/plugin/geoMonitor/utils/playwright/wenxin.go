package playwright

// CollectWenxin 通过 Playwright 访问文心一言网页版并采集回答
func CollectWenxin(webUrl string, prompt string, screenshotPath string) (*PageCollectResult, error) {
	return collectWithPage(webUrl, prompt, screenshotPath, "textarea, [contenteditable], .yiyan-input")
}

// TestWenxin 通过 Playwright 访问文心一言网页版
func TestWenxin(webUrl string) (bool, error) {
	_, err := CollectWenxin(webUrl, "ping", "")
	return err == nil, err
}
