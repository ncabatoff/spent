package main

import (
	"encoding/csv"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/ncabatoff/spent"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// User is idle when idletime exceeds idleCutoff.
	idleCutoff = 180 * time.Second
	// pollInterval is how often we check to see if anything has changed.
	pollInterval = 5 * time.Second
	// writeInterval is how often a line is written in the absence of change.
	writeInterval = 1 * time.Second
)

var (
	idleSeconds = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "idle_seconds_total",
		Help: "seconds idle, i.e screensaver active or enough time elapsed without keyboard/mouse activity",
	})
	activeSeconds = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "active_seconds_total",
		Help: "seconds idle, i.e screensaver active or enough time elapsed without keyboard/mouse activity",
	}, []string{
		"app",
		"appcontext",
	})
)

func init() {
	// Register the summary and the histogram with Prometheus's default registry.
	prometheus.MustRegister(activeSeconds)
	prometheus.MustRegister(idleSeconds)
}

func main() {
	var (
		addr = flag.String("listen-address", ":9357", "The address to listen on for HTTP requests.")
	)

	go gather()
	// Expose the registered metrics via HTTP.
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func gather() {
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
			if rpt.Title == "idle" {
				idleSeconds.Add(rpt.Elapsed.Seconds())
			} else if rpt.App != "" {
				activeSeconds.WithLabelValues(rpt.App, rpt.AppContext).Add(rpt.Elapsed.Seconds())
			}
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
