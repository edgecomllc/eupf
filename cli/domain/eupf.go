package domain

import (
	"context"
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
	HuaweiSupport           bool           `json:"huawei_support"`
	TraceAssociation        bool           `json:"trace_association"`
	TraceHeartbeat          bool           `json:"trace_heartbeat"`
	TraceBlocked            bool           `json:"trace_blocked"`
	TraceMaxDumpFiles       int            `json:"trace_files"`
	TraceMaxDumpSize        int            `json:"trace_max_size"`
	TraceMaxDumpPackets     int            `json:"trace_max_packets"`
	IP6RaSupport            bool           `json:"ip6_ra_support"`
	Ip6RaPrefix             string         `json:"ip6_ra_prefix"`
}

type EupfLocalRepository interface {
	BackupList(ctx context.Context) ([]BackupRecord, error)
	SaveUpfConfig(ctx context.Context, config *UpfConfig) error
	ReadUpfConfig(ctx context.Context, name string) (*UpfConfig, error)
}

type EupfRepository interface {
	TraceList(ctx context.Context, baseURL string) ([]TraceRecord, error)
	StartTrace(ctx context.Context, imsi, msisdn *string, baseURL string) error
	StopTrace(ctx context.Context, imsi, msisdn *string, baseURL string) error

	GetUpfConfig(ctx context.Context, baseURL string) (*UpfConfig, error)
	RestoreConfigLoggingLevel(ctx context.Context, baseURL string, logLevel string) error
	RestoreConfigLoggingCaller(ctx context.Context, baseURL string, logCaller bool) error
	RestoreConfigDataPlaneEbpf(ctx context.Context, baseURL string, interfaceName []string, xdpAttachMode string) error
	RestoreConfigDataPlaneAddresses(ctx context.Context, baseURL string, n3Address string, n9Address string) error
	RestoreConfigPFCPN4(ctx context.Context, baseURL string, pfcpAddress string, pfcpNodeId string, pfcpRemoteNode []string) error
	RestoreConfigPFCPSxa(ctx context.Context, baseURL string, sxaAddress string, sxaNodeId string, sxaRemoteNode []string) error
	RestoreConfigPFCPSxb(ctx context.Context, baseURL string, sxbAddress string, sxbNodeId string, sxbRemoteNode []string) error
	RestoreConfigPFCPTimers(ctx context.Context, baseURL string, associationSetupTimeout uint32, heartbeatTimeout uint32) error
	RestoreConfigGTPPath(ctx context.Context, baseURL string, gtpPeer []string, gtpEchoInterval uint32) error
}

type EupfUseCase interface {
	ConfigSetNewEUPFBaseURL(baseURL string) error
	ConfigShowEUPFBaseURL() (string, error)

	TraceList(ctx context.Context, tempBaseURL string) ([]TraceRecord, error)
	StartTrace(ctx context.Context, imsi, msisdn *string, tempBaseURL string) error
	StopTrace(ctx context.Context, imsi, msisdn *string, tempBaseURL string) error

	BackupList(ctx context.Context) ([]BackupRecord, error)
	BackupCreate(ctx context.Context, baseURL string) error
	BackupRestore(ctx context.Context, baseURL string, name string) error
}
