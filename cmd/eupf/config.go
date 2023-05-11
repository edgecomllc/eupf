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
	QerMapSize     uint32   `mapstructure:"qer_map_size" validate:"min=1"`
	FarMapSize     uint32   `mapstructure:"far_map_size" validate:"min=1"`
	PdrMapSize     uint32   `mapstructure:"pdr_map_size" validate:"min=1"`
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
	// pflags defaults are ignored in this setup
	pflag.StringArray("iface", []string{}, "Interface list to bind XDP program to")
	pflag.String("attach", "", "XDP attach mode")
	pflag.String("aaddr", "", "Address to bind api server to")
	pflag.String("paddr", "", "Address to bind PFCP server to")
	pflag.String("nodeid", "", "PFCP Server Node ID")
	pflag.String("maddr", "", "Address to bind metrics server to")
	pflag.String("n3addr", "", "Address for communication over N3 interface")
	pflag.String("qersize", "", "Size of the QER ebpf map")
	pflag.String("farsize", "", "Size of the FAR ebpf map")
	pflag.String("pdrsize", "", "Size of the PDR ebpf map")
	pflag.Parse()

	// Bind flag errors only when flag is nil, and we ignore empty cli args
	_ = viper.BindPFlag("interface_name", pflag.Lookup("iface"))
	_ = viper.BindPFlag("xdp_attach_mode", pflag.Lookup("attach"))
	_ = viper.BindPFlag("api_address", pflag.Lookup("aaddr"))
	_ = viper.BindPFlag("pfcp_address", pflag.Lookup("paddr"))
	_ = viper.BindPFlag("pfcp_node_id", pflag.Lookup("nodeid"))
	_ = viper.BindPFlag("metrics_address", pflag.Lookup("maddr"))
	_ = viper.BindPFlag("n3_address", pflag.Lookup("n3addr"))
	_ = viper.BindPFlag("qer_map_size", pflag.Lookup("qersize"))
	_ = viper.BindPFlag("far_map_size", pflag.Lookup("farsize"))
	_ = viper.BindPFlag("pdr_map_size", pflag.Lookup("pdrsize"))

	viper.SetDefault("interface_name", "lo")
	viper.SetDefault("xdp_attach_mode", "generic")
	viper.SetDefault("api_address", ":8080")
	viper.SetDefault("pfcp_address", ":8805")
	viper.SetDefault("pfcp_node_id", "127.0.0.1")
	viper.SetDefault("metrics_address", ":9090")
	viper.SetDefault("n3_address", "127.0.0.1")
	viper.SetDefault("qer_map_size", "1024")
	viper.SetDefault("far_map_size", "1024")
	viper.SetDefault("pdr_map_size", "1024")

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
