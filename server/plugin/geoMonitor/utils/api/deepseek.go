package api

func CollectDeepSeek(apiBase string, apiKey string, prompt string) (*APICollectResult, error) {
	return collectWithOpenAICompatible(apiBase, apiKey, "deepseek-chat", prompt, true)
}

func TestDeepSeek(apiBase string, apiKey string) (bool, error) {
	_, err := CollectDeepSeek(apiBase, apiKey, "ping")
	return err == nil, err
}
