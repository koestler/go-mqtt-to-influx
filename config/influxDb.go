package config

import (
	"errors"
	"fmt"
)

type InfluxDbConfig struct {
	Addr     string
	User     string
	Password string
	Database string
}

func GetInfluxDbConfig() (influxDbConfig *InfluxDbConfig, err error) {
	influxDbConfig = &InfluxDbConfig{
		Addr:     "",
		User:     "",
		Password: "",
		Database: "go-mqtt-to-influxdb",
	}

	// check if mqttClient sections exists
	_, err = config.GetSection("InfluxDbClient")
	if err != nil {
		return nil, errors.New("no InfluxDbClient configuration found")
	}

	err = config.Section("InfluxDbClient").MapTo(influxDbConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot read InfluxDbClient configuration: %v", err)
	}

	if len(influxDbConfig.Addr) < 1 {
		return nil, errors.New("Addr not specified")
	}

	return
}
