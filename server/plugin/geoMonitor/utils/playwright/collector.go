package playwright

import "fmt"

func CollectByCode(code string, webURL string, prompt string, screenshotPath string, storageStatePath string) (*PageCollectResult, error) {
	selector, err := selectorForCode(code)
	if err != nil {
		return nil, err
	}
	return collectWithPageState(webURL, prompt, screenshotPath, selector, storageStatePath, code)
}

func selectorForCode(code string) (string, error) {
	switch code {
	case "deepseek":
		return "#chat-input, textarea[placeholder], .chat-input", nil
	case "qwen":
		return "textarea, [contenteditable], .editor", nil
	case "zhipu":
		return "textarea, [contenteditable], .chat-input", nil
	case "doubao":
		return "textarea, [contenteditable], .semi-input-textarea", nil
	case "kimi":
		return "textarea, [contenteditable], .chat-input", nil
	case "wenxin":
		return "textarea, [contenteditable], .chat-input", nil
	case "yuanbao":
		return "textarea, [contenteditable], .chat-input", nil
	default:
		return "", fmt.Errorf("不支持的 Playwright 平台: %s", code)
	}
}
