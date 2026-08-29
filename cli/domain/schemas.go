package domain

import (
	"time"
)

type TraceRecord struct {
	Imsi      string    `json:"imsi"`
	Msisdn    string    `json:"msisdn"`
	Timestamp time.Time `json:"timestamp"`
	ConnAddr  string    `json:"connAddr"`
}

type BackupRecord struct {
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
}

type UpfConfig struct {
	InterfaceName           []string       `json:"interface_name"`
	XDPAttachMode           string         `json:"xdp_attach_mode"`
	ApiAddress              string         `json:"api_address"`
	PfcpAddress             string         `json:"pfcp_address"`
	PfcpNodeId              string         `json:"pfcp_node_id"`
	PfcpRemoteNode          []string       `json:"pfcp_node"`
	SxaRemoteNode           []string       `json:"sxa_node"`
	SxbRemoteNode           []string       `json:"sxb_node"`
	SxaLocalAddress         string         `json:"sxa_address"`
	SxbLocalAddress         string         `json:"sxb_address"`
	SxaLocalNodeId          string         `json:"sxa_node_id"`
	SxbLocalNodeId          string         `json:"sxb_node_id"`
	AssociationSetupTimeout uint32         `json:"association_setup_timeout"`
	MetricsAddress          string         `json:"metrics_address"`
	N3Address               string         `json:"n3_address"`
	N9Address               string         `json:"n9_address"`
	S1UAddress              string         `json:"s1u_address"`
	S5S8Address             string         `json:"s5s8_address"`
	PAAddress               string         `json:"pa_address"`
	GtpPeer                 []string       `json:"gtp_peer"`
	GtpEchoInterval         uint32         `json:"gtp_echo_interval"`
	QerMapSize              uint32         `json:"qer_map_size"`
	FarMapSize              uint32         `json:"far_map_size"`
	UrrMapSize              uint32         `json:"urr_map_size"`
	PdrMapSize              uint32         `json:"pdr_map_size"`
	EbpfMapResize           bool           `json:"resize_ebpf_maps"`
	HeartbeatRetries        uint32         `json:"heartbeat_retries"`
	HeartbeatInterval       uint32         `json:"heartbeat_interval"`
	HeartbeatTimeout        uint32         `json:"heartbeat_timeout"`
	LoggingLevel            string         `json:"logging_level"`
	LoggingCaller           bool           `json:"logging_caller"`
	UEIPPool                string         `json:"ueip_pool"`
	FTEIDPool               uint32         `json:"teid_pool"`
	FeatureUEIP             bool           `json:"feature_ueip"`
	FeatureFTUP             bool           `json:"feature_ftup"`
	Qci2DscpMapping         map[string]int `json:"qci_dscp_mapping"`
	AllowedApns             string         `json:"allowed_apns"`
	DeniedApns              string         `json:"denied_apns"`
	PfcpProfile             string         `json:"pfcp_profile"`
	TraceAssociation        bool           `json:"trace_association"`
	TraceHeartbeat          bool           `json:"trace_heartbeat"`
	TraceBlocked            bool           `json:"trace_blocked"`
	TraceMaxDumpFiles       int            `json:"trace_files"`
	TraceMaxDumpSize        int            `json:"trace_max_size"`
	TraceMaxDumpPackets     int            `json:"trace_max_packets"`
	IP6RaSupport            bool           `json:"ip6_ra_support"`
	Ip6RaPrefix             string         `json:"ip6_ra_prefix"`
}

type PfcpSession struct {
	LocalSEID   uint64
	RemoteSEID  uint64
	IMSI        string
	MSISDN      string
	PDRs        map[string]SPDRInfo
	FARs        map[string]SFarInfo
	QERs        map[string]SQerInfo
	URRs        map[string]SUrrInfo
	URRSequence uint32
	Traced      bool
}

type SPDRInfo struct {
	PdrID           uint32
	PdrInfo         PdrInfo
	Teid            uint32
	Ipv4            string
	Ipv6            string
	NetworkInstance string
	Allocated       bool
	PCCInfo         *PCCInfo
}

type PCCInfo struct {
	PCCName      string
	Notify       bool
	RawSDFFilter string
	SDFFilter    SdfFilter
	FAR          FarInfo
	QER          QerInfo
}

type SFarInfo struct {
	FarInfo  FarInfo
	GlobalId uint32
}

type SQerInfo struct {
	QerInfo  QerInfo
	GlobalId uint32
}

type SUrrInfo struct {
	UrrInfo         UrrInfo
	GlobalId        uint32
	ReportSeqNumber uint32
}

type PdrInfo struct {
	OuterHeaderRemoval uint8
	FarId              uint32
	QerId              uint32
	Urr1Id             uint32
	Urr2Id             uint32
	SdfFilter          []SdfFilter
	TraceFlag          bool
	NotifyFlag         bool
}

type SdfFilter struct {
	Protocol     uint8
	SrcAddress   IpWMask
	SrcPortRange PortRange
	DstAddress   IpWMask
	DstPortRange PortRange
}

type IpWMask struct {
	Type uint8
	Ip   string
	Mask string
}

type PortRange struct {
	LowerBound uint16
	UpperBound uint16
}

type FarInfo struct {
	Action                uint8  `json:"action"`
	OuterHeaderCreation   uint8  `json:"outer_header_creation"`
	Teid                  uint32 `json:"teid"`
	RemoteIP              string `json:"remote_ip"`
	Trigger               uint8  `json:"trigger"`
	TransportLevelMarking uint16 `json:"transport_level_marking"`
}

type QerInfo struct {
	GateStatusUL uint8
	GateStatusDL uint8
	Qfi          uint8
	Dscp         uint8
	MaxBitrateUL uint64
	MaxBitrateDL uint64
}

type UrrInfo struct {
	UplinkVolume    uint64
	DownlinkVolume  uint64
	VolumeThreshold uint64
	TimeThreshold   float64
}
