package config

import (
	"errors"
	"fmt"
	"time"
)

type InfluxDbConfigRead struct {
	Addr          string
	User          string
	Password      string
	Database      string
	WriteInterval string
}

type InfluxDbConfig struct {
	Addr          string
	User          string
	Password      string
	Database      string
	WriteInterval time.Duration
}

func GetInfluxDbConfig() (influxDbConfig *InfluxDbConfig, err error) {
	influxDbConfigRead := &InfluxDbConfigRead{
		Addr:          "",
		User:          "",
		Password:      "",
		Database:      "go-mqtt-to-influxdb",
		WriteInterval: "1s",
	}

	// check if mqttClient sections exists
	_, err = config.GetSection("InfluxDbClient")
	if err != nil {
		return nil, errors.New("no InfluxDbClient configuration found")
	}

	err = config.Section("InfluxDbClient").MapTo(influxDbConfigRead)
	if err != nil {
		return nil, fmt.Errorf("cannot read InfluxDbClient configuration: %v", err)
	}

	if len(influxDbConfigRead.Addr) < 1 {
		return nil, errors.New("Addr not specified")
	}

	writeInterval, err := time.ParseDuration(influxDbConfigRead.WriteInterval)
	if err != nil {
		return nil, fmt.Errorf("cannot read writeInterval: %v", err)
	}

	return &InfluxDbConfig{
		Addr:          influxDbConfigRead.Addr,
		User:          influxDbConfigRead.User,
		Password:      influxDbConfigRead.Password,
		Database:      influxDbConfigRead.Database,
		WriteInterval: writeInterval,
	}, nil
}
