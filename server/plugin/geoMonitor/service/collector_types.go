package service

type CollectOutput struct {
	Answer         string         `json:"answer"`
	Citations      string         `json:"citations"`
	CitationItems  []CitationItem `json:"citationItems"`
	AnalysisJSON   string         `json:"analysisJson"`
	ScreenshotPath string         `json:"screenshotPath"`
	DurationMs     int            `json:"durationMs"`
	RawResponse    string         `json:"rawResponse"`
	RunLog         string         `json:"runLog"`
	ErrorMsg       string         `json:"errorMsg"`
}

const (
	CollectModeAPI        = "api"
	CollectModePlaywright = "playwright"

	TaskStatusRunning = "running"
	TaskStatusDone    = "done"
	TaskStatusFailed  = "failed"
	TaskStatusPartial = "partial"

	ResultStatusSuccess = "success"
	ResultStatusFailed  = "failed"
)
