package shared

import "time"

type WatcherEvent struct {
	WatcherID string         `json:"watcher_id"`
	ProcessID int            `json:"process_id"`
	Timestamp time.Time      `json:"timestamp"`
	EventType string         `json:"event_type"`
	AppName   string         `json:"app_name"`
	Context   map[string]any `json:"context"`
}
