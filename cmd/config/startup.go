package config

import (
	"log"
)

var Conf UpfConfig

// Init init config for eupf package
func Init() {
	if err := Conf.Unmarshal(); err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}

	if err := Conf.Validate(); err != nil {
		log.Fatalf("eUPF config is invalid: %v", err)
	}

	log.Printf("Apply eUPF config: %+v", Conf)
}

func GetUpdatedConf() (UpfConfig, error) {
	var updatedConf UpfConfig
	if err := updatedConf.Unmarshal(); err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
		return UpfConfig{}, err
	}

	if err := updatedConf.Validate(); err != nil {
		log.Fatalf("eUPF config is invalid: %v", err)
		return UpfConfig{}, err
	}

	log.Printf("Apply updated eUPF config: %+v", updatedConf)
	return updatedConf, nil
}
