package config

import (
	"log"
)

var Conf UpfConfig
var PCCConf PCCRulesConfig

// Init init config for eupf package
func Init() {
	initialize()

	if err := Conf.Unmarshal(); err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}

	if err := Conf.Validate(); err != nil {
		log.Fatalf("eUPF config is invalid: %v", err)
	}

	if err := PCCConf.Unmarshal(); err != nil {
		log.Fatalf("PCC config is invalid: %v", err)
	}

	if err := PCCConf.Validate(); err != nil {
		log.Fatalf("PCC config is invalid: %v", err)
	}

	log.Printf("Apply eUPF config: %+v", Conf)
}
