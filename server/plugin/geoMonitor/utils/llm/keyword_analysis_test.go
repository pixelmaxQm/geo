package llm

import "testing"

func TestParseKeywordRanking(t *testing.T) {
	got, err := parseKeywordRanking("```json\n[\"性能\",\"Gin\",\"框架\"]\n```")
	if err != nil {
		t.Fatalf("parse keyword ranking: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 items, got %d", len(got))
	}
	if got[0] != "性能" || got[1] != "Gin" || got[2] != "框架" {
		t.Fatalf("unexpected parsed ranking: %+v", got)
	}
}
