package api

func CollectDoubao(apiBase string, apiKey string, prompt string) (*APICollectResult, error) {
	return collectWithOpenAICompatible(apiBase, apiKey, "deepseek-v3-250324", prompt, true)
}

func TestDoubao(apiBase string, apiKey string) (bool, error) {
	_, err := CollectDoubao(apiBase, apiKey, "ping")
	return err == nil, err
}
