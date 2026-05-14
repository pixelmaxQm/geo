# 持续监控分析工具 (geo-monitor)

## 基本信息

- 提出日期：2026-05-11
- 当前状态：`active`
- 需求类型：新增插件
- 优先级：高

## 用户原始意图摘要

构建一个针对国内主流 AI 平台（豆包、千问、智谱、文心、元宝、Kimi、DeepSeek）的持续监控分析工具，支持 API 模式和 Playwright 真实浏览器采集两种模式，用于对比分析各平台回答质量、引用来源和响应趋势。

## 影响范围

- 后端：`server/plugin/geo-monitor/` (新增)
- 前端：`web/src/plugin/geo-monitor/` (新增)
- 文档：`aiDoc/modules/plugin-geo-monitor.md`
- 插件 / 模块：`geo-monitor`

## 涉及对象

- 模块：7 个国内 AI 平台
- 功能板块（6 个）：平台管理、问题采集、采集历史、定时采集、跨平台对比分析、系统配置
- 菜单：AI 监控分析（父级）→ 平台配置 / 问题采集 / 采集历史 / 对比分析 / 定时任务 / 系统配置
- 配置：各平台 API Key、定时任务 cron 表达式、系统参数

## 已确认约束

- 分 API 模式（低级，开 search）和 Playwright 模式（高级，真实浏览器）
- 使用 playwright-go（`github.com/playwright-community/playwright-go`）
- 遵循项目插件规范，自包含在 `server/plugin/geoMonitor/`
- 遵循统一响应 `{ code, data, msg }` 和分页 `{ page, pageSize, total, list }`
- 先做国内平台，后续可扩展国际平台
- 按功能板块组织开发，非按技术分层

## 当前进展

- 已完成模块规格文档（`aiDoc/modules/plugin-geo-monitor.md`）
- 已完成功能板块依赖分析（`aiDoc/modules/plugin-geo-monitor-analysis.md`），按 6 个功能板块组织
- GVA 生成器已创建插件骨架（13 个文件位于 `server/plugin/geoMonitor/`）
- 板块一已完成：平台管理（2026-05-11）
- 板块一增强：平台配置增加 api/playwright 双模式（2026-05-13）
  - Platform 模型新增 `Mode` 字段，`(code, mode)` 复合唯一索引
  - 同一个平台可同时存在 API 和 Playwright 两个渠道
  - utils 拆分为 `utils/api/`（8 个文件）和 `utils/playwright/`（8 个文件）两个子包
  - 前端表单/搜索/表格/连通性测试全部适配双模式，切换模式自动替换默认地址
  - Playwright 连通性测试使用 playwright-go 真实浏览器访问，每个平台独立 selector 等待
  - 分析文档 `aiDoc/modules/plugin-geo-monitor-analysis.md` 已同步更新设计决策
- 板块三采集增强需求已确认（2026-05-14）
  - `/geoMonitor/collect/run` 需要完整记录 API 联网搜索/引用结果：平台原始响应保存到 `rawResponse`，标准化有序引用源/搜索列表保存到 `citations`
  - Playwright 采集失败也必须落库并保存失败日志、页面诊断和失败截图
  - 新增二维码优先的 Playwright 登录授权能力，扫码成功后保存 `storageState`，后续采集自动加载登录态
  - 设计文档：`docs/superpowers/specs/2026-05-14-geo-monitor-collect-run-reliability-design.md`

## 后续待办（按功能板块顺序）

- [x] 板块一：平台管理 — model + CRUD + 种子数据 + 连通性测试 + 7 个 SDK 工具文件（已完成 2026-05-11）
- [ ] 板块二：系统配置 — config 结构体 + 读写接口（0.3 天）
- [ ] 板块三：问题采集 — Collector 接口 + API 适配器 + Playwright 适配器（3-4 天）
- [ ] 板块四：采集历史 — 列表查询 + 详情 + 删除（0.5 天）
- [ ] 板块五：对比分析 — 多平台对比接口（1 天）
- [ ] 板块六：定时采集 — cron 任务管理 + 启停（1 天）
