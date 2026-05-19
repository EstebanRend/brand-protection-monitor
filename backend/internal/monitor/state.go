package monitor

import "time"

type State struct {
    LastTreeSize       int64      `json:"lastTreeSize"`
    LastProcessedAt    *time.Time `json:"lastProcessedAt"`
    ProcessedLastCycle int        `json:"processedLastCycle"`
    Status             string     `json:"status"`
}
