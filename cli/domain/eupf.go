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

type EupfRepository interface {
	TraceList(ctx context.Context, baseURL string) ([]TraceRecord, error)
	StartTrace(ctx context.Context, imsi, msisdn *string, baseURL string) error
	StopTrace(ctx context.Context, imsi, msisdn *string, baseURL string) error
}

type EupfUseCase interface {
	TraceList(ctx context.Context, tempBaseURL string) ([]TraceRecord, error)
	StartTrace(ctx context.Context, imsi, msisdn *string, tempBaseURL string) error
	StopTrace(ctx context.Context, imsi, msisdn *string, tempBaseURL string) error
}
