package playwright

import (
	"fmt"
	"time"

	pw "github.com/playwright-community/playwright-go"
)

// TestQwen 通过 Playwright 访问通义千问网页版
func TestQwen(webUrl string) (bool, error) {
	page, cleanup, err := NewPage()
	if err != nil {
		return false, err
	}
	defer cleanup()

	if _, err := page.Goto(webUrl, pw.PageGotoOptions{
		Timeout:   pw.Float(30000),
		WaitUntil: pw.WaitUntilStateDomcontentloaded,
	}); err != nil {
		return false, fmt.Errorf("通义千问页面加载失败: %w", err)
	}

	time.Sleep(3 * time.Second)
	if _, err := page.WaitForSelector("textarea, [contenteditable], .editor", pw.PageWaitForSelectorOptions{
		Timeout: pw.Float(15000),
	}); err != nil {
		return false, fmt.Errorf("通义千问页面结构异常: %w", err)
	}

	return true, nil
}
