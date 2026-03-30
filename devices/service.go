package devices

import "time"

// CalculateUptime calculates the percentage of minutes a device was online
// based on the number of heartbeats received vs the total minutes elapsed
// between the first and last heartbeat.
// Formula: (sumHeartbeats / numMinutesBetweenFirstAndLastHeartbeat) * 100
func CalculateUptime(device *Device) float64 {
	// need at least 2 heartbeats to calculate a time range
	if len(device.SentAt) < 2 {
		return 0
	}

	// number of heartbeats received
	sumHeartbeats := float64(len(device.SentAt))

	// time elapsed between first and last heartbeat in minutes
	first := device.SentAt[0]
	last := device.SentAt[len(device.SentAt)-1]
	// subtract first heartbeat from last to get total elapsed time, converted to minutes
	numMinutes := last.Sub(first).Minutes()

	// avoid division by zero
	if numMinutes == 0 {
		return 0
	}

	return (sumHeartbeats / numMinutes) * 100
}

// CalculateAvgUploadTime calculates the average upload duration
// across all recorded upload times for a device and returns it
// as a human-readable duration string e.g. "5m10s"
func CalculateAvgUploadTime(device *Device) string {
	// no data yet
	if len(device.UploadTime) == 0 {
		return "0s"
	}

	// sum all upload durations
	var total time.Duration
	for _, d := range device.UploadTime {
		total += d
	}

	// divide by count to get average
	avg := total / time.Duration(len(device.UploadTime))

	// time.Duration.String() returns human-readable format e.g. "5m10s"
	return avg.String()
}
