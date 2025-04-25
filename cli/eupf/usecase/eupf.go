package usecase

import (
	"context"
	"fmt"
	"net/url"

	"github.com/edgecomllc/eupf/cli/config"
	"github.com/edgecomllc/eupf/cli/domain"
	"github.com/rs/zerolog/log"
)

type Eupf struct {
	eupfRepo domain.EupfRepository
	fileRepo domain.EupfLocalRepository
	cfg      *config.Config
}

func NewEupf(eupfRepo domain.EupfRepository, storage domain.EupfLocalRepository, cfg *config.Config) domain.EupfUseCase {
	return &Eupf{
		eupfRepo: eupfRepo,
		fileRepo: storage,
		cfg:      cfg,
	}
}

func validateURL(baseURL string) (string, error) {
	if baseURL == "" {
		return "", nil
	}

	parsed, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %s", baseURL)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("unsupported scheme in base URL: %s", parsed.Scheme)
	}

	if parsed.Host == "" {
		return "", fmt.Errorf("invalid base URL: missing host")
	}

	return baseURL, nil
}

func (e *Eupf) TraceList(ctx context.Context, tempBaseURL string) ([]domain.TraceRecord, error) {
	baseURL, err := validateURL(tempBaseURL)
	if err != nil {
		return nil, err
	}

	return e.eupfRepo.TraceList(ctx, baseURL)
}

func (e *Eupf) StartTrace(ctx context.Context, imsi, msisdn *string, tempBaseURL string) error {
	baseURL, err := validateURL(tempBaseURL)
	if err != nil {
		return err
	}

	return e.eupfRepo.StartTrace(ctx, imsi, msisdn, baseURL)
}

func (e *Eupf) StopTrace(ctx context.Context, imsi, msisdn *string, tempBaseURL string) error {
	baseURL, err := validateURL(tempBaseURL)
	if err != nil {
		return err
	}

	return e.eupfRepo.StopTrace(ctx, imsi, msisdn, baseURL)
}

func (e *Eupf) BackupList(ctx context.Context) ([]domain.BackupRecord, error) {
	return e.fileRepo.BackupList(ctx)
}

func (e *Eupf) BackupCreate(ctx context.Context, tempBaseURL string) error {
	baseURL, err := validateURL(tempBaseURL)
	if err != nil {
		return err
	}

	config, err := e.eupfRepo.GetUpfConfig(ctx, baseURL)
	if err != nil {
		return err
	}

	return e.fileRepo.SaveUpfConfig(ctx, config)
}

func (e *Eupf) BackupRestore(ctx context.Context, tempBaseURL string, name string) error {
	baseURL, err := validateURL(tempBaseURL)
	if err != nil {
		return err
	}

	config, err := e.fileRepo.ReadUpfConfig(ctx, name)
	if err != nil {
		return err
	}

	if err := e.eupfRepo.RestoreConfigLoggingLevel(ctx, baseURL, config.LoggingLevel); err != nil {
		log.Warn().Msgf("Failed to restore logging level: %s", err)
	}

	if err := e.eupfRepo.RestoreConfigLoggingCaller(ctx, baseURL, config.LoggingCaller); err != nil {
		log.Warn().Msgf("Failed to restore logging caller: %s", err)
	}

	if err := e.eupfRepo.RestoreConfigDataPlaneEbpf(ctx, baseURL, config.InterfaceName, config.XDPAttachMode); err != nil {
		log.Warn().Msgf("Failed to restore data plane ebpf: %s", err)
	}

	if err := e.eupfRepo.RestoreConfigDataPlaneAddresses(ctx, baseURL, config.N3Address, config.N9Address); err != nil {
		log.Warn().Msgf("Failed to restore data plane addresses: %s", err)
	}

	if err := e.eupfRepo.RestoreConfigPFCPN4(ctx, baseURL, config.PfcpAddress, config.PfcpNodeId, config.PfcpRemoteNode); err != nil {
		log.Warn().Msgf("Failed to restore PFCP N4: %s", err)
	}

	if err := e.eupfRepo.RestoreConfigPFCPSxa(ctx, baseURL, config.SxaLocalAddress, config.SxaLocalNodeId, config.SxaRemoteNode); err != nil {
		log.Warn().Msgf("Failed to restore PFCP SXA: %s", err)
	}

	if err := e.eupfRepo.RestoreConfigPFCPSxb(ctx, baseURL, config.SxbLocalAddress, config.SxbLocalNodeId, config.SxbRemoteNode); err != nil {
		log.Warn().Msgf("Failed to restore PFCP SXB: %s", err)
	}

	if err := e.eupfRepo.RestoreConfigPFCPTimers(ctx, baseURL, config.AssociationSetupTimeout, config.HeartbeatTimeout); err != nil {
		log.Warn().Msgf("Failed to restore PFCP timers: %s", err)
	}

	if err := e.eupfRepo.RestoreConfigGTPPath(ctx, baseURL, config.GtpPeer, config.GtpEchoInterval); err != nil {
		log.Warn().Msgf("Failed to restore GTP path: %s", err)
	}

	return nil
}
