package main

import (
	"flag"
	"fmt"
	"safelyyou-assessment/devices"
)

func main() {
	// parse for devices.csv path using flag
	csvPath := flag.String("device_path", "devices.csv", "path to devices csv file")
	flag.Parse()
	csvfile, err := devices.ReadCsvFile(csvPath)
	if err != nil {
		fmt.Printf("error reading csv file %v", err)
	}
	fmt.Printf("csv file contents %v\n", csvfile)

}
