package http

import (
	"context"
	"fmt"
	"github.com/edgecomllc/eupf/cli/domain"
	"github.com/go-resty/resty/v2"
	"time"
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
