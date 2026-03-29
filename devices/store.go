package devices

import (
	"encoding/csv"
	"log"
	"os"
	"time"
)

type Device struct {
	DeviceId string
	UploadTime []time.Duration
	SentAt []time.Time
}

type Store struct {
	Devices map[string]*Device
} 

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

// reads the CSV and returns a populated *Store
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