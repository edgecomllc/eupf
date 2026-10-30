// SPDX-License-Identifier: Apache-2.0
// Copyright 2023-2026 Edgecom LLC

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
