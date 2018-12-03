package main

import (
	"github.com/jessevdk/go-flags"
	"github.com/koestler/go-mqtt-to-influxdb/config"
	"github.com/koestler/go-mqtt-to-influxdb/converter"
	"github.com/koestler/go-mqtt-to-influxdb/influxDbClient"
	"github.com/koestler/go-mqtt-to-influxdb/mqttClient"
	"log"
	"os"
)

type CmdOptions struct {
	Config flags.Filename `short:"c" long:"config" description:"Config File in ini format" default:"./config.ini"`
}

var cmdOptions CmdOptions

var mqttClientInstance *mqttClient.MqttClient
var influxDbClientInstance *influxDbClient.InfluxDbClient

func main() {
	log.Print("main: start go-mqtt-to-influxdb...")

	setupConfig()
	setupMqttClient()
	setupInfluxDbClient()
	setupConverters()

	log.Print("main: start completed; run until kill signal is received")

	select {}
}

func setupConfig() {
	log.Printf("main: setup config")

	// parse command line options
	parser := flags.NewParser(&cmdOptions, flags.Default)
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}
	// initialize config library
	config.Setup(string(cmdOptions.Config))
}

func setupMqttClient() {
	mqttClientConfig, err := config.GetMqttClientConfig()
	if err == nil {
		log.Printf(
			"main: start mqtt client, broker=%v, clientId=%v",
			mqttClientConfig.Broker, mqttClientConfig.ClientId,
		)
		mqttClientInstance = mqttClient.Run(mqttClientConfig)
	} else {
		log.Fatalf("main: skip mqtt client, err=%v", err)
	}
}

func setupInfluxDbClient() {
	influxDbClientConfig, err := config.GetInfluxDbConfig()
	if err == nil {
		log.Printf(
			"main: start influxDB client, addr=%v",
			influxDbClientConfig.Addr,
		)
		influxDbClientInstance = influxDbClient.Run(influxDbClientConfig)
	} else {
		log.Fatalf("main: skip influxDb client, err=%v", err)
	}
}

func setupConverters() {
	converterConfigs := config.GetConvertConfigs()
	for _, convertConfig := range converterConfigs {
		log.Printf("main: start converter %s", convertConfig.Name)
		if err := converter.RunConverter(convertConfig, mqttClientInstance, influxDbClientInstance); err != nil {
			log.Fatalf("main: cannot get converter; err=%s", err)
		}
	}
}
