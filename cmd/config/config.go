package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/rs/zerolog/log"

	"reflect"
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	commonConfigPathName = "./config.yml"
	pccConfigPathName    = "./pcc.yaml"
	validate             *validator.Validate
	commonConfigV        = viper.New()
	pccConfigV           = viper.New()
	SDFFilterRegex       = regexp.MustCompile(`^permit (out|in) (icmp|ip|tcp|udp|\d+) from (any|[\d.]+|[\da-fA-F:]+)(?:/(\d+))?(?: (\d+|\d+-\d+))? to (assigned|any|[\d.]+|[\da-fA-F:]+)(?:/(\d+))?(?: (\d+|\d+-\d+))?$`)
)

type PCCRulesConfig struct {
	PccRules []PccRule `mapstructure:"pcc_rules" json:"pcc_rules" validate:"dive"`
}

// PccRule describes the structure of the pcc rule representation
type PccRule struct {
	PccName   string `mapstructure:"pcc_name" validate:"required"`
	Notify    bool   `mapstructure:"notify"`
	SdfFilter string `mapstructure:"sdf_filter" validate:"required,sdfFilter"`
	Far       Far    `mapstructure:"far"`
	Qer       Qer    `mapstructure:"qer"`
	Urr       Urr    `mapstructure:"urr"`
}

type Urr struct {
	Urrid uint32 `mapstructure:"urrid" default:"0"`
}

// Far Forwarding Action Rule in the pcc config view
type Far struct {
	Action                uint8  `mapstructure:"action"`
	OuterHeaderCreation   uint8  `mapstructure:"outer_header_creation"`
	Teid                  uint32 `mapstructure:"teid"`
	RemoteIP              uint32 `mapstructure:"remote_ip"`
	TransportLevelMarking uint16 `mapstructure:"transport_level_marking"`
}

// Qer QoS Enforcement Rule in the pcc config view
type Qer struct {
	Qfi          uint8  `mapstructure:"qfi" validate:"min=1,max=255"`
	MaxBitrateUl uint32 `mapstructure:"max_bitrate_ul"`
	MaxBitrateDl uint32 `mapstructure:"max_bitrate_dl"`
}

type UpfConfig struct {
	InterfaceName           []string       `mapstructure:"interface_name" json:"interface_name"`
	XDPAttachMode           string         `mapstructure:"xdp_attach_mode" validate:"oneof=generic native offload" json:"xdp_attach_mode"`
	ApiAddress              string         `mapstructure:"api_address" validate:"hostname_port" json:"api_address"`
	PfcpAddress             string         `mapstructure:"pfcp_address" validate:"hostname_port" json:"pfcp_address"`
	PfcpNodeId              string         `mapstructure:"pfcp_node_id" validate:"hostname|ip" json:"pfcp_node_id"`
	PfcpRemoteNode          []string       `mapstructure:"pfcp_remote_node" validate:"omitempty,dive,hostname|ip" json:"pfcp_node"`
	SxaRemoteNode           []string       `mapstructure:"sxa_remote_node" validate:"omitempty,dive,hostname|ip" json:"sxa_node"`
	SxbRemoteNode           []string       `mapstructure:"sxb_remote_node" validate:"omitempty,dive,hostname|ip" json:"sxb_node"`
	SxaLocalAddress         string         `mapstructure:"sxa_address" validate:"hostname_port" json:"sxa_address"`
	SxbLocalAddress         string         `mapstructure:"sxb_address" validate:"hostname_port" json:"sxb_address"`
	SxaLocalNodeId          string         `mapstructure:"sxa_node_id" validate:"hostname|ip" json:"sxa_node_id"`
	SxbLocalNodeId          string         `mapstructure:"sxb_node_id" validate:"hostname|ip" json:"sxb_node_id"`
	AssociationSetupTimeout uint32         `mapstructure:"association_setup_timeout" json:"association_setup_timeout"`
	MetricsAddress          string         `mapstructure:"metrics_address" validate:"hostname_port" json:"metrics_address"`
	N3Address               string         `mapstructure:"n3_address" validate:"ipv4" json:"n3_address"`
	N9Address               string         `mapstructure:"n9_address" validate:"ipv4" json:"n9_address"`
	S1UAddress              string         `mapstructure:"s1u_address" validate:"ipv4" json:"s1u_address"`
	S5S8Address             string         `mapstructure:"s5s8_address" validate:"ipv4" json:"s5s8_address"`
	PAAddress               string         `mapstructure:"pa_address" validate:"ipv4" json:"pa_address"`
	GtpPeer                 []string       `mapstructure:"gtp_peer" validate:"omitempty,dive,hostname_port" json:"gtp_peer"`
	GtpEchoInterval         uint32         `mapstructure:"gtp_echo_interval" validate:"min=1" json:"gtp_echo_interval"`
	QerMapSize              uint32         `mapstructure:"qer_map_size" validate:"min=1" json:"qer_map_size"`
	FarMapSize              uint32         `mapstructure:"far_map_size" validate:"min=1" json:"far_map_size"`
	UrrMapSize              uint32         `mapstructure:"urr_map_size" validate:"min=1" json:"urr_map_size"`
	PdrMapSize              uint32         `mapstructure:"pdr_map_size" validate:"min=1" json:"pdr_map_size"`
	EbpfMapResize           bool           `mapstructure:"resize_ebpf_maps" json:"resize_ebpf_maps"`
	HeartbeatRetries        uint32         `mapstructure:"heartbeat_retries" json:"heartbeat_retries"`
	HeartbeatInterval       uint32         `mapstructure:"heartbeat_interval" json:"heartbeat_interval"`
	HeartbeatTimeout        uint32         `mapstructure:"heartbeat_timeout" json:"heartbeat_timeout"`
	LoggingLevel            string         `mapstructure:"logging_level" validate:"required" json:"logging_level"`
	LoggingCaller           bool           `mapstructure:"logging_caller" json:"logging_caller"`
	UEIPPool                string         `mapstructure:"ueip_pool" validate:"cidr" json:"ueip_pool"`
	FTEIDPool               uint32         `mapstructure:"teid_pool" json:"teid_pool"`
	FeatureUEIP             bool           `mapstructure:"feature_ueip" json:"feature_ueip"`
	FeatureFTUP             bool           `mapstructure:"feature_ftup" json:"feature_ftup"`
	Qci2DscpMapping         map[string]int `mapstructure:"qci_dscp_mapping" json:"qci_dscp_mapping"`
	AllowedApns             string         `mapstructure:"allowed_apns" json:"allowed_apns"`
	DeniedApns              string         `mapstructure:"denied_apns" json:"denied_apns"`
	HuaweiSupport           bool           `mapstructure:"huawei_support" json:"huawei_support"`
	TraceAssociation        bool           `mapstructure:"trace_association" json:"trace_association"`
	TraceHeartbeat          bool           `mapstructure:"trace_heartbeat" json:"trace_heartbeat"`
	TraceBlocked            bool           `mapstructure:"trace_blocked" json:"trace_blocked"`
	TraceMaxDumpFiles       int            `mapstructure:"trace_files" json:"trace_files"`
	TraceMaxDumpSize        int            `mapstructure:"trace_max_size" json:"trace_max_size"`
	TraceMaxDumpPackets     int            `mapstructure:"trace_max_packets" json:"trace_max_packets"`
	IP6RaSupport            bool           `mapstructure:"ip6_ra_support" json:"ip6_ra_support"`
	Ip6RaPrefix             string         `mapstructure:"ip6_ra_prefix" validate:"cidrv6" json:"ip6_ra_prefix"`
}

func initialize() {
	defineFlags()
	initValidator()
	initCommonConfig()
	initPccConfig()
}

func defineFlags() {
	// PCC flags
	pflag.String("pcc-config", pccConfigPathName, "Path to PCC config file")

	// Common flags
	pflag.String("config", commonConfigPathName, "Path to config file")
	// pflags defaults are ignored in this setup
	pflag.StringArray("iface", []string{}, "Interface list to bind XDP program to")
	pflag.String("attach", "generic", "XDP attach mode")
	pflag.String("aaddr", ":8080", "Address to bind api server to")
	pflag.String("paddr", "127.0.0.1:8805", "Address to bind PFCP server to")
	pflag.String("nodeid", "127.0.0.1", "PFCP Server Node ID")
	pflag.String("maddr", ":9090", "Address to bind metrics server to")
	pflag.String("n3addr", "127.0.0.1", "Address for communication over N3 interface")
	pflag.String("n9addr", "n3addr", "Address for communication over N9 interface")
	pflag.String("s1uaddr", "127.0.0.1", "Address for communication over S1-U interface")
	pflag.String("s5s8addr", "127.0.0.1", "Address for communication over S5/S8 interface")
	pflag.String("paaddr", "127.0.0.1", "Address for communication over PA interface")
	pflag.StringArray("peer", []string{}, "Address of GTP peer")
	pflag.Uint32("echo", 10, "Interval of sending echo requests in seconds")
	pflag.Uint32("qersize", 1024, "Size of the QER ebpf map")
	pflag.Uint32("farsize", 1024, "Size of the FAR ebpf map")
	pflag.Uint32("urrsize", 1024, "Size of the URR ebpf map")
	pflag.Uint32("pdrsize", 1024, "Size of the PDR ebpf map")
	pflag.Bool("mapresize", false, "Enable or disable ebpf map resizing")
	pflag.Uint32("hbretries", 3, "Number of heartbeat retries")
	pflag.Uint32("hbinterval", 5, "Heartbeat interval in seconds")
	pflag.Uint32("hbtimeout", 5, "Heartbeat timeout in seconds")
	pflag.String("loglvl", "info", "Logging level")
	pflag.Bool("logcaller", false, "Enable or disable logging caller")
	pflag.Bool("ueip", false, "Enable or disable UEIP feature")
	pflag.Bool("ftup", false, "Enable or disable FTUP feature")
	pflag.String("ueippool", "10.60.0.0/24", "IP pool for UEIP feature")
	pflag.Uint32("teidpool", 65535, "TEID pool for FTUP feature")
	pflag.StringArray("pfcprnode", []string{}, "Address of remote PFCP node")
	pflag.StringArray("sxanode", []string{}, "Address of remote Sxa node")
	pflag.StringArray("sxbnode", []string{}, "Address of remote Sxb node")
	pflag.String("sxaaddr", "127.0.0.2:8805", "Sxa Address to bind PFCP server to")
	pflag.String("sxbaddr", "127.0.0.3:8805", "Sxb Address to bind PFCP server to")
	pflag.String("sxanodeid", "127.0.0.2", "Sxa Server Node ID")
	pflag.String("sxbnodeid", "127.0.0.3", "Sxb Server Node ID")
	pflag.Uint32("astimeout", 5, "Association setup timeout in seconds")
	pflag.StringToInt("qdmap", map[string]int{}, "QCI to DSCP binding")
	pflag.String("aapns", ".*", "Allowed APNs mask")
	pflag.String("dapns", "", "Denied APNs mask")
	pflag.Bool("huasupp", true, "Enable or disable huawei support")
	pflag.Bool("traceassoc", true, "Trace PFCP Association messages (Establish/Modify/Release)")
	pflag.Bool("tracehb", false, "Trace PFCP Heartbeat messages")
	pflag.Bool("traceblock", true, "Trace dropped dataplane packets")
	pflag.Int("tracefcnt", 10, "Maximum number of rotated trace dump files")
	pflag.Int("tracefsize", 10*1024*1024, "Maximum size (in bytes) of one trace dump file")
	pflag.Int("tracefpackets", 100000, "Maximum number of packets in one trace dump file")
	pflag.Bool("ip6ra", false, "Enable or disable IPv6 Router Advertisement support")
	pflag.String("ip6rapref", "2a03:d000:29a0:509::/64", "Subscriber IPv6 prefix")

	pflag.Parse()
}

func initPccConfig() {
	configPath := pflag.Lookup("pcc-config").Value.String()

	pccConfigV.SetDefault("pcc_rules", []PccRule{})

	pccConfigV.SetConfigFile(configPath)
	pccConfigV.SetEnvPrefix("pcc")
	pccConfigV.AutomaticEnv()

	if err := pccConfigV.ReadInConfig(); err != nil {
		if errors.As(err, &viper.ConfigFileNotFoundError{}) || errors.Is(err, os.ErrNotExist) {
			log.Print("PCC config file not found. Using defaults")
		} else {
			log.Printf("Unable to read PCC config file: %v", err)
		}
	}

	log.Printf("Startup PCC config: %+v", pccConfigV.AllSettings())
}

func initCommonConfig() {
	configPath := pflag.Lookup("config").Value.String()
	// Bind flag errors only when flag is nil, and we ignore empty cli args
	_ = commonConfigV.BindPFlag("interface_name", pflag.Lookup("iface"))
	_ = commonConfigV.BindPFlag("xdp_attach_mode", pflag.Lookup("attach"))
	_ = commonConfigV.BindPFlag("api_address", pflag.Lookup("aaddr"))
	_ = commonConfigV.BindPFlag("pfcp_address", pflag.Lookup("paddr"))
	_ = commonConfigV.BindPFlag("pfcp_node_id", pflag.Lookup("nodeid"))
	_ = commonConfigV.BindPFlag("pfcp_remote_node", pflag.Lookup("pfcprnode"))
	_ = commonConfigV.BindPFlag("sxa_remote_node", pflag.Lookup("sxanode"))
	_ = commonConfigV.BindPFlag("sxb_remote_node", pflag.Lookup("sxbnode"))
	_ = commonConfigV.BindPFlag("sxa_address", pflag.Lookup("sxaaddr"))
	_ = commonConfigV.BindPFlag("sxb_address", pflag.Lookup("sxbaddr"))
	_ = commonConfigV.BindPFlag("sxa_node_id", pflag.Lookup("sxanodeid"))
	_ = commonConfigV.BindPFlag("sxb_node_id", pflag.Lookup("sxbnodeid"))
	_ = commonConfigV.BindPFlag("association_setup_timeout", pflag.Lookup("astimeout"))
	_ = commonConfigV.BindPFlag("metrics_address", pflag.Lookup("maddr"))
	_ = commonConfigV.BindPFlag("n3_address", pflag.Lookup("n3addr"))
	_ = commonConfigV.BindPFlag("n9_address", pflag.Lookup("n9addr"))
	_ = commonConfigV.BindPFlag("s1u_address", pflag.Lookup("s1uaddr"))
	_ = commonConfigV.BindPFlag("s5s8_address", pflag.Lookup("s5s8addr"))
	_ = commonConfigV.BindPFlag("pa_address", pflag.Lookup("paaddr"))
	_ = commonConfigV.BindPFlag("gtp_peer", pflag.Lookup("peer"))
	_ = commonConfigV.BindPFlag("gtp_echo_interval", pflag.Lookup("echo"))
	_ = commonConfigV.BindPFlag("qer_map_size", pflag.Lookup("qersize"))
	_ = commonConfigV.BindPFlag("far_map_size", pflag.Lookup("farsize"))
	_ = commonConfigV.BindPFlag("urr_map_size", pflag.Lookup("urrsize"))
	_ = commonConfigV.BindPFlag("pdr_map_size", pflag.Lookup("pdrsize"))
	_ = commonConfigV.BindPFlag("resize_ebpf_maps", pflag.Lookup("mapresize"))
	_ = commonConfigV.BindPFlag("heartbeat_retries", pflag.Lookup("hbretries"))
	_ = commonConfigV.BindPFlag("heartbeat_interval", pflag.Lookup("hbinterval"))
	_ = commonConfigV.BindPFlag("heartbeat_timeout", pflag.Lookup("hbtimeout"))
	_ = commonConfigV.BindPFlag("logging_level", pflag.Lookup("loglvl"))
	_ = commonConfigV.BindPFlag("logging_caller", pflag.Lookup("logcaller"))
	_ = commonConfigV.BindPFlag("feature_ueip", pflag.Lookup("ueip"))
	_ = commonConfigV.BindPFlag("feature_ftup", pflag.Lookup("ftup"))
	_ = commonConfigV.BindPFlag("ueip_pool", pflag.Lookup("ueippool"))
	_ = commonConfigV.BindPFlag("teid_pool", pflag.Lookup("teidpool"))
	_ = commonConfigV.BindPFlag("qci_dscp_mapping", pflag.Lookup("qdmap"))
	_ = commonConfigV.BindPFlag("allowed_apns", pflag.Lookup("aapns"))
	_ = commonConfigV.BindPFlag("denied_apns", pflag.Lookup("dapns"))
	_ = commonConfigV.BindPFlag("huawei_support", pflag.Lookup("huasupp"))
	_ = commonConfigV.BindPFlag("trace_association_disable", pflag.Lookup("traceassoc"))
	_ = commonConfigV.BindPFlag("trace_heartbeat_disable", pflag.Lookup("tracehb"))
	_ = commonConfigV.BindPFlag("trace_blocked_disable", pflag.Lookup("traceblock"))
	_ = commonConfigV.BindPFlag("trace_files", pflag.Lookup("tracefcnt"))
	_ = commonConfigV.BindPFlag("trace_max_size", pflag.Lookup("tracefsize"))
	_ = commonConfigV.BindPFlag("trace_max_packets", pflag.Lookup("tracefpackets"))
	_ = commonConfigV.BindPFlag("ip6_ra_support", pflag.Lookup("ip6ra"))
	_ = commonConfigV.BindPFlag("ip6_ra_prefix", pflag.Lookup("ip6rapref"))

	commonConfigV.SetDefault("n9_address", commonConfigV.GetString("n3_address"))

	commonConfigV.SetConfigFile(configPath)

	commonConfigV.SetEnvPrefix("upf")
	commonConfigV.AutomaticEnv()

	if err := commonConfigV.ReadInConfig(); err != nil {
		if errors.As(err, &viper.ConfigFileNotFoundError{}) || errors.Is(err, os.ErrNotExist) {
			// Config file not found; ignore error if desired
			log.Print("Config file not found. Using defaults")
		} else {
			// Config file was found but another error was produced
			log.Printf("Unable to read config file: %s commonConfigV. Using defaults", err)
		}
	}

	log.Printf("Startup config: %+v", commonConfigV.AllSettings())
}

func (c *UpfConfig) GetDscpMarkByQci(qci uint8) uint8 {
	dscp := c.Qci2DscpMapping[strconv.Itoa(int(qci))]
	return uint8(dscp)
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
	return commonConfigV.UnmarshalExact(c)
}

func validateSdfFilter(fl validator.FieldLevel) bool {
	if fl.Field().Type() != reflect.TypeOf("") {
		return false
	}

	return SDFFilterRegex.MatchString(fl.Field().String())
}

func initValidator() {
	validate = validator.New()

	err := validate.RegisterValidation("sdfFilter", validateSdfFilter)
	if err != nil {
		log.Error().Msgf("error register sdfFilter validator: %v", err)
	}
}

func (pcr *PCCRulesConfig) Validate() error {
	if err := validate.Struct(pcr); err != nil {
		return err
	}

	return nil
}

// Unmarshal data from config file
func (pcr *PCCRulesConfig) Unmarshal() error {
	return pccConfigV.UnmarshalExact(pcr)
}

func (c *UpfConfig) UpdateFile(params map[string]interface{}) error {
	for key, value := range params {
		commonConfigV.Set(key, value)
	}

	if err := commonConfigV.WriteConfigAs(commonConfigPathName); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	log.Info().Msgf("Config file updated successfully: %s", commonConfigPathName)
	return nil
}
