package usecase

import (
	"context"
	"fmt"
	"github.com/edgecomllc/eupf/cli/config"
	"github.com/edgecomllc/eupf/cli/domain"
	"net/url"
)

type Eupf struct {
	eupfRepo domain.EupfRepository
	cfg      *config.Config
}

func NewEupf(eupfRepo domain.EupfRepository, cfg *config.Config) domain.EupfUseCase {
	return &Eupf{
		eupfRepo: eupfRepo,
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
