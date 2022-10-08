package main

import (
	"github.com/koestler/go-mqtt-to-influx/config"
	"github.com/koestler/go-mqtt-to-influx/httpServer"
	"github.com/koestler/go-mqtt-to-influx/statistics"
	"log"
)

func runHttpServer(cfg *config.Config, statisticsInstance statistics.Statistics) *httpServer.HttpServer {
	httpCfg := cfg.HttpServer()

	if !httpCfg.Enabled() {
		return nil
	}

	if cfg.LogWorkerStart() {
		log.Printf("httpServer: start: bind=%s, port=%d", httpCfg.Bind(), httpCfg.Port())
	}

	return httpServer.Run(
		cfg.HttpServer(),
		&httpServer.Environment{
			Statistics: statisticsInstance,
		},
	)
}
