package types

import "time"

type ScannedNodeModule struct {
	Path         string
	Staleness    int64 // In days
	Size         int64
	LastModified time.Time
}

type ScanInfo struct {
	TotalSize    int64
	AvgStaleness float64
	ScanDuration time.Duration
}
