package main

import (
	"github.com/koestler/go-mqtt-to-influx/v2/config"
	"github.com/koestler/go-mqtt-to-influx/v2/influxClient"
	LocalDb "github.com/koestler/go-mqtt-to-influx/v2/localDb"
	"github.com/koestler/go-mqtt-to-influx/v2/statistics"
	"log"
)

func runInfluxClient(
	cfg *config.Config,
	localDbInstance LocalDb.LocalDb,
	statisticsInstance statistics.Statistics,
	initiateShutdown chan<- error,
) (influxClientPoolInstance *influxClient.ClientPool) {
	influxClientPoolInstance = influxClient.RunPool()

	// convert []*config.InfluxAuxiliaryTags to []influxClient.AuxiliaryTag
	auxiliaryTags := make([]influxClient.AuxiliaryTag, len(cfg.InfluxAuxiliaryTags()))
	for i, t := range cfg.InfluxAuxiliaryTags() {
		auxiliaryTags[i] = t
	}

	for _, influxClientConfig := range cfg.InfluxClients() {
		if cfg.LogWorkerStart() {
			log.Printf(
				"influxClient[%s]: start: url='%s', len(token)=%d, org='%s', bucket='%s'",
				influxClientConfig.Name(),
				influxClientConfig.Url(),
				len(influxClientConfig.Token()),
				influxClientConfig.Org(),
				influxClientConfig.Bucket(),
			)
		}

		client := influxClient.RunClient(influxClientConfig, auxiliaryTags, localDbInstance, statisticsInstance)

		influxClientPoolInstance.AddClient(client)
		if cfg.LogWorkerStart() {
			log.Printf("influxClient[%s]: started", influxClientConfig.Name())
		}
	}

	return
}
