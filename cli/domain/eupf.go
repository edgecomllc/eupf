package domain

import (
	"context"
)

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
	RestoreConfigLoggingLevel(ctx context.Context, logLevel string, baseURL string) error
	RestoreConfigLoggingCaller(ctx context.Context, logCaller bool, baseURL string) error
	RestoreConfigDataPlaneEbpf(ctx context.Context, interfaceName []string, xdpAttachMode string, baseURL string) error
	RestoreConfigDataPlaneAddresses(ctx context.Context, n3Address string, n9Address string, baseURL string) error
	RestoreConfigPFCPN4(ctx context.Context, pfcpAddress string, pfcpNodeId string, pfcpRemoteNode []string, baseURL string) error
	RestoreConfigPFCPSxa(ctx context.Context, sxaAddress string, sxaNodeId string, sxaRemoteNode []string, baseURL string) error
	RestoreConfigPFCPSxb(ctx context.Context, sxbAddress string, sxbNodeId string, sxbRemoteNode []string, baseURL string) error
	RestoreConfigPFCPTimers(ctx context.Context, associationSetupTimeout uint32, heartbeatTimeout uint32, baseURL string) error
	RestoreConfigGTPPath(ctx context.Context, gtpPeer []string, gtpEchoInterval uint32, baseURL string) error

	SessionShow(ctx context.Context, ip, teid, baseURL string) ([]PfcpSession, error)
	SessionReleaseByIMSI(ctx context.Context, imsi, baseURL string) error
	SessionReleaseByMSISDN(ctx context.Context, msisdn, baseURL string) error
	SessionReleaseByID(ctx context.Context, id, baseURL string) error
}

type EupfUseCase interface {
	CliConfigSetNewEUPFBaseURL(baseURL string) error
	CliConfigShowEUPFBaseURL() (string, error)

	BackupList(ctx context.Context) ([]BackupRecord, error)
	BackupCreate(ctx context.Context, baseURL string) error
	BackupRestore(ctx context.Context, baseURL string, name string) error

	TraceList(ctx context.Context, tempBaseURL string) ([]TraceRecord, error)
	StartTrace(ctx context.Context, imsi, msisdn *string, tempBaseURL string) error
	StopTrace(ctx context.Context, imsi, msisdn *string, tempBaseURL string) error

	SessionShow(ctx context.Context, ip, teid, baseURL string) ([]PfcpSession, error)
	SessionRelease(ctx context.Context, imsi, msisdn, id, baseURL string) error

	SetLoggingLevel(ctx context.Context, logLevel string, baseURL string) error
	SetLoggingCaller(ctx context.Context, logCaller bool, baseURL string) error
	SetDataPlaneEbpf(ctx context.Context, interfaceName []string, xdpAttachMode string, baseURL string) error
	SetDataPlaneAddresses(ctx context.Context, n3Address string, n9Address string, baseURL string) error
	SetPFCPN4(ctx context.Context, pfcpAddress string, pfcpNodeId string, pfcpRemoteNode []string, baseURL string) error
	SetPFCPSxa(ctx context.Context, sxaAddress string, sxaNodeId string, sxaRemoteNode []string, baseURL string) error
	SetPFCPSxb(ctx context.Context, sxbAddress string, sxbNodeId string, sxbRemoteNode []string, baseURL string) error
	SetPFCPTimers(ctx context.Context, associationSetupTimeout uint32, heartbeatTimeout uint32, baseURL string) error
	SetGTPPath(ctx context.Context, gtpPeer []string, gtpEchoInterval uint32, baseURL string) error

	GetConfig(ctx context.Context, baseURL string) (*UpfConfig, error)
}
