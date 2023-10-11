package core

import (
	"fmt"

	"github.com/edgecomllc/eupf/cmd/config"

	"github.com/rs/zerolog"
)

func SetConfig(conf config.UpfConfig) error {
	// For now only logging_level parameter update in config.UpfConfig is supported.
	if err := SetLoggerLevel(conf.LoggingLevel); err != nil {
		return fmt.Errorf("Logger configuring error: %s. Using '%s' level",
			err.Error(), zerolog.GlobalLevel().String())
	}
	return nil
}
