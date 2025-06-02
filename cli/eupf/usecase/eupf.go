package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/edgecomllc/eupf/cli/config"
	"github.com/edgecomllc/eupf/cli/domain"
	"github.com/rs/zerolog/log"
)

const (
	baseURLConfigKey = "eupf_addr"
)

var (
	ErrNotFoundConfig = errors.New("config not found")
	ErrNoEUPFAddr     = errors.New("no eupf address")

	ErrInvalidURLFormat  = errors.New("invalid URL format")
	ErrUnsupportedScheme = errors.New("unsupported scheme: must be http or https")
	ErrMissingHostInURL  = errors.New("missing host in URL")
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

	if err := e.eupfRepo.RestoreConfigLoggingLevel(ctx, config.LoggingLevel, baseURL); err != nil {
		log.Warn().Msgf("Failed to restore logging level: %s", err)
	}

	if err := e.eupfRepo.RestoreConfigLoggingCaller(ctx, config.LoggingCaller, baseURL); err != nil {
		log.Warn().Msgf("Failed to restore logging caller: %s", err)
	}

	if err := e.eupfRepo.RestoreConfigDataPlaneEbpf(ctx, config.InterfaceName, config.XDPAttachMode, baseURL); err != nil {
		log.Warn().Msgf("Failed to restore data plane ebpf: %s", err)
	}

	if err := e.eupfRepo.RestoreConfigDataPlaneAddresses(ctx, config.N3Address, config.N9Address, baseURL); err != nil {
		log.Warn().Msgf("Failed to restore data plane addresses: %s", err)
	}

	if err := e.eupfRepo.RestoreConfigPFCPN4(ctx, config.PfcpAddress, config.PfcpNodeId, config.PfcpRemoteNode, baseURL); err != nil {
		log.Warn().Msgf("Failed to restore PFCP N4: %s", err)
	}

	if err := e.eupfRepo.RestoreConfigPFCPSxa(ctx, config.SxaLocalAddress, config.SxaLocalNodeId, config.SxaRemoteNode, baseURL); err != nil {
		log.Warn().Msgf("Failed to restore PFCP SXA: %s", err)
	}

	if err := e.eupfRepo.RestoreConfigPFCPSxb(ctx, config.SxbLocalAddress, config.SxbLocalNodeId, config.SxbRemoteNode, baseURL); err != nil {
		log.Warn().Msgf("Failed to restore PFCP SXB: %s", err)
	}

	if err := e.eupfRepo.RestoreConfigPFCPTimers(ctx, config.AssociationSetupTimeout, config.HeartbeatTimeout, baseURL); err != nil {
		log.Warn().Msgf("Failed to restore PFCP timers: %s", err)
	}

	if err := e.eupfRepo.RestoreConfigGTPPath(ctx, config.GtpPeer, config.GtpEchoInterval, baseURL); err != nil {
		log.Warn().Msgf("Failed to restore GTP path: %s", err)
	}

	return nil
}

// validateBaseURL checks if the provided URL is a valid HTTP(S) address.
func validateBaseURL(u string) error {
	parsed, err := url.ParseRequestURI(u)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidURLFormat, err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ErrUnsupportedScheme
	}

	if parsed.Host == "" {
		return ErrMissingHostInURL
	}

	return nil
}

func (e *Eupf) CliConfigSetNewEUPFBaseURL(baseURL string) error {
	if err := validateBaseURL(baseURL); err != nil {
		return err
	}

	return e.cfg.UpdateFile(map[string]interface{}{baseURLConfigKey: baseURL})
}

func (e *Eupf) CliConfigShowEUPFBaseURL() (string, error) {
	if e.cfg == nil {
		return "", ErrNotFoundConfig
	}

	if e.cfg.EupfAddr == "" {
		return "", ErrNoEUPFAddr
	}

	return e.cfg.EupfAddr, nil
}

func (e *Eupf) SessionShow(ctx context.Context, ip, teid, tempBaseURL string) ([]domain.PfcpSession, error) {
	baseURL, err := validateURL(tempBaseURL)
	if err != nil {
		return nil, err
	}

	return e.eupfRepo.SessionShow(ctx, ip, teid, baseURL)
}

func (e *Eupf) SessionRelease(ctx context.Context, imsi, msisdn, id, tempBaseURL string) error {
	baseURL, err := validateURL(tempBaseURL)
	if err != nil {
		return err
	}

	if imsi != "" {
		return e.eupfRepo.SessionReleaseByIMSI(ctx, imsi, baseURL)
	}

	if msisdn != "" {
		return e.eupfRepo.SessionReleaseByMSISDN(ctx, msisdn, baseURL)
	}

	if id != "" {
		return e.eupfRepo.SessionReleaseByID(ctx, id, baseURL)
	}

	return nil
}

func (e *Eupf) LoggingLevel(ctx context.Context, logLevel string, tempBaseURL string) error {
	tempBaseURL, err := validateURL(tempBaseURL)
	if err != nil {
		return err
	}

	return e.eupfRepo.RestoreConfigLoggingLevel(ctx, logLevel, tempBaseURL)
}

func (e *Eupf) LoggingCaller(ctx context.Context, logCaller bool, tempBaseURL string) error {
	tempBaseURL, err := validateURL(tempBaseURL)
	if err != nil {
		return err
	}

	return e.eupfRepo.RestoreConfigLoggingCaller(ctx, logCaller, tempBaseURL)
}

func (e *Eupf) DataPlaneEbpf(ctx context.Context, interfaceName []string, xdpAttachMode string, tempBaseURL string) error {
	tempBaseURL, err := validateURL(tempBaseURL)
	if err != nil {
		return err
	}

	return e.eupfRepo.RestoreConfigDataPlaneEbpf(ctx, interfaceName, xdpAttachMode, tempBaseURL)
}

func (e *Eupf) DataPlaneAddresses(ctx context.Context, n3Address string, n9Address string, tempBaseURL string) error {
	tempBaseURL, err := validateURL(tempBaseURL)
	if err != nil {
		return err
	}

	return e.eupfRepo.RestoreConfigDataPlaneAddresses(ctx, n3Address, n9Address, tempBaseURL)
}

func (e *Eupf) PFCPN4(ctx context.Context, pfcpAddress string, pfcpNodeId string, pfcpRemoteNode []string, tempBaseURL string) error {
	tempBaseURL, err := validateURL(tempBaseURL)
	if err != nil {
		return err
	}

	return e.eupfRepo.RestoreConfigPFCPN4(ctx, pfcpAddress, pfcpNodeId, pfcpRemoteNode, tempBaseURL)
}

func (e *Eupf) PFCPSxa(ctx context.Context, sxaAddress string, sxaNodeId string, sxaRemoteNode []string, tempBaseURL string) error {
	tempBaseURL, err := validateURL(tempBaseURL)
	if err != nil {
		return err
	}

	return e.eupfRepo.RestoreConfigPFCPSxa(ctx, sxaAddress, sxaNodeId, sxaRemoteNode, tempBaseURL)
}

func (e *Eupf) PFCPSxb(ctx context.Context, sxbAddress string, sxbNodeId string, sxbRemoteNode []string, tempBaseURL string) error {
	tempBaseURL, err := validateURL(tempBaseURL)
	if err != nil {
		return err
	}

	return e.eupfRepo.RestoreConfigPFCPSxb(ctx, sxbAddress, sxbNodeId, sxbRemoteNode, tempBaseURL)
}

func (e *Eupf) PFCPTimers(ctx context.Context, associationSetupTimeout uint32, heartbeatTimeout uint32, tempBaseURL string) error {
	tempBaseURL, err := validateURL(tempBaseURL)
	if err != nil {
		return err
	}

	return e.eupfRepo.RestoreConfigPFCPTimers(ctx, associationSetupTimeout, heartbeatTimeout, tempBaseURL)
}

func (e *Eupf) GTPPath(ctx context.Context, gtpPeer []string, gtpEchoInterval uint32, tempBaseURL string) error {
	tempBaseURL, err := validateURL(tempBaseURL)
	if err != nil {
		return err
	}

	return e.eupfRepo.RestoreConfigGTPPath(ctx, gtpPeer, gtpEchoInterval, tempBaseURL)
}
