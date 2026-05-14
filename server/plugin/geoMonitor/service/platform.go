package service

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

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
	if err != nil {
		return nil, 0, err
	}
	if err = Service.PlaywrightSession.AttachCurrentAuthorizedSession(list); err != nil {
		return nil, 0, err
	}
	return list, total, err
}

func (s *platform) GetPlatform(id uint) (info model.Platform, err error) {
	err = global.GVA_DB.Where("id = ?", id).First(&info).Error
	if err != nil {
		return
	}
	err = Service.PlaywrightSession.AttachCurrentAuthorizedSessionToPlatform(&info)
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
func (s *platform) TestConnectivity(id uint) (PlatformTestResult, error) {
	p, err := s.GetPlatform(id)
	if err != nil {
		return PlatformTestResult{}, fmt.Errorf("平台不存在: %w", err)
	}
	result := PlatformTestResult{ID: p.ID, Name: p.Name, Code: p.Code}
	if p.Mode == "playwright" {
		return s.testPlaywrightWithSnapshot(p), nil
	}
	if p.ApiKey == "" {
		result.Ok = false
		result.Status = "unconfigured"
		result.Message = "请先配置 API Key"
		return result, nil
	}
	ok, err := s.doTest(p)
	result.Ok = ok
	if ok {
		result.Status = "connected"
		result.Message = "连接成功"
		return result, nil
	}
	result.Status = "failed"
	if err != nil {
		result.Message = err.Error()
		return result, nil
	}
	result.Message = "连接失败"
	return result, nil
}

func (s *platform) testPlaywrightWithSnapshot(p model.Platform) PlatformTestResult {
	result := PlatformTestResult{ID: p.ID, Name: p.Name, Code: p.Code}
	if p.ApiBase == "" {
		result.Ok = false
		result.Status = "unconfigured"
		result.Message = "未配置网页地址"
		return result
	}
	screenshotPath := filepath.ToSlash(filepath.Join("uploads", "file", fmt.Sprintf("gm-test-platform-%d-%d.png", p.ID, time.Now().UnixNano())))
	collectResult, err := playwright.CollectByCode(p.Code, p.ApiBase, "ping", screenshotPath, "")
	result.ScreenshotPath = screenshotPath
	if collectResult != nil {
		result.ScreenshotPath = collectResult.ScreenshotPath
		result.RawResponse = collectResult.RawResponse
	}
	if err != nil {
		result.Ok = false
		result.Status = "failed"
		result.Message = err.Error()
		return result
	}
	result.Ok = true
	result.Status = "connected"
	result.Message = "网页可达"
	return result
}

func (s *platform) testPlaywright(p model.Platform) (bool, error) {
	result := s.testPlaywrightWithSnapshot(p)
	if result.Ok {
		return true, nil
	}
	if result.Message == "" {
		return false, fmt.Errorf("连接失败")
	}
	return false, errors.New(result.Message)
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

// PlatformTestResult 连通性测试结果
type PlatformTestResult struct {
	ID             uint   `json:"id"`
	Name           string `json:"name"`
	Code           string `json:"code"`
	Ok             bool   `json:"ok"`
	Status         string `json:"status"` // connected / failed / unconfigured
	Message        string `json:"message"`
	ScreenshotPath string `json:"screenshotPath,omitempty"`
	RawResponse    string `json:"rawResponse,omitempty"`
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
			results = append(results, s.testPlaywrightWithSnapshot(p))
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
