package main

import (
	"github.com/koestler/go-mqtt-to-influxdb/config"
	"github.com/koestler/go-mqtt-to-influxdb/statistics"
	"log"
)

func runStatistics(cfg *config.Config) statistics.Statistics {
	if cfg.LogWorkerStart && cfg.Statistics.Enabled() {
		log.Printf("main: start Statistisc module")
	}

	return statistics.Run(cfg.Statistics)
}
