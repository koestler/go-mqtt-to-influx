package main

import (
	"github.com/koestler/go-mqtt-to-influx/config"
	"github.com/koestler/go-mqtt-to-influx/statistics"
	"log"
)

func runStatistics(cfg *config.Config) statistics.Statistics {
	statCfg := cfg.Statistics()

	if cfg.LogWorkerStart() && statCfg.Enabled() {
		log.Printf("statitics: start")
	}

	return statistics.Run(statCfg)
}
