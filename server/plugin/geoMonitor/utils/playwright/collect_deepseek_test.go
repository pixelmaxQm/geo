package playwright

import "testing"

func TestBuildDeepSeekCitationItemsFromCards(t *testing.T) {
	items := buildDeepSeekCitationItems([]deepSeekCitationCard{{
		Title:   "Go Packages",
		URL:     "https://pkg.go.dev/github.com/mingrammer/go-web-framework-stars",
		Icon:    "https://cdn.deepseek.com/site-icons/pkg.go.dev",
		Summary: "gin | 86685 | 8466 | 869 | Gin is a high-performance HTTP web framework written in Go.",
	}})
	if len(items) != 1 {
		t.Fatalf("expected 1 citation item, got %d", len(items))
	}
	if items[0].Icon == "" {
		t.Fatalf("expected icon to be preserved")
	}
	if items[0].Snippet == "" {
		t.Fatalf("expected summary to become snippet")
	}
}

func TestNeedsDedicatedFlow(t *testing.T) {
	if !needsDedicatedFlow("deepseek") {
		t.Fatalf("expected deepseek to use dedicated flow")
	}
	if needsDedicatedFlow("qwen") {
		t.Fatalf("expected qwen not to use dedicated flow yet")
	}
}

func TestDeepSeekCitationTriggerSelectors(t *testing.T) {
	selectors := deepSeekCitationTriggerSelectors()
	if len(selectors) < 3 {
		t.Fatalf("expected multiple fallback selectors, got %d", len(selectors))
	}
}
