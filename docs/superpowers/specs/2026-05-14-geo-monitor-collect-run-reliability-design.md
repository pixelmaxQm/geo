# geoMonitor collect/run 结果完整性与 Playwright 登录态设计

## 背景

`/api/geoMonitor/collect/run` 当前需要补强两个问题：

1. API 模式没有完整记录平台联网搜索/引用返回内容，也没有保持平台返回顺序。
2. Playwright 模式失败时缺少可诊断日志；失败可能由网页登录态缺失导致，需要新增二维码优先的登录授权能力。

本设计按两阶段实施，但在同一轮能力建设中完成：先保证采集结果完整和失败可诊断，再接入二维码登录态。

## 目标

- API 模式保留平台原始响应结构到 `rawResponse`。
- API 模式将有序搜索/引用列表标准化后保存到 `citations`。
- Playwright 模式失败也落库，保存错误、运行日志、失败截图和页面摘要。
- 新增 Playwright 平台登录授权能力，二维码优先，授权成功后保存可复用 `storageState`。
- 后续 Playwright 采集自动加载已授权 session。

## 非目标

- 不把登录流程塞进 `/collect/run`，避免采集接口被授权状态卡住。
- 不伪造搜索结果；平台响应里没有可提取字段时，`citations` 保持空数组。
- 不做复杂账号密码托管，第一版以二维码扫码登录为主。

## 总体架构

沿用现有分层：`Router -> API -> Service -> Model`。

```text
collect/run
  -> CollectorService
    -> api_collector
      -> utils/api
      -> CollectionResult(rawResponse, citations, runLog)
    -> playwright_collector
      -> utils/playwright
      -> CollectionResult(errorMsg, screenshotPath, rawResponse, runLog)

playwright/session/*
  -> PlaywrightSessionService
    -> utils/playwright auth task
    -> PlaywrightAuthSession(statePath, qrImagePath, status)
```

`/geoMonitor/collect/run` 继续只负责执行采集。网页登录授权由独立 session 接口负责，采集时只读取已保存的授权状态。

## 数据模型

### CollectionResult 增强

保留现有字段：

- `Answer`
- `Citations`
- `ScreenshotPath`
- `DurationMs`
- `RawResponse`
- `ErrorMsg`

新增字段：

- `RunLog string json:"runLog" gorm:"type:longtext;comment:运行日志JSON"`

字段语义：

- `rawResponse`：保存平台原始响应结构或 Playwright 页面诊断信息，不截断。
- `citations`：保存标准化后的有序搜索/引用列表。
- `runLog`：保存采集过程阶段日志。

`citations` 标准结构：

```json
[
  {
    "index": 1,
    "title": "标题",
    "url": "https://example.com",
    "snippet": "摘要或命中文本",
    "source": "qwen",
    "raw": {}
  }
]
```

列表顺序以平台返回顺序为准，不做去重重排。

### PlaywrightAuthSession

新增登录授权模型：

- `PlatformID uint`
- `PlatformCode string`
- `Status string`：`pending / waiting_scan / authorized / failed / expired`
- `LoginURL string`
- `QrImagePath string`
- `ScreenshotPath string`
- `StatePath string`
- `ErrorMsg string`
- `ExpiresAt *time.Time`
- `LastUsedAt *time.Time`
- `CreatedBy uint`
- `CreatedByName string`

`statePath` 指向后端保存的 Playwright `storageState` 文件，采集时按平台加载。

## 接口设计

### 采集接口

现有接口保持不变：

```text
POST /geoMonitor/collect/run
```

返回结果增强：每条 `CollectionResult` 包含 `citations/rawResponse/runLog/screenshotPath/errorMsg`。

### 登录授权接口

新增接口：

```text
POST /geoMonitor/playwright/session/start
参数: { platformId }
返回: { sessionId, status, qrImagePath, screenshotPath }

GET /geoMonitor/playwright/session/:id
返回: { id, platformId, status, qrImagePath, screenshotPath, errorMsg }

POST /geoMonitor/playwright/session/:id/refresh
作用: 重新截图或刷新二维码状态

DELETE /geoMonitor/playwright/session/:id
作用: 删除授权记录和登录态文件
```

权限沿用当前插件私有路由和用户权限隔离。非超级管理员只能查看和操作自己创建的授权任务。

## API 模式流程

1. 创建请求日志：平台、模型、接口地址、开始时间。
2. 发起平台 API 请求。
3. 保存完整原始响应到 `rawResponse`。
4. 提取主回答到 `answer`。
5. 从搜索结果、引用、工具调用、metadata 等候选字段中提取引用源。
6. 按平台返回顺序写入 `citations`。
7. 写入阶段日志到 `runLog`。

平台 SDK 如果无法保留搜索结果字段，则对该平台改为原生 HTTP 请求，确保原始响应不被 SDK 丢字段。平台不返回搜索/引用字段时，`citations` 为 `[]`。

API 失败时仍创建 `CollectionResult`：

- `status=failed`
- `errorMsg=具体错误`
- `rawResponse=HTTP body 或错误响应结构`
- `runLog=请求阶段、耗时、错误摘要`

## Playwright 模式流程

统一包装每个平台采集步骤：

1. 创建 context，加载平台已授权 `storageState`。
2. 打开平台网页。
3. 检测是否需要登录。
4. 等待输入框 selector。
5. 输入 prompt。
6. 提交。
7. 等待回答区域稳定。
8. 提取回答和页面文本。
9. 截图。

每一步追加日志项：

```json
{"step":"goto","status":"success","message":"页面加载完成","durationMs":1234}
```

任一步失败时：

- 截图保存到 `screenshotPath`。
- 保存当前 URL、页面标题、body 摘要到 `rawResponse`。
- 保存完整 `runLog`。
- `errorMsg` 写明失败阶段。
- 如果检测到登录页、二维码或手机号输入框，错误信息明确写“需要登录授权”。

## 二维码登录授权流程

1. 前端在 Playwright 平台行展示“登录授权”。
2. 用户点击后调用 `POST /geoMonitor/playwright/session/start`。
3. 后端打开平台登录页。
4. 优先识别二维码区域并截图；识别不到则截整页。
5. 前端弹窗展示二维码或登录页截图，并轮询 `GET /session/:id`。
6. 用户扫码登录。
7. 后端检测登录成功后保存 `storageState`。
8. session 状态更新为 `authorized`。
9. 后续 Playwright 采集加载该平台最新 authorized session。
10. session 失效时，采集结果失败并提示重新授权。

## 前端设计

### 采集页

- 结果表新增“引用源/搜索结果”查看入口。
- 结果表新增“运行日志”查看入口。
- 失败结果显示 `errorMsg`，可展开查看 `runLog` 和截图。
- API 结果可展开查看完整有序 `citations`。

### 登录授权弹窗

- 入口：平台配置页或采集页的 Playwright 平台旁。
- 弹窗内容：二维码截图、当前状态、错误提示、刷新按钮。
- 状态轮询：`waiting_scan` 时持续轮询，`authorized/failed/expired` 停止。

## 错误处理

- 平台不支持：明确返回“不支持的平台”。
- API 鉴权失败：`errorMsg` 标记 API Key 无效。
- API 搜索字段缺失：不报错，`citations=[]`，保留完整 `rawResponse`。
- Playwright 未登录：保存失败结果，提示需要登录授权。
- Playwright selector 变化：保存失败步骤、截图和页面摘要。
- 授权二维码识别失败：返回整页截图，让用户仍可扫码或判断页面状态。
- 登录态过期：标记 session expired，并要求重新授权。

## 测试与验收

### 后端测试

- API 成功：`rawResponse` 不为空，`citations` 保持平台返回顺序。
- API 失败：仍创建失败结果，`errorMsg/runLog/rawResponse` 可诊断。
- Playwright 失败：仍创建失败结果，`runLog` 包含失败步骤，必要时有截图路径。
- 部分平台失败：任务状态为 `partial`。
- session 创建：状态可从 `pending -> waiting_scan -> authorized/failed` 流转。
- 权限隔离：非超级管理员不能访问他人的 session。

### 手动验收

1. 选择 API 平台执行 `/collect/run`。
2. 查看结果：`rawResponse` 有完整原始响应，`citations` 有有序引用/搜索列表。
3. 选择未登录 Playwright 平台执行采集。
4. 查看失败结果：有错误日志、失败截图、需要登录提示。
5. 点击登录授权，扫码完成。
6. 再次执行 Playwright 采集，确认使用登录态。

## 实施顺序

1. 增强 `CollectionResult`，补充 `runLog`。
2. 建立采集日志结构和 API 引用提取结构。
3. 改造 API collector：保留原始响应，提取有序 citations。
4. 改造 Playwright collector：失败落库、失败截图、页面诊断、登录检测。
5. 新增 PlaywrightAuthSession 模型、迁移、服务和路由。
6. 新增二维码登录任务和 session 保存能力。
7. 前端采集结果增加 citations/runLog 查看。
8. 前端增加 Playwright 登录授权弹窗。
9. 补充测试并手动验收。
