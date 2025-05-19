package config

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
)

const configFileName = "config-cli.yaml"

var (
	cfgViper  = viper.New()
	validate  *validator.Validate
	flagGroup = pflag.NewFlagSet("cli-flags", pflag.ContinueOnError)
)

// Config defines CLI-related config
type Config struct {
	LogLevel   string `mapstructure:"log_level" validate:"omitempty,oneof=debug info warn error"`
	EupfAddr   string `mapstructure:"eupf_addr" validate:"omitempty,url"`
	BackupPath string `mapstructure:"backup_path" validate:"required"`
}

// NewConfig initializes config: flags, env, file
func NewConfig() (*Config, error) {
	defineFlags()

	cfgViper.SetDefault("log_level", "debug")
	cfgViper.SetDefault("backup_path", "/var/lib/eupf/backups")
	cfgViper.SetDefault("eupf_addr", "http://127.0.0.1:8081/api/v1")

	cfgViper.SetConfigFile(configFileName)
	cfgViper.SetConfigType("yaml")
	cfgViper.SetEnvPrefix("cli")
	cfgViper.AutomaticEnv()

	// Bind flags
	_ = cfgViper.BindPFlags(flagGroup)

	if err := cfgViper.ReadInConfig(); err != nil {
		if errors.As(err, &viper.ConfigFileNotFoundError{}) || errors.Is(err, os.ErrNotExist) {
			log.Warn().Msg("CLI config file not found. Using defaults and flags")
		} else {
			log.Error().Err(err).Msg("Failed to read CLI config file")
		}
	}

	var cfg Config
	if err := cfgViper.UnmarshalExact(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal CLI config: %w", err)
	}

	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	log.Info().Msgf("Loaded CLI config: %+v", cfg)

	return &cfg, nil
}

// UpdateFile updates config file values based on provided map
func (c *Config) UpdateFile(params map[string]interface{}) error {
	for key, value := range params {
		cfgViper.Set(key, value)
	}

	if err := cfgViper.WriteConfigAs(configFileName); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	if err := cfgViper.UnmarshalExact(c); err != nil {
		return fmt.Errorf("failed to sync config struct after write: %w", err)
	}

	log.Info().Msgf("CLI config file updated successfully: %s", configFileName)

	return nil
}

func defineFlags() {
	flagGroup.String("log_level", "", "Log level (debug, info, warn, error)")
	flagGroup.String("eupf_addr", "", "Base URL for eupf API (e.g. http://localhost:8081/api/v1/)")
	flagGroup.String("backup_path", "", "Directory to store backups")

	_ = flagGroup.Parse(os.Args[1:])
}

func validateConfig(cfg *Config) error {
	validate = validator.New()
	if err := validate.Struct(cfg); err != nil {
		return err
	}

	return nil
}
