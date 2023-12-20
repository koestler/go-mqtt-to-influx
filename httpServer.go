package main

import (
	"github.com/koestler/go-mqtt-to-influx/v2/config"
	"github.com/koestler/go-mqtt-to-influx/v2/httpServer"
	"github.com/koestler/go-mqtt-to-influx/v2/statistics"
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
		httpCfg,
		&httpServer.Environment{
			Statistics: statisticsInstance,
		},
	)
}
