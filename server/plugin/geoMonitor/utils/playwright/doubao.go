package playwright

import (
	"fmt"
	"time"

	pw "github.com/playwright-community/playwright-go"
)

// TestDoubao 通过 Playwright 访问豆包网页版
func TestDoubao(webUrl string) (bool, error) {
	page, cleanup, err := NewPage()
	if err != nil {
		return false, err
	}
	defer cleanup()

	if _, err := page.Goto(webUrl, pw.PageGotoOptions{
		Timeout:   pw.Float(30000),
		WaitUntil: pw.WaitUntilStateDomcontentloaded,
	}); err != nil {
		return false, fmt.Errorf("豆包页面加载失败: %w", err)
	}

	time.Sleep(3 * time.Second)
	if _, err := page.WaitForSelector("textarea, [contenteditable], .chat-input-box", pw.PageWaitForSelectorOptions{
		Timeout: pw.Float(15000),
	}); err != nil {
		return false, fmt.Errorf("豆包页面结构异常: %w", err)
	}

	return true, nil
}
