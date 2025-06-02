package http

import (
	"context"
	"fmt"
	"time"

	"github.com/edgecomllc/eupf/cli/domain"
	"github.com/edgecomllc/eupf/cli/eupf/models"
	"github.com/go-resty/resty/v2"
)

type EupfHttpRepository struct {
	client  *resty.Client
	baseURL string
}

func NewEupfHttpRepository(baseURL string) domain.EupfRepository {
	client := resty.New().
		SetTimeout(5 * time.Second).
		SetBaseURL(baseURL)

	return &EupfHttpRepository{
		client:  client,
		baseURL: baseURL,
	}
}

func (r *EupfHttpRepository) getClient(baseURL string) *resty.Client {
	if baseURL != "" {
		return r.client.Clone().SetBaseURL(baseURL)
	}

	return r.client
}

func buildTraceRequest(client *resty.Client, ctx context.Context, imsi, msisdn *string) *resty.Request {
	req := client.R().SetContext(ctx)
	if imsi != nil {
		req.SetQueryParam("imsi", *imsi)
	}

	if msisdn != nil {
		req.SetQueryParam("msisdn", *msisdn)
	}

	return req
}

func (r *EupfHttpRepository) TraceList(ctx context.Context, baseURL string) ([]domain.TraceRecord, error) {
	client := r.getClient(baseURL)

	resp, err := client.R().
		SetContext(ctx).
		SetResult([]domain.TraceRecord{}).
		Get("/subscriber_trace/")
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("trace list failed: %s", resp.Status())
	}

	records := *resp.Result().(*[]domain.TraceRecord)

	return records, nil
}

func (r *EupfHttpRepository) StartTrace(ctx context.Context, imsi, msisdn *string, baseURL string) error {
	client := r.getClient(baseURL)
	req := buildTraceRequest(client, ctx, imsi, msisdn)

	resp, err := req.Post("/subscriber_trace/")
	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("start trace failed: %s", resp.Status())
	}

	return nil
}

func (r *EupfHttpRepository) StopTrace(ctx context.Context, imsi, msisdn *string, baseURL string) error {
	client := r.getClient(baseURL)
	req := buildTraceRequest(client, ctx, imsi, msisdn)

	resp, err := req.Delete("/subscriber_trace/")
	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("stop trace failed: %s", resp.Status())
	}

	return nil
}

func (r *EupfHttpRepository) GetUpfConfig(ctx context.Context, baseURL string) (*domain.UpfConfig, error) {
	client := r.getClient(baseURL)

	resp, err := client.R().
		SetContext(ctx).
		SetResult(domain.UpfConfig{}).
		Get("/config/")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("get upf config failed: %s", resp.Status())
	}

	return resp.Result().(*domain.UpfConfig), nil
}

func (r *EupfHttpRepository) sendRestoreRequest(ctx context.Context, baseURL, path string, body any) error {
	client := r.getClient(baseURL)

	req := client.R().
		SetContext(ctx).
		SetBody(body)

	resp, err := req.Post(path)
	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("request to %s failed: %s", path, resp.String())
	}

	return nil
}
func (r *EupfHttpRepository) RestoreConfigLoggingLevel(ctx context.Context, logLevel string, baseURL string) error {
	body := models.LoggingLevelConfig{
		LoggingLevel: logLevel,
	}

	return r.sendRestoreRequest(ctx, baseURL, "/config/logging_level", body)
}

func (r *EupfHttpRepository) RestoreConfigLoggingCaller(ctx context.Context, logCaller bool, baseURL string) error {
	body := models.LoggingCallerConfig{
		LoggingCaller: logCaller,
	}

	return r.sendRestoreRequest(ctx, baseURL, "/config/logging_caller", body)

}

func (r *EupfHttpRepository) RestoreConfigDataPlaneEbpf(ctx context.Context, interfaceName []string, xdpAttachMode string, baseURL string) error {
	body := models.DataPlaneEbpfConfig{
		InterfaceName: interfaceName,
		XDPAttachMode: xdpAttachMode,
	}

	return r.sendRestoreRequest(ctx, baseURL, "/config/dataplane_ebpf", body)

}

func (r *EupfHttpRepository) RestoreConfigDataPlaneAddresses(ctx context.Context, n3Address string, n9Address string, baseURL string) error {
	body := models.DataPlaneAddressesConfig{
		N3Address: n3Address,
		N9Address: n9Address,
	}

	return r.sendRestoreRequest(ctx, baseURL, "/config/dataplane_addresses", body)

}

func (r *EupfHttpRepository) RestoreConfigPFCPN4(ctx context.Context, pfcpAddress string, pfcpNodeId string, pfcpRemoteNode []string, baseURL string) error {
	body := models.PFCPN4Config{
		PFCPAddress:    pfcpAddress,
		PFCPNodeID:     pfcpNodeId,
		PFCPRemoteNode: pfcpRemoteNode,
	}

	return r.sendRestoreRequest(ctx, baseURL, "/config/pfcp_n4", body)

}

func (r *EupfHttpRepository) RestoreConfigPFCPSxa(ctx context.Context, sxaAddress string, sxaNodeId string, sxaRemoteNode []string, baseURL string) error {
	body := models.PFCPSxaConfig{
		SXAAddress:    sxaAddress,
		SXANodeID:     sxaNodeId,
		SXARemoteNode: sxaRemoteNode,
	}

	return r.sendRestoreRequest(ctx, baseURL, "/config/pfcp_sxa", body)

}

func (r *EupfHttpRepository) RestoreConfigPFCPSxb(ctx context.Context, sxbAddress string, sxbNodeId string, sxbRemoteNode []string, baseURL string) error {
	body := models.PFCPSxbConfig{
		SXBAddress:    sxbAddress,
		SXBNodeID:     sxbNodeId,
		SXBRemoteNode: sxbRemoteNode,
	}

	return r.sendRestoreRequest(ctx, baseURL, "/config/pfcp_sxb", body)

}

func (r *EupfHttpRepository) RestoreConfigPFCPTimers(ctx context.Context, associationSetupTimeout uint32, heartbeatTimeout uint32, baseURL string) error {
	body := models.PFCPTimersConfig{
		AssociationSetupTimeout: associationSetupTimeout,
		HeartbeatTimeout:        heartbeatTimeout,
	}

	return r.sendRestoreRequest(ctx, baseURL, "/config/pfcp_timers", body)
}

func (r *EupfHttpRepository) RestoreConfigGTPPath(ctx context.Context, gtpPeer []string, gtpEchoInterval uint32, baseURL string) error {
	body := models.GTPPathConfig{
		GtpPeer:         gtpPeer,
		GtpEchoInterval: gtpEchoInterval,
	}

	return r.sendRestoreRequest(ctx, baseURL, "/config/gtp_path", body)
}

func (r *EupfHttpRepository) SessionShow(ctx context.Context, ip, teid, baseURL string) ([]domain.PfcpSession, error) {
	client := r.getClient(baseURL)
	req := client.R().SetContext(ctx)

	if ip != "" {
		req.SetQueryParam("ip", ip)
	}

	if teid != "" {
		req.SetQueryParam("teid", teid)
	}

	var sessions []domain.PfcpSession

	resp, err := req.
		SetResult(&sessions).
		ForceContentType("application/json").
		Get("/pfcp_sessions/")
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("get pfcp sessions failed: %s", resp.Status())
	}

	return sessions, nil
}

func (r *EupfHttpRepository) SessionReleaseByIMSI(ctx context.Context, imsi, baseURL string) error {
	client := r.getClient(baseURL)
	req := client.R().SetContext(ctx).SetQueryParam("imsi", imsi)

	resp, err := req.Delete("/pfcp_sessions/")
	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("delete pfcp sessions by imsi failed: %s", resp.Status())
	}

	return nil
}

func (r *EupfHttpRepository) SessionReleaseByMSISDN(ctx context.Context, msisdn, baseURL string) error {
	client := r.getClient(baseURL)
	req := client.R().SetContext(ctx).SetQueryParam("msisdn", msisdn)

	resp, err := req.Delete("/pfcp_sessions/")
	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("delete pfcp sessions by msisdn failed: %s", resp.Status())
	}

	return nil
}

func (r *EupfHttpRepository) SessionReleaseByID(ctx context.Context, id, baseURL string) error {
	client := r.getClient(baseURL)
	req := client.R().SetContext(ctx).SetQueryParam("id", id)

	resp, err := req.Delete("/pfcp_sessions/")
	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("delete pfcp sessions by session id failed: %s", resp.Status())
	}

	return nil
}
