package tracing

import (
	"fmt"
	"time"
)

type TraceRecordStorage interface {
	GetTraceRecords() ([]TraceRecord, error)
	GetTraceRecordByImsi(imsi string) (TraceRecord, error)
	GetTraceRecordByMsisdn(msisdn string) (TraceRecord, error)

	AddTraceRecordByImsi(imsi string) error
	AddTraceRecordByMsisdn(msisdn string) error
	DeleteTraceRecords() ([]TraceRecord, error)
	DeleteTraceRecordByImsi(imsi string) (TraceRecord, error)
	DeleteTraceRecordByMsisdn(msisdn string) (TraceRecord, error)
}

type TraceRecord struct {
	Imsi      string
	Msisdn    string
	Timestamp time.Time
	ConnAddr  string
}

func NewTraceRecord(imsi, msisdn, connAddr string) *TraceRecord {
	return &TraceRecord{
		Imsi:      imsi,
		Msisdn:    msisdn,
		Timestamp: time.Now(),
		ConnAddr:  connAddr,
	}
}

type ErrSubscriberNotFound struct {
	IdType string // IMSI or MSISDN
	Val    string
}

func (e ErrSubscriberNotFound) Error() string {
	return fmt.Sprintf("subscriber not found by %s: %s", e.IdType, e.Val)
}
