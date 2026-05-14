# geo-monitor 插件模块依赖分析（按功能板块）

## 当前骨架

GVA 生成器已创建 13 个文件，插件注册、路由分组、配置注入的基础框架已就绪。

```
server/plugin/geoMonitor/
├── plugin.go                   # 插件注册入口
├── plugin/plugin.go            # 插件全局变量
├── config/config.go            # 配置结构体（空）
├── api/enter.go                # API 层入口（空）
├── service/enter.go            # Service 层入口（空）
├── router/enter.go             # Router 层入口（空）
├── initialize/
│   ├── router.go               # public/private 路由分组已搭好
│   ├── gorm.go                 # AutoMigrate 调用（空）
│   ├── viper.go                # Viper 配置注入（key: geoMonitor）
│   ├── api.go                  # API 权限注册（空）
│   ├── menu.go                 # 菜单注册（空）
│   └── dictionary.go           # 字典注册（空）
└── gen/gen.go                  # GORM Gen 生成器（空模板）
```

---

## 功能板块一：平台管理 ✅ 已完成

> 2026-05-11 完成基础 CRUD。2026-05-13 完成 api/playwright 双模式改造。

### 对应后台页面

**平台配置页** — 列表展示所有平台，支持新增/编辑/删除。每条记录定义接入一个 AI 平台的渠道：API 模式（官方 API + Key）或 Playwright 模式（网页地址 + 真实浏览器抓取）。

### 页面功能

- 平台列表（表格）：名称、code、采集模式（API/Playwright 标签）、地址、状态、连通状态、排序、创建时间
- 搜索筛选：关键词、**采集模式**、启用/停用
- 新增/编辑表单：
  - 平台渠道下拉（7 个预设平台，编辑时锁定）
  - **采集模式 Radio**（API 模式 / Playwright 模式），切换时自动替换对应默认地址
  - 地址字段（动态标签：API 模式下"API 地址"，Playwright 下"网页地址"）
  - API Key 字段（**仅 API 模式下显示**）
  - 状态开关、排序、备注
- 删除（软删除）
- **单平台连通性测试**：API 模式发 chat ping 验证 Key，Playwright 模式用 playwright-go 真实浏览器访问
- **一键测试所有**：批量返回每个平台的连通状态

### 需要的后端文件

```
model/platform.go            # 数据模型（含 Mode 字段 + 复合唯一索引）
model/request/platform.go    # 请求参数结构体（含 Mode 筛选）
service/platform.go          # 业务逻辑（含 doTest/testPlaywright 双模式分发）
api/platform.go              # HTTP 接口
initialize/gorm.go           # 旧索引迁移 + AutoMigrate
```

### 数据模型

```go
type Platform struct {
    global.GVA_MODEL
    Code    string `gorm:"uniqueIndex:idx_code_mode"` // 平台唯一标识
    Name    string // 显示名称
    Mode    string `gorm:"default:api;uniqueIndex:idx_code_mode"` // api / playwright
    ApiBase string // API 端点地址 或 网页 URL（视 Mode 而定）
    ApiKey  string // API Key（API 模式使用，Playwright 模式隐藏）
    Status  int    `gorm:"default:1"` // 1启用 0停用
    Sort    int    `gorm:"default:0"`
    Remark  string
}
```

**关键设计决策：**
- `(code, mode)` 复合唯一索引 — 同一平台可同时存在 api 和 playwright 两条记录，视为两个独立渠道
- `ApiBase` 字段复用：api 模式下存 API 端点，playwright 模式下存网页 URL；前端表单标签随 mode 动态切换
- 旧版单列唯一索引 `idx_gva_geo_monitor_platforms_code` 在 initialize/gorm.go 中自动检测并删除

### 后端接口

```
GET    /geoMonitor/platform/list       — 平台列表（分页 + code/name/mode/status 筛选）
GET    /geoMonitor/platform/:id        — 平台详情
POST   /geoMonitor/platform            — 新增平台
PUT    /geoMonitor/platform/:id        — 编辑平台
DELETE /geoMonitor/platform/:id        — 删除平台
POST   /geoMonitor/platform/test/:id   — 连通性测试（自动按 mode 分发）
POST   /geoMonitor/platform/testAll    — 一键测试所有平台
```

### 连通性测试分发逻辑

```
TestConnectivity(id)
  └→ p.Mode == "api"       → doTest()    → api.Test{Platform}(apiBase, apiKey)
  └→ p.Mode == "playwright" → testPlaywright() → playwright.Test{Platform}(webUrl)
                                                    ↓
                                         NewPage() → pw.Run() → Chromium headless
                                                    → page.Goto() → WaitForSelector()
```

### 依赖的 GVA 模块

| 模块 | 用途 |
|---|---|
| `global.GVA_DB` | GORM CRUD + index migration |
| `global.GVA_MODEL` | 嵌入基础字段 |
| `utils/request` | 分页请求结构 |
| `utils/response` | 统一响应 `{ code, data, msg }` |

### 依赖的第三方库

| 库 | 用途 | 模式 |
|---|---|---|
| `github.com/sashabaranov/go-openai` | OpenAI 兼容 SDK（DeepSeek/千问/智谱/豆包/Kimi） | API |
| `github.com/playwright-community/playwright-go` | 驱动真实 Chromium 浏览器 | Playwright |

### 工具包结构

```
utils/
├── api/                         # API SDK 模式工具（package api）
│   ├── auth.go                  # isAuthError 鉴权错误检测
│   ├── deepseek.go              # TestDeepSeek(apiBase, apiKey)
│   ├── qwen.go                  # TestQwen(apiBase, apiKey)
│   ├── zhipu.go                 # TestZhipu(apiBase, apiKey)
│   ├── doubao.go                # TestDoubao(apiBase, apiKey)
│   ├── kimi.go                  # TestKimi(apiBase, apiKey)
│   ├── wenxin.go                # TestWenxin(apiBase, apiKey)
│   └── yuanbao.go               # TestYuanbao(apiBase, apiKey)
└── playwright/                  # Playwright 浏览器模式工具（package playwright）
    ├── playwright.go            # 浏览器生命周期：Install/Launch/NewPage/Close
    ├── deepseek.go              # TestDeepSeek(webUrl) → 等待 #chat-input
    ├── qwen.go                  # TestQwen(webUrl) → 等待 .editor
    ├── zhipu.go                 # TestZhipu(webUrl) → 等待 .input-area
    ├── doubao.go                # TestDoubao(webUrl) → 等待 .chat-input-box
    ├── kimi.go                  # TestKimi(webUrl) → 等待 .chat-input
    ├── wenxin.go                # TestWenxin(webUrl) → 等待 .yiyan-input
    └── yuanbao.go               # TestYuanbao(webUrl) → 等待 .input-box
```

> 每个平台有独立的 CSS selector 等待逻辑。导入别名 `pw "github.com/playwright-community/playwright-go"` 避免与包名冲突。

### 种子数据

插件安装时预置 7 个平台的 **API 模式** 基础信息（不含 API Key，由管理员后续填入）：

| code | name | mode | api_base |
|---|---|---|---|
| deepseek | DeepSeek | api | `https://api.deepseek.com` |
| qwen | 通义千问 | api | `https://dashscope.aliyuncs.com/compatible-mode/v1` |
| zhipu | 智谱GLM | api | `https://open.bigmodel.cn/api/paas/v4` |
| doubao | 豆包 | api | `https://ark.cn-beijing.volces.com/api/v3` |
| kimi | Kimi | api | `https://api.moonshot.cn` |
| wenxin | 文心一言 | api | `https://qianfan.baidubce.com/v2` |
| yuanbao | 元宝 | api | `https://hunyuan.tencentcloudapi.com` |

种子检查条件已适配复合唯一键：`WHERE code = ? AND mode = ?`。

### 前端预设（platformConfig.js）

每个预设平台同时提供 API 和 Playwright 默认地址，切换模式时自动填充：

| code | apiBase (API 端点) | webBase (Playwright 网页) |
|---|---|---|
| deepseek | `https://api.deepseek.com` | `https://chat.deepseek.com` |
| qwen | `https://dashscope.aliyuncs.com/compatible-mode/v1` | `https://www.qianwen.com/` |
| zhipu | `https://open.bigmodel.cn/api/paas/v4` | `https://chatglm.cn` |
| doubao | `https://ark.cn-beijing.volces.com/api/v3` | `https://www.doubao.com/chat/` |
| kimi | `https://api.moonshot.cn` | `https://www.kimi.com/` |
| wenxin | `https://qianfan.baidubce.com/v2` | `https://yiyan.baidu.com` |
| yuanbao | `https://hunyuan.tencentcloudapi.com` | `https://yuanbao.tencent.com` |

---

## 功能板块二：问题采集

### 对应后台页面

**问题采集页** — 核心操作页。输入问题（或选择问题模板），选择平台和采集模式，执行采集，查看结果与分析。

### 页面功能

- 问题输入区：文本输入框 + 问题模板下拉选择
- 平台选择：多选（至少选一个）
- 模式选择：API 模式 / Playwright 高级模式
- 执行按钮：开始采集
- 结果展示区：
  - 各平台回答内容（Tab 切换或并排展示）
  - 引用排名：各平台引用了哪些来源 URL，按出现频次排序
  - 来源分析：引用来源的域名分布、类型分布
  - 关键词统计：用户指定的关键词在各平台回答中出现的次数
- 截图查看（Playwright 模式下）
- 导出按钮

### 需要的后端文件

```
model/question_template.go       # 问题模板数据模型
model/collection_task.go         # 采集任务模型
model/collection_result.go       # 采集结果模型
model/request/template.go        # 模板请求参数
model/request/task.go            # 任务请求参数
model/request/result.go          # 结果查询参数
service/template.go              # 模板管理业务
service/collector.go             # 采集调度核心（接口定义 + 分发）
service/api_collector.go         # API 模式采集器
service/playwright_collector.go  # Playwright 模式采集器
service/analyzer.go              # 引用分析、关键词统计
api/template.go                  # 模板接口
api/task.go                      # 采集任务接口
api/result.go                    # 结果查询+分析接口
```

### 数据模型

```go
// QuestionTemplate 问题模板
type QuestionTemplate struct {
    global.GVA_MODEL
    Name    string // 模板名称
    Content string // 问题内容（可含变量占位符）
    Tags    string // 标签（JSON 数组）
    Remark  string // 备注
}

// CollectionTask 采集任务
type CollectionTask struct {
    global.GVA_MODEL
    PlatformID  uint      // 关联平台ID
    TemplateID  uint      // 关联模板ID（0 表示手动输入）
    Prompt      string    // 实际发送的问题文本
    Mode        string    // 采集模式: "api" / "playwright"
    Status      string    // 状态: pending / running / done / failed
    CronExpr    string    // 定时表达式（为空表示手动单次）
    StartedAt   time.Time // 开始时间
    FinishedAt  time.Time // 结束时间
    ErrorMsg    string    // 错误信息
}

// CollectionResult 采集结果
type CollectionResult struct {
    global.GVA_MODEL
    TaskID         uint   // 关联任务ID
    PlatformID     uint   // 冗余平台ID（便于查询）
    Prompt         string // 实际发送的问题
    Answer         string // AI 回答内容（完整文本）
    Citations      string // 引用来源（JSON: [{"url":"...", "title":"..."}]）
    ScreenshotPath string // 截图路径（Playwright 模式）
    DurationMs     int    // 响应耗时（毫秒）
    RawResponse    string // 原始响应（调试用）
}
```

### 后端接口

**模板管理：**
```
GET    /geoMonitor/template/list     — 模板列表（分页）
POST   /geoMonitor/template          — 创建模板
PUT    /geoMonitor/template/:id      — 更新模板
DELETE /geoMonitor/template/:id      — 删除模板
```

**采集执行：**
```
POST   /geoMonitor/collect/run       — 手动执行一次采集
       参数: { platform_ids: [], template_id?, prompt?, mode: "api"|"playwright" }
       返回: { task_id, results: [...] }
```

**结果查询与分析：**
```
GET    /geoMonitor/result/list        — 结果列表（分页+筛选: 平台/时间范围/模式）
GET    /geoMonitor/result/:id         — 结果详情（含回答全文、引用JSON、截图）
GET    /geoMonitor/result/analysis    — 单次采集的分析数据
       参数: { task_id }
       返回: {
         citations_ranking: [{url, title, count}],  // 引用排名
         source_domains: [{domain, count}],           // 来源域名分布
         keywords: [{word, platform, count}]          // 关键词出现次数
       }
```

### 采集器架构（对接板块一的双模式设计）

板块一已完成 `utils/api/` 和 `utils/playwright/` 两个工具包。板块二的 Collector 实现直接复用此结构：

```
Collector 接口
  ├── ApiCollector         → 复用 utils/api/    中各平台的函数签名 (apiBase, apiKey)
  │     ├── DeepSeekAdapter   ← api.TestDeepSeek 模式 → 扩展为完整采集
  │     ├── QwenAdapter       ← api.TestQwen
  │     ├── ZhipuAdapter      ← api.TestZhipu
  │     ├── DoubaoAdapter     ← api.TestDoubao
  │     ├── KimiAdapter       ← api.TestKimi
  │     ├── WenxinAdapter     ← api.TestWenxin
  │     └── YuanbaoAdapter    ← api.TestYuanbao
  └── PlaywrightCollector  → 复用 utils/playwright/ 的各平台访问模式
        ├── DeepSeekPageAdapter   ← playwright.TestDeepSeek 模式 → 扩展为完整采集+截图
        ├── QwenPageAdapter       ← playwright.TestQwen
        ├── ZhipuPageAdapter      ← playwright.TestZhipu
        ├── DoubaoPageAdapter     ← playwright.TestDoubao
        ├── KimiPageAdapter       ← playwright.TestKimi
        ├── WenxinPageAdapter     ← playwright.TestWenxin
        └── YuanbaoPageAdapter    ← playwright.TestYuanbao
```

**对接方式：**
- API Collector 从 DB 查 `Mode == "api"` 的平台，取 `ApiBase` + `ApiKey`，调用 `utils/api/` 的对应函数，只需扩写请求体（把 "ping" 替换为真实 prompt + search 参数）
- Playwright Collector 从 DB 查 `Mode == "playwright"` 的平台，取 `ApiBase`(网页 URL)，复用 `utils/playwright/playwright.go` 的浏览器生命周期（Launch/NewPage/Close），每个平台的 selector 已在连通性测试中验证，直接在此基础上添加输入→等待回答→提取→截图逻辑
- 两种 Collector 通过 `Platform.Mode` 字段区分，无需额外路由或工厂模式

### 依赖的 GVA 模块

| 模块 | 用途 |
|---|---|
| `global.GVA_DB` | CRUD |
| `global.GVA_MODEL` | 基础模型 |
| `global.GVA_LOG` | 采集过程日志 |
| `global.GVA_CONFIG` | 获取系统 RouterPrefix |
| `utils/request` | 分页 |
| `utils/response` | 统一响应 |

### 依赖的第三方库

| 库 | 用途 | 模式 |
|---|---|---|
| `github.com/playwright-community/playwright-go` | 驱动真实浏览器 | 高级模式 |
| `github.com/google/uuid` | 任务唯一标识 | 通用 |
| `golang.org/x/sync/errgroup` | 多平台并发采集、错误聚合 | 通用 |

### 依赖的标准库

| 包 | 用途 |
|---|---|
| `net/http` | 调各平台 API |
| `encoding/json` | 序列化/反序列化 |
| `context` | 超时控制、取消传播 |
| `strings` | 关键词统计、文本处理 |
| `regexp` | 引用 URL 提取 |

---

## 功能板块三：采集历史

### 对应后台页面

**采集历史页** — 以列表+筛选的方式展示所有历史采集记录。点击某条记录进入详情，查看完整结果和分析数据。

### 页面功能

- 筛选条件：平台、模式、状态、时间范围
- 列表展示：序号、问题摘要（截断）、平台、模式、状态标签、耗时、时间
- 点击行展开/跳转详情
- 详情页：
  - 基本信息：问题全文、平台、模式、耗时、状态
  - 回答全文（支持 Markdown 渲染）
  - 引用列表（可点击跳转）
  - 截图（Playwright 模式）
  - 分析数据（引用排名、来源域名、关键词统计）
- 删除（软删除）
- 批量删除

### 需要的后端文件

```
api/result.go                  # 结果列表+详情查询接口
service/analyzer.go            # 分析数据计算（由板块二共用）
model/collection_result.go     # 结果模型（由板块二共用）
model/collection_task.go       # 任务模型（由板块二共用）
```

> 此板块主要复用板块二的 model 和 service，新增内容集中在 api 层的列表/详情查询。

### 后端接口

```
GET    /geoMonitor/history/list       — 采集历史列表
       参数: { page, pageSize, platform_id?, mode?, status?, start_time?, end_time? }
       返回: { list: [{id, prompt_abstract, platform_name, mode, status, duration_ms, created_at}] }

GET    /geoMonitor/history/:id        — 采集历史详情
       返回: { task_info, result: { answer, citations, screenshot, analysis } }

DELETE /geoMonitor/history/:id        — 删除单条记录
DELETE /geoMonitor/history/batch      — 批量删除
       参数: { ids: [] }
```

### 依赖的 GVA 模块

| 模块 | 用途 |
|---|---|
| `global.GVA_DB` | 列表查询（分页+多条件筛选） |
| `global.GVA_MODEL` | 软删除（DeletedAt） |
| `utils/request` | 分页 |
| `utils/response` | 统一响应 |

### 依赖的第三方库

无额外依赖，复用板块二的模型和 service。

---

## 功能板块四：定时采集

### 对应后台页面

**定时任务管理页** — 配置周期性自动采集任务。可创建/编辑/启停定时规则。

### 页面功能

- 定时任务列表：任务名、平台、问题模板、cron 表达式、下次执行时间、状态
- 新增/编辑定时任务：选择平台、选择模板、选择模式、设置 cron 表达式
- 快捷 cron 选择：每 30 分钟 / 每 1 小时 / 每 6 小时 / 每天
- 手动触发一次
- 启用/停用
- 最近执行记录（关联到采集历史）

### 需要的后端文件

```
model/scheduled_task.go          # 定时任务模型
model/request/scheduled_task.go  # 定时任务请求参数
service/scheduler.go             # 定时任务管理（cron 增删改查）
api/scheduled_task.go            # 定时任务接口
```

### 数据模型

```go
// ScheduledTask 定时采集任务
type ScheduledTask struct {
    global.GVA_MODEL
    Name       string // 任务名称
    PlatformIDs string // 平台ID列表（JSON: [1,2,3]）
    TemplateID uint   // 问题模板ID
    Mode       string // 采集模式
    CronExpr   string // cron 表达式
    Status     int    // 1启用 0停用
    LastRunAt  time.Time // 上次执行时间
    NextRunAt  time.Time // 下次执行时间
}
```

### 后端接口

```
GET    /geoMonitor/schedule/list       — 定时任务列表
POST   /geoMonitor/schedule            — 创建定时任务
PUT    /geoMonitor/schedule/:id        — 编辑定时任务
DELETE /geoMonitor/schedule/:id        — 删除定时任务
POST   /geoMonitor/schedule/:id/toggle — 启用/停用
POST   /geoMonitor/schedule/:id/run    — 手动触发一次
```

### 依赖的 GVA 模块

| 模块 | 用途 |
|---|---|
| `global.GVA_DB` | CRUD |
| `global.GVA_TIMER` | 基于 robfig/cron 的任务调度器，`AddTaskByFunc` 注册、`RemoveTask` 移除 |
| `global.GVA_LOG` | 调度日志 |
| `utils/request` / `utils/response` | 分页 + 统一响应 |

### 依赖的第三方库

| 库 | 用途 |
|---|---|
| `github.com/robfig/cron/v3` | cron 表达式解析与调度（项目已引入，复用在 `global.GVA_TIMER` 中） |

---

## 功能板块五：跨平台对比分析

### 对应后台页面

**对比分析页** — 选择一个历史采集任务（涉及多平台的），查看该问题在各平台的表现对比。

### 页面功能

- 选择对比任务（下拉，筛选多平台任务）
- 对比视图：
  - 回答并排展示（最多 4 列，超出滚动）
  - 回答长度统计柱状图
  - 响应耗时对比柱状图
  - 引用来源重叠度（韦恩图或矩阵）
  - 引用域名分布对比
  - 关键词出现位置高亮对比

### 后端接口

```
GET    /geoMonitor/compare/:task_id    — 获取某次多平台采集的对比数据
       返回: {
         platforms: [{id, name}],
         answers: [{platform_id, answer, duration_ms, word_count}],
         citations_overlap: { platform_A: [urls], platform_B: [urls], overlap: [urls] },
         domain_distribution: [{platform_id, domains: [{domain, count}]}],
         keyword_hits: [{keyword, platform_hits: [{platform_id, count}]}]
       }
```

### 依赖

复用板块二的 model 和 service/analyzer.go，仅新增 api 层接口。依赖标准库 `sort`、`strings` 做文本处理，无额外第三方库。

---

## 功能板块六：系统配置

### 对应后台页面

**系统配置页** — 插件级全局参数配置。

### 页面功能

- 表单配置项：
  - 采集超时时间（秒）
  - 并发采集上限
  - 失败重试次数
  - Playwright 模式开关
  - 浏览器路径
  - 截图存储路径

### 需要的后端文件

```
config/config.go              # 配置结构体定义
api/config.go                 # 配置读写接口
```

### 后端接口

```
GET    /geoMonitor/config            — 获取当前配置
PUT    /geoMonitor/config            — 更新配置
```

### 依赖的 GVA 模块

| 模块 | 用途 |
|---|---|
| `global.GVA_VP` | Viper 配置读写（`UnmarshalKey("geoMonitor")` 读取，`Set` + `WriteConfig` 回写） |
| `config/config.go` | 已有插件配置结构体 |

### config.yaml 中需新增段

```yaml
geoMonitor:
  timeout: 60
  parallel-limit: 3
  retry-count: 2
  playwright:
    headless: true
    browser-path: ""
    pool-size: 2
  screenshot-dir: "uploads/geo-monitor/"
```

### 配置结构体

```go
type Config struct {
    Timeout       int              `mapstructure:"timeout" yaml:"timeout"`
    ParallelLimit int              `mapstructure:"parallel-limit" yaml:"parallel-limit"`
    RetryCount    int              `mapstructure:"retry-count" yaml:"retry-count"`
    Playwright    PlaywrightConfig `mapstructure:"playwright" yaml:"playwright"`
    ScreenshotDir string           `mapstructure:"screenshot-dir" yaml:"screenshot-dir"`
}

type PlaywrightConfig struct {
    Headless   bool   `mapstructure:"headless" yaml:"headless"`
    BrowserPath string `mapstructure:"browser-path" yaml:"browser-path"`
    PoolSize   int    `mapstructure:"pool-size" yaml:"pool-size"`
}
```

---

## 下拉/筛选字典

以下系统字典需要在 `initialize/dictionary.go` 中注册，供前端下拉框使用：

| 字典类型 | 字典值 |
|---|---|
| `geo_monitor_mode` | `api` — API 模式 / `playwright` — 高级模式 |
| `geo_monitor_task_status` | `pending` — 等待中 / `running` — 采集中 / `done` — 已完成 / `failed` — 失败 |
| `geo_monitor_platform_status` | `1` — 启用 / `0` — 停用 |

---

## 菜单结构

```
AI 监控分析 (父级, icon: monitor)
├── 平台配置          → /geo-monitor/platform
├── 问题采集          → /geo-monitor/collect
├── 采集历史          → /geo-monitor/history
├── 对比分析          → /geo-monitor/compare
├── 定时任务          → /geo-monitor/schedule
└── 系统配置          → /geo-monitor/config
```

---

## 开发顺序（按功能板块）

| 顺序 | 板块 | 预估 | 状态 |
|---|---|---|---|
| **1** | 平台管理 | 0.5 天 | ✅ 已完成（2026-05-13） |
| **2** | 系统配置 | 0.3 天 | 待开发 |
| **3** | 问题采集（核心） | 3-4 天 | 待开发（基础工具包已就绪） |
| **4** | 采集历史 | 0.5 天 | 待开发 |
| **5** | 对比分析 | 1 天 | 待开发 |
| **6** | 定时采集 | 1 天 | 待开发 |

---

## 全部依赖汇总

### GVA 内置模块
`global.GVA_DB` · `global.GVA_LOG` · `global.GVA_VP` · `global.GVA_TIMER` · `global.GVA_CONFIG` · `global.GVA_MODEL` · `middleware.JWTAuth` · `middleware.CasbinHandler` · `plugin-tool/utils` (RegisterApis / RegisterMenus / RegisterDictionaries) · `utils/request` · `utils/response` · `model/system` (SysApi / SysBaseMenu / SysDictionary)

### 新增第三方库
`github.com/playwright-community/playwright-go` (v0.5700.1) · `github.com/sashabaranov/go-openai` · `github.com/google/uuid` · `golang.org/x/sync/errgroup`

### 复用已有第三方库
`github.com/robfig/cron/v3` · `github.com/gin-gonic/gin` · `gorm.io/gorm` · `go.uber.org/zap` · `github.com/spf13/viper` · `github.com/pkg/errors`

### 标准库
`net/http` · `encoding/json` · `context` · `strings` · `regexp` · `sort` · `time` · `sync`

---

## 骨架文件与功能板块的对应关系

| 骨架文件 | 需改动 | 服务于哪些板块 |
|---|---|---|
| `plugin.go` | 是 — 启用 Viper/Api/Menu/Dict 注册调用 | 全部 |
| `plugin/plugin.go` | 否 | 全部（全局 Config 引用） |
| `config/config.go` | 是 — 填入配置结构体 | 板块六 |
| `api/enter.go` | 是 — 挂载各板块 api 实例 | 全部 |
| `service/enter.go` | 是 — 挂载各板块 service 实例 | 全部 |
| `router/enter.go` | 是 — 挂载全部路由 | 全部 |
| `initialize/router.go` | 是 — 挂载路由组 | 全部 |
| `initialize/gorm.go` | 是 — 填入 AutoMigrate 表 | 板块一/二/三/四 |
| `initialize/viper.go` | 否（已正确） | 板块六 |
| `initialize/api.go` | 是 — 填入 API 注册列表 | 全部 |
| `initialize/menu.go` | 是 — 填入菜单注册 | 全部 |
| `initialize/dictionary.go` | 是 — 填入字典注册 | 全部 |
| `gen/gen.go` | 是 — 填入模型列表 | 板块一/二/四 |

> 骨架 13 个文件全部需要修改，但骨架结构正确，按板块逐步填充即可。
