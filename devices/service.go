package devices

import "time"

// CalculateUptime returns the percentage of minutes a device was online.
//
// Formula: (sumHeartbeats / numMinutesBetweenFirstAndLastHeartbeat) * 100
//
// Uses pre-computed aggregates from AddHeartbeat, so this runs in O(1)
// regardless of how many heartbeats have been recorded.
//
// Edge cases:
//   - Fewer than 2 heartbeats → 0 (no time range to measure)
//   - First == Last timestamp  → 0 (zero-length window)
func CalculateUptime(d *Device) float64 {
	count, first, last := d.UptimeAggregates()

	if count < 2 {
		return 0
	}

	minutes := last.Sub(first).Minutes()
	if minutes == 0 {
		return 0
	}

	return (float64(count) / minutes) * 100
}

// CalculateAvgUploadTime returns the mean upload duration as a human-readable
// string (e.g. "3m7.893s"). Returns "0s" when no uploads have been recorded.
//
// Uses pre-computed aggregates from AddUpload, so this runs in O(1)
// regardless of how many uploads have been recorded.
func CalculateAvgUploadTime(d *Device) string {
	count, sum := d.UploadAggregates()

	if count == 0 {
		return "0s"
	}

	avg := sum / time.Duration(count)
	return avg.String()
}
