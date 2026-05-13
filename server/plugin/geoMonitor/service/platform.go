package service

import (
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/model"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/model/request"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/utils/api"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/utils/playwright"
)

var Platform = new(platform)

type platform struct{}

// seedPlatforms 7 个真实平台预设（不含 API Key，需管理员手动填入）
var seedPlatforms = []model.Platform{
	{Code: "deepseek", Name: "DeepSeek", Mode: "api", ApiBase: "https://api.deepseek.com", Status: 1, Sort: 1},
	{Code: "qwen", Name: "通义千问", Mode: "api", ApiBase: "https://dashscope.aliyuncs.com/compatible-mode/v1", Status: 1, Sort: 2},
	{Code: "zhipu", Name: "智谱GLM", Mode: "api", ApiBase: "https://open.bigmodel.cn/api/paas/v4", Status: 1, Sort: 3},
	{Code: "doubao", Name: "豆包", Mode: "api", ApiBase: "https://ark.cn-beijing.volces.com/api/v3", Status: 1, Sort: 4},
	{Code: "kimi", Name: "Kimi", Mode: "api", ApiBase: "https://api.moonshot.cn", Status: 1, Sort: 5},
	{Code: "wenxin", Name: "文心一言", Mode: "api", ApiBase: "https://qianfan.baidubce.com/v2", Status: 1, Sort: 6},
	{Code: "yuanbao", Name: "元宝", Mode: "api", ApiBase: "https://hunyuan.tencentcloudapi.com", Status: 1, Sort: 7},
}

// InitSeedData 安装时写入 7 个平台基础信息，已存在则跳过
func (s *platform) InitSeedData() error {
	for _, p := range seedPlatforms {
		var count int64
		if err := global.GVA_DB.Model(&model.Platform{}).Where("code = ? AND mode = ?", p.Code, p.Mode).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err := global.GVA_DB.Create(&p).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *platform) GetPlatformList(info request.PlatformSearch) (list []model.Platform, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)

	db := global.GVA_DB.Model(&model.Platform{})
	if info.Code != "" {
		db = db.Where("code LIKE ?", "%"+info.Code+"%")
	}
	if info.Name != "" {
		db = db.Where("name LIKE ?", "%"+info.Name+"%")
	}
	if info.Mode != "" {
		db = db.Where("mode = ?", info.Mode)
	}
	if info.Status != nil {
		db = db.Where("status = ?", *info.Status)
	}

	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = db.Limit(limit).Offset(offset).Order("sort asc, id desc").Find(&list).Error
	return list, total, err
}

func (s *platform) GetPlatform(id uint) (info model.Platform, err error) {
	err = global.GVA_DB.Where("id = ?", id).First(&info).Error
	return
}

func (s *platform) CreatePlatform(info *model.Platform) error {
	return global.GVA_DB.Create(info).Error
}

func (s *platform) UpdatePlatform(info *model.Platform) error {
	return global.GVA_DB.Model(&model.Platform{}).Where("id = ?", info.ID).Updates(info).Error
}

func (s *platform) DeletePlatform(id uint) error {
	return global.GVA_DB.Delete(&model.Platform{}, id).Error
}

// TestConnectivity 根据平台配置的模式进行连通性测试
func (s *platform) TestConnectivity(id uint) (bool, error) {
	p, err := s.GetPlatform(id)
	if err != nil {
		return false, fmt.Errorf("平台不存在: %w", err)
	}
	if p.Mode == "playwright" {
		return s.testPlaywright(p)
	}
	if p.ApiKey == "" {
		return false, fmt.Errorf("请先配置 API Key")
	}
	return s.doTest(p)
}

func (s *platform) doTest(p model.Platform) (bool, error) {
	switch p.Code {
	case "deepseek":
		return api.TestDeepSeek(p.ApiBase, p.ApiKey)
	case "qwen":
		return api.TestQwen(p.ApiBase, p.ApiKey)
	case "zhipu":
		return api.TestZhipu(p.ApiBase, p.ApiKey)
	case "doubao":
		return api.TestDoubao(p.ApiBase, p.ApiKey)
	case "kimi":
		return api.TestKimi(p.ApiBase, p.ApiKey)
	case "wenxin":
		return api.TestWenxin(p.ApiBase, p.ApiKey)
	case "yuanbao":
		return api.TestYuanbao(p.ApiBase, p.ApiKey)
	default:
		return false, fmt.Errorf("不支持的平台: %s", p.Code)
	}
}

// testPlaywright Playwright 模式连通性测试：使用 playwright-go 真实浏览器访问
func (s *platform) testPlaywright(p model.Platform) (bool, error) {
	if p.ApiBase == "" {
		return false, fmt.Errorf("请先配置网页地址")
	}
	switch p.Code {
	case "deepseek":
		return playwright.TestDeepSeek(p.ApiBase)
	case "qwen":
		return playwright.TestQwen(p.ApiBase)
	case "zhipu":
		return playwright.TestZhipu(p.ApiBase)
	case "doubao":
		return playwright.TestDoubao(p.ApiBase)
	case "kimi":
		return playwright.TestKimi(p.ApiBase)
	case "wenxin":
		return playwright.TestWenxin(p.ApiBase)
	case "yuanbao":
		return playwright.TestYuanbao(p.ApiBase)
	default:
		return false, fmt.Errorf("不支持的平台: %s", p.Code)
	}
}

// PlatformTestResult 连通性测试结果
type PlatformTestResult struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Code    string `json:"code"`
	Ok      bool   `json:"ok"`
	Status  string `json:"status"` // connected / failed / unconfigured
	Message string `json:"message"`
}

// TestAllConnectivity 一键测试所有平台连通性
func (s *platform) TestAllConnectivity() ([]PlatformTestResult, error) {
	var platforms []model.Platform
	if err := global.GVA_DB.Order("sort asc, id asc").Find(&platforms).Error; err != nil {
		return nil, err
	}

	results := make([]PlatformTestResult, 0, len(platforms))
	for _, p := range platforms {
		if p.Mode == "playwright" {
			if p.ApiBase == "" {
				results = append(results, PlatformTestResult{
					ID: p.ID, Name: p.Name, Code: p.Code,
					Ok: false, Status: "unconfigured", Message: "未配置网页地址",
				})
				continue
			}
			ok, err := s.testPlaywright(p)
			if ok {
				results = append(results, PlatformTestResult{
					ID: p.ID, Name: p.Name, Code: p.Code,
					Ok: true, Status: "connected", Message: "网页可达",
				})
			} else {
				results = append(results, PlatformTestResult{
					ID: p.ID, Name: p.Name, Code: p.Code,
					Ok: false, Status: "failed", Message: err.Error(),
				})
			}
			continue
		}
		if p.ApiKey == "" {
			results = append(results, PlatformTestResult{
				ID: p.ID, Name: p.Name, Code: p.Code,
				Ok: false, Status: "unconfigured", Message: "未配置 API Key",
			})
			continue
		}
		ok, err := s.doTest(p)
		if ok {
			results = append(results, PlatformTestResult{
				ID: p.ID, Name: p.Name, Code: p.Code,
				Ok: true, Status: "connected", Message: "连接成功",
			})
		} else {
			results = append(results, PlatformTestResult{
				ID: p.ID, Name: p.Name, Code: p.Code,
				Ok: false, Status: "failed", Message: err.Error(),
			})
		}
	}
	return results, nil
}
