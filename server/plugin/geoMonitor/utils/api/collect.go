package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type APICollectResult struct {
	Answer      string
	Citations   string
	RawResponse string
}

func collectWithOpenAICompatible(apiBase string, apiKey string, modelName string, prompt string, appendV1 bool) (*APICollectResult, error) {
	baseURL := strings.TrimRight(apiBase, "/")
	if appendV1 {
		baseURL += "/v1"
	}
	endpoint := baseURL + "/chat/completions"

	body := map[string]any{
		"model": modelName,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"enable_search": true,
		"search":        true,
		"web_search":    true,
		"search_options": map[string]any{
			"enable": true,
		},
	}
	payload, _ := json.Marshal(body)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("创建采集请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("采集失败: %w", err)
	}
	defer resp.Body.Close()
	rawBytes, _ := io.ReadAll(resp.Body)
	raw := string(rawBytes)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return &APICollectResult{Citations: "[]", RawResponse: raw}, fmt.Errorf("API Key 无效: %s", raw)
		}
		return &APICollectResult{Citations: "[]", RawResponse: raw}, fmt.Errorf("采集失败: HTTP %d: %s", resp.StatusCode, raw)
	}

	answer := extractOpenAICompatibleAnswer(rawBytes)
	return &APICollectResult{Answer: answer, Citations: "[]", RawResponse: raw}, nil
}

func extractOpenAICompatibleAnswer(raw []byte) string {
	var data struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &data); err != nil || len(data.Choices) == 0 {
		return ""
	}
	return data.Choices[0].Message.Content
}
