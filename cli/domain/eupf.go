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
	RestoreConfigLoggingLevel(ctx context.Context, baseURL string, logLevel string) error
	RestoreConfigLoggingCaller(ctx context.Context, baseURL string, logCaller bool) error
	RestoreConfigDataPlaneEbpf(ctx context.Context, baseURL string, interfaceName []string, xdpAttachMode string) error
	RestoreConfigDataPlaneAddresses(ctx context.Context, baseURL string, n3Address string, n9Address string) error
	RestoreConfigPFCPN4(ctx context.Context, baseURL string, pfcpAddress string, pfcpNodeId string, pfcpRemoteNode []string) error
	RestoreConfigPFCPSxa(ctx context.Context, baseURL string, sxaAddress string, sxaNodeId string, sxaRemoteNode []string) error
	RestoreConfigPFCPSxb(ctx context.Context, baseURL string, sxbAddress string, sxbNodeId string, sxbRemoteNode []string) error
	RestoreConfigPFCPTimers(ctx context.Context, baseURL string, associationSetupTimeout uint32, heartbeatTimeout uint32) error
	RestoreConfigGTPPath(ctx context.Context, baseURL string, gtpPeer []string, gtpEchoInterval uint32) error

	SessionShow(ctx context.Context, ip, teid, baseURL string) ([]PfcpSession, error)
	SessionReleaseByIMSI(ctx context.Context, imsi, baseURL string) error
	SessionReleaseByMSISDN(ctx context.Context, msisdn, baseURL string) error
	SessionReleaseByID(ctx context.Context, id, baseURL string) error
}

type EupfUseCase interface {
	ConfigSetNewEUPFBaseURL(baseURL string) error
	ConfigShowEUPFBaseURL() (string, error)

	BackupList(ctx context.Context) ([]BackupRecord, error)
	BackupCreate(ctx context.Context, baseURL string) error
	BackupRestore(ctx context.Context, baseURL string, name string) error

	TraceList(ctx context.Context, tempBaseURL string) ([]TraceRecord, error)
	StartTrace(ctx context.Context, imsi, msisdn *string, tempBaseURL string) error
	StopTrace(ctx context.Context, imsi, msisdn *string, tempBaseURL string) error

	SessionShow(ctx context.Context, ip, teid, baseURL string) ([]PfcpSession, error)
	SessionRelease(ctx context.Context, imsi, msisdn, id, baseURL string) error
}
