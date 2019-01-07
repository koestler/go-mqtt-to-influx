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
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
)

type CmdOptions struct {
	Config     flags.Filename `short:"c" long:"config" description:"Config File in yaml format" default:"./config.yaml"`
	CpuProfile flags.Filename `long:"cpuprofile" description:"write cpu profile to <file>"`
	MemProfile flags.Filename `long:"memprofile" description:"write memory profile to <file>"`
}

var (
	cmdOptions                 CmdOptions
	cfg                        config.Config
	statisticsInstance         *statistics.Statistics
	mqttClientInstances        map[string]*mqttClient.MqttClient
	influxDbClientPoolInstance *influxDbClient.ClientPool
	httpServerInstance         *httpServer.HttpServer
)

func main() {
	setupConfig()

	if cfg.LogWorkerStart {
		log.Print("main: start go-mqtt-to-influxdb...")
	}

	setupCpuProfile()
	defer pprof.StopCPUProfile()

	setupStatistics()
	setupMqttClient()
	setupInfluxDbClient()
	setupConverters()
	setupHttpServer()

	if cfg.LogWorkerStart {
		log.Print("main: start completed; run until SIGTERM or SIGINT is received")
	}

	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	sig := <-gracefulStop
	if cfg.LogWorkerStart {
		log.Printf("main: caught signal: %+v; shutdown", sig)
	}

	writeMemProfile()

	// shutdown all workers
	for _, client := range mqttClientInstances {
		client.Shutdown()
	}
	httpServerInstance.Shutdown()
	influxDbClientPoolInstance.Shutdown()

	if cfg.LogWorkerStart {
		log.Print("main: shutdown completed; exit")
	}
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

func setupCpuProfile() {
	if cmdOptions.CpuProfile == "" {
		return
	}

	f, err := os.Create(string(cmdOptions.CpuProfile))
	if err != nil {
		log.Fatal("main: could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("main: could not start CPU profile: ", err)
	}
	log.Print("main: started CPU profile")
}

func writeMemProfile() {
	if cmdOptions.MemProfile == "" {
		return
	}

	f, err := os.Create(string(cmdOptions.MemProfile))
	if err != nil {
		log.Fatal("main: could not create memory profile: ", err)
	}
	runtime.GC() // get up-to-date statistics
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatal("main: could not write memory profile: ", err)
	}
	log.Print("main: wrote memory profile")
	f.Close()
}

func setupStatistics() {
	if cfg.LogWorkerStart && cfg.Statistics.Enabled() {
		log.Printf("main: start Statistisc module")
	}

	statisticsInstance = statistics.Run(cfg.Statistics)
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
				"main: start Mqtt client, Name='%s', Broker='%s', ClientId='%s'",
				mqttClientConfig.Name(), mqttClientConfig.Broker(), mqttClientConfig.ClientId(),
			)
		}
		mqttClientInstances[mqttClientConfig.Name()] = mqttClient.Run(mqttClientConfig, statisticsInstance)
	}
}

func setupInfluxDbClient() {
	influxDbClientPoolInstance = influxDbClient.RunPool()

	for _, influxDbClientConfig := range cfg.InfluxDbClients {
		if cfg.LogWorkerStart {
			log.Printf(
				"main: start IinfluxDB client, Name='%s' Address='%s'",
				influxDbClientConfig.Name(),
				influxDbClientConfig.Address(),
			)
		}
		influxDbClientPoolInstance.AddClient(
			influxDbClient.RunClient(influxDbClientConfig, statisticsInstance),
		)
	}
}

func setupConverters() {
	for _, converterConfig := range cfg.Converters {
		handleFunc, err := converter.GetHandler(converterConfig.Implementation())
		messageHandler := getMqttMessageHandler(converterConfig, handleFunc)

		if err != nil {
			log.Printf("Converter[%s]: cannot start: %s", converterConfig.Name(), err)
			continue
		}

		for _, mqttClientInstance := range getMqttClient(converterConfig.MqttClients()) {
			if cfg.LogWorkerStart {
				log.Printf(
					"main: start Converter[%s], Implementation='%s', MqttClient='%s', InfluxDbClients=%v",
					converterConfig.Name(),
					converterConfig.Implementation(),
					mqttClientInstance.Name(),
					converterConfig.InfluxDbClients(),
				)
			}

			for _, mqttTopic := range converterConfig.MqttTopics() {
				if err := mqttClientInstance.Subscribe(mqttTopic, messageHandler); err != nil {
					log.Printf("Converter[%s]: error while subscribing: %s", converterConfig.Name(), err)
				}
			}
		}
	}
}

func getMqttMessageHandler(config converter.Config, handleFunc converter.HandleFunc) mqtt.MessageHandler {
	return func(client mqtt.Client, message mqtt.Message) {
		if config.LogHandleOnce() {
			converter.LogTopicOnce(config.Name(), message)
		}
		statisticsInstance.IncrementOne("converter", config.Name(), message.Topic())
		handleFunc(
			config,
			message,
			func(output converter.Output) {
				influxDbClientPoolInstance.WritePoint(
					output,
					config.InfluxDbClients(),
				)
			},
		)
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

func setupHttpServer() {
	if !cfg.HttpServer.Enabled() {
		return
	}

	if cfg.LogWorkerStart {
		log.Printf("main: start HttpServer, bind=%s, port=%d", cfg.HttpServer.Bind(), cfg.HttpServer.Port())
	}

	httpServerInstance = httpServer.Run(
		&cfg.HttpServer,
		&httpServer.Environment{
			Statistics: statisticsInstance,
		},
	)
}
