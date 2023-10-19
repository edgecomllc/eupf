package config

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"log"
)

type Config struct {
	InterfaceName     []string `mapstructure:"interface_name" json:"interface_name"`
	XDPAttachMode     string   `mapstructure:"xdp_attach_mode" validate:"oneof=generic native offload" json:"xdp_attach_mode"`
	ApiAddress        string   `mapstructure:"api_address" validate:"hostname_port" json:"api_address"`
	PfcpAddress       string   `mapstructure:"pfcp_address" validate:"hostname_port" json:"pfcp_address"`
	PfcpNodeId        string   `mapstructure:"pfcp_node_id" validate:"hostname|ip" json:"pfcp_node_id"`
	MetricsAddress    string   `mapstructure:"metrics_address" validate:"hostname_port" json:"metrics_address"`
	N3Address         string   `mapstructure:"n3_address" validate:"ipv4" json:"n3_address"`
	QerMapSize        uint32   `mapstructure:"qer_map_size" validate:"min=1" json:"qer_map_size"`
	FarMapSize        uint32   `mapstructure:"far_map_size" validate:"min=1" json:"far_map_size"`
	PdrMapSize        uint32   `mapstructure:"pdr_map_size" validate:"min=1" json:"pdr_map_size"`
	EbpfMapResize     bool     `mapstructure:"resize_ebpf_maps" json:"resize_ebpf_maps"`
	HeartbeatRetries  uint32   `mapstructure:"heartbeat_retries" json:"heartbeat_retries"`
	HeartbeatInterval uint32   `mapstructure:"heartbeat_interval" json:"heartbeat_interval"`
	HeartbeatTimeout  uint32   `mapstructure:"heartbeat_timeout" json:"heartbeat_timeout"`
	LoggingLevel      string   `mapstructure:"logging_level" validate:"required" json:"logging_level"`
}

func New() (*Config, error) {
	var v = viper.GetViper()
	var cfg *Config
	var err error

	if err = v.UnmarshalExact(cfg); err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to decode into struct, %v", err))
	}
	if err = validator.New().Struct(cfg); err != nil {
		return nil, errors.New(fmt.Sprintf("eUPF config is invalid: %v", err))
	}

	log.Printf("Apply eUPF config: %+v", cfg)

	return cfg, nil
}
