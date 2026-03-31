package devices

import (
	"testing"
	"time"
)

func BenchmarkAddHeartbeat(b *testing.B) {
	d := NewDevice("bench")
	t := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.AddHeartbeat(t.Add(time.Duration(i) * time.Minute))
	}
}

func BenchmarkAddUpload(b *testing.B) {
	d := NewDevice("bench")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.AddUpload(time.Duration(i) * time.Millisecond)
	}
}

func BenchmarkCalculateUptime(b *testing.B) {
	d := NewDevice("bench")
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 500; i++ {
		d.AddHeartbeat(start.Add(time.Duration(i) * time.Minute))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateUptime(d)
	}
}

func BenchmarkCalculateAvgUploadTime(b *testing.B) {
	d := NewDevice("bench")
	for i := 0; i < 100; i++ {
		d.AddUpload(time.Duration(i) * time.Second)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateAvgUploadTime(d)
	}
}
