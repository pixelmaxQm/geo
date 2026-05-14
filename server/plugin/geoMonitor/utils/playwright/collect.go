package playwright

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	pw "github.com/playwright-community/playwright-go"
)

type PageCollectResult struct {
	Answer         string
	ScreenshotPath string
	RawResponse    string
	RunLog         string
}

type pageRunLogItem struct {
	Step       string `json:"step"`
	Status     string `json:"status"`
	Message    string `json:"message"`
	DurationMs int64  `json:"durationMs"`
	Time       string `json:"time"`
}

type pageRunLog struct {
	items []pageRunLogItem
}

func newPageRunLog() *pageRunLog {
	return &pageRunLog{items: make([]pageRunLogItem, 0)}
}

func (l *pageRunLog) add(step string, status string, message string, started time.Time) {
	l.items = append(l.items, pageRunLogItem{Step: step, Status: status, Message: message, DurationMs: time.Since(started).Milliseconds(), Time: time.Now().Format(time.RFC3339)})
}

func (l *pageRunLog) json() string {
	if l == nil || len(l.items) == 0 {
		return "[]"
	}
	data, err := json.Marshal(l.items)
	if err != nil {
		return "[]"
	}
	return string(data)
}

type pageDiagnostics struct {
	URL     string `json:"url"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	Login   bool   `json:"login"`
	Message string `json:"message"`
}

func collectWithPage(webURL string, prompt string, screenshotPath string, selector string) (*PageCollectResult, error) {
	return collectWithPageState(webURL, prompt, screenshotPath, selector, "")
}

func collectWithPageState(webURL string, prompt string, screenshotPath string, selector string, storageStatePath string) (*PageCollectResult, error) {
	runLog := newPageRunLog()
	page, cleanup, err := NewPageWithStorageState(storageStatePath)
	if err != nil {
		runLog.add("new_page", "failed", err.Error(), time.Now())
		return &PageCollectResult{ScreenshotPath: screenshotPath, RawResponse: buildDiagnostics(nil, err.Error()), RunLog: runLog.json()}, err
	}
	defer cleanup()

	started := time.Now()
	if _, err := page.Goto(webURL, pw.PageGotoOptions{Timeout: pw.Float(30000), WaitUntil: pw.WaitUntilStateDomcontentloaded}); err != nil {
		runLog.add("goto", "failed", err.Error(), started)
		captureScreenshot(page, screenshotPath)
		return &PageCollectResult{ScreenshotPath: screenshotPath, RawResponse: buildDiagnostics(page, err.Error()), RunLog: runLog.json()}, fmt.Errorf("页面加载失败: %w", err)
	}
	runLog.add("goto", "success", "页面加载完成", started)

	time.Sleep(3 * time.Second)
	started = time.Now()
	if _, err := page.WaitForSelector(selector, pw.PageWaitForSelectorOptions{Timeout: pw.Float(15000)}); err != nil {
		runLog.add("wait_input", "failed", err.Error(), started)
		captureScreenshot(page, screenshotPath)
		diagnostics := buildDiagnostics(page, "页面结构异常，可能需要登录授权")
		return &PageCollectResult{ScreenshotPath: screenshotPath, RawResponse: diagnostics, RunLog: runLog.json()}, fmt.Errorf("页面结构异常，可能需要登录授权: %w", err)
	}
	runLog.add("wait_input", "success", "输入框已出现", started)

	locator := page.Locator(selector).First()
	started = time.Now()
	if err := locator.Fill(prompt); err != nil {
		if clickErr := locator.Click(); clickErr != nil {
			runLog.add("fill_prompt", "failed", clickErr.Error(), started)
			captureScreenshot(page, screenshotPath)
			return &PageCollectResult{ScreenshotPath: screenshotPath, RawResponse: buildDiagnostics(page, clickErr.Error()), RunLog: runLog.json()}, fmt.Errorf("激活输入框失败: %w", clickErr)
		}
		if fillErr := locator.Fill(prompt); fillErr != nil {
			runLog.add("fill_prompt", "failed", fillErr.Error(), started)
			captureScreenshot(page, screenshotPath)
			return &PageCollectResult{ScreenshotPath: screenshotPath, RawResponse: buildDiagnostics(page, fillErr.Error()), RunLog: runLog.json()}, fmt.Errorf("输入提示词失败: %w", fillErr)
		}
	}
	runLog.add("fill_prompt", "success", "提示词输入完成", started)

	started = time.Now()
	if err := locator.Press("Enter"); err != nil {
		runLog.add("submit", "failed", err.Error(), started)
		captureScreenshot(page, screenshotPath)
		return &PageCollectResult{ScreenshotPath: screenshotPath, RawResponse: buildDiagnostics(page, err.Error()), RunLog: runLog.json()}, fmt.Errorf("提交提示词失败: %w", err)
	}
	runLog.add("submit", "success", "提示词已提交", started)

	time.Sleep(8 * time.Second)
	started = time.Now()
	answer, err := page.Locator("body").InnerText()
	if err != nil {
		runLog.add("extract_answer", "failed", err.Error(), started)
		captureScreenshot(page, screenshotPath)
		return &PageCollectResult{ScreenshotPath: screenshotPath, RawResponse: buildDiagnostics(page, err.Error()), RunLog: runLog.json()}, fmt.Errorf("提取回答失败: %w", err)
	}
	answer = strings.TrimSpace(answer)
	if answer == "" {
		runLog.add("extract_answer", "failed", "未获取到回答内容", started)
		captureScreenshot(page, screenshotPath)
		return &PageCollectResult{ScreenshotPath: screenshotPath, RawResponse: buildDiagnostics(page, "未获取到回答内容"), RunLog: runLog.json()}, fmt.Errorf("未获取到回答内容")
	}
	runLog.add("extract_answer", "success", "回答提取完成", started)

	started = time.Now()
	if screenshotPath != "" {
		_ = os.MkdirAll(filepath.Dir(screenshotPath), 0755)
		if _, err := page.Screenshot(pw.PageScreenshotOptions{Path: pw.String(screenshotPath), FullPage: pw.Bool(true)}); err != nil {
			runLog.add("screenshot", "failed", err.Error(), started)
			return &PageCollectResult{Answer: answer, ScreenshotPath: screenshotPath, RawResponse: buildDiagnostics(page, err.Error()), RunLog: runLog.json()}, fmt.Errorf("截图失败: %w", err)
		}
		runLog.add("screenshot", "success", "截图完成", started)
	}

	return &PageCollectResult{Answer: answer, ScreenshotPath: screenshotPath, RawResponse: answer, RunLog: runLog.json()}, nil
}

func captureScreenshot(page pw.Page, screenshotPath string) {
	if page == nil || screenshotPath == "" {
		return
	}
	_ = os.MkdirAll(filepath.Dir(screenshotPath), 0755)
	_, _ = page.Screenshot(pw.PageScreenshotOptions{Path: pw.String(screenshotPath), FullPage: pw.Bool(true)})
}

func buildDiagnostics(page pw.Page, message string) string {
	diagnostics := pageDiagnostics{Message: message}
	if page != nil {
		diagnostics.URL = page.URL()
		if title, err := page.Title(); err == nil {
			diagnostics.Title = title
		}
		if body, err := page.Locator("body").InnerText(pw.LocatorInnerTextOptions{Timeout: pw.Float(1000)}); err == nil {
			diagnostics.Body = truncateText(strings.TrimSpace(body), 4000)
			diagnostics.Login = looksLikeLogin(body)
		}
	}
	data, err := json.Marshal(diagnostics)
	if err != nil {
		return message
	}
	return string(data)
}

func looksLikeLogin(text string) bool {
	lower := strings.ToLower(text)
	return strings.Contains(text, "登录") || strings.Contains(text, "扫码") || strings.Contains(text, "手机号") || strings.Contains(lower, "login") || strings.Contains(lower, "sign in")
}

func truncateText(text string, max int) string {
	if len(text) <= max {
		return text
	}
	return text[:max]
}
