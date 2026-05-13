package playwright

import (
	"fmt"
	"time"

	pw "github.com/playwright-community/playwright-go"
)

// TestZhipu 通过 Playwright 访问智谱GLM网页版
func TestZhipu(webUrl string) (bool, error) {
	page, cleanup, err := NewPage()
	if err != nil {
		return false, err
	}
	defer cleanup()

	if _, err := page.Goto(webUrl, pw.PageGotoOptions{
		Timeout:   pw.Float(30000),
		WaitUntil: pw.WaitUntilStateDomcontentloaded,
	}); err != nil {
		return false, fmt.Errorf("智谱GLM页面加载失败: %w", err)
	}

	time.Sleep(3 * time.Second)
	if _, err := page.WaitForSelector("textarea, [contenteditable], .input-area", pw.PageWaitForSelectorOptions{
		Timeout: pw.Float(15000),
	}); err != nil {
		return false, fmt.Errorf("智谱GLM页面结构异常: %w", err)
	}

	return true, nil
}
