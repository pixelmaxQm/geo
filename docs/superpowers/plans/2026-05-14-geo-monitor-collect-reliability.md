# Geo Monitor Collect Reliability Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `/geoMonitor/collect/run` preserve complete API responses and ordered references, persist Playwright failure diagnostics, and add QR-first Playwright login sessions.

**Architecture:** Keep collection and login authorization separate. Collection writes richer `CollectionResult` records (`citations`, `rawResponse`, `runLog`, screenshots), while the new Playwright session service manages browser login tasks and saved storage state that collection can load later.

**Tech Stack:** Go, Gin, GORM, playwright-go, gin-vue-admin response/request patterns, Vue 3, Element Plus.

---

## File Structure

### Backend create
- `server/plugin/geoMonitor/model/playwright_auth_session.go` — persisted login authorization session state.
- `server/plugin/geoMonitor/model/request/playwright_session.go` — request structs for session endpoints.
- `server/plugin/geoMonitor/service/collector_log.go` — reusable run-log and citation JSON helpers.
- `server/plugin/geoMonitor/service/playwright_session.go` — session CRUD/start/refresh/delete service.
- `server/plugin/geoMonitor/api/playwright_session.go` — HTTP handlers for session endpoints.
- `server/plugin/geoMonitor/router/playwright_session.go` — route registration.
- `server/plugin/geoMonitor/service/collector_log_test.go` — tests for run-log and ordered citation helpers.
- `server/plugin/geoMonitor/service/playwright_session_test.go` — tests for session ownership and status creation.

### Backend modify
- `server/plugin/geoMonitor/model/collection_result.go` — add `RunLog`.
- `server/plugin/geoMonitor/initialize/gorm.go` — migrate `PlaywrightAuthSession` and new `RunLog` column.
- `server/plugin/geoMonitor/initialize/api.go` — register session endpoint permissions.
- `server/plugin/geoMonitor/api/enter.go` — expose session API.
- `server/plugin/geoMonitor/router/enter.go` — expose session router.
- `server/plugin/geoMonitor/initialize/router.go` — mount session routes.
- `server/plugin/geoMonitor/service/enter.go` — expose session service.
- `server/plugin/geoMonitor/service/collector.go` — persist `RunLog` for every result and keep failed records diagnostic.
- `server/plugin/geoMonitor/service/api_collector.go` — produce run logs and ordered citations.
- `server/plugin/geoMonitor/service/playwright_collector.go` — load session state, produce failure diagnostics.
- `server/plugin/geoMonitor/utils/api/collect.go` — preserve raw response and parse ordered references from raw JSON.
- `server/plugin/geoMonitor/utils/playwright/playwright.go` — allow `NewPage` with optional storage state.
- `server/plugin/geoMonitor/utils/playwright/collect.go` — collect run logs, screenshots, login detection, diagnostics.

### Frontend create
- `web/src/plugin/geoMonitor/api/playwrightSession.js` — session endpoint wrappers.

### Frontend modify
- `web/src/plugin/geoMonitor/view/collect.vue` — show citations and run logs in dialogs.
- `web/src/plugin/geoMonitor/view/platform.vue` — add Playwright login authorization button/dialog.

---

## Task 1: Add diagnostic fields and helper types

**Files:**
- Modify: `server/plugin/geoMonitor/model/collection_result.go`
- Create: `server/plugin/geoMonitor/service/collector_log.go`
- Create: `server/plugin/geoMonitor/service/collector_log_test.go`

- [ ] **Step 1: Add failing tests for ordered citations and logs**

Create `server/plugin/geoMonitor/service/collector_log_test.go`:

```go
package service

import (
    "encoding/json"
    "testing"
)

func TestBuildOrderedCitationsJSONPreservesOrder(t *testing.T) {
    citations := []CitationItem{
        {Title: "first", URL: "https://a.example", Snippet: "A", Source: "qwen", Raw: map[string]any{"rank": float64(2)}},
        {Title: "second", URL: "https://b.example", Snippet: "B", Source: "qwen", Raw: map[string]any{"rank": float64(1)}},
    }

    got := BuildOrderedCitationsJSON(citations)

    var decoded []CitationItem
    if err := json.Unmarshal([]byte(got), &decoded); err != nil {
        t.Fatalf("unmarshal citations: %v", err)
    }
    if len(decoded) != 2 {
        t.Fatalf("len(decoded) = %d, want 2", len(decoded))
    }
    if decoded[0].Index != 1 || decoded[0].URL != "https://a.example" {
        t.Fatalf("first citation = %+v", decoded[0])
    }
    if decoded[1].Index != 2 || decoded[1].URL != "https://b.example" {
        t.Fatalf("second citation = %+v", decoded[1])
    }
}

func TestRunLogJSONRecordsSteps(t *testing.T) {
    log := NewRunLog()
    log.Add("goto", "success", "页面加载完成", 15)
    log.Add("submit", "failed", "提交失败", 9)

    got := log.JSON()

    var decoded []RunLogItem
    if err := json.Unmarshal([]byte(got), &decoded); err != nil {
        t.Fatalf("unmarshal run log: %v", err)
    }
    if decoded[0].Step != "goto" || decoded[1].Status != "failed" {
        t.Fatalf("decoded = %+v", decoded)
    }
}
```

- [ ] **Step 2: Run tests to verify failure**

Run: `go test ./plugin/geoMonitor/service -run 'TestBuildOrderedCitationsJSONPreservesOrder|TestRunLogJSONRecordsSteps' -v`
Expected: FAIL because helper types/functions do not exist.

- [ ] **Step 3: Implement helpers and model field**

Create `server/plugin/geoMonitor/service/collector_log.go`:

```go
package service

import (
    "encoding/json"
    "time"
)

type CitationItem struct {
    Index   int            `json:"index"`
    Title   string         `json:"title"`
    URL     string         `json:"url"`
    Snippet string         `json:"snippet"`
    Source  string         `json:"source"`
    Raw     map[string]any `json:"raw,omitempty"`
}

type RunLogItem struct {
    Step       string `json:"step"`
    Status     string `json:"status"`
    Message    string `json:"message"`
    DurationMs int64  `json:"durationMs"`
    Time       string `json:"time"`
}

type RunLog struct {
    items []RunLogItem
}

func NewRunLog() *RunLog {
    return &RunLog{items: make([]RunLogItem, 0)}
}

func (l *RunLog) Add(step string, status string, message string, durationMs int64) {
    l.items = append(l.items, RunLogItem{Step: step, Status: status, Message: message, DurationMs: durationMs, Time: time.Now().Format(time.RFC3339)})
}

func (l *RunLog) JSON() string {
    if l == nil || len(l.items) == 0 {
        return "[]"
    }
    data, err := json.Marshal(l.items)
    if err != nil {
        return "[]"
    }
    return string(data)
}

func BuildOrderedCitationsJSON(items []CitationItem) string {
    if len(items) == 0 {
        return "[]"
    }
    normalized := make([]CitationItem, len(items))
    copy(normalized, items)
    for i := range normalized {
        normalized[i].Index = i + 1
        if normalized[i].Raw == nil {
            normalized[i].Raw = map[string]any{}
        }
    }
    data, err := json.Marshal(normalized)
    if err != nil {
        return "[]"
    }
    return string(data)
}
```

Modify `server/plugin/geoMonitor/model/collection_result.go` and add:

```go
RunLog string `json:"runLog" gorm:"type:longtext;comment:运行日志JSON"`
```

- [ ] **Step 4: Run tests**

Run: `go test ./plugin/geoMonitor/service -run 'TestBuildOrderedCitationsJSONPreservesOrder|TestRunLogJSONRecordsSteps' -v`
Expected: PASS.

---

## Task 2: Persist run logs from `/collect/run`

**Files:**
- Modify: `server/plugin/geoMonitor/service/collector_types.go`
- Modify: `server/plugin/geoMonitor/service/collector.go`
- Modify: `server/plugin/geoMonitor/service/collector_run_test.go`

- [ ] **Step 1: Add output field and persistence test**

Extend `CollectOutput` with:

```go
RunLog string `json:"runLog"`
```

Add a test in `collector_run_test.go` that migrates `CollectionResult`, inserts a platform with unsupported code, runs collection, and asserts the failed result has non-empty `ErrorMsg` and `RunLog`.

- [ ] **Step 2: Run failing test**

Run: `go test ./plugin/geoMonitor/service -run TestCollectorRunPersistsFailureDiagnostics -v`
Expected: FAIL because `RunLog` is not persisted.

- [ ] **Step 3: Persist run log**

In `collector.go`, set `RunLog: output.RunLog` when building `model.CollectionResult`. If `collectErr != nil` and `output.RunLog == ""`, set `RunLog` to a JSON log with `collect` failed.

- [ ] **Step 4: Run tests**

Run: `go test ./plugin/geoMonitor/service -run 'TestCollectorRunPersistsFailureDiagnostics|TestSummarizeTaskStatus' -v`
Expected: PASS.

---

## Task 3: Preserve API raw responses and ordered citations

**Files:**
- Modify: `server/plugin/geoMonitor/utils/api/collect.go`
- Modify: `server/plugin/geoMonitor/service/api_collector.go`
- Create/modify tests: `server/plugin/geoMonitor/service/collector_log_test.go`

- [ ] **Step 1: Add citation extraction test**

Add a test with raw JSON containing `search_results` array. Assert extracted citations preserve order and include title/url/snippet/raw.

- [ ] **Step 2: Run failing test**

Run: `go test ./plugin/geoMonitor/service -run TestExtractCitationsFromRawResponsePreservesOrder -v`
Expected: FAIL until extractor exists.

- [ ] **Step 3: Implement extraction**

Add `ExtractCitationsFromRawResponse(raw string, source string) []CitationItem` in `collector_log.go`. It should parse common arrays named `search_results`, `searchResults`, `citations`, `references`, or `web_search`. For each object, map `title`, `url/link`, and `snippet/content/summary`, preserve order, and store the original object in `Raw`.

- [ ] **Step 4: Wire API collector**

In `api_collector.go`, after each API call, compute:

```go
citations := ExtractCitationsFromRawResponse(result.RawResponse, platform.Code)
runLog := NewRunLog()
runLog.Add("api_request", "success", "API 采集完成", int64(time.Since(started).Milliseconds()))
```

Return `Citations: BuildOrderedCitationsJSON(citations)` and `RunLog: runLog.JSON()`.

- [ ] **Step 5: Run tests**

Run: `go test ./plugin/geoMonitor/service -run 'TestExtractCitationsFromRawResponsePreservesOrder|TestBuildOrderedCitationsJSONPreservesOrder' -v`
Expected: PASS.

---

## Task 4: Add Playwright failure diagnostics and optional storage state

**Files:**
- Modify: `server/plugin/geoMonitor/utils/playwright/playwright.go`
- Modify: `server/plugin/geoMonitor/utils/playwright/collect.go`
- Modify: `server/plugin/geoMonitor/service/playwright_collector.go`

- [ ] **Step 1: Add browser context option**

Add `NewPageWithStorageState(storageStatePath string) (pw.Page, func(), error)` and let `NewPage()` call it with an empty string.

- [ ] **Step 2: Add diagnostic result fields**

Extend `PageCollectResult` with `RunLog string` and `ErrorRawResponse string`.

- [ ] **Step 3: Capture failure diagnostics**

In `collectWithPage`, maintain `RunLog`, take screenshots on error when `screenshotPath` is set, and return errors with a partial `PageCollectResult` through a new custom error type or service wrapper. Include current URL, title, and body text excerpt in raw diagnostics.

- [ ] **Step 4: Wire service result**

In `playwright_collector.go`, load latest authorized session state path when available, call the storage-state page collector, and return `CollectOutput` with `RunLog` and `RawResponse` even when failures occur.

- [ ] **Step 5: Run backend tests**

Run: `go test ./plugin/geoMonitor/service ./plugin/geoMonitor/utils/playwright -v`
Expected: PASS.

---

## Task 5: Add Playwright login session backend

**Files:**
- Create: `server/plugin/geoMonitor/model/playwright_auth_session.go`
- Create: `server/plugin/geoMonitor/model/request/playwright_session.go`
- Create: `server/plugin/geoMonitor/service/playwright_session.go`
- Create: `server/plugin/geoMonitor/api/playwright_session.go`
- Create: `server/plugin/geoMonitor/router/playwright_session.go`
- Modify: `server/plugin/geoMonitor/service/enter.go`
- Modify: `server/plugin/geoMonitor/api/enter.go`
- Modify: `server/plugin/geoMonitor/router/enter.go`
- Modify: `server/plugin/geoMonitor/initialize/router.go`
- Modify: `server/plugin/geoMonitor/initialize/gorm.go`
- Modify: `server/plugin/geoMonitor/initialize/api.go`

- [ ] **Step 1: Add model and request structs**

Define `PlaywrightAuthSession` with platform/user/status/path fields from the spec. Define `StartPlaywrightSessionRequest{PlatformID uint}`.

- [ ] **Step 2: Add service tests**

Test that creating a session for a Playwright platform stores `waiting_scan`, and non-owner lookup is rejected for non-super-admin.

- [ ] **Step 3: Implement service**

Add `Start`, `Get`, `Refresh`, and `Delete`. First version starts browser page, captures QR or full-page screenshot path, and records status. Authorized detection can be implemented by checking login selectors disappear and then saving `storageState`.

- [ ] **Step 4: Add API and router**

Register routes:

```text
POST geoMonitor/playwright/session/start
GET geoMonitor/playwright/session/:id
POST geoMonitor/playwright/session/:id/refresh
DELETE geoMonitor/playwright/session/:id
```

- [ ] **Step 5: Register migration and permissions**

Add `new(model.PlaywrightAuthSession)` to AutoMigrate and four `SysApi` entries.

- [ ] **Step 6: Run tests**

Run: `go test ./plugin/geoMonitor/service -run 'TestPlaywrightSession' -v`
Expected: PASS.

---

## Task 6: Add frontend result viewers

**Files:**
- Modify: `web/src/plugin/geoMonitor/view/collect.vue`

- [ ] **Step 1: Add dialogs**

Add two dialogs: one for parsed `citations`, one for `runLog`. Parse JSON safely in the component.

- [ ] **Step 2: Add table actions**

Add buttons in each result row: “引用源” and “日志”. Disable or show empty state when JSON arrays are empty.

- [ ] **Step 3: Run frontend check**

Run: `npm run lint` from `web` if configured; otherwise run `npm run build` from `web`.
Expected: command succeeds.

---

## Task 7: Add frontend login authorization dialog

**Files:**
- Create: `web/src/plugin/geoMonitor/api/playwrightSession.js`
- Modify: `web/src/plugin/geoMonitor/view/platform.vue`

- [ ] **Step 1: Add API wrapper**

Create functions: `startPlaywrightSession`, `getPlaywrightSession`, `refreshPlaywrightSession`, `deletePlaywrightSession`.

- [ ] **Step 2: Add platform table action**

For rows with `mode === 'playwright'`, show “登录授权”. On click, call start endpoint and open dialog.

- [ ] **Step 3: Add dialog and polling**

Display QR/screenshot image, status, error message, and refresh button. Poll every 2 seconds while status is `pending` or `waiting_scan`; stop on close or terminal status.

- [ ] **Step 4: Run frontend build**

Run: `npm run build` from `web`.
Expected: build succeeds.

---

## Task 8: Final validation

**Files:**
- Modify docs only if implementation changes API contracts from this plan.

- [ ] **Step 1: Run backend tests**

Run: `go test ./plugin/geoMonitor/...`
Expected: PASS.

- [ ] **Step 2: Run frontend build**

Run: `npm run build` from `web`.
Expected: PASS.

- [ ] **Step 3: Manual verification**

Start server and web app if available. Verify API mode result has `rawResponse`, `citations`, and `runLog`; verify Playwright failure creates failed result with `errorMsg`, `runLog`, and screenshot; verify login dialog can create and poll a session.

---

## Self-Review

- Spec coverage: API raw response, ordered citations, Playwright diagnostics, QR-first session endpoints, frontend viewers, and tests are all mapped to tasks.
- Placeholder scan: No TBD/TODO placeholders remain. Task 5 keeps authorization detection scoped to login selector disappearance and storageState save.
- Type consistency: `RunLog`, `CitationItem`, `PlaywrightAuthSession`, and session endpoints are named consistently across tasks.
