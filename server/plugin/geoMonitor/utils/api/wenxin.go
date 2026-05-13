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

func TestWenxin(apiBase string, apiKey string) (bool, error) {
	url := strings.TrimRight(apiBase, "/") + "/chat/completions?model=ernie-speed"

	body := map[string]any{
		"messages": []map[string]string{
			{"role": "user", "content": "ping"},
		},
		"max_tokens": 1,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return false, fmt.Errorf("构造请求失败: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return false, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("请求发送失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusUnauthorized, http.StatusForbidden:
		return false, fmt.Errorf("API Key 无效 (HTTP %d): %s", resp.StatusCode, string(respBody))
	default:
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return true, nil
		}
		return false, fmt.Errorf("连接失败 (HTTP %d): %s", resp.StatusCode, string(respBody))
	}
}
