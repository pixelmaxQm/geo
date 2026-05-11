package utils

import (
	"context"
	"fmt"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// TestDoubaoConnectivity 测试豆包平台连通性
// 火山引擎 ARK API v3 兼容 OpenAI 格式，使用 go-openai SDK
func TestDoubaoConnectivity(apiBase string, apiKey string) (bool, error) {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = strings.TrimRight(apiBase, "/") + "/v1"
	client := openai.NewClientWithConfig(config)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: "deepseek-v3-250324",
		Messages: []openai.ChatCompletionMessage{
			{Role: "user", Content: "ping"},
		},
		MaxTokens: 1,
	})
	if err != nil {
		if isAuthError(err) {
			return false, fmt.Errorf("API Key 无效: %w", err)
		}
		return false, fmt.Errorf("连接失败: %w", err)
	}
	return true, nil
}
