package core

import (
	"fmt"
	"os"

	"github.com/edgecomllc/eupf/cmd/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func InitLogger() {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006/01/02 15:04:05"}
	log.Logger = zerolog.New(output).With().Timestamp().Logger()
}

func SetLoggerLevel(loggingLevel string) error {
	if loggingLevel == "" {
		return fmt.Errorf("logging level can't be empty")
	}
	if loglvl, err := zerolog.ParseLevel(loggingLevel); err == nil {
		zerolog.SetGlobalLevel(loglvl)
		config.Conf.LoggingLevel = zerolog.GlobalLevel().String()
	} else {
		return fmt.Errorf("can't parse logging level: '%s'", loggingLevel)
	}
	return nil
}
