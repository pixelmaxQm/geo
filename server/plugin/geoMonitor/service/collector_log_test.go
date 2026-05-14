package service

import (
	"encoding/json"
	"testing"
)

func TestBuildOrderedCitationsJSONPreservesOrder(t *testing.T) {
	citations := []CitationItem{
		{Title: "first", URL: "https://a.example", Snippet: "A", Source: "qwen", Raw: map[string]any{"rank": float64(2)}},
		{Title: "second", URL: "https://b.example", Snippet: "B", Source: "qwen", Raw: map[string]any{"rank": float64(1)}},
	}

	got := BuildOrderedCitationsJSON(citations)

	var decoded []CitationItem
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("unmarshal citations: %v", err)
	}
	if len(decoded) != 2 {
		t.Fatalf("len(decoded) = %d, want 2", len(decoded))
	}
	if decoded[0].Index != 1 || decoded[0].URL != "https://a.example" {
		t.Fatalf("first citation = %+v", decoded[0])
	}
	if decoded[1].Index != 2 || decoded[1].URL != "https://b.example" {
		t.Fatalf("second citation = %+v", decoded[1])
	}
}

func TestExtractCitationsFromRawResponsePreservesOrder(t *testing.T) {
	raw := `{"search_results":[{"title":"first","url":"https://a.example","snippet":"A"},{"title":"second","link":"https://b.example","content":"B"}]}`

	got := ExtractCitationsFromRawResponse(raw, "qwen")

	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}
	if got[0].Title != "first" || got[0].URL != "https://a.example" || got[0].Source != "qwen" {
		t.Fatalf("first citation = %+v", got[0])
	}
	if got[1].Title != "second" || got[1].URL != "https://b.example" || got[1].Snippet != "B" {
		t.Fatalf("second citation = %+v", got[1])
	}
}

func TestRunLogJSONRecordsSteps(t *testing.T) {
	log := NewRunLog()
	log.Add("goto", "success", "页面加载完成", 15)
	log.Add("submit", "failed", "提交失败", 9)

	got := log.JSON()

	var decoded []RunLogItem
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("unmarshal run log: %v", err)
	}
	if decoded[0].Step != "goto" || decoded[1].Status != "failed" {
		t.Fatalf("decoded = %+v", decoded)
	}
}
