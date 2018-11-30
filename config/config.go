package config

import (
"github.com/go-ini/ini"
"log"
"path/filepath"
)

var config *ini.File
var configDir string

func Setup(source string) {
	configDir = filepath.Dir(source) + "/"

	log.Printf("config: load configuration source=%v, configDir=%v", source, configDir)

	var err error
	config, err = ini.Load(source)
	if err != nil {
		log.Fatalf("config: cannot load configuration: %v", err)
	}
}

