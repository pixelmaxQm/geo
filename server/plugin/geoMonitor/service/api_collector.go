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
		runLog := NewRunLog()
		runLog.Add("api_request", "failed", err.Error(), int64(time.Since(started).Milliseconds()))
		output := CollectOutput{Citations: "[]", DurationMs: int(time.Since(started).Milliseconds()), RunLog: runLog.JSON(), ErrorMsg: err.Error()}
		if result != nil {
			output.Answer = result.Answer
			output.RawResponse = result.RawResponse
			output.Citations = BuildOrderedCitationsJSON(ExtractCitationsFromRawResponse(result.RawResponse, platform.Code))
		}
		return output, err
	}

	citations := ExtractCitationsFromRawResponse(result.RawResponse, platform.Code)
	runLog := NewRunLog()
	runLog.Add("api_request", "success", "API 采集完成", int64(time.Since(started).Milliseconds()))

	return CollectOutput{
		Answer:      result.Answer,
		Citations:   BuildOrderedCitationsJSON(citations),
		DurationMs:  int(time.Since(started).Milliseconds()),
		RawResponse: result.RawResponse,
		RunLog:      runLog.JSON(),
	}, nil
}
