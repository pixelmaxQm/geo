package service

import (
	"encoding/json"
	"testing"
)

func TestMarshalKeywordAnalysisJSON(t *testing.T) {
	got := marshalKeywordAnalysisJSON([]string{"性能", "Gin", "框架"})
	var decoded []string
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("unmarshal analysis json: %v", err)
	}
	if len(decoded) != 3 {
		t.Fatalf("expected 3 ranked keywords, got %d", len(decoded))
	}
	if decoded[0] != "性能" || decoded[1] != "Gin" || decoded[2] != "框架" {
		t.Fatalf("unexpected keyword ranking: %+v", decoded)
	}
}
