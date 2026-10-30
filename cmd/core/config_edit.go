// SPDX-License-Identifier: Apache-2.0
// Copyright 2023-2026 Edgecom LLC

package core

import (
	"fmt"

	"github.com/edgecomllc/eupf/cmd/config"

	"github.com/rs/zerolog"
)

func SetConfig(conf config.UpfConfig) error {
	// For now only logging_level parameter update in config.UpfConfig is supported.
	if err := SetLoggerLevel(conf.LoggingLevel); err != nil {
		return fmt.Errorf("logger configuring error: %s. Using '%s' level",
			err.Error(), zerolog.GlobalLevel().String())
	}
	return nil
}
