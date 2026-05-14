package playwright

// CollectZhipu 通过 Playwright 访问智谱GLM网页版并采集回答
func CollectZhipu(webUrl string, prompt string, screenshotPath string) (*PageCollectResult, error) {
	return collectWithPage(webUrl, prompt, screenshotPath, "textarea, [contenteditable], .input-area")
}

// TestZhipu 通过 Playwright 访问智谱GLM网页版
func TestZhipu(webUrl string) (bool, error) {
	_, err := CollectZhipu(webUrl, "ping", "")
	return err == nil, err
}
