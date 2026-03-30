package main

import (
	"flag"
	"fmt"
	"log"
	"safelyyou-assessment/devices"
	"safelyyou-assessment/server"
)

func main() {
	// configure CLI flags
	csvPath := flag.String("device_path", "devices.csv", "path to devices csv file")
	addr := flag.String("addr", ":6733", "address to listen on")
	flag.Parse()
	// load known devices from CSV into memory
	store, err := devices.LoadFromCSV(csvPath)
	if err != nil {
		fmt.Printf("error getting store contents from CSV file: %v", err)
	}
	
	log.Printf("server starting on %s", *addr)
	s := server.New(store)
	err = s.Run(*addr); if err != nil {
		fmt.Printf("error starting server: %v", err)
	}

}
