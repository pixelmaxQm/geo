package playwright

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	pw "github.com/playwright-community/playwright-go"
)

var (
	pwInstance *pw.Playwright
	browser    pw.Browser
)

// Install 确保 Playwright 浏览器已安装（首次运行时调用）
func Install() error {
	return pw.Install()
}

// Launch 启动浏览器，应在应用启动时调用一次
func Launch() error {
	var err error
	pwInstance, err = pw.Run()
	if err != nil {
		return fmt.Errorf("启动 Playwright 失败: %w", err)
	}

	browser, err = pwInstance.Chromium.Launch(pw.BrowserTypeLaunchOptions{
		Headless: pw.Bool(true),
		Args: []string{
			"--no-sandbox",
			"--disable-setuid-sandbox",
		},
	})
	if err != nil {
		return fmt.Errorf("启动 Chromium 失败: %w", err)
	}

	// 监听进程退出信号，优雅关闭浏览器
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		Close()
		os.Exit(0)
	}()

	return nil
}

// NewPage 创建新页面，返回 page 和 cleanup 函数
func NewPage() (pw.Page, func(), error) {
	if browser == nil {
		return nil, nil, fmt.Errorf("浏览器未启动，请先调用 Launch()")
	}
	ctx, err := browser.NewContext(pw.BrowserNewContextOptions{
		UserAgent: pw.String("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("创建浏览器上下文失败: %w", err)
	}
	page, err := ctx.NewPage()
	if err != nil {
		ctx.Close()
		return nil, nil, fmt.Errorf("创建页面失败: %w", err)
	}
	cleanup := func() {
		page.Close()
		ctx.Close()
	}
	return page, cleanup, nil
}

// Close 关闭浏览器和 Playwright 实例
func Close() {
	if browser != nil {
		browser.Close()
	}
	if pwInstance != nil {
		pwInstance.Stop()
	}
}
