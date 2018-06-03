package main

import (
	"encoding/csv"
	"log"
	"os"
	"time"

	"github.com/ncabatoff/spent"
)

const (
	// User is idle when idletime exceeds idleCutoff.
	idleCutoff = 180 * time.Second
	// pollInterval is how often we check to see if anything has changed.
	pollInterval = 5 * time.Second
	// writeInterval is how often a line is written in the absence of change.
	writeInterval = 15 * time.Minute
)

func main() {
	wcsv := csv.NewWriter(os.Stdout)
	rpt := spent.NewReporter(writeInterval)
	for {
		idle, err := spent.GetIdleTime()
		if err != nil {
			log.Fatalf("error getting idle time via xprintidle: %v", err)
		}
		title, err := spent.GetActiveWindow()
		if err != nil {
			log.Fatalf("Unable to read active window: %v", err)
		}
		saverOn, err := spent.GetScreensaverOn()
		if err != nil {
			log.Fatalf("Unable to read screensaver status: %v", err)
		}
		if saverOn || idle > idleCutoff {
			title = "idle"
		}
		row := rpt.GetReport(title)
		if row != nil {
			writeReport(wcsv, row)
		}
		time.Sleep(pollInterval)
	}
}

func writeReport(wcsv *csv.Writer, row []string) {
	err := wcsv.Write(row)
	if err != nil {
		log.Fatalf("Unable to write CSV: %v", err)
	}
	wcsv.Flush()
}
