package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func CollectWenxin(apiBase string, apiKey string, prompt string) (*APICollectResult, error) {
	url := strings.TrimRight(apiBase, "/") + "/chat/completions?model=ernie-speed"
	body := map[string]any{
		"messages": []map[string]string{{"role": "user", "content": prompt}},
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("构造请求失败: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求发送失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("API Key 无效 (HTTP %d): %s", resp.StatusCode, string(respBody))
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("采集失败 (HTTP %d): %s", resp.StatusCode, string(respBody))
	}
	return &APICollectResult{Answer: string(respBody), Citations: "[]", RawResponse: string(respBody)}, nil
}

func TestWenxin(apiBase string, apiKey string) (bool, error) {
	_, err := CollectWenxin(apiBase, apiKey, "ping")
	return err == nil, err
}
