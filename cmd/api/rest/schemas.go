package rest

type LoggingLevelConfig struct {
	LoggingLevel string `json:"logging_level" binding:"required"`
}

type LoggingCallerConfig struct {
	LoggingCaller bool `json:"logging_caller"`
}

type DataPlaneEbpfConfig struct {
	InterfaceName []string `json:"interface_name" binding:"required"`
	XDPAttachMode string   `json:"xdp_attach_mode" binding:"required,oneof=generic native offload"`
}

type DataPlaneAddressesConfig struct {
	N3Address string `json:"n3_address" binding:"required,ipv4"`
	N9Address string `json:"n9_address" binding:"required,ipv4"`
}

type PFCPN4Config struct {
	PFCPAddress    string   `json:"pfcp_address" binding:"required"`
	PFCPNodeID     string   `json:"pfcp_node_id" binding:"required"`
	PFCPRemoteNode []string `json:"pfcp_remote_node" binding:"required"`
}

type PFCPSxaConfig struct {
	SXAAddress    string   `json:"sxa_address" binding:"required"`
	SXANodeID     string   `json:"sxa_node_id" binding:"required"`
	SXARemoteNode []string `json:"sxa_remote_node" binding:"required"`
}

type PFCPSxbConfig struct {
	SXBAddress    string   `json:"sxb_address" binding:"required"`
	SXBNodeID     string   `json:"sxb_node_id" binding:"required"`
	SXBRemoteNode []string `json:"sxb_remote_node" binding:"required"`
}

type PFCPTimersConfig struct {
	AssociationSetupTimeout uint32 `json:"association_setup_timeout" binding:"required,gte=0"`
	HeartbeatTimeout        uint32 `json:"heartbeat_timeout" binding:"required,gte=0"`
}

type GTPPathConfig struct {
	GtpPeer         []string `json:"gtp_peer" binding:"required"`
	GtpEchoInterval uint32   `json:"gtp_echo_interval" binding:"required,gte=1"`
}
