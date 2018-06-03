package main

import (
	"encoding/csv"
	"fmt"
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
		rpt := rpt.GetReport(title)
		if rpt != nil {
			writeReport(wcsv, rpt)
		}
		time.Sleep(pollInterval)
	}
}

func writeReport(wcsv *csv.Writer, rpt *spent.Report) {
	row := []string{
		rpt.At.Format(time.RFC3339),
		fmt.Sprintf("%.0f", rpt.Elapsed.Seconds()),
		rpt.Title,
	}
	if rpt.App != "" {
		row = append(row, rpt.App)
		if rpt.AppContext != "" {
			row = append(row, rpt.AppContext)
			if rpt.AppDetail != "" {
				row = append(row, rpt.AppDetail)
			}
		}
	}
	err := wcsv.Write(row)
	if err != nil {
		log.Fatalf("Unable to write CSV: %v", err)
	}
	wcsv.Flush()
}
