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
	Citations      []CitationItem
	ScreenshotPath string
	RawResponse    string
	RunLog         string
}

type CitationItem struct {
	Title   string
	URL     string
	Icon    string
	Snippet string
	Source  string
}

type deepSeekCitationCard struct {
	Title   string
	URL     string
	Icon    string
	Summary string
}

type pageRunLogItem struct {
	Step           string         `json:"step"`
	Status         string         `json:"status"`
	Message        string         `json:"message"`
	DurationMs     int64          `json:"durationMs"`
	Time           string         `json:"time"`
	ScreenshotPath string         `json:"screenshotPath,omitempty"`
	Meta           map[string]any `json:"meta,omitempty"`
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

func (l *pageRunLog) addWithDetails(step string, status string, message string, started time.Time, screenshotPath string, meta map[string]any) {
	l.items = append(l.items, pageRunLogItem{Step: step, Status: status, Message: message, DurationMs: time.Since(started).Milliseconds(), Time: time.Now().Format(time.RFC3339), ScreenshotPath: screenshotPath, Meta: meta})
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
	return collectWithPageState(webURL, prompt, screenshotPath, selector, "", "")
}

func collectWithPageState(webURL string, prompt string, screenshotPath string, selector string, storageStatePath string, code string) (*PageCollectResult, error) {
	runLog := newPageRunLog()
	page, cleanup, err := NewPageWithStorageState(storageStatePath)
	if err != nil {
		runLog.add("new_page", "failed", err.Error(), time.Now())
		return &PageCollectResult{ScreenshotPath: screenshotPath, RawResponse: buildDiagnostics(nil, err.Error()), RunLog: runLog.json()}, err
	}
	defer cleanup()

	started := time.Now()
	if _, err := page.Goto(webURL, pw.PageGotoOptions{Timeout: pw.Float(30000), WaitUntil: pw.WaitUntilStateDomcontentloaded}); err != nil {
		stepShot := stepScreenshotPath(screenshotPath, "goto-failed")
		captureScreenshot(page, stepShot)
		runLog.addWithDetails("goto", "failed", err.Error(), started, stepShot, map[string]any{"url": webURL})
		return &PageCollectResult{ScreenshotPath: screenshotPath, RawResponse: buildDiagnostics(page, err.Error()), RunLog: runLog.json()}, fmt.Errorf("页面加载失败: %w", err)
	}
	runLog.addWithDetails("goto", "success", "页面加载完成", started, captureStepScreenshot(page, screenshotPath, "goto"), map[string]any{"url": webURL})

	if needsDedicatedFlow(code) {
		if err := preparePlatformConversation(code, page, runLog, screenshotPath); err != nil {
			captureScreenshot(page, screenshotPath)
			return &PageCollectResult{ScreenshotPath: screenshotPath, RawResponse: buildDiagnostics(page, err.Error()), RunLog: runLog.json()}, err
		}
	}

	started = time.Now()
	if _, err := page.WaitForSelector(selector, pw.PageWaitForSelectorOptions{Timeout: pw.Float(15000)}); err != nil {
		stepShot := stepScreenshotPath(screenshotPath, "wait-input-failed")
		captureScreenshot(page, stepShot)
		diagnostics := buildDiagnostics(page, "页面结构异常，可能需要登录授权")
		runLog.addWithDetails("wait_input", "failed", err.Error(), started, stepShot, map[string]any{"selector": selector})
		return &PageCollectResult{ScreenshotPath: screenshotPath, RawResponse: diagnostics, RunLog: runLog.json()}, fmt.Errorf("页面结构异常，可能需要登录授权: %w", err)
	}
	runLog.addWithDetails("wait_input", "success", "输入框已出现", started, captureStepScreenshot(page, screenshotPath, "wait-input"), map[string]any{"selector": selector})

	locator := page.Locator(selector).First()
	started = time.Now()
	if err := locator.Fill(prompt); err != nil {
		if clickErr := locator.Click(); clickErr != nil {
			stepShot := stepScreenshotPath(screenshotPath, "fill-failed")
			captureScreenshot(page, stepShot)
			runLog.addWithDetails("fill_prompt", "failed", clickErr.Error(), started, stepShot, map[string]any{"selector": selector})
			return &PageCollectResult{ScreenshotPath: screenshotPath, RawResponse: buildDiagnostics(page, clickErr.Error()), RunLog: runLog.json()}, fmt.Errorf("激活输入框失败: %w", clickErr)
		}
		if fillErr := locator.Fill(prompt); fillErr != nil {
			stepShot := stepScreenshotPath(screenshotPath, "fill-failed")
			captureScreenshot(page, stepShot)
			runLog.addWithDetails("fill_prompt", "failed", fillErr.Error(), started, stepShot, map[string]any{"selector": selector})
			return &PageCollectResult{ScreenshotPath: screenshotPath, RawResponse: buildDiagnostics(page, fillErr.Error()), RunLog: runLog.json()}, fmt.Errorf("输入提示词失败: %w", fillErr)
		}
	}
	runLog.addWithDetails("fill_prompt", "success", "提示词输入完成", started, captureStepScreenshot(page, screenshotPath, "fill"), map[string]any{"selector": selector, "promptLength": len(prompt)})

	started = time.Now()
	if err := locator.Press("Enter"); err != nil {
		stepShot := stepScreenshotPath(screenshotPath, "submit-failed")
		captureScreenshot(page, stepShot)
		runLog.addWithDetails("submit", "failed", err.Error(), started, stepShot, map[string]any{"selector": selector})
		return &PageCollectResult{ScreenshotPath: screenshotPath, RawResponse: buildDiagnostics(page, err.Error()), RunLog: runLog.json()}, fmt.Errorf("提交提示词失败: %w", err)
	}
	runLog.addWithDetails("submit", "success", "提示词已提交", started, captureStepScreenshot(page, screenshotPath, "submit"), map[string]any{"selector": selector})

	started = time.Now()
	answer, err := waitForAnswerStable(page, code)
	if err != nil {
		stepShot := stepScreenshotPath(screenshotPath, "answer-failed")
		captureScreenshot(page, stepShot)
		runLog.addWithDetails("extract_answer", "failed", err.Error(), started, stepShot, nil)
		return &PageCollectResult{ScreenshotPath: screenshotPath, RawResponse: buildDiagnostics(page, err.Error()), RunLog: runLog.json()}, fmt.Errorf("提取回答失败: %w", err)
	}
	answer = strings.TrimSpace(answer)
	if answer == "" {
		stepShot := stepScreenshotPath(screenshotPath, "answer-empty")
		captureScreenshot(page, stepShot)
		runLog.addWithDetails("extract_answer", "failed", "未获取到回答内容", started, stepShot, nil)
		return &PageCollectResult{ScreenshotPath: screenshotPath, RawResponse: buildDiagnostics(page, "未获取到回答内容"), RunLog: runLog.json()}, fmt.Errorf("未获取到回答内容")
	}
	runLog.addWithDetails("extract_answer", "success", "回答提取完成", started, captureStepScreenshot(page, screenshotPath, "answer"), map[string]any{"answerLength": len(answer)})

	citations := make([]CitationItem, 0)
	if needsDedicatedFlow(code) {
		started = time.Now()
		cards, citationErr := collectPlatformCitations(code, page)
		if citationErr != nil {
			stepShot := stepScreenshotPath(screenshotPath, "citations-failed")
			captureScreenshot(page, stepShot)
			runLog.addWithDetails("extract_citations", "failed", citationErr.Error(), started, stepShot, nil)
		} else {
			citations = buildCitationItems(code, cards)
			runLog.addWithDetails("extract_citations", "success", fmt.Sprintf("提取到 %d 条引用", len(citations)), started, captureStepScreenshot(page, screenshotPath, "citations"), map[string]any{"count": len(citations)})
		}
	}

	started = time.Now()
	if screenshotPath != "" {
		_ = os.MkdirAll(filepath.Dir(screenshotPath), 0755)
		if _, err := page.Screenshot(pw.PageScreenshotOptions{Path: pw.String(screenshotPath), FullPage: pw.Bool(true)}); err != nil {
			stepShot := stepScreenshotPath(screenshotPath, "final-failed")
			captureScreenshot(page, stepShot)
			runLog.addWithDetails("screenshot", "failed", err.Error(), started, stepShot, nil)
			return &PageCollectResult{Answer: answer, Citations: citations, ScreenshotPath: screenshotPath, RawResponse: buildDiagnostics(page, err.Error()), RunLog: runLog.json()}, fmt.Errorf("截图失败: %w", err)
		}
		runLog.addWithDetails("screenshot", "success", "截图完成", started, screenshotPath, nil)
	}

	return &PageCollectResult{Answer: answer, Citations: citations, ScreenshotPath: screenshotPath, RawResponse: answer, RunLog: runLog.json()}, nil
}

func stepScreenshotPath(finalPath string, suffix string) string {
	if finalPath == "" {
		return ""
	}
	ext := filepath.Ext(finalPath)
	base := strings.TrimSuffix(finalPath, ext)
	if ext == "" {
		ext = ".png"
	}
	return base + "-" + suffix + ext
}

func captureStepScreenshot(page pw.Page, finalPath string, suffix string) string {
	path := stepScreenshotPath(finalPath, suffix)
	captureScreenshot(page, path)
	return path
}

func waitForAnswerStable(page pw.Page, code string) (string, error) {
	if code == "deepseek" {
		return waitForDeepSeekAnswerStable(page)
	}
	time.Sleep(8 * time.Second)
	return page.Locator("body").InnerText()
}

func waitForDeepSeekAnswerStable(page pw.Page) (string, error) {
	var lastText string
	stableCount := 0
	deadline := time.Now().Add(45 * time.Second)
	for time.Now().Before(deadline) {
		text, err := page.Locator("body").InnerText(pw.LocatorInnerTextOptions{Timeout: pw.Float(2000)})
		if err == nil {
			trimmed := strings.TrimSpace(text)
			if trimmed != "" {
				if trimmed == lastText {
					stableCount++
				} else {
					stableCount = 0
					lastText = trimmed
				}
				if stableCount >= 2 && !deepSeekIsGenerating(page) {
					return trimmed, nil
				}
			}
		}
		time.Sleep(1200 * time.Millisecond)
	}
	if strings.TrimSpace(lastText) == "" {
		return "", fmt.Errorf("等待 DeepSeek 回答超时")
	}
	return lastText, nil
}

func deepSeekIsGenerating(page pw.Page) bool {
	selectors := []string{
		"button:has-text('停止')",
		"button:has-text('Stop')",
		"[aria-label*='停止']",
		"[aria-label*='Stop']",
	}
	for _, selector := range selectors {
		locator := page.Locator(selector).First()
		if err := locator.WaitFor(pw.LocatorWaitForOptions{State: pw.WaitForSelectorStateVisible, Timeout: pw.Float(300)}); err == nil {
			return true
		}
	}
	return false
}

func needsDedicatedFlow(code string) bool {
	switch code {
	case "deepseek":
		return true
	default:
		return false
	}
}

func preparePlatformConversation(code string, page pw.Page, runLog *pageRunLog, screenshotPath string) error {
	switch code {
	case "deepseek":
		return prepareDeepSeekConversation(page, runLog, screenshotPath)
	default:
		return nil
	}
}

func collectPlatformCitations(code string, page pw.Page) ([]deepSeekCitationCard, error) {
	switch code {
	case "deepseek":
		return collectDeepSeekCitations(page)
	default:
		return nil, nil
	}
}

func buildCitationItems(code string, cards []deepSeekCitationCard) []CitationItem {
	switch code {
	case "deepseek":
		return buildDeepSeekCitationItems(cards)
	default:
		return nil
	}
}

func prepareDeepSeekConversation(page pw.Page, runLog *pageRunLog, screenshotPath string) error {
	selectors := []string{
		"div[tabindex='0']:has-text('开启新对话')",
		"button:has-text('开启新对话')",
		"text=开启新对话",
	}
	started := time.Now()
	for _, selector := range selectors {
		locator := page.Locator(selector).First()
		if err := locator.WaitFor(pw.LocatorWaitForOptions{State: pw.WaitForSelectorStateVisible, Timeout: pw.Float(1500)}); err == nil {
			if err := locator.Click(); err == nil {
				stepShot := captureStepScreenshot(page, screenshotPath, "new-conversation")
				runLog.addWithDetails("new_conversation", "success", "已开启新对话", started, stepShot, map[string]any{"selector": selector})
				time.Sleep(1500 * time.Millisecond)
				return nil
			}
		}
	}
	stepShot := captureStepScreenshot(page, screenshotPath, "new-conversation-failed")
	runLog.addWithDetails("new_conversation", "failed", "未找到开启新对话按钮", started, stepShot, nil)
	return fmt.Errorf("未找到开启新对话按钮")
}

func deepSeekCitationTriggerSelectors() []string {
	return []string{
		"text=/\\d+\\s*个网页/",
		"span:has-text('个网页')",
		"div:has-text('个网页')",
		"text=/\\d+\\s*web pages/i",
	}
}

func collectDeepSeekCitations(page pw.Page) ([]deepSeekCitationCard, error) {
	opened := false
	for _, selector := range deepSeekCitationTriggerSelectors() {
		locator := page.Locator(selector).First()
		if err := locator.WaitFor(pw.LocatorWaitForOptions{State: pw.WaitForSelectorStateVisible, Timeout: pw.Float(1500)}); err == nil {
			if err := locator.Click(); err == nil {
				opened = true
				break
			}
		}
	}
	if !opened {
		return nil, fmt.Errorf("未找到引用抽屉入口")
	}

	cardSelector := "a.c64652fe"
	if _, err := page.WaitForSelector(cardSelector, pw.PageWaitForSelectorOptions{Timeout: pw.Float(5000)}); err != nil {
		return nil, fmt.Errorf("未找到引用卡片: %w", err)
	}

	cards := page.Locator(cardSelector)
	count, err := cards.Count()
	if err != nil {
		return nil, err
	}
	items := make([]deepSeekCitationCard, 0, count)
	for i := 0; i < count; i++ {
		card := cards.Nth(i)
		href, _ := card.GetAttribute("href")
		title, _ := card.Locator("._8d3001c").First().InnerText()
		summary, _ := card.Locator(".d153a905").First().InnerText()
		icon, _ := card.Locator("img.site_logo_img").First().GetAttribute("src")
		items = append(items, deepSeekCitationCard{
			Title:   strings.TrimSpace(title),
			URL:     strings.TrimSpace(href),
			Icon:    strings.TrimSpace(icon),
			Summary: strings.TrimSpace(summary),
		})
	}
	return items, nil
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
		if body, err := page.Locator("body").InnerText(pw.LocatorInnerTextOptions{Timeout: pw.Float(1000)}) ; err == nil {
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

func buildDeepSeekCitationItems(cards []deepSeekCitationCard) []CitationItem {
	items := make([]CitationItem, 0, len(cards))
	for _, card := range cards {
		if strings.TrimSpace(card.Title) == "" && strings.TrimSpace(card.URL) == "" && strings.TrimSpace(card.Summary) == "" {
			continue
		}
		items = append(items, CitationItem{
			Title:   strings.TrimSpace(card.Title),
			URL:     strings.TrimSpace(card.URL),
			Icon:    strings.TrimSpace(card.Icon),
			Snippet: strings.TrimSpace(card.Summary),
			Source:  "deepseek",
		})
	}
	return items
}
