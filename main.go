package main

import (
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/jessevdk/go-flags"
	"github.com/koestler/go-mqtt-to-influxdb/config"
	"github.com/koestler/go-mqtt-to-influxdb/converter"
	"github.com/koestler/go-mqtt-to-influxdb/httpServer"
	"github.com/koestler/go-mqtt-to-influxdb/influxDbClient"
	"github.com/koestler/go-mqtt-to-influxdb/mqttClient"
	"github.com/koestler/go-mqtt-to-influxdb/statistics"
	"log"
	"os"
)

type CmdOptions struct {
	Config flags.Filename `short:"c" long:"config" description:"Config File in yaml format" default:"./config.yaml"`
}

var (
	cmdOptions                 CmdOptions
	mqttClientInstances        map[string]*mqttClient.MqttClient
	influxDbClientPoolInstance *influxDbClient.ClientPool
	httpServerInstance         *httpServer.HttpServer
	statisticsInstance         *statistics.Statistics

	cfg config.Config
)

func main() {
	setupConfig()

	if cfg.LogWorkerStart {
		log.Print("main: start go-mqtt-to-influxdb...")
	}

	setupMqttClient()
	setupInfluxDbClient()
	setupConverters()
	setupStatistics()
	setupHttpServer()

	if cfg.LogWorkerStart {
		log.Print("main: start completed; run until kill signal is received")
	}

	select {}
}

func setupConfig() {
	// parse command line options
	parser := flags.NewParser(&cmdOptions, flags.Default)
	parser.Usage = "[-c <path to yaml config file>]"
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	// read, transform and validate configuration
	var err []error
	cfg, err = config.ReadConfigFile(parser.Name, string(cmdOptions.Config))
	if len(err) > 0 {
		for _, e := range err {
			log.Printf("config: error: %v", e)
		}
		os.Exit(2)
	}

	if cfg.LogConfig {
		cfg.PrintConfig()
	}
}

func setupMqttClient() {
	mqtt.ERROR = log.New(os.Stdout, "MqttDebugLog: ", log.LstdFlags)
	if cfg.LogMqttDebug {
		mqtt.DEBUG = log.New(os.Stdout, "MqttDebugLog: ", log.LstdFlags)
	}

	mqttClientInstances = make(map[string]*mqttClient.MqttClient)

	for _, mqttClientConfig := range cfg.MqttClients {
		if cfg.LogWorkerStart {
			log.Printf(
				"main: start mqtt client, name='%s', broker='%s', clientId='%s'",
				mqttClientConfig.Name, mqttClientConfig.Broker, mqttClientConfig.ClientId,
			)
		}
		mqttClientInstances[mqttClientConfig.Name] = mqttClient.Run(mqttClientConfig)
	}
}

func setupInfluxDbClient() {
	influxDbClientPoolInstance = influxDbClient.RunPool()

	for _, influxDbClientConfig := range cfg.InfluxDbClients {
		if cfg.LogWorkerStart {
			log.Printf(
				"main: start influxDB client, name='%s' addr='%s'",
				influxDbClientConfig.Name,
				influxDbClientConfig.Address,
			)
		}
		influxDbClientPoolInstance.AddClient(
			influxDbClient.RunClient(influxDbClientConfig),
		)
	}
}

func setupConverters() {
	for _, convertConfig := range cfg.Converters {
		for _, clientInstance := range getMqttClient(convertConfig.MqttClients) {
			if cfg.LogWorkerStart {
				log.Printf(
					"main: start converter name='%s', implementation='%s', mqttClient='%s', influxDbClients=%v",
					convertConfig.Name,
					convertConfig.Implementation,
					clientInstance.GetName(),
					convertConfig.InfluxDbClients,
				)
			}

			if err := converter.RunConverter(convertConfig, clientInstance, influxDbClientPoolInstance); err != nil {
				log.Fatalf("main: cannot start converter: %s", err)
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
			i++
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

func setupStatistics() {
	if !cfg.Statistics.Enabled {
		return
	}

	if cfg.LogWorkerStart {
		log.Printf("main: start Statistisc module")
	}

	statisticsInstance = statistics.Run()
}

func setupHttpServer() {
	if !cfg.HttpServer.Enabled() {
		return
	}

	if cfg.LogWorkerStart {
		log.Printf("main: start Http server, bind=%s, port=%d", cfg.HttpServer.Bind(), cfg.HttpServer.Port())
	}

	httpServerInstance = httpServer.Run(
		cfg.HttpServer,
		&httpServer.Environment{
			Statistics: statisticsInstance,
		},
	)
}
