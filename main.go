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
	Config flags.Filename `short:"c" long:"config" description:"Config File in ini format" default:"./config.yaml"`
}

var (
	cmdOptions             CmdOptions
	mqttClientInstance     *mqttClient.MqttClient
	influxDbClientInstance *influxDbClient.InfluxDbClient
	configInstance         config.Config
)

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

	// read, transform and validate configuration
	configInstance = config.ReadConfig(string(cmdOptions.Config))
}

func setupMqttClient() {
	log.Printf(
		"main: start mqtt client, broker=%v, clientId=%v",
		configInstance.MqttClient.Broker, configInstance.MqttClient.ClientId,
	)
	mqttClientInstance = mqttClient.Run(&configInstance.MqttClient)
}

func setupInfluxDbClient() {
	log.Printf(
		"main: start influxDB client, addr=%v",
		configInstance.InfluxDbClient.Address,
	)
	influxDbClientInstance = influxDbClient.Run(&configInstance.InfluxDbClient)
}

func setupConverters() {
	for _, convertConfig := range configInstance.Converters {
		log.Printf("main: start converter %s", convertConfig.Name)
		if err := converter.RunConverter(&convertConfig, mqttClientInstance, influxDbClientInstance); err != nil {
			log.Fatalf("main: cannot get converter; err=%s", err)
		}
	}
}
