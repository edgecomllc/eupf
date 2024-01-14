package config

import (
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var v = viper.GetViper()

type UpfConfig struct {
	InterfaceName     []string `mapstructure:"interface_name" json:"interface_name"`
	XDPAttachMode     string   `mapstructure:"xdp_attach_mode" validate:"oneof=generic native offload" json:"xdp_attach_mode"`
	ApiAddress        string   `mapstructure:"api_address" validate:"hostname_port" json:"api_address"`
	PfcpAddress       string   `mapstructure:"pfcp_address" validate:"hostname_port" json:"pfcp_address"`
	PfcpNodeId        string   `mapstructure:"pfcp_node_id" validate:"hostname|ip" json:"pfcp_node_id"`
	MetricsAddress    string   `mapstructure:"metrics_address" validate:"hostname_port" json:"metrics_address"`
	N3Address         string   `mapstructure:"n3_address" validate:"ipv4" json:"n3_address"`
	GtpPeer           []string `mapstructure:"gtp_peer" validate:"omitempty,dive,hostname_port" json:"gtp_peer"`
	EchoInterval      uint32   `mapstructure:"echo_interval" validate:"min=1" json:"echo_interval"`
	QerMapSize        uint32   `mapstructure:"qer_map_size" validate:"min=1" json:"qer_map_size"`
	FarMapSize        uint32   `mapstructure:"far_map_size" validate:"min=1" json:"far_map_size"`
	PdrMapSize        uint32   `mapstructure:"pdr_map_size" validate:"min=1" json:"pdr_map_size"`
	EbpfMapResize     bool     `mapstructure:"resize_ebpf_maps" json:"resize_ebpf_maps"`
	HeartbeatRetries  uint32   `mapstructure:"heartbeat_retries" json:"heartbeat_retries"`
	HeartbeatInterval uint32   `mapstructure:"heartbeat_interval" json:"heartbeat_interval"`
	HeartbeatTimeout  uint32   `mapstructure:"heartbeat_timeout" json:"heartbeat_timeout"`
	LoggingLevel      string   `mapstructure:"logging_level" validate:"required" json:"logging_level"`
	UEIPPool          string   `mapstructure:"ueip_pool" validate:"cidr" json:"ueip_pool"`
	FTEIDPool         uint32   `mapstructure:"teid_pool" json:"teid_pool"`
	FeatureUEIP       bool     `mapstructure:"feature_ueip" json:"feature_ueip"`
	FeatureFTUP       bool     `mapstructure:"feature_ftup" json:"feature_ftup"`
}

func init() {
	var configPath = pflag.String("config", "./config.yml", "Path to config file")
	// pflags defaults are ignored in this setup
	pflag.StringArray("iface", []string{}, "Interface list to bind XDP program to")
	pflag.String("attach", "generic", "XDP attach mode")
	pflag.String("aaddr", ":8080", "Address to bind api server to")
	pflag.String("paddr", "127.0.0.1:8805", "Address to bind PFCP server to")
	pflag.String("nodeid", "127.0.0.1", "PFCP Server Node ID")
	pflag.String("maddr", ":9090", "Address to bind metrics server to")
	pflag.String("n3addr", "127.0.0.1", "Address for communication over N3 interface")
	pflag.StringArray("peer", []string{}, "Address of GTP peer")
	pflag.Uint32("echo", 10, "Interval of sending echo requests in seconds")
	pflag.Uint32("qersize", 1024, "Size of the QER ebpf map")
	pflag.Uint32("farsize", 1024, "Size of the FAR ebpf map")
	pflag.Uint32("pdrsize", 1024, "Size of the PDR ebpf map")
	pflag.Bool("mapresize", false, "Enable or disable ebpf map resizing")
	pflag.Uint32("hbretries", 3, "Number of heartbeat retries")
	pflag.Uint32("hbinterval", 5, "Heartbeat interval in seconds")
	pflag.Uint32("hbtimeout", 5, "Heartbeat timeout in seconds")
	pflag.String("loglvl", "info", "Logging level")
	pflag.Bool("ueip", false, "Enable or disable UEIP feature")
	pflag.Bool("ftup", false, "Enable or disable FTUP feature")
	pflag.String("ueippool", "10.60.0.0/24", "IP pool for UEIP feature")
	pflag.Uint32("teidpool", 65535, "TEID pool for FTUP feature")
	pflag.Parse()

	// Bind flag errors only when flag is nil, and we ignore empty cli args
	_ = v.BindPFlag("interface_name", pflag.Lookup("iface"))
	_ = v.BindPFlag("xdp_attach_mode", pflag.Lookup("attach"))
	_ = v.BindPFlag("api_address", pflag.Lookup("aaddr"))
	_ = v.BindPFlag("pfcp_address", pflag.Lookup("paddr"))
	_ = v.BindPFlag("pfcp_node_id", pflag.Lookup("nodeid"))
	_ = v.BindPFlag("metrics_address", pflag.Lookup("maddr"))
	_ = v.BindPFlag("n3_address", pflag.Lookup("n3addr"))
	_ = v.BindPFlag("gtp_peer", pflag.Lookup("peer"))
	_ = v.BindPFlag("echo_interval", pflag.Lookup("echo"))
	_ = v.BindPFlag("qer_map_size", pflag.Lookup("qersize"))
	_ = v.BindPFlag("far_map_size", pflag.Lookup("farsize"))
	_ = v.BindPFlag("pdr_map_size", pflag.Lookup("pdrsize"))
	_ = v.BindPFlag("resize_ebpf_maps", pflag.Lookup("mapresize"))
	_ = v.BindPFlag("heartbeat_retries", pflag.Lookup("hbretries"))
	_ = v.BindPFlag("heartbeat_interval", pflag.Lookup("hbinterval"))
	_ = v.BindPFlag("heartbeat_timeout", pflag.Lookup("hbtimeout"))
	_ = v.BindPFlag("logging_level", pflag.Lookup("loglvl"))
	_ = v.BindPFlag("feature_ueip", pflag.Lookup("ueip"))
	_ = v.BindPFlag("feature_ftup", pflag.Lookup("ftup"))
	_ = v.BindPFlag("ueip_pool", pflag.Lookup("ueippool"))
	_ = v.BindPFlag("teid_pool", pflag.Lookup("teidpool"))

	v.SetDefault("interface_name", "lo")
	v.SetDefault("xdp_attach_mode", "generic")
	v.SetDefault("api_address", ":8080")
	v.SetDefault("pfcp_address", "127.0.0.1:8805")
	v.SetDefault("pfcp_node_id", "127.0.0.1")
	v.SetDefault("metrics_address", ":9090")
	v.SetDefault("n3_address", "127.0.0.1")
	v.SetDefault("echo_interval", 10)
	v.SetDefault("qer_map_size", 1024)
	v.SetDefault("far_map_size", 1024)
	v.SetDefault("pdr_map_size", 1024)
	v.SetDefault("resize_ebpf_maps", false)
	v.SetDefault("heartbeat_retries", 3)
	v.SetDefault("heartbeat_interval", 5)
	v.SetDefault("heartbeat_timeout", 5)
	v.SetDefault("logging_level", "info")
	v.SetDefault("feature_ueip", false)
	v.SetDefault("feature_ftup", false)
	v.SetDefault("ueip_pool", "10.60.0.0/24")
	v.SetDefault("teid_pool", 65535)

	v.SetConfigFile(*configPath)

	v.SetEnvPrefix("upf")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			log.Print("Config file not found. Using defaults")
		} else {
			// Config file was found but another error was produced
			log.Printf("Unable to read config file: %v. Using defaults", err)
		}
	}

	log.Printf("Startup config: %+v", v.AllSettings())
}

func (c *UpfConfig) Validate() error {
	if err := validator.New().Struct(c); err != nil {
		return err
	}

	if !c.FeatureFTUP {
		c.FTEIDPool = 0
	}

	if !c.FeatureUEIP {
		c.UEIPPool = ""
	}

	return nil
}

// Unmarshal data from config file
func (c *UpfConfig) Unmarshal() error {
	return v.UnmarshalExact(c)
}
