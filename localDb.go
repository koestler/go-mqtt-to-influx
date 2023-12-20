package main

import (
	"github.com/koestler/go-mqtt-to-influx/v2/config"
	LocalDb "github.com/koestler/go-mqtt-to-influx/v2/localDb"
	"log"
)

func runLocalDb(cfg *config.Config) LocalDb.LocalDb {
	localDbCfg := cfg.LocalDb()

	if cfg.LogWorkerStart() && localDbCfg.Enabled() {
		log.Printf("localDb: start: path=%s", localDbCfg.Path())
	}

	return LocalDb.Run(localDbCfg)
}
