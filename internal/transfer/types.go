package transfer

// ProgressInfo contains transfer progress information
type ProgressInfo struct {
	Rate         float64 // bytes per second
	BytesTotal   int64
	BytesCurrent int64
	Progress     float64 // 0-1
}
