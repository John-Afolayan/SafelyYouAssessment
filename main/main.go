package main

import (
	"flag"
	"log"

	"safelyyou-assessment/devices"
	"safelyyou-assessment/server"
)

func main() {
	csvPath := flag.String("device_path", "devices.csv", "path to devices CSV file")
	addr := flag.String("addr", "127.0.0.1:6733", "address to listen on")
	flag.Parse()

	store, err := devices.LoadFromCSV(*csvPath)
	if err != nil {
		log.Fatalf("failed to load devices: %v", err)
	}

	log.Printf("loaded %d devices from %s", len(store.DeviceIDs()), *csvPath)
	log.Printf("server starting on %s", *addr)

	s := server.New(store)
	if err := s.Run(*addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
