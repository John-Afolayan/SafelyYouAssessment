package devices

import (
	"encoding/csv"
	"fmt"
	"os"
	"sync"
	"time"
)

// Device holds raw telemetry data collected for a single device.
// All fields are unexported; access is through thread-safe methods.
//
// Running aggregates (heartbeatCount, firstBeat, lastBeat, uploadCount,
// uploadSum) are maintained on each write so that stats reads are O(1)
// instead of O(n). The raw slices are retained for auditability and
// potential future use (e.g. percentile calculations).
type Device struct {
	mu       sync.RWMutex
	deviceID string

	// raw data (append-only)
	heartbeats []time.Time
	uploads    []time.Duration

	// running aggregates, updated on every write
	heartbeatCount int
	firstBeat      time.Time
	lastBeat       time.Time
	uploadCount    int
	uploadSum      time.Duration
}

// NewDevice creates a Device with the given ID and pre-allocated slices.
func NewDevice(id string) *Device {
	return &Device{
		deviceID:   id,
		uploads:    make([]time.Duration, 0, 128),
		heartbeats: make([]time.Time, 0, 512),
	}
}

// ID returns the device identifier.
func (d *Device) ID() string { return d.deviceID }

// AddHeartbeat safely appends a heartbeat timestamp and updates aggregates.
func (d *Device) AddHeartbeat(t time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.heartbeats = append(d.heartbeats, t)
	d.heartbeatCount++
	if d.heartbeatCount == 1 {
		d.firstBeat = t
	}
	d.lastBeat = t
}

// AddUpload safely appends an upload duration and updates aggregates.
func (d *Device) AddUpload(dur time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.uploads = append(d.uploads, dur)
	d.uploadCount++
	d.uploadSum += dur
}

// UptimeAggregates returns the values needed to compute uptime in O(1).
// Returns (count, first, last) under a read lock.
func (d *Device) UptimeAggregates() (int, time.Time, time.Time) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.heartbeatCount, d.firstBeat, d.lastBeat
}

// UploadAggregates returns the values needed to compute average upload time in O(1).
// Returns (count, sum) under a read lock.
func (d *Device) UploadAggregates() (int, time.Duration) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.uploadCount, d.uploadSum
}

// Heartbeats returns a snapshot of all heartbeat timestamps.
// Useful for detailed analysis; prefer UptimeAggregates for stats.
func (d *Device) Heartbeats() []time.Time {
	d.mu.RLock()
	defer d.mu.RUnlock()
	out := make([]time.Time, len(d.heartbeats))
	copy(out, d.heartbeats)
	return out
}

// Uploads returns a snapshot of all upload durations.
// Useful for detailed analysis; prefer UploadAggregates for stats.
func (d *Device) Uploads() []time.Duration {
	d.mu.RLock()
	defer d.mu.RUnlock()
	out := make([]time.Duration, len(d.uploads))
	copy(out, d.uploads)
	return out
}

// Store holds all known devices keyed by device ID.
// The map itself is read-only after initialization, so no mutex is needed on it.
type Store struct {
	devices map[string]*Device
}

// NewStore creates an empty Store.
func NewStore() *Store {
	return &Store{devices: make(map[string]*Device)}
}

// Register adds a device to the store. Must only be called during initialization.
func (s *Store) Register(id string) {
	s.devices[id] = NewDevice(id)
}

// Lookup returns the device for the given ID, or nil if not found.
func (s *Store) Lookup(id string) (*Device, bool) {
	d, ok := s.devices[id]
	return d, ok
}

// DeviceIDs returns all registered device IDs (useful for health checks / listing).
func (s *Store) DeviceIDs() []string {
	ids := make([]string, 0, len(s.devices))
	for id := range s.devices {
		ids = append(ids, id)
	}
	return ids
}

// LoadFromCSV reads device definitions from a CSV file and returns an initialized Store.
func LoadFromCSV(path string) (*Store, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening device CSV %q: %w", path, err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parsing device CSV %q: %w", path, err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("device CSV %q is empty", path)
	}

	store := NewStore()
	for _, row := range records[1:] { // skip header
		if len(row) == 0 || row[0] == "" {
			continue
		}
		store.Register(row[0])
	}

	return store, nil
}
