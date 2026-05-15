package service

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/model"
	geoReq "github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/model/request"
	pwutils "github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/utils/playwright"
	pw "github.com/playwright-community/playwright-go"
	"gorm.io/gorm"
)

const (
	PlaywrightSessionStatusPending     = "pending"
	PlaywrightSessionStatusWaitingScan = "waiting_scan"
	PlaywrightSessionStatusAuthorized  = "authorized"
	PlaywrightSessionStatusFailed      = "failed"
	PlaywrightSessionStatusExpired     = "expired"
)

type playwrightSession struct{}

type StartPlaywrightSessionResponse struct {
	SessionID      uint   `json:"sessionId"`
	Status         string `json:"status"`
	QrImagePath    string `json:"qrImagePath"`
	ScreenshotPath string `json:"screenshotPath"`
}

type playwrightSessionRuntime struct {
	sessionID  uint
	platformID uint
	page       pw.Page
	cleanup    func()
	stopCh     chan struct{}
	doneCh     chan struct{}
}

type playwrightRuntimeManager struct {
	mu           sync.Mutex
	bySessionID  map[uint]*playwrightSessionRuntime
	byPlatformID map[uint]*playwrightSessionRuntime
}

var sessionRuntimeManager = &playwrightRuntimeManager{
	bySessionID:  map[uint]*playwrightSessionRuntime{},
	byPlatformID: map[uint]*playwrightSessionRuntime{},
}

func (m *playwrightRuntimeManager) add(runtime *playwrightSessionRuntime) {
	if runtime == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bySessionID[runtime.sessionID] = runtime
	m.byPlatformID[runtime.platformID] = runtime
}

func (m *playwrightRuntimeManager) getBySessionID(sessionID uint) *playwrightSessionRuntime {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.bySessionID[sessionID]
}

func (m *playwrightRuntimeManager) remove(sessionID uint, platformID uint) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.bySessionID, sessionID)
	if platformID != 0 {
		delete(m.byPlatformID, platformID)
	}
}

func (m *playwrightRuntimeManager) stopBySessionID(sessionID uint) {
	runtime := m.getBySessionID(sessionID)
	if runtime == nil {
		return
	}
	select {
	case <-runtime.doneCh:
		return
	default:
	}
	select {
	case <-runtime.stopCh:
	default:
		close(runtime.stopCh)
	}
	<-runtime.doneCh
}

func (s *playwrightSession) GetCurrentAuthorizedByPlatform(platformID uint) (*model.PlaywrightAuthSession, error) {
	var session model.PlaywrightAuthSession
	err := global.GVA_DB.
		Where("platform_id = ? AND status = ? AND state_path <> ''", platformID, PlaywrightSessionStatusAuthorized).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Order("id desc").
		First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

func buildStartPlaywrightSessionResponse(session *model.PlaywrightAuthSession) StartPlaywrightSessionResponse {
	if session == nil {
		return StartPlaywrightSessionResponse{}
	}
	return StartPlaywrightSessionResponse{
		SessionID:      session.ID,
		Status:         session.Status,
		QrImagePath:    session.QrImagePath,
		ScreenshotPath: session.ScreenshotPath,
	}
}

func buildPlatformSessionSummary(session *model.PlaywrightAuthSession) *model.PlatformSessionSummary {
	if session == nil {
		return nil
	}
	return &model.PlatformSessionSummary{
		ID:             session.ID,
		Status:         session.Status,
		ScreenshotPath: session.ScreenshotPath,
		QrImagePath:    session.QrImagePath,
		StatePath:      session.StatePath,
		ExpiresAt:      session.ExpiresAt,
		UpdatedAt:      session.UpdatedAt,
	}
}

func (s *playwrightSession) AttachCurrentAuthorizedSession(platforms []model.Platform) error {
	for i := range platforms {
		if platforms[i].Mode != CollectModePlaywright {
			continue
		}
		session, err := s.GetCurrentAuthorizedByPlatform(platforms[i].ID)
		if err != nil {
			return err
		}
		platforms[i].CurrentAuthorizedSession = buildPlatformSessionSummary(session)
	}
	return nil
}

func (s *playwrightSession) AttachCurrentAuthorizedSessionToPlatform(platform *model.Platform) error {
	if platform == nil || platform.Mode != CollectModePlaywright {
		return nil
	}
	session, err := s.GetCurrentAuthorizedByPlatform(platform.ID)
	if err != nil {
		return err
	}
	platform.CurrentAuthorizedSession = buildPlatformSessionSummary(session)
	return nil
}

func (s *playwrightSession) Start(req geoReq.StartPlaywrightSessionRequest, userID uint, userName string, authorityID uint) (StartPlaywrightSessionResponse, error) {
	var platform model.Platform
	if err := global.GVA_DB.Where("id = ? AND mode = ?", req.PlatformID, CollectModePlaywright).First(&platform).Error; err != nil {
		return StartPlaywrightSessionResponse{}, err
	}
	currentSession, err := s.GetCurrentAuthorizedByPlatform(platform.ID)
	if err != nil {
		return StartPlaywrightSessionResponse{}, err
	}
	if currentSession != nil {
		return buildStartPlaywrightSessionResponse(currentSession), nil
	}

	now := time.Now()
	session := model.PlaywrightAuthSession{
		PlatformID:     platform.ID,
		PlatformCode:   platform.Code,
		Status:         PlaywrightSessionStatusPending,
		LoginURL:       platform.ApiBase,
		ScreenshotPath: filepath.ToSlash(filepath.Join("uploads", "file", fmt.Sprintf("gm-platform-%d-session-%d.png", platform.ID, now.UnixNano()))),
		StatePath:      filepath.ToSlash(filepath.Join("uploads", "file", fmt.Sprintf("gm-platform-%d-state-%d.json", platform.ID, now.UnixNano()))),
		CreatedBy:      userID,
		CreatedByName:  userName,
	}
	if err := global.GVA_DB.Create(&session).Error; err != nil {
		return StartPlaywrightSessionResponse{}, err
	}
	if err := s.startRuntime(&session); err != nil {
		s.failSession(session.ID, err.Error())
		return StartPlaywrightSessionResponse{}, err
	}
	updated, err := s.Get(session.ID, userID, authorityID)
	if err != nil {
		return StartPlaywrightSessionResponse{}, err
	}
	return buildStartPlaywrightSessionResponse(&updated), nil
}

func (s *playwrightSession) startRuntime(session *model.PlaywrightAuthSession) error {
	if session == nil {
		return nil
	}
	sessionRuntimeManager.stopBySessionID(session.ID)
	page, cleanup, err := pwutilsNewPageForSession()
	if err != nil {
		return err
	}
	runtime := &playwrightSessionRuntime{
		sessionID:  session.ID,
		platformID: session.PlatformID,
		page:       page,
		cleanup:    cleanup,
		stopCh:     make(chan struct{}),
		doneCh:     make(chan struct{}),
	}
	sessionRuntimeManager.add(runtime)
	go s.runSessionRuntime(runtime, *session)
	return nil
}

func (s *playwrightSession) runSessionRuntime(runtime *playwrightSessionRuntime, session model.PlaywrightAuthSession) {
	defer func() {
		if runtime.cleanup != nil {
			runtime.cleanup()
		}
		sessionRuntimeManager.remove(runtime.sessionID, runtime.platformID)
		close(runtime.doneCh)
	}()

	page := runtime.page
	if _, err := page.Goto(session.LoginURL, pw.PageGotoOptions{WaitUntil: pw.WaitUntilStateDomcontentloaded, Timeout: pw.Float(30000)}); err != nil {
		s.failSession(session.ID, fmt.Sprintf("打开登录页失败: %v", err))
		return
	}

	deadline := time.Now().Add(5 * time.Minute)
	for time.Now().Before(deadline) {
		select {
		case <-runtime.stopCh:
			return
		default:
		}

		status, qrImagePath, err := s.captureRuntimeSnapshot(&session, page)
		if err != nil {
			s.failSession(session.ID, err.Error())
			return
		}
		if err := global.GVA_DB.Model(&model.PlaywrightAuthSession{}).Where("id = ?", session.ID).Updates(map[string]any{
			"status":          status,
			"screenshot_path": session.ScreenshotPath,
			"qr_image_path":   qrImagePath,
			"error_msg":       "",
		}).Error; err != nil {
			s.failSession(session.ID, err.Error())
			return
		}
		if status == PlaywrightSessionStatusAuthorized {
			return
		}
		time.Sleep(2 * time.Second)
	}
	_ = global.GVA_DB.Model(&model.PlaywrightAuthSession{}).Where("id = ? AND status <> ?", session.ID, PlaywrightSessionStatusAuthorized).Update("status", PlaywrightSessionStatusExpired).Error
}

func (s *playwrightSession) captureRuntimeSnapshot(session *model.PlaywrightAuthSession, page pw.Page) (string, string, error) {
	if session == nil {
		return PlaywrightSessionStatusFailed, "", nil
	}
	inputVisible := hasVisibleLocator(page, inputSelectorForPlatform(session.PlatformCode))
	if isAuthorizedSession(page, session.PlatformCode) {
		if err := s.saveAuthorizedSession(session.ID, session, page); err != nil {
			return PlaywrightSessionStatusFailed, "", err
		}
		_ = global.GVA_DB.Model(&model.PlaywrightAuthSession{}).Where("id = ?", session.ID).Update("error_msg", buildSessionDiagnosticsMessage(PlaywrightSessionStatusAuthorized, page.URL(), pageTitle(page), inputVisible)).Error
		return PlaywrightSessionStatusAuthorized, "", nil
	}
	if isUnauthorizedSession(page, session.PlatformCode) {
		waitForLoginQRCode(page)
	}
	if err := ensureParentDir(session.ScreenshotPath); err == nil && session.ScreenshotPath != "" {
		_, _ = page.Screenshot(pw.PageScreenshotOptions{Path: pw.String(session.ScreenshotPath), FullPage: pw.Bool(true)})
	}
	diagnostics := buildSessionDiagnosticsMessage(PlaywrightSessionStatusWaitingScan, page.URL(), pageTitle(page), inputVisible)
	_ = global.GVA_DB.Model(&model.PlaywrightAuthSession{}).Where("id = ?", session.ID).Update("error_msg", diagnostics).Error
	return PlaywrightSessionStatusWaitingScan, session.ScreenshotPath, nil
}

func (s *playwrightSession) saveAuthorizedSession(id uint, session *model.PlaywrightAuthSession, page pw.Page) error {
	markCurrentAuthorizedSessionExpiredByPlatform(session.PlatformID)
	statePath := session.StatePath
	if statePath == "" {
		statePath = filepath.ToSlash(filepath.Join("uploads", "file", fmt.Sprintf("gm-platform-%d-state-%d.json", session.PlatformID, time.Now().UnixNano())))
	}
	if err := ensureParentDir(statePath); err != nil {
		s.failSession(id, fmt.Sprintf("创建登录态目录失败: %v", err))
		return err
	}
	if err := ensureParentDir(session.ScreenshotPath); err == nil && session.ScreenshotPath != "" {
		_, _ = page.Screenshot(pw.PageScreenshotOptions{Path: pw.String(session.ScreenshotPath), FullPage: pw.Bool(true)})
	}
	if _, err := page.Context().StorageState(statePath); err != nil {
		s.failSession(id, fmt.Sprintf("保存登录态失败: %v", err))
		return err
	}
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	if err := global.GVA_DB.Model(&model.PlaywrightAuthSession{}).Where("id = ?", id).Updates(map[string]any{
		"status":          PlaywrightSessionStatusAuthorized,
		"state_path":      statePath,
		"expires_at":      &expiresAt,
		"screenshot_path": session.ScreenshotPath,
		"qr_image_path":   "",
		"error_msg":       "",
	}).Error; err != nil {
		return err
	}
	session.StatePath = statePath
	session.ExpiresAt = &expiresAt
	return nil
}

func isUnauthorizedSession(page pw.Page, platformCode string) bool {
	switch platformCode {
	case "deepseek":
		return page.URL() == "https://chat.deepseek.com/sign_in"
	default:
		return false
	}
}

func markCurrentAuthorizedSessionExpiredByPlatform(platformID uint) {
	if global.GVA_DB == nil || platformID == 0 {
		return
	}
	_ = global.GVA_DB.Model(&model.PlaywrightAuthSession{}).
		Where("platform_id = ? AND status = ?", platformID, PlaywrightSessionStatusAuthorized).
		Updates(map[string]any{"status": PlaywrightSessionStatusExpired, "expires_at": time.Now()}).Error
}

func markSessionExpiredByStatePath(statePath string) {
	if global.GVA_DB == nil || statePath == "" {
		return
	}
	_ = global.GVA_DB.Model(&model.PlaywrightAuthSession{}).
		Where("state_path = ? AND status = ?", statePath, PlaywrightSessionStatusAuthorized).
		Updates(map[string]any{"status": PlaywrightSessionStatusExpired, "expires_at": time.Now()}).Error
}

func IsPlaywrightAuthExpired(platformCode string, raw string) bool {
	if platformCode != "deepseek" || raw == "" {
		return false
	}
	return strings.Contains(raw, "https://chat.deepseek.com/sign_in") || strings.Contains(raw, "可能需要登录授权")
}

func (s *playwrightSession) Get(id uint, userID uint, authorityID uint) (model.PlaywrightAuthSession, error) {
	var session model.PlaywrightAuthSession
	db := global.GVA_DB.Where("id = ?", id)
	if authorityID != superAdminAuthorityID {
		db = db.Where("created_by = ?", userID)
	}
	err := db.First(&session).Error
	return session, err
}

func (s *playwrightSession) Refresh(id uint, userID uint, authorityID uint) (model.PlaywrightAuthSession, error) {
	session, err := s.Get(id, userID, authorityID)
	if err != nil {
		return model.PlaywrightAuthSession{}, err
	}
	sessionRuntimeManager.stopBySessionID(session.ID)
	now := time.Now()
	relogin := model.PlaywrightAuthSession{
		PlatformID:     session.PlatformID,
		PlatformCode:   session.PlatformCode,
		Status:         PlaywrightSessionStatusPending,
		LoginURL:       session.LoginURL,
		ScreenshotPath: filepath.ToSlash(filepath.Join("uploads", "file", fmt.Sprintf("gm-platform-%d-session-%d.png", session.PlatformID, now.UnixNano()))),
		StatePath:      filepath.ToSlash(filepath.Join("uploads", "file", fmt.Sprintf("gm-platform-%d-state-%d.json", session.PlatformID, now.UnixNano()))),
		CreatedBy:      session.CreatedBy,
		CreatedByName:  session.CreatedByName,
	}
	if err := global.GVA_DB.Create(&relogin).Error; err != nil {
		return model.PlaywrightAuthSession{}, err
	}
	if err := s.startRuntime(&relogin); err != nil {
		s.failSession(relogin.ID, err.Error())
		return model.PlaywrightAuthSession{}, err
	}
	return relogin, nil
}

func (s *playwrightSession) Delete(id uint, userID uint, authorityID uint) error {
	session, err := s.Get(id, userID, authorityID)
	if err != nil {
		return err
	}
	sessionRuntimeManager.stopBySessionID(session.ID)
	return global.GVA_DB.Delete(&session).Error
}

func (s *playwrightSession) failSession(id uint, msg string) {
	if global.GVA_DB == nil || id == 0 {
		return
	}
	_ = global.GVA_DB.Model(&model.PlaywrightAuthSession{}).Where("id = ?", id).Updates(map[string]any{"status": PlaywrightSessionStatusFailed, "error_msg": msg}).Error
}

func pwutilsNewPageForSession() (pw.Page, func(), error) {
	return pwutils.NewPage()
}

func isAuthorizedSession(page pw.Page, platformCode string) bool {
	switch platformCode {
	case "deepseek":
		return !strings.Contains(page.URL(), "/sign_in") && hasVisibleLocator(page, inputSelectorForPlatform(platformCode))
	default:
		return false
	}
}

func inputSelectorForPlatform(platformCode string) string {
	switch platformCode {
	case "deepseek":
		return "textarea[placeholder='给 DeepSeek 发送消息 '], #chat-input, textarea[placeholder], .chat-input"
	default:
		return ""
	}
}

func buildSessionDiagnosticsMessage(status string, url string, title string, inputVisible bool) string {
	return fmt.Sprintf("status=%s url=%s title=%s inputVisible=%t", status, url, title, inputVisible)
}

func pageTitle(page pw.Page) string {
	title, err := page.Title()
	if err != nil {
		return ""
	}
	return title
}

func hasVisibleLocator(page pw.Page, selector string) bool {
	locator := page.Locator(selector).First()
	if err := locator.WaitFor(pw.LocatorWaitForOptions{State: pw.WaitForSelectorStateVisible, Timeout: pw.Float(1000)}); err != nil {
		return false
	}
	return true
}

func waitForLoginQRCode(page pw.Page) {
	_ = page.WaitForLoadState(pw.PageWaitForLoadStateOptions{State: pw.LoadStateNetworkidle, Timeout: pw.Float(5000)})
	selectors := []string{
		"canvas",
		"img[src*='qr']",
		"img[alt*='二维码']",
		"img[alt*='QR']",
		"[class*='qr']",
		"[class*='QRCode']",
		"[class*='qrcode']",
	}
	for _, selector := range selectors {
		locator := page.Locator(selector).First()
		if err := locator.WaitFor(pw.LocatorWaitForOptions{State: pw.WaitForSelectorStateVisible, Timeout: pw.Float(1500)}); err == nil {
			time.Sleep(1 * time.Second)
			return
		}
	}
	time.Sleep(3 * time.Second)
}

func ensureParentDir(path string) error {
	if path == "" {
		return nil
	}
	return os.MkdirAll(filepath.Dir(path), 0755)
}
