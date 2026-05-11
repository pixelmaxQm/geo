# 持续监控分析工具 (geo-monitor)

## 概述

geo-monitor 是一个持续监控与分析插件，用于对国内主流 AI 对话平台进行持续抓取、对比和趋势分析。

## 监控目标平台

| 平台 | 所属公司 | 备注 |
|---|---|---|
| 豆包 (Doubao) | 字节跳动 | |
| 千问 (Qwen) | 阿里 | 通义千问 |
| 智谱 (Zhipu) | 智谱AI | ChatGLM |
| 文心 (Wenxin) | 百度 | 文心一言 |
| 元宝 (Yuanbao) | 腾讯 | |
| Kimi | 月之暗面 (Moonshot) | |
| DeepSeek | 深度求索 | |

## 两种采集模式

### 模式一：API 模式（低级模式）

- 调用各平台的官方 API，开启 search/web 搜索能力
- 适合高频轮询、低成本监控
- 结果反映的是 "API 视角" 下的回答

### 模式二：Playwright 模式（高级模式）

- 使用 playwright-go 驱动真实浏览器，模拟用户访问各平台 Web 端
- 真实输入、真实等待、真实截图
- 结果反映的是 "真实用户视角" 下的回答
- 适合抽样审核、质量验证

## 核心能力

1. **定时轮询**: 按可配置的时间间隔自动对目标平台发起提问
2. **问题模板管理**: 预定义一组标准问题，按模板生成每次提问内容
3. **结果存储**: 将每次请求的 prompt、回答、来源引用、耗时、截图等存入数据库
4. **变更对比**: 对同一问题的多次回答进行 diff 对比，检测内容变化
5. **趋势分析**: 统计各平台的引用偏好、回答风格变化、响应时间趋势
6. **告警通知**: 关键指标变化时触发告警（如某平台不可用、回答中出现特定关键词等）

## 后端插件结构

遵循项目插件规范，后端位于 `server/plugin/geo-monitor/`：

```
server/plugin/geo-monitor/
├── plugin.go                    # 插件注册入口
├── config/
│   └── config.go               # 插件配置结构
├── initialize/
│   ├── api.go                  # 注册 API 到全局路由
│   ├── gorm.go                 # 注册模型自动迁移
│   ├── menu.go                 # 注册菜单
│   ├── router.go               # 注册插件路由
│   └── viper.go                # 注册 viper 配置
├── model/
│   ├── platform.go             # 平台定义模型
│   ├── question_template.go    # 问题模板模型
│   ├── collection_task.go      # 采集任务记录
│   ├── collection_result.go    # 采集结果（回答）模型
│   └── request/                # 请求模型
├── api/
│   ├── platform.go             # 平台管理接口
│   ├── template.go             # 问题模板接口
│   ├── task.go                 # 采集任务接口
│   └── result.go               # 结果查询/分析接口
├── service/
│   ├── platform.go             # 平台管理业务
│   ├── template.go             # 模板管理业务
│   ├── collector.go            # 采集调度核心
│   ├── api_collector.go        # API 模式采集实现
│   ├── playwright_collector.go # Playwright 模式采集实现
│   └── analyzer.go             # 分析对比业务
└── router/
    └── enter.go                # 路由注册
```

## 数据模型设计要点

### collection_task (采集任务)
- 平台ID
- 问题模板ID
- 采集模式 (api / playwright)
- 状态 (pending / running / done / failed)
- 开始时间
- 结束时间
- 耗时

### collection_result (采集结果)
- 关联任务ID
- 实际发送的 prompt
- AI 回答内容
- 引用来源 (JSON)
- 截图路径 (playwright 模式)
- 响应耗时
- 创建时间

## 核心接口设计

### 平台管理
- `GET /geo-monitor/platforms` - 平台列表
- `GET /geo-monitor/platforms/:id` - 平台详情
- `PUT /geo-monitor/platforms/:id` - 更新平台配置（含 API Key）

### 问题模板
- `GET /geo-monitor/templates` - 模板列表（分页）
- `POST /geo-monitor/templates` - 创建模板
- `PUT /geo-monitor/templates/:id` - 更新模板
- `DELETE /geo-monitor/templates/:id` - 删除模板

### 采集任务
- `POST /geo-monitor/tasks` - 创建采集任务
- `GET /geo-monitor/tasks` - 任务列表（分页+筛选）
- `POST /geo-monitor/tasks/run` - 手动触发一次采集
- `POST /geo-monitor/tasks/cron` - 设置/修改定时采集规则
- `GET /geo-monitor/tasks/:id/results` - 查看任务结果

### 分析接口
- `GET /geo-monitor/analysis/compare` - 同问题多平台对比
- `GET /geo-monitor/analysis/trend` - 单平台/单问题趋势
- `GET /geo-monitor/analysis/overview` - 总览仪表盘数据

## 采集器抽象接口

```go
// Collector 采集器接口
type Collector interface {
    // Collect 执行一次采集，返回结果
    Collect(ctx context.Context, platform Platform, prompt string) (*CollectResult, error)
    // Name 采集器名称
    Name() string
}

// ApiCollector API 模式采集器
type ApiCollector struct { ... }
// PlaywrightCollector Playwright 模式采集器
type PlaywrightCollector struct { ... }
```

每个平台需要各自的 **平台适配器**（PlatformAdapter），负责：
- API 模式: 构造请求、调官方 SDK/HTTP、解析响应
- Playwright 模式: 封装页面定位器、输入策略、等待策略、提取回答

## 开发阶段规划

### 阶段一：基础框架（约 3-4 天）

1. 插件脚手架搭建（plugin.go、config、initialize、router）
2. 数据模型设计与 GORM 迁移
3. 平台表初始化（7 个平台的基础数据 + 种子）
4. 平台管理 CRUD 接口
5. 问题模板 CRUD 接口

### 阶段二：API 模式采集（约 4-5 天）

1. 定义 Collector 抽象接口和通用采集流程
2. 实现各平台 API 适配器（按优先级逐个接入）：
   - DeepSeek: 官方 API 有 search 能力
   - 千问 (Qwen): DashScope API + 搜索插件
   - 智谱 (Zhipu): GLM API + web_search
   - 豆包 (Doubao): 火山引擎豆包 API
   - Kimi: Moonshot API
   - 文心 (Wenxin): 千帆 API
   - 元宝 (Yuanbao): 腾讯混元 API
3. 定时任务调度（基于 cron 或内置 timer）
4. 采集结果存储与基础查询

### 阶段三：Playwright 模式采集（约 4-5 天）

1. playwright-go 环境搭建与浏览器池管理
2. 实现通用浏览器采集流程（输入 → 等待回答 → 提取 → 截图）
3. 实现各平台 Web 端适配器（页面元素定位、登录态管理）
4. 与 API 模式的对比验证

### 阶段四：分析与可视化（约 3-4 天）

1. 前端页面搭建（插件前端入口）
2. 对比分析接口（同问题多平台 diff）
3. 趋势分析接口（时序数据聚合）
4. 仪表盘概览页面
5. 结果详情页（含截图查看、引用来源展示）

### 阶段五：告警与运营（约 2-3 天）

1. 告警规则配置
2. 告警触发逻辑
3. 通知渠道（邮件/站内信）

## 关键技术决策

1. **Go 的 Playwright 库**: 使用 `github.com/playwright-community/playwright-go`，是官方 Playwright 的 Go 绑定
2. **定时任务**: 优先复用项目已有的 `initialize.Timer()` 机制，或引入 `robfig/cron`
3. **各平台 API 接入建议先统一抽象接口（Collector），再逐平台实现，降低耦合**
4. **Playwright 模式需要独立的浏览器进程管理**，考虑浏览器实例池复用，避免每次采集都启动新浏览器
5. **截图存储**: 本地文件 + 数据库存路径，后续可扩展至 OSS

## 依赖与约束

- Go 1.21+
- playwright-go 需要安装 Playwright 依赖浏览器
- 各平台的 API Key 通过配置管理（支持加密存储或环境变量注入）
- 插件自包含，不侵入主系统核心逻辑
- 遵循项目统一响应结构 `{ code, data, msg }`
- 遵循项目统一分页结构 `{ page, pageSize, total, list }`

## 前端入口

- 插件前端位于 `web/src/plugin/geo-monitor/`
- 主页面包含：仪表盘、平台配置、问题模板、采集任务、结果对比、趋势分析
- 遵循前端插件结构规范，参考 `aiDoc/examples/plugin/`
