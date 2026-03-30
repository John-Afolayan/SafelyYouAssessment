package devices

import (
	"encoding/csv"
	"log"
	"os"
	"sync"
	"time"
)

// Device struct holds raw telemetry data collected for a single device
type Device struct {
	mu sync.Mutex
	DeviceId string
	UploadTime []time.Duration
	SentAt []time.Time
}

// Store holds all known devices keyed by device ID
type Store struct {
	Devices map[string]*Device
}

// ReadCsvFile reads a CSV file and returns its raw rows
func ReadCsvFile(csvPath *string) ([][]string, error) {
	f, err := os.Open(*csvPath)
	if err != nil {
		log.Printf("Unable to read input file "+*csvPath, err)
		return nil, err
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Printf("Unable to parse file as CSV for "+*csvPath, err)
		return nil, err
	}

	return records, nil
}

// LoadFromCSV reads device definitions from a CSV and returns an initialized Store
func LoadFromCSV(csvPath *string) (*Store, error) {
	csvContents, err := ReadCsvFile(csvPath); if err != nil {
		log.Printf("Unable to parse file as CSV for "+*csvPath, err)
		return nil, err
	}

	m := make(map[string]*Device)
	// iterate through devices csv, skipping header
	for i := 1; i < len(csvContents); i++ {
		s := csvContents[i][0]
		m[s] = &Device{
			DeviceId: s,
			UploadTime: []time.Duration{},
			SentAt: []time.Time{},
		}
	}
	return &Store{Devices: m}, nil
}

// AddHeartbeat safely appends a heartbeat timestamp to the device
func (d *Device) AddHeartbeat(t time.Time) {
    d.mu.Lock()
    defer d.mu.Unlock()
    d.SentAt = append(d.SentAt, t)
}
