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
	cmdOptions                 CmdOptions
	mqttClientInstances        map[string]*mqttClient.MqttClient
	influxDbClientPoolInstance *influxDbClient.InfluxDbClientPool
	configInstance             config.Config
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
	mqttClientInstances = make(map[string]*mqttClient.MqttClient)

	for _, mqttClientConfig := range configInstance.MqttClients {
		log.Printf(
			"main: start mqtt client, name=%s, broker=%s, clientId=%s",
			mqttClientConfig.Name, mqttClientConfig.Broker, mqttClientConfig.ClientId,
		)
		mqttClientInstances[mqttClientConfig.Name] = mqttClient.Run(mqttClientConfig)
	}
}

func setupInfluxDbClient() {
	influxDbClientPoolInstance = influxDbClient.RunPool()

	for _, influxDbClientConfig := range configInstance.InfluxDbClients {
		log.Printf(
			"main: start influxDB client, name=%s addr=%v",
			influxDbClientConfig.Name,
			influxDbClientConfig.Address,
		)
		influxDbClientPoolInstance.AddClient(
			influxDbClient.RunClient(influxDbClientConfig),
		)
	}
}

func setupConverters() {
	for _, convertConfig := range configInstance.Converters {
		for _, clientInstance := range getMqttClient(convertConfig.MqttClients) {
			log.Printf(
				"main: start converter name=%s, implementation=%s, mqttClient=%s, influxDbClients=%v",
				convertConfig.Name,
				convertConfig.Implementation,
				clientInstance.GetName(),
				convertConfig.InfluxDbClients,
			)

			if err := converter.RunConverter(convertConfig, clientInstance, influxDbClientPoolInstance); err != nil {
				log.Fatalf("main: cannot start converter; err=%s", err)
			}
		}
	}
}

func getMqttClient(clientNames []string) (clients []*mqttClient.MqttClient) {
	if len(clientNames) < 1 {
		clients = make([]*mqttClient.MqttClient, len(mqttClientInstances))
		i := 0
		for _, c := range mqttClientInstances {
			clients[i] = c
			i += 1
		}
		return
	}

	for _, clientName := range clientNames {
		if receiver, ok := mqttClientInstances[clientName]; ok {
			clients = append(clients, receiver)
		}
	}

	return
}
