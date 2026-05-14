package api

func CollectKimi(apiBase string, apiKey string, prompt string) (*APICollectResult, error) {
	return collectWithOpenAICompatible(apiBase, apiKey, "moonshot-v1-8k", prompt, true)
}

func TestKimi(apiBase string, apiKey string) (bool, error) {
	_, err := CollectKimi(apiBase, apiKey, "ping")
	return err == nil, err
}
