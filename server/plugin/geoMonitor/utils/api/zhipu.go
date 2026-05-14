package api

func CollectZhipu(apiBase string, apiKey string, prompt string) (*APICollectResult, error) {
	return collectWithOpenAICompatible(apiBase, apiKey, "glm-4-flash", prompt, false)
}

func TestZhipu(apiBase string, apiKey string) (bool, error) {
	_, err := CollectZhipu(apiBase, apiKey, "ping")
	return err == nil, err
}
