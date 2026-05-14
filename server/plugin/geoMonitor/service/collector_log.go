package service

import (
	"encoding/json"
	"time"
)

type CitationItem struct {
	Index   int            `json:"index"`
	Title   string         `json:"title"`
	URL     string         `json:"url"`
	Snippet string         `json:"snippet"`
	Source  string         `json:"source"`
	Raw     map[string]any `json:"raw,omitempty"`
}

type RunLogItem struct {
	Step       string `json:"step"`
	Status     string `json:"status"`
	Message    string `json:"message"`
	DurationMs int64  `json:"durationMs"`
	Time       string `json:"time"`
}

type RunLog struct {
	items []RunLogItem
}

func NewRunLog() *RunLog {
	return &RunLog{items: make([]RunLogItem, 0)}
}

func (l *RunLog) Add(step string, status string, message string, durationMs int64) {
	l.items = append(l.items, RunLogItem{Step: step, Status: status, Message: message, DurationMs: durationMs, Time: time.Now().Format(time.RFC3339)})
}

func (l *RunLog) JSON() string {
	if l == nil || len(l.items) == 0 {
		return "[]"
	}
	data, err := json.Marshal(l.items)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func BuildOrderedCitationsJSON(items []CitationItem) string {
	if len(items) == 0 {
		return "[]"
	}
	normalized := make([]CitationItem, len(items))
	copy(normalized, items)
	for i := range normalized {
		normalized[i].Index = i + 1
		if normalized[i].Raw == nil {
			normalized[i].Raw = map[string]any{}
		}
	}
	data, err := json.Marshal(normalized)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func ExtractCitationsFromRawResponse(raw string, source string) []CitationItem {
	var data any
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil
	}
	items := make([]CitationItem, 0)
	walkCitationValue(data, source, &items)
	return items
}

func walkCitationValue(value any, source string, items *[]CitationItem) {
	switch typed := value.(type) {
	case map[string]any:
		for key, val := range typed {
			if isCitationArrayKey(key) {
				appendCitationArray(val, source, items)
				continue
			}
			walkCitationValue(val, source, items)
		}
	case []any:
		for _, item := range typed {
			walkCitationValue(item, source, items)
		}
	}
}

func appendCitationArray(value any, source string, items *[]CitationItem) {
	list, ok := value.([]any)
	if !ok {
		return
	}
	for _, item := range list {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		citation := CitationItem{
			Title:   firstString(obj, "title", "name"),
			URL:     firstString(obj, "url", "link", "href"),
			Snippet: firstString(obj, "snippet", "content", "summary", "text"),
			Source:  source,
			Raw:     obj,
		}
		if citation.Title == "" && citation.URL == "" && citation.Snippet == "" {
			continue
		}
		*items = append(*items, citation)
	}
}

func isCitationArrayKey(key string) bool {
	switch key {
	case "search_results", "searchResults", "citations", "references", "web_search", "webSearch", "search", "sources":
		return true
	default:
		return false
	}
}

func firstString(obj map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := obj[key].(string); ok {
			return value
		}
	}
	return ""
}
