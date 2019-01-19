package main

import (
	"github.com/koestler/go-mqtt-to-influxdb/config"
	"github.com/koestler/go-mqtt-to-influxdb/httpServer"
	"github.com/koestler/go-mqtt-to-influxdb/statistics"
	"log"
)

func runHttpServer(cfg *config.Config, statisticsInstance statistics.Statistics) *httpServer.HttpServer {
	if !cfg.HttpServer.Enabled() {
		return nil
	}

	if cfg.LogWorkerStart {
		log.Printf("httpServer: start: bind=%s, port=%d", cfg.HttpServer.Bind(), cfg.HttpServer.Port())
	}

	return httpServer.Run(
		&cfg.HttpServer,
		&httpServer.Environment{
			Statistics: statisticsInstance,
		},
	)
}
