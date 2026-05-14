package api

func CollectQwen(apiBase string, apiKey string, prompt string) (*APICollectResult, error) {
	return collectWithOpenAICompatible(apiBase, apiKey, "qwen-turbo", prompt, false)
}

func TestQwen(apiBase string, apiKey string) (bool, error) {
	_, err := CollectQwen(apiBase, apiKey, "ping")
	return err == nil, err
}
