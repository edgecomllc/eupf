package main

import (
	"fmt"
	"github.com/edgecomllc/eupf/cli/config"
	"github.com/edgecomllc/eupf/cli/eupf/injection"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
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
	} else {
		return fmt.Errorf("can't parse logging level: '%s'", loggingLevel)
	}

	return nil
}

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading config")
	}

	InitLogger()

	err = SetLoggerLevel(cfg.LogLevel)
	if err != nil {
		log.Fatal().Err(err).Msg("Error setting log level")
	}

	root := injection.InitEupfCLI(cfg.EupfAddr, cfg)
	err = root.Execute()
	if err != nil {
		log.Fatal().Err(err).Msg("Error executing command")
	}
}
