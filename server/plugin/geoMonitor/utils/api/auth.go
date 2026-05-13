package api

import (
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// isAuthError 判断是否为鉴权类错误（API Key 无效）
func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	authKeywords := []string{"401", "403", "unauthorized", "authentication", "invalid api key", "incorrect api key", "auth", "access denied", "permission denied", "invalid token", "token expired"}
	for _, kw := range authKeywords {
		if strings.Contains(errStr, kw) {
			return true
		}
	}
	if apiErr, ok := err.(*openai.APIError); ok {
		code := apiErr.HTTPStatusCode
		if code == 401 || code == 403 {
			return true
		}
	}
	return false
}
