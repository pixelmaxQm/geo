package api

import (
	"context"
	"fmt"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

func TestQwen(apiBase string, apiKey string) (bool, error) {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = strings.TrimRight(apiBase, "/")
	client := openai.NewClientWithConfig(config)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: "qwen-turbo",
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
