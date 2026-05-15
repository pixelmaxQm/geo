package llm

import (
	"encoding/json"
	"strings"

	apiutils "github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/utils/api"
)

type KeywordAnalysisInput struct {
	Question  string
	Answer    string
	Citations string
	APIBase   string
	APIKey    string
}

func RankQuestionKeywordsWithGLM(input KeywordAnalysisInput) ([]string, error) {
	prompt := buildKeywordAnalysisPrompt(input)
	result, err := apiutils.CollectZhipu(input.APIBase, input.APIKey, prompt)
	if err != nil {
		return nil, err
	}
	return parseKeywordRanking(result.Answer)
}

func buildKeywordAnalysisPrompt(input KeywordAnalysisInput) string {
	return strings.TrimSpace("请从问题中提取关键词，并根据这些关键词在回答中的相关性从高到低排序。只返回 JSON 字符串数组，不要返回其他内容。\\n问题：" + input.Question + "\\n回答：" + input.Answer + "\\n引用：" + input.Citations)
}

func parseKeywordRanking(raw string) ([]string, error) {
	cleaned := strings.TrimSpace(raw)
	result := make([]string, 0)
	if err := json.Unmarshal([]byte(cleaned), &result); err == nil {
		return result, nil
	}
	start := strings.Index(cleaned, "[")
	end := strings.LastIndex(cleaned, "]")
	if start >= 0 && end > start {
		if err := json.Unmarshal([]byte(cleaned[start:end+1]), &result); err == nil {
			return result, nil
		}
	}
	return []string{}, nil
}
