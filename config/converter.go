package config

import (
	"fmt"
	"log"
	"strings"
)

type ConvertConfig struct {
	Name           string
	Implementation string
	MqttTopic      string
}

const convertPrefix = "Convert."

func GetConvertConfig(sectionName string) (convertConfig *ConvertConfig, err error) {
	convertConfig = &ConvertConfig{
		Name:           sectionName[len(convertPrefix):],
		Implementation: "",
		MqttTopic:      "",
	}

	_, err = config.GetSection(sectionName)
	if err != nil {
		return nil, fmt.Errorf("no %s configuration found", sectionName)
	}

	err = config.Section(sectionName).MapTo(convertConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s configuration: %v", sectionName, err)
	}

	// replace * by # in MqttTopic names
	convertConfig.MqttTopic = strings.Replace(convertConfig.MqttTopic, "*", "#",16)

	return
}

func GetConvertConfigs() (convertConfigs []*ConvertConfig) {
	sections := config.SectionStrings()
	for _, sectionName := range sections {
		if !strings.HasPrefix(sectionName, convertPrefix) {
			continue
		}
		if config, err := GetConvertConfig(sectionName); err == nil {
			convertConfigs = append(convertConfigs, config)
		} else {
			log.Printf("converterConfig: cannot add section %s; err=%s", sectionName, err.Error())
		}
	}

	return
}
