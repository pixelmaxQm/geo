package service

import (
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/model"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/model/request"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/utils"
)

var Platform = new(platform)

type platform struct{}

// seedPlatforms 7 个真实平台预设（不含 API Key，需管理员手动填入）
var seedPlatforms = []model.Platform{
	{Code: "deepseek", Name: "DeepSeek", ApiBase: "https://api.deepseek.com", Status: 1, Sort: 1},
	{Code: "qwen", Name: "通义千问", ApiBase: "https://dashscope.aliyuncs.com/compatible-mode/v1", Status: 1, Sort: 2},
	{Code: "zhipu", Name: "智谱GLM", ApiBase: "https://open.bigmodel.cn/api/paas/v4", Status: 1, Sort: 3},
	{Code: "doubao", Name: "豆包", ApiBase: "https://ark.cn-beijing.volces.com/api/v3", Status: 1, Sort: 4},
	{Code: "kimi", Name: "Kimi", ApiBase: "https://api.moonshot.cn", Status: 1, Sort: 5},
	{Code: "wenxin", Name: "文心一言", ApiBase: "https://qianfan.baidubce.com/v2", Status: 1, Sort: 6},
	{Code: "yuanbao", Name: "元宝", ApiBase: "https://hunyuan.tencentcloudapi.com", Status: 1, Sort: 7},
}

// InitSeedData 安装时写入 7 个平台基础信息，已存在则跳过
func (s *platform) InitSeedData() error {
	for _, p := range seedPlatforms {
		var count int64
		if err := global.GVA_DB.Model(&model.Platform{}).Where("code = ?", p.Code).Count(&count).Error; err != nil {
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

// TestConnectivity 根据平台 code 分发到对应的 SDK 工具进行连通性测试
func (s *platform) TestConnectivity(id uint) (bool, error) {
	p, err := s.GetPlatform(id)
	if err != nil {
		return false, fmt.Errorf("平台不存在: %w", err)
	}
	if p.ApiKey == "" {
		return false, fmt.Errorf("请先配置 API Key")
	}
	return s.doTest(p)
}

func (s *platform) doTest(p model.Platform) (bool, error) {
	switch p.Code {
	case "deepseek":
		return utils.TestDeepSeekConnectivity(p.ApiBase, p.ApiKey)
	case "qwen":
		return utils.TestQwenConnectivity(p.ApiBase, p.ApiKey)
	case "zhipu":
		return utils.TestZhipuConnectivity(p.ApiBase, p.ApiKey)
	case "doubao":
		return utils.TestDoubaoConnectivity(p.ApiBase, p.ApiKey)
	case "kimi":
		return utils.TestKimiConnectivity(p.ApiBase, p.ApiKey)
	case "wenxin":
		return utils.TestWenxinConnectivity(p.ApiBase, p.ApiKey)
	case "yuanbao":
		return utils.TestYuanbaoConnectivity(p.ApiBase, p.ApiKey)
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
	Status  string `json:"status"`  // connected / failed / unconfigured
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
