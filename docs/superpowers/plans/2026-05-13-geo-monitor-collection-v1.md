# Geo Monitor Collection V1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first geo-monitor collection flow that lets a user manually execute one monitor topic against multiple platforms in one mode, persist one task plus many per-platform results, and inspect the stored results without adding analysis or scheduling.

**Architecture:** Add two new persistence models (`CollectionTask`, `CollectionResult`), a thin collector service layer that wraps the existing non-unified `utils/api` and `utils/playwright` packages into one `CollectOutput` shape, and three APIs (`run`, `task detail`, `result list`). The execution path is topic-driven, enforces topic ownership, runs platforms concurrently with isolated failures, and stores all outcomes for later UI inspection.

**Tech Stack:** Go, Gin, GORM, gin-vue-admin response/request patterns, `golang.org/x/sync/errgroup`, Vue 3, Element Plus.

---

## File Structure

### Backend files to create
- `server/plugin/geoMonitor/model/collection_task.go` — task model for one manual execution.
- `server/plugin/geoMonitor/model/collection_result.go` — per-platform result model.
- `server/plugin/geoMonitor/model/request/collect.go` — request/query structs for run/detail/result-list APIs.
- `server/plugin/geoMonitor/service/collector_types.go` — unified collector output types and status constants.
- `server/plugin/geoMonitor/service/collector.go` — orchestration: validate topic/platforms, create task, run concurrent collection, persist results, finalize task.
- `server/plugin/geoMonitor/service/api_collector.go` — mode=`api` adapter from `utils/api/*` to unified `CollectOutput`.
- `server/plugin/geoMonitor/service/playwright_collector.go` — mode=`playwright` adapter from `utils/playwright/*` to unified `CollectOutput`.
- `server/plugin/geoMonitor/api/collect.go` — run task, get task detail, get result list.
- `server/plugin/geoMonitor/router/collect.go` — collector route registration.

### Backend files to modify
- `server/plugin/geoMonitor/service/enter.go` — register collector service.
- `server/plugin/geoMonitor/api/enter.go` — register collector API.
- `server/plugin/geoMonitor/router/enter.go` — register collector router.
- `server/plugin/geoMonitor/initialize/router.go` — ensure collect routes mount with JWT/casbin.
- `server/plugin/geoMonitor/initialize/gorm.go` — migrate new models.
- `server/plugin/geoMonitor/initialize/api.go` — register new API permissions.
- `server/plugin/geoMonitor/initialize/menu.go` — add first collection page menu.
- `server/plugin/geoMonitor/plugin.go` — initialize Playwright browser lifecycle if needed for collection mode.
- `server/plugin/geoMonitor/service/topic.go` — if needed, expose or reuse ownership-aware topic lookup for execution.
- `server/plugin/geoMonitor/utils/api/*.go` — only where needed to add real collection functions alongside existing test functions.
- `server/plugin/geoMonitor/utils/playwright/*.go` — only where needed to add real collection functions alongside existing test functions.

### Frontend files to create
- `web/src/plugin/geoMonitor/api/collect.js` — collection API wrappers.
- `web/src/plugin/geoMonitor/view/collect.vue` — minimal manual execution page with result list.

### Frontend files to modify
- `web/src/plugin/geoMonitor/api/topic.js` — if a lightweight select/list helper is needed.
- `web/src/plugin/geoMonitor/api/platform.js` — if a lightweight enabled-platform query helper is needed.

---

### Task 1: Add collection models and migrate them

**Files:**
- Create: `server/plugin/geoMonitor/model/collection_task.go`
- Create: `server/plugin/geoMonitor/model/collection_result.go`
- Modify: `server/plugin/geoMonitor/initialize/gorm.go`
- Test: use `go test ./server/plugin/geoMonitor/...`

- [ ] **Step 1: Write the model files**

```go
// server/plugin/geoMonitor/model/collection_task.go
package model

import (
    "time"

    "github.com/flipped-aurora/gin-vue-admin/server/global"
)

type CollectionTask struct {
    global.GVA_MODEL
    TopicID          uint       `json:"topicId" gorm:"index;comment:监控话题ID"`
    TopicName        string     `json:"topicName" gorm:"comment:监控话题名称"`
    Prompt           string     `json:"prompt" gorm:"type:text;comment:本次执行提示词"`
    Mode             string     `json:"mode" gorm:"index;comment:采集模式"`
    Status           string     `json:"status" gorm:"index;comment:任务状态"`
    PlatformIDs      string     `json:"platformIds" gorm:"type:text;comment:平台ID列表JSON"`
    RequestedBy      uint       `json:"requestedBy" gorm:"index;comment:发起人ID"`
    RequestedByName  string     `json:"requestedByName" gorm:"comment:发起人用户名"`
    StartedAt        *time.Time `json:"startedAt" gorm:"comment:开始时间"`
    FinishedAt       *time.Time `json:"finishedAt" gorm:"comment:结束时间"`
    ErrorMsg         string     `json:"errorMsg" gorm:"type:text;comment:任务级错误信息"`
}

func (CollectionTask) TableName() string {
    return "gva_geo_monitor_collection_tasks"
}
```

```go
// server/plugin/geoMonitor/model/collection_result.go
package model

import "github.com/flipped-aurora/gin-vue-admin/server/global"

type CollectionResult struct {
    global.GVA_MODEL
    TaskID          uint   `json:"taskId" gorm:"index;comment:任务ID"`
    PlatformID      uint   `json:"platformId" gorm:"index;comment:平台ID"`
    PlatformName    string `json:"platformName" gorm:"comment:平台名称"`
    PlatformCode    string `json:"platformCode" gorm:"index;comment:平台编码"`
    Mode            string `json:"mode" gorm:"index;comment:采集模式"`
    Prompt          string `json:"prompt" gorm:"type:text;comment:本次执行提示词"`
    Answer          string `json:"answer" gorm:"type:longtext;comment:回答内容"`
    Status          string `json:"status" gorm:"index;comment:执行结果状态"`
    ErrorMsg        string `json:"errorMsg" gorm:"type:text;comment:错误信息"`
    Citations       string `json:"citations" gorm:"type:longtext;comment:引用信息JSON"`
    ScreenshotPath  string `json:"screenshotPath" gorm:"comment:截图路径"`
    DurationMs      int    `json:"durationMs" gorm:"comment:耗时毫秒"`
    RawResponse     string `json:"rawResponse" gorm:"type:longtext;comment:原始响应"`
}

func (CollectionResult) TableName() string {
    return "gva_geo_monitor_collection_results"
}
```

- [ ] **Step 2: Register the new models in AutoMigrate**

```go
// server/plugin/geoMonitor/initialize/gorm.go
err := global.GVA_DB.WithContext(ctx).AutoMigrate(
    new(model.Platform),
    new(model.MonitorTopic),
    new(model.CollectionTask),
    new(model.CollectionResult),
)
```

- [ ] **Step 3: Run package tests/compile check**

Run:
```bash
go test ./server/plugin/geoMonitor/...
```

Expected: package compile passes or reports the next missing types/imports only from later planned files.

- [ ] **Step 4: Commit**

```bash
git add server/plugin/geoMonitor/model/collection_task.go server/plugin/geoMonitor/model/collection_result.go server/plugin/geoMonitor/initialize/gorm.go
git commit -m "feat: add geo monitor collection models"
```

### Task 2: Add collection request and response shapes

**Files:**
- Create: `server/plugin/geoMonitor/model/request/collect.go`
- Create: `server/plugin/geoMonitor/service/collector_types.go`
- Modify: `server/plugin/geoMonitor/service/enter.go`
- Modify: `server/plugin/geoMonitor/api/enter.go`
- Test: use `go test ./server/plugin/geoMonitor/...`

- [ ] **Step 1: Add request/query structs**

```go
// server/plugin/geoMonitor/model/request/collect.go
package request

import common "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"

type RunCollectionRequest struct {
    TopicID     uint   `json:"topicId" binding:"required"`
    PlatformIDs []uint `json:"platformIds" binding:"required,min=1"`
    Mode        string `json:"mode" binding:"required,oneof=api playwright"`
}

type CollectionResultSearch struct {
    common.PageInfo
    TaskID uint `json:"taskId" form:"taskId" binding:"required"`
}
```

- [ ] **Step 2: Add unified collector types and constants**

```go
// server/plugin/geoMonitor/service/collector_types.go
package service

type CollectOutput struct {
    Answer         string `json:"answer"`
    Citations      string `json:"citations"`
    ScreenshotPath string `json:"screenshotPath"`
    DurationMs     int    `json:"durationMs"`
    RawResponse    string `json:"rawResponse"`
    ErrorMsg       string `json:"errorMsg"`
}

const (
    CollectModeAPI        = "api"
    CollectModePlaywright = "playwright"

    TaskStatusRunning = "running"
    TaskStatusDone    = "done"
    TaskStatusFailed  = "failed"
    TaskStatusPartial = "partial"

    ResultStatusSuccess = "success"
    ResultStatusFailed  = "failed"
)
```

- [ ] **Step 3: Register collector in enter files**

```go
// server/plugin/geoMonitor/service/enter.go
type service struct {
    Platform  platform
    Topic     topic
    Collector collector
}
```

```go
// server/plugin/geoMonitor/api/enter.go
var (
    Api              = new(api)
    platformService  = service.Service.Platform
    topicService     = service.Service.Topic
    collectorService = service.Service.Collector
)

type api struct {
    Platform  platform
    Topic     topic
    Collect   collect
}
```

- [ ] **Step 4: Run package tests/compile check**

Run:
```bash
go test ./server/plugin/geoMonitor/...
```

Expected: compile passes or only fails on the still-missing `collector`/`collect` concrete files planned next.

- [ ] **Step 5: Commit**

```bash
git add server/plugin/geoMonitor/model/request/collect.go server/plugin/geoMonitor/service/collector_types.go server/plugin/geoMonitor/service/enter.go server/plugin/geoMonitor/api/enter.go
git commit -m "feat: add geo monitor collect request types"
```

### Task 3: Build the API-mode collector adapter

**Files:**
- Create: `server/plugin/geoMonitor/service/api_collector.go`
- Modify: `server/plugin/geoMonitor/utils/api/deepseek.go`
- Modify: `server/plugin/geoMonitor/utils/api/qwen.go`
- Modify: `server/plugin/geoMonitor/utils/api/zhipu.go`
- Modify: `server/plugin/geoMonitor/utils/api/doubao.go`
- Modify: `server/plugin/geoMonitor/utils/api/kimi.go`
- Modify: `server/plugin/geoMonitor/utils/api/wenxin.go`
- Modify: `server/plugin/geoMonitor/utils/api/yuanbao.go`
- Test: use `go test ./server/plugin/geoMonitor/...`

- [ ] **Step 1: Add a real collect helper next to one existing test helper and use that pattern for all providers**

```go
// example shape to add in each server/plugin/geoMonitor/utils/api/<provider>.go
package api

import (
    "context"
    "encoding/json"
    "strings"
    "time"

    openai "github.com/sashabaranov/go-openai"
)

type APICollectResult struct {
    Answer      string
    Citations   string
    RawResponse string
}

func CollectDeepSeek(apiBase string, apiKey string, prompt string) (*APICollectResult, error) {
    config := openai.DefaultConfig(apiKey)
    config.BaseURL = strings.TrimRight(apiBase, "/") + "/v1"
    client := openai.NewClientWithConfig(config)

    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()

    resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
        Model: "deepseek-chat",
        Messages: []openai.ChatCompletionMessage{{Role: "user", Content: prompt}},
    })
    if err != nil {
        return nil, err
    }

    raw, _ := json.Marshal(resp)
    answer := ""
    if len(resp.Choices) > 0 {
        answer = resp.Choices[0].Message.Content
    }

    return &APICollectResult{
        Answer:      answer,
        Citations:   "[]",
        RawResponse: string(raw),
    }, nil
}
```

- [ ] **Step 2: Add the service adapter that normalizes provider outputs**

```go
// server/plugin/geoMonitor/service/api_collector.go
package service

import (
    "fmt"
    "time"

    "github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/model"
    apiutils "github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/utils/api"
)

func (s *collector) collectWithAPI(platform model.Platform, prompt string) (CollectOutput, error) {
    started := time.Now()

    var (
        result *apiutils.APICollectResult
        err    error
    )

    switch platform.Code {
    case "deepseek":
        result, err = apiutils.CollectDeepSeek(platform.ApiBase, platform.ApiKey, prompt)
    case "qwen":
        result, err = apiutils.CollectQwen(platform.ApiBase, platform.ApiKey, prompt)
    case "zhipu":
        result, err = apiutils.CollectZhipu(platform.ApiBase, platform.ApiKey, prompt)
    case "doubao":
        result, err = apiutils.CollectDoubao(platform.ApiBase, platform.ApiKey, prompt)
    case "kimi":
        result, err = apiutils.CollectKimi(platform.ApiBase, platform.ApiKey, prompt)
    case "wenxin":
        result, err = apiutils.CollectWenxin(platform.ApiBase, platform.ApiKey, prompt)
    case "yuanbao":
        result, err = apiutils.CollectYuanbao(platform.ApiBase, platform.ApiKey, prompt)
    default:
        return CollectOutput{}, fmt.Errorf("不支持的 API 平台: %s", platform.Code)
    }
    if err != nil {
        return CollectOutput{}, err
    }

    return CollectOutput{
        Answer:         result.Answer,
        Citations:      result.Citations,
        DurationMs:     int(time.Since(started).Milliseconds()),
        RawResponse:    result.RawResponse,
        ScreenshotPath: "",
    }, nil
}
```

- [ ] **Step 3: Run package tests/compile check**

Run:
```bash
go test ./server/plugin/geoMonitor/...
```

Expected: compile passes or only fails on the still-missing Playwright/collector orchestration files.

- [ ] **Step 4: Commit**

```bash
git add server/plugin/geoMonitor/service/api_collector.go server/plugin/geoMonitor/utils/api/*.go
git commit -m "feat: add geo monitor api collector adapters"
```

### Task 4: Build the Playwright-mode collector adapter

**Files:**
- Create: `server/plugin/geoMonitor/service/playwright_collector.go`
- Modify: `server/plugin/geoMonitor/utils/playwright/playwright.go`
- Modify: `server/plugin/geoMonitor/utils/playwright/deepseek.go`
- Modify: `server/plugin/geoMonitor/utils/playwright/qwen.go`
- Modify: `server/plugin/geoMonitor/utils/playwright/zhipu.go`
- Modify: `server/plugin/geoMonitor/utils/playwright/doubao.go`
- Modify: `server/plugin/geoMonitor/utils/playwright/kimi.go`
- Modify: `server/plugin/geoMonitor/utils/playwright/wenxin.go`
- Modify: `server/plugin/geoMonitor/utils/playwright/yuanbao.go`
- Modify: `server/plugin/geoMonitor/plugin.go`
- Test: use `go test ./server/plugin/geoMonitor/...`

- [ ] **Step 1: Extend the Playwright utility layer with a real collect result shape**

```go
// add in server/plugin/geoMonitor/utils/playwright/playwright.go
package playwright

type PageCollectResult struct {
    Answer         string
    ScreenshotPath string
    RawResponse    string
}
```

```go
// example pattern to add in each server/plugin/geoMonitor/utils/playwright/<provider>.go
func CollectDeepSeek(webURL string, prompt string, screenshotPath string) (*PageCollectResult, error) {
    page, cleanup, err := NewPage()
    if err != nil {
        return nil, err
    }
    defer cleanup()

    if _, err := page.Goto(webURL, pw.PageGotoOptions{Timeout: pw.Float(30000), WaitUntil: pw.WaitUntilStateDomcontentloaded}); err != nil {
        return nil, err
    }

    input := page.Locator("#chat-input, textarea[placeholder], .chat-input").First()
    if err := input.Fill(prompt); err != nil {
        return nil, err
    }
    if err := input.Press("Enter"); err != nil {
        return nil, err
    }

    answerNode := page.Locator(".markdown, .assistant-message, .message-content").Last()
    if _, err := answerNode.WaitFor(pw.LocatorWaitForOptions{Timeout: pw.Float(60000)}); err != nil {
        return nil, err
    }
    answer, err := answerNode.InnerText()
    if err != nil {
        return nil, err
    }
    if screenshotPath != "" {
        if _, err := page.Screenshot(pw.PageScreenshotOptions{Path: pw.String(screenshotPath), FullPage: pw.Bool(true)}); err != nil {
            return nil, err
        }
    }

    return &PageCollectResult{Answer: answer, ScreenshotPath: screenshotPath, RawResponse: answer}, nil
}
```

- [ ] **Step 2: Initialize Playwright once during plugin registration**

```go
// server/plugin/geoMonitor/plugin.go
func (p *plugin) Register(group *gin.Engine) {
    ctx := context.Background()
    initialize.Viper()
    initialize.Api(ctx)
    initialize.Menu(ctx)
    initialize.Dictionary(ctx)
    initialize.Gorm(ctx)
    initialize.Router(group)
    service.Service.Platform.InitSeedData()
    _ = service.Service.Collector.InitPlaywright()
}
```

```go
// add to server/plugin/geoMonitor/service/playwright_collector.go
func (s *collector) InitPlaywright() error {
    if err := playwright.Install(); err != nil {
        return err
    }
    return playwright.Launch()
}
```

- [ ] **Step 3: Add the service adapter that normalizes Playwright outputs**

```go
// server/plugin/geoMonitor/service/playwright_collector.go
package service

import (
    "fmt"
    "path/filepath"
    "time"

    "github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/model"
    pwutils "github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/utils/playwright"
)

func (s *collector) collectWithPlaywright(platform model.Platform, prompt string, taskID uint) (CollectOutput, error) {
    started := time.Now()
    screenshotPath := filepath.ToSlash(filepath.Join("uploads", "geo-monitor", fmt.Sprintf("task-%d-platform-%d.png", taskID, platform.ID)))

    var (
        result *pwutils.PageCollectResult
        err    error
    )

    switch platform.Code {
    case "deepseek":
        result, err = pwutils.CollectDeepSeek(platform.ApiBase, prompt, screenshotPath)
    case "qwen":
        result, err = pwutils.CollectQwen(platform.ApiBase, prompt, screenshotPath)
    case "zhipu":
        result, err = pwutils.CollectZhipu(platform.ApiBase, prompt, screenshotPath)
    case "doubao":
        result, err = pwutils.CollectDoubao(platform.ApiBase, prompt, screenshotPath)
    case "kimi":
        result, err = pwutils.CollectKimi(platform.ApiBase, prompt, screenshotPath)
    case "wenxin":
        result, err = pwutils.CollectWenxin(platform.ApiBase, prompt, screenshotPath)
    case "yuanbao":
        result, err = pwutils.CollectYuanbao(platform.ApiBase, prompt, screenshotPath)
    default:
        return CollectOutput{}, fmt.Errorf("不支持的 Playwright 平台: %s", platform.Code)
    }
    if err != nil {
        return CollectOutput{}, err
    }

    return CollectOutput{
        Answer:         result.Answer,
        Citations:      "[]",
        ScreenshotPath: result.ScreenshotPath,
        DurationMs:     int(time.Since(started).Milliseconds()),
        RawResponse:    result.RawResponse,
    }, nil
}
```

- [ ] **Step 4: Run package tests/compile check**

Run:
```bash
go test ./server/plugin/geoMonitor/...
```

Expected: compile passes or only fails on the still-missing top-level collector orchestration/API files.

- [ ] **Step 5: Commit**

```bash
git add server/plugin/geoMonitor/service/playwright_collector.go server/plugin/geoMonitor/utils/playwright/*.go server/plugin/geoMonitor/plugin.go
git commit -m "feat: add geo monitor playwright collector adapters"
```

### Task 5: Build the collection orchestration service

**Files:**
- Create: `server/plugin/geoMonitor/service/collector.go`
- Modify: `server/plugin/geoMonitor/service/topic.go`
- Test: use `go test ./server/plugin/geoMonitor/...`

- [ ] **Step 1: Add ownership-aware topic lookup helper if needed**

```go
// add in server/plugin/geoMonitor/service/topic.go
func (s *topic) GetExecutableTopic(id uint, userID uint, authorityID uint) (info model.MonitorTopic, err error) {
    return s.GetTopic(id, userID, authorityID)
}
```

- [ ] **Step 2: Add the collector orchestration service**

```go
// server/plugin/geoMonitor/service/collector.go
package service

import (
    "encoding/json"
    "fmt"
    "sync"
    "time"

    "github.com/flipped-aurora/gin-vue-admin/server/global"
    "github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/model"
    geoReq "github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/model/request"
    "golang.org/x/sync/errgroup"
)

type collector struct{}

type RunCollectionResponse struct {
    TaskID  uint                     `json:"taskId"`
    Status  string                   `json:"status"`
    Results []model.CollectionResult `json:"results"`
}

func (s *collector) Run(req geoReq.RunCollectionRequest, userID uint, userName string, authorityID uint) (RunCollectionResponse, error) {
    topic, err := Service.Topic.GetExecutableTopic(req.TopicID, userID, authorityID)
    if err != nil {
        return RunCollectionResponse{}, err
    }

    var platforms []model.Platform
    if err := global.GVA_DB.Where("id IN ? AND mode = ? AND status = ?", req.PlatformIDs, req.Mode, 1).Find(&platforms).Error; err != nil {
        return RunCollectionResponse{}, err
    }
    if len(platforms) == 0 {
        return RunCollectionResponse{}, fmt.Errorf("未找到可执行平台")
    }

    platformIDsJSON, _ := json.Marshal(req.PlatformIDs)
    startedAt := time.Now()
    task := model.CollectionTask{
        TopicID:         topic.ID,
        TopicName:       topic.Name,
        Prompt:          topic.Prompt,
        Mode:            req.Mode,
        Status:          TaskStatusRunning,
        PlatformIDs:     string(platformIDsJSON),
        RequestedBy:     userID,
        RequestedByName: userName,
        StartedAt:       &startedAt,
    }
    if err := global.GVA_DB.Create(&task).Error; err != nil {
        return RunCollectionResponse{}, err
    }

    results := make([]model.CollectionResult, 0, len(platforms))
    var mu sync.Mutex
    g := new(errgroup.Group)
    g.SetLimit(3)

    for _, platform := range platforms {
        platform := platform
        g.Go(func() error {
            output, collectErr := s.collectOne(task.ID, platform, topic.Prompt, req.Mode)
            record := model.CollectionResult{
                TaskID:         task.ID,
                PlatformID:     platform.ID,
                PlatformName:   platform.Name,
                PlatformCode:   platform.Code,
                Mode:           req.Mode,
                Prompt:         topic.Prompt,
                Answer:         output.Answer,
                Status:         ResultStatusSuccess,
                ErrorMsg:       output.ErrorMsg,
                Citations:      output.Citations,
                ScreenshotPath: output.ScreenshotPath,
                DurationMs:     output.DurationMs,
                RawResponse:    output.RawResponse,
            }
            if collectErr != nil {
                record.Status = ResultStatusFailed
                record.ErrorMsg = collectErr.Error()
            }
            if err := global.GVA_DB.Create(&record).Error; err != nil {
                return err
            }
            mu.Lock()
            results = append(results, record)
            mu.Unlock()
            return nil
        })
    }

    groupErr := g.Wait()
    finishedAt := time.Now()
    status := summarizeTaskStatus(results)
    update := map[string]any{"status": status, "finished_at": &finishedAt}
    if groupErr != nil {
        update["status"] = TaskStatusPartial
        update["error_msg"] = groupErr.Error()
    }
    if err := global.GVA_DB.Model(&model.CollectionTask{}).Where("id = ?", task.ID).Updates(update).Error; err != nil {
        return RunCollectionResponse{}, err
    }

    return RunCollectionResponse{TaskID: task.ID, Status: update["status"].(string), Results: results}, nil
}

func (s *collector) collectOne(taskID uint, platform model.Platform, prompt string, mode string) (CollectOutput, error) {
    switch mode {
    case CollectModeAPI:
        return s.collectWithAPI(platform, prompt)
    case CollectModePlaywright:
        return s.collectWithPlaywright(platform, prompt, taskID)
    default:
        return CollectOutput{}, fmt.Errorf("不支持的采集模式: %s", mode)
    }
}

func summarizeTaskStatus(results []model.CollectionResult) string {
    if len(results) == 0 {
        return TaskStatusFailed
    }
    successCount := 0
    for _, result := range results {
        if result.Status == ResultStatusSuccess {
            successCount++
        }
    }
    switch {
    case successCount == len(results):
        return TaskStatusDone
    case successCount == 0:
        return TaskStatusFailed
    default:
        return TaskStatusPartial
    }
}

func (s *collector) GetTask(id uint, userID uint, authorityID uint) (model.CollectionTask, error) {
    var task model.CollectionTask
    db := global.GVA_DB.Where("id = ?", id)
    if authorityID != superAdminAuthorityID {
        db = db.Where("requested_by = ?", userID)
    }
    err := db.First(&task).Error
    return task, err
}

func (s *collector) GetResultList(taskID uint, page int, pageSize int, userID uint, authorityID uint) ([]model.CollectionResult, int64, error) {
    task, err := s.GetTask(taskID, userID, authorityID)
    if err != nil {
        return nil, 0, err
    }
    _ = task

    var list []model.CollectionResult
    var total int64
    db := global.GVA_DB.Model(&model.CollectionResult{}).Where("task_id = ?", taskID)
    if err := db.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    err = db.Order("id desc").Limit(pageSize).Offset(pageSize * (page - 1)).Find(&list).Error
    return list, total, err
}
```

- [ ] **Step 3: Run package tests/compile check**

Run:
```bash
go test ./server/plugin/geoMonitor/...
```

Expected: compile passes or only fails on the still-missing API/router/frontend files.

- [ ] **Step 4: Commit**

```bash
git add server/plugin/geoMonitor/service/collector.go server/plugin/geoMonitor/service/topic.go
git commit -m "feat: add geo monitor collection orchestration"
```

### Task 6: Expose the collection APIs and routes

**Files:**
- Create: `server/plugin/geoMonitor/api/collect.go`
- Create: `server/plugin/geoMonitor/router/collect.go`
- Modify: `server/plugin/geoMonitor/router/enter.go`
- Modify: `server/plugin/geoMonitor/initialize/api.go`
- Modify: `server/plugin/geoMonitor/initialize/menu.go`
- Test: use `go test ./server/plugin/geoMonitor/...`

- [ ] **Step 1: Add the collector API handlers**

```go
// server/plugin/geoMonitor/api/collect.go
package api

import (
    "strconv"

    "github.com/flipped-aurora/gin-vue-admin/server/global"
    "github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
    geoReq "github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/model/request"
    "github.com/flipped-aurora/gin-vue-admin/server/utils"
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

var Collect = new(collect)

type collect struct{}

func (a *collect) Run(c *gin.Context) {
    var req geoReq.RunCollectionRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.FailWithMessage(err.Error(), c)
        return
    }
    data, err := collectorService.Run(req, utils.GetUserID(c), utils.GetUserName(c), utils.GetUserAuthorityId(c))
    if err != nil {
        global.GVA_LOG.Error("执行采集失败!", zap.Error(err))
        response.FailWithMessage("执行失败", c)
        return
    }
    response.OkWithDetailed(data, "执行成功", c)
}

func (a *collect) GetTask(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        response.FailWithMessage("参数错误", c)
        return
    }
    data, err := collectorService.GetTask(uint(id), utils.GetUserID(c), utils.GetUserAuthorityId(c))
    if err != nil {
        global.GVA_LOG.Error("获取任务失败!", zap.Error(err))
        response.FailWithMessage("获取失败", c)
        return
    }
    response.OkWithData(data, c)
}

func (a *collect) GetResultList(c *gin.Context) {
    var req geoReq.CollectionResultSearch
    if err := c.ShouldBindQuery(&req); err != nil {
        response.FailWithMessage(err.Error(), c)
        return
    }
    if req.Page == 0 {
        req.Page = 1
    }
    if req.PageSize == 0 {
        req.PageSize = 10
    }
    list, total, err := collectorService.GetResultList(req.TaskID, req.Page, req.PageSize, utils.GetUserID(c), utils.GetUserAuthorityId(c))
    if err != nil {
        global.GVA_LOG.Error("获取结果列表失败!", zap.Error(err))
        response.FailWithMessage("获取失败", c)
        return
    }
    response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, "获取成功", c)
}
```

- [ ] **Step 2: Add the router and register it**

```go
// server/plugin/geoMonitor/router/collect.go
package router

import "github.com/gin-gonic/gin"

type collect struct{}

func (r *collect) Init(public *gin.RouterGroup, private *gin.RouterGroup) {
    group := private.Group("geoMonitor")
    group.POST("collect/run", apiCollect.Run)
    group.GET("collect/task/:id", apiCollect.GetTask)
    group.GET("collect/result/list", apiCollect.GetResultList)
}
```

```go
// server/plugin/geoMonitor/router/enter.go
var (
    Router      = new(router)
    apiPlatform = api.Api.Platform
    apiTopic    = api.Api.Topic
    apiCollect  = api.Api.Collect
)

type router struct {
    Platform platform
    Topic    topic
    Collect  collect
}
```

- [ ] **Step 3: Register API permissions and menu entry**

```go
// add in server/plugin/geoMonitor/initialize/api.go
{
    Path: "/geoMonitor/collect/run", Description: "执行采集", ApiGroup: "geoMonitor采集", Method: "POST",
},
{
    Path: "/geoMonitor/collect/task/:id", Description: "采集任务详情", ApiGroup: "geoMonitor采集", Method: "GET",
},
{
    Path: "/geoMonitor/collect/result/list", Description: "采集结果列表", ApiGroup: "geoMonitor采集", Method: "GET",
},
```

```go
// add in server/plugin/geoMonitor/initialize/menu.go
{
    ParentId: 0,
    Path: "geoMonitorCollect",
    Name: "geoMonitorCollect",
    Hidden: false,
    Component: "plugin/geoMonitor/view/collect.vue",
    Sort: 3,
    Meta: model.Meta{Title: "问题采集", Icon: "tickets"},
},
```

- [ ] **Step 4: Run package tests/compile check**

Run:
```bash
go test ./server/plugin/geoMonitor/...
```

Expected: backend package compile passes.

- [ ] **Step 5: Commit**

```bash
git add server/plugin/geoMonitor/api/collect.go server/plugin/geoMonitor/router/collect.go server/plugin/geoMonitor/router/enter.go server/plugin/geoMonitor/initialize/api.go server/plugin/geoMonitor/initialize/menu.go
git commit -m "feat: expose geo monitor collection apis"
```

### Task 7: Add the frontend execution page and API client

**Files:**
- Create: `web/src/plugin/geoMonitor/api/collect.js`
- Create: `web/src/plugin/geoMonitor/view/collect.vue`
- Modify: `web/src/plugin/geoMonitor/api/topic.js`
- Modify: `web/src/plugin/geoMonitor/api/platform.js`
- Test: browser manual test + frontend build if available

- [ ] **Step 1: Add the collection API client**

```js
// web/src/plugin/geoMonitor/api/collect.js
import service from '@/utils/request'

export const runCollection = (data) => {
  return service({
    url: '/geoMonitor/collect/run',
    method: 'post',
    data
  })
}

export const getCollectionTask = (id) => {
  return service({
    url: `/geoMonitor/collect/task/${id}`,
    method: 'get'
  })
}

export const getCollectionResultList = (params) => {
  return service({
    url: '/geoMonitor/collect/result/list',
    method: 'get',
    params
  })
}
```

- [ ] **Step 2: Add lightweight selectors for enabled topics/platforms if current APIs are too table-oriented**

```js
// in web/src/plugin/geoMonitor/api/topic.js
export const getAllTopics = () => {
  return service({
    url: '/geoMonitor/topic/list',
    method: 'get',
    params: { page: 1, pageSize: 1000, status: 1 }
  })
}
```

```js
// in web/src/plugin/geoMonitor/api/platform.js
export const getEnabledPlatformsByMode = (mode) => {
  return service({
    url: '/geoMonitor/platform/list',
    method: 'get',
    params: { page: 1, pageSize: 1000, mode, status: 1 }
  })
}
```

- [ ] **Step 3: Build the minimal execution page**

```vue
<!-- web/src/plugin/geoMonitor/view/collect.vue -->
<template>
  <div>
    <div class="gva-search-box">
      <el-form :model="formData" label-position="top">
        <el-form-item label="监控话题">
          <el-select v-model="formData.topicId" placeholder="请选择话题" style="width: 100%">
            <el-option v-for="item in topicOptions" :key="item.ID" :label="item.name" :value="item.ID" />
          </el-select>
        </el-form-item>
        <el-form-item label="采集模式">
          <el-radio-group v-model="formData.mode" @change="loadPlatforms">
            <el-radio value="api">API</el-radio>
            <el-radio value="playwright">Playwright</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="选择平台">
          <el-select v-model="formData.platformIds" multiple placeholder="请选择平台" style="width: 100%">
            <el-option v-for="item in platformOptions" :key="item.ID" :label="item.name" :value="item.ID" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="running" @click="submitRun">执行采集</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="gva-table-box" v-if="taskId">
      <div class="mb-4">任务ID：{{ taskId }} / 状态：{{ taskStatus }}</div>
      <el-table :data="resultList">
        <el-table-column prop="platformName" label="平台" min-width="140" />
        <el-table-column prop="status" label="状态" width="100" />
        <el-table-column prop="durationMs" label="耗时(ms)" width="120" />
        <el-table-column prop="answer" label="回答摘要" min-width="320" show-overflow-tooltip />
        <el-table-column label="截图" width="120">
          <template #default="scope">
            <el-link v-if="scope.row.screenshotPath" :href="scope.row.screenshotPath" target="_blank">查看</el-link>
            <span v-else>-</span>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getAllTopics } from '@/plugin/geoMonitor/api/topic'
import { getEnabledPlatformsByMode } from '@/plugin/geoMonitor/api/platform'
import { runCollection, getCollectionResultList } from '@/plugin/geoMonitor/api/collect'

defineOptions({ name: 'GeoMonitorCollect' })

const formData = ref({ topicId: '', mode: 'api', platformIds: [] })
const topicOptions = ref([])
const platformOptions = ref([])
const resultList = ref([])
const taskId = ref(0)
const taskStatus = ref('')
const running = ref(false)

const loadTopics = async () => {
  const res = await getAllTopics()
  if (res.code === 0) topicOptions.value = res.data.list
}

const loadPlatforms = async () => {
  formData.value.platformIds = []
  const res = await getEnabledPlatformsByMode(formData.value.mode)
  if (res.code === 0) platformOptions.value = res.data.list
}

const submitRun = async () => {
  if (!formData.value.topicId || formData.value.platformIds.length === 0) {
    ElMessage({ type: 'warning', message: '请先选择话题和平台' })
    return
  }
  running.value = true
  const res = await runCollection(formData.value)
  running.value = false
  if (res.code !== 0) return
  taskId.value = res.data.taskId
  taskStatus.value = res.data.status
  await loadResults()
  ElMessage({ type: 'success', message: '执行完成' })
}

const loadResults = async () => {
  const res = await getCollectionResultList({ taskId: taskId.value, page: 1, pageSize: 100 })
  if (res.code === 0) resultList.value = res.data.list
}

onMounted(async () => {
  await loadTopics()
  await loadPlatforms()
})
</script>
```

- [ ] **Step 4: Start the frontend dev server and verify the golden path in a browser**

Run:
```bash
npm run dev
```

Manual test:
1. Open the geo-monitor collection page.
2. Confirm topics only include topics the current user can see.
3. Switch mode between `api` and `playwright` and confirm platform options reload.
4. Run one task against at least two platforms.
5. Verify one task ID is returned and multiple results appear.
6. If using Playwright mode, verify screenshot links render for successful results.

Expected: the page loads, executes one task, and displays persisted results.

- [ ] **Step 5: Commit**

```bash
git add web/src/plugin/geoMonitor/api/collect.js web/src/plugin/geoMonitor/view/collect.vue web/src/plugin/geoMonitor/api/topic.js web/src/plugin/geoMonitor/api/platform.js
git commit -m "feat: add geo monitor manual collection page"
```

### Task 8: End-to-end verification and cleanup

**Files:**
- Modify: any touched files from earlier tasks if test fixes are needed
- Test: backend compile + frontend manual flow

- [ ] **Step 1: Run the backend package tests**

Run:
```bash
go test ./server/plugin/geoMonitor/...
```

Expected: PASS for the geo-monitor package tree.

- [ ] **Step 2: Run a repo-level targeted build if available**

Run:
```bash
go test ./server/...
```

Expected: PASS, or only unrelated pre-existing failures outside geo-monitor.

- [ ] **Step 3: Re-run the frontend manual scenario**

Manual test checklist:
1. Create or pick a topic owned by the current user.
2. Run API mode with one configured platform and one intentionally misconfigured platform.
3. Confirm task status becomes `partial`.
4. Confirm one successful result row and one failed result row are stored.
5. Open `GET /geoMonitor/collect/task/:id` and `GET /geoMonitor/collect/result/list?taskId=...` through the app flow and confirm persisted data matches the UI.

Expected: per-platform failures are isolated and persisted; task status summarizes the batch correctly.

- [ ] **Step 4: Commit any verification fixes**

```bash
git add server/plugin/geoMonitor web/src/plugin/geoMonitor
git commit -m "fix: polish geo monitor collection flow"
```

---

## Self-Review

- Spec coverage checked: manual run, one task-many results, topic ownership, no analysis, no scheduling, concurrent execution, unified service wrapper around non-unified utils, storage, and minimal UI are all covered.
- Placeholder scan checked: no TBD/TODO markers remain in the plan body.
- Type consistency checked: `RunCollectionRequest`, `CollectOutput`, task/result statuses, and API paths are consistent across tasks.
