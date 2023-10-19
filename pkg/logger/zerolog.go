package logger

import (
	"fmt"
	"github.com/rs/zerolog"
	"log"
	"os"
)

type ZeroLogger struct {
	log *zerolog.Logger
}

func NewZeroLogger(loggingLevel string) *ZeroLogger {
	if loggingLevel == "" {
		log.Fatal("Logging level can't be empty")
	}
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006/01/02 15:04:05"}
	logger := zerolog.New(output).With().Timestamp().Logger()
	level, err := zerolog.ParseLevel(loggingLevel)
	if err != nil {
		log.Fatal(err)
	}
	logger.Level(level)

	return &ZeroLogger{log: &logger}
}

func (z *ZeroLogger) SetLoggerLevel(level string) error {
	if level == "" {
		return fmt.Errorf("Logging level can't be empty")
	}
	if loggingLevel, err := zerolog.ParseLevel(level); err == nil {
		zerolog.SetGlobalLevel(loggingLevel)
	} else {
		return fmt.Errorf("Can't parse logging level: '%s'", loggingLevel)
	}
	return nil
}

func (z *ZeroLogger) Fatal(msg string) {
	z.log.Fatal().Msg(msg)
}

func (z *ZeroLogger) Fatalf(format string, v ...interface{}) {
	z.log.Fatal().Msgf(format, v...)
}

func (z *ZeroLogger) Fatale(err error) {
	z.log.Fatal().Err(err)
}

func (z *ZeroLogger) Error(err error) {
	z.log.Err(err)
}

func (z *ZeroLogger) Info(msg string) {
	z.log.Info().Msg(msg)
}

func (z *ZeroLogger) Infof(format string, v ...interface{}) {
	z.log.Info().Msgf(format, v...)
}

func (z *ZeroLogger) Panicf(format string, v ...interface{}) {
	z.log.Panic().Msgf(format, v...)
}

func (z *ZeroLogger) Printf(format string, v ...interface{}) {
	z.log.Printf(format, v)
}

//Methods on (*Logger):
//UpdateContext(update func(c zerolog.Context) zerolog.Context)
//Trace() *zerolog.Event
//Debug() *zerolog.Event
//Info() *zerolog.Event
//Warn() *zerolog.Event
//Error() *zerolog.Event
//Err(err error) *zerolog.Event
//Fatal() *zerolog.Event
//Panic() *zerolog.Event
//WithLevel(level zerolog.Level) *zerolog.Event
//Log() *zerolog.Event
//Print(v ...interface{})
//Printf(format string, v ...interface{})
//newEvent(level zerolog.Level, done func(string)) *zerolog.Event
//should(lvl zerolog.Level) bool
