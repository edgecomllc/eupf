package config

import (
	"fmt"
	"github.com/spf13/viper"
)

const configFileName = "config-cli.yaml"

type Config struct {
	LogLevel string `mapstructure:"log_level" yaml:"log_level"`
	EupfAddr string `mapstructure:"eupf_addr" yaml:"eupf_addr"`
}

var cfgViper = viper.New()

func NewConfig() (*Config, error) {
	cfgViper.SetDefault("log_level", "debug")
	cfgViper.SetDefault("eupf_addr", "http://127.0.0.1/api/vi/")

	cfgViper.SetConfigFile(configFileName)
	cfgViper.SetConfigType("yaml")

	_ = cfgViper.ReadInConfig()

	var cfg Config
	if err := cfgViper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
