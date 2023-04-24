package main

import (
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type UpfConfig struct {
	InterfaceName  []string `mapstructure:"interface_name"`
	XDPAttachMode  string   `mapstructure:"xdp_attach_mode" validate:"oneof=generic native offload"`
	ApiAddress     string   `mapstructure:"api_address" validate:"hostname_port"`
	PfcpAddress    string   `mapstructure:"pfcp_address" validate:"hostname_port"`
	PfcpNodeId     string   `mapstructure:"pfcp_node_id" validate:"ipv4"`
	MetricsAddress string   `mapstructure:"metrics_address" validate:"hostname_port"`
	N3Address      string   `mapstructure:"n3_address" validate:"ipv4"`
}

func (c *UpfConfig) Validate() error {
	if err := validator.New().Struct(c); err != nil {
		return err
	}

	return nil
}

var config UpfConfig

func LoadConfig() error {
	var configPath = pflag.String("config", "./config.yml", "Path to config file")
	pflag.StringArray("iface", []string{"lo"}, "Interface list to bind XDP program to")
	pflag.String("attach", "generic", "XDP attach mode")
	pflag.String("aaddr", ":8080", "Address to bind api server to")
	pflag.String("paddr", ":8805", "Address to bind PFCP server to")
	pflag.String("nodeid", "localhost", "PFCP Server Node ID")
	pflag.String("maddr", ":9090", "Address to bind metrics server to")
	pflag.String("n3addr", "127.0.0.1", "Address for communication over N3 interface")
	pflag.Parse()

	viper.BindPFlag("interface_name", pflag.Lookup("iface"))
	viper.BindPFlag("xdp_attach_mode", pflag.Lookup("attach"))
	viper.BindPFlag("api_address", pflag.Lookup("aaddr"))
	viper.BindPFlag("pfcp_address", pflag.Lookup("paddr"))
	viper.BindPFlag("pfcp_node_id", pflag.Lookup("nodeid"))
	viper.BindPFlag("metrics_address", pflag.Lookup("maddr"))
	viper.BindPFlag("n3_address", pflag.Lookup("n3addr"))

	viper.SetDefault("interface_name", "lo")
	viper.SetDefault("xdp_attach_mode", "generic")
	viper.SetDefault("api_address", ":8080")
	viper.SetDefault("pfcp_address", ":8805")
	viper.SetDefault("pfcp_node_id", "upf.edgecom.ru")
	viper.SetDefault("metrics_address", ":9090")
	viper.SetDefault("n3_address", "127.0.0.1")

	viper.SetConfigFile(*configPath)

	viper.SetEnvPrefix("upf")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Unable to read config file, %v", err)
	}

	log.Printf("Get raw config: %+v", viper.AllSettings())
	if err := viper.UnmarshalExact(&config); err != nil {
		log.Printf("Unable to decode into struct, %v", err)
		return err
	}

	if err := config.Validate(); err != nil {
		log.Printf("eUPF config is invalid: %v", err)
		return err
	}

	log.Printf("Apply eUPF config: %+v", config)
	return nil
}
