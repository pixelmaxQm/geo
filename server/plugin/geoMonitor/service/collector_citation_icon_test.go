package service

import (
	"strings"
	"testing"
)

func TestBuildOrderedCitationsJSONPreservesIconAndSummary(t *testing.T) {
	got := BuildOrderedCitationsJSON([]CitationItem{{
		Title:   "Go Packages",
		URL:     "https://pkg.go.dev/github.com/mingrammer/go-web-framework-stars",
		Icon:    "https://cdn.deepseek.com/site-icons/pkg.go.dev",
		Snippet: "gin | 86685 | 8466 | 869 | Gin is a high-performance HTTP web framework written in Go.",
		Source:  "deepseek",
	}})
	if got == "[]" {
		t.Fatalf("expected citation JSON to include icon and snippet")
	}
	checks := []string{
		"\"title\":\"Go Packages\"",
		"\"url\":\"https://pkg.go.dev/github.com/mingrammer/go-web-framework-stars\"",
		"\"icon\":\"https://cdn.deepseek.com/site-icons/pkg.go.dev\"",
		"\"snippet\":\"gin | 86685 | 8466 | 869 | Gin is a high-performance HTTP web framework written in Go.\"",
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("expected citation JSON to contain %q, got %s", check, got)
		}
	}
}
