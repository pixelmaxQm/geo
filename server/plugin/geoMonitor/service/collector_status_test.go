package service

import (
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/model"
)

func TestSummarizeTaskStatus(t *testing.T) {
	tests := []struct {
		name    string
		results []model.CollectionResult
		want    string
	}{
		{
			name:    "returns failed when no results",
			results: nil,
			want:    TaskStatusFailed,
		},
		{
			name: "returns done when all results succeed",
			results: []model.CollectionResult{
				{Status: ResultStatusSuccess},
				{Status: ResultStatusSuccess},
			},
			want: TaskStatusDone,
		},
		{
			name: "returns failed when all results fail",
			results: []model.CollectionResult{
				{Status: ResultStatusFailed},
				{Status: ResultStatusFailed},
			},
			want: TaskStatusFailed,
		},
		{
			name: "returns partial when success and failed are mixed",
			results: []model.CollectionResult{
				{Status: ResultStatusSuccess},
				{Status: ResultStatusFailed},
			},
			want: TaskStatusPartial,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := summarizeTaskStatus(tt.results)
			if got != tt.want {
				t.Fatalf("summarizeTaskStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}
