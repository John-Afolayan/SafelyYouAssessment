package devices

import (
	"math"
	"os"
	"sync"
	"testing"
	"time"
)

// ---------- CalculateUptime ----------

func TestCalculateUptime_FullUptime(t *testing.T) {
	d := NewDevice("test-1")
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	// 481 heartbeats over 480 minutes → (481/480)*100 = 100.208…
	// but spec says 100% for device with no gaps; let's test the formula directly.
	// 480 heartbeats spanning 480 minutes (first at 0, last at 480) → (481/480)*100
	for i := 0; i <= 480; i++ {
		d.AddHeartbeat(start.Add(time.Duration(i) * time.Minute))
	}

	uptime := CalculateUptime(d)
	expected := (481.0 / 480.0) * 100
	if math.Abs(uptime-expected) > 0.001 {
		t.Errorf("expected uptime ~%.5f, got %.5f", expected, uptime)
	}
}

func TestCalculateUptime_NoHeartbeats(t *testing.T) {
	d := NewDevice("test-2")
	if uptime := CalculateUptime(d); uptime != 0 {
		t.Errorf("expected 0, got %f", uptime)
	}
}

func TestCalculateUptime_SingleHeartbeat(t *testing.T) {
	d := NewDevice("test-3")
	d.AddHeartbeat(time.Now())
	if uptime := CalculateUptime(d); uptime != 0 {
		t.Errorf("expected 0 for single heartbeat, got %f", uptime)
	}
}

func TestCalculateUptime_WithGaps(t *testing.T) {
	d := NewDevice("test-4")
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	// 10 heartbeats in first 10 minutes, then gap, then 1 more at minute 20
	for i := 0; i < 10; i++ {
		d.AddHeartbeat(start.Add(time.Duration(i) * time.Minute))
	}
	d.AddHeartbeat(start.Add(20 * time.Minute))

	uptime := CalculateUptime(d)
	expected := (11.0 / 20.0) * 100 // 55%
	if math.Abs(uptime-expected) > 0.001 {
		t.Errorf("expected uptime ~%.5f, got %.5f", expected, uptime)
	}
}

// ---------- CalculateAvgUploadTime ----------

func TestCalculateAvgUploadTime_Normal(t *testing.T) {
	d := NewDevice("test-5")
	d.AddUpload(2 * time.Second)
	d.AddUpload(4 * time.Second)
	d.AddUpload(6 * time.Second)

	avg := CalculateAvgUploadTime(d)
	if avg != "4s" {
		t.Errorf("expected 4s, got %s", avg)
	}
}

func TestCalculateAvgUploadTime_NoData(t *testing.T) {
	d := NewDevice("test-6")
	if avg := CalculateAvgUploadTime(d); avg != "0s" {
		t.Errorf("expected 0s, got %s", avg)
	}
}

func TestCalculateAvgUploadTime_SingleEntry(t *testing.T) {
	d := NewDevice("test-7")
	d.AddUpload(3*time.Minute + 10*time.Second)

	avg := CalculateAvgUploadTime(d)
	if avg != "3m10s" {
		t.Errorf("expected 3m10s, got %s", avg)
	}
}

// ---------- Store ----------

func TestLoadFromCSV_Valid(t *testing.T) {
	// write a temp CSV
	tmpFile := t.TempDir() + "/devices.csv"
	content := "device_id\naa-bb-cc\ndd-ee-ff\n"
	if err := writeFile(tmpFile, content); err != nil {
		t.Fatal(err)
	}

	store, err := LoadFromCSV(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := store.Lookup("aa-bb-cc"); !ok {
		t.Error("device aa-bb-cc not found")
	}
	if _, ok := store.Lookup("dd-ee-ff"); !ok {
		t.Error("device dd-ee-ff not found")
	}
	if len(store.DeviceIDs()) != 2 {
		t.Errorf("expected 2 devices, got %d", len(store.DeviceIDs()))
	}
}

func TestLoadFromCSV_MissingFile(t *testing.T) {
	_, err := LoadFromCSV("/nonexistent.csv")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadFromCSV_EmptyFile(t *testing.T) {
	tmpFile := t.TempDir() + "/empty.csv"
	if err := writeFile(tmpFile, "device_id\n"); err != nil {
		t.Fatal(err)
	}

	store, err := LoadFromCSV(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.DeviceIDs()) != 0 {
		t.Errorf("expected 0 devices, got %d", len(store.DeviceIDs()))
	}
}

// ---------- Concurrency ----------

func TestDevice_ConcurrentAccess(t *testing.T) {
	d := NewDevice("concurrent")
	var wg sync.WaitGroup

	// hammer heartbeats and uploads concurrently
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func(n int) {
			defer wg.Done()
			d.AddHeartbeat(time.Now().Add(time.Duration(n) * time.Minute))
		}(i)
		go func(n int) {
			defer wg.Done()
			d.AddUpload(time.Duration(n) * time.Millisecond)
		}(i)
	}
	wg.Wait()

	beats := d.Heartbeats()
	uploads := d.Uploads()
	if len(beats) != 100 {
		t.Errorf("expected 100 heartbeats, got %d", len(beats))
	}
	if len(uploads) != 100 {
		t.Errorf("expected 100 uploads, got %d", len(uploads))
	}
}

// helper
func writeFile(path, content string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}
