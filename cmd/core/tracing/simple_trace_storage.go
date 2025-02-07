package tracing

import (
	"sync"
)

type SimpleTraceRecordStorage struct {
	subsImsi   map[string]*TraceRecord
	subsMsisdn map[string]*TraceRecord
	mu         sync.RWMutex
}

func NewSimpleTraceRecordStorage() *SimpleTraceRecordStorage {
	return &SimpleTraceRecordStorage{
		subsImsi:   make(map[string]*TraceRecord),
		subsMsisdn: make(map[string]*TraceRecord),
	}
}

func (ms *SimpleTraceRecordStorage) GetTraceRecords() ([]TraceRecord, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	subsList := make([]TraceRecord, 0, len(ms.subsImsi)+len(ms.subsMsisdn))
	for _, s := range ms.subsImsi {
		subsList = append(subsList, *s)
	}

	for _, s := range ms.subsMsisdn {
		subsList = append(subsList, *s)
	}

	return subsList, nil
}

func (ms *SimpleTraceRecordStorage) GetTraceRecordByImsi(imsi string) (TraceRecord, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if s := ms.subsImsi[imsi]; s != nil {
		return *s, nil
	} else {
		return TraceRecord{}, ErrSubscriberNotFound{IdType: "IMSI", Val: imsi}
	}
}

func (ms *SimpleTraceRecordStorage) GetTraceRecordByMsisdn(msisdn string) (TraceRecord, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if s := ms.subsMsisdn[msisdn]; s != nil {
		return *s, nil
	} else {
		return TraceRecord{}, ErrSubscriberNotFound{IdType: "MSISDN", Val: msisdn}
	}
}

func (ms *SimpleTraceRecordStorage) AddTraceRecordByImsi(imsi string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.subsImsi[imsi] = NewTraceRecord(imsi, "", "")

	return nil
}

func (ms *SimpleTraceRecordStorage) AddTraceRecordByMsisdn(msisdn string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.subsMsisdn[msisdn] = NewTraceRecord("", msisdn, "")

	return nil
}

func (ms *SimpleTraceRecordStorage) DeleteTraceRecords() ([]TraceRecord, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	subsList := make([]TraceRecord, 0, len(ms.subsImsi)+len(ms.subsMsisdn))
	for key, s := range ms.subsImsi {
		subsList = append(subsList, *s)
		delete(ms.subsImsi, key)
	}

	for key, s := range ms.subsMsisdn {
		subsList = append(subsList, *s)
		delete(ms.subsMsisdn, key)
	}

	return subsList, nil
}

func (ms *SimpleTraceRecordStorage) DeleteTraceRecordByImsi(imsi string) (TraceRecord, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if s := ms.subsImsi[imsi]; s != nil {
		result := *s
		delete(ms.subsImsi, imsi)
		return result, nil
	} else {
		return TraceRecord{}, ErrSubscriberNotFound{IdType: "IMSI", Val: imsi}
	}
}

func (ms *SimpleTraceRecordStorage) DeleteTraceRecordByMsisdn(msisdn string) (TraceRecord, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if s := ms.subsMsisdn[msisdn]; s != nil {
		result := *s
		delete(ms.subsMsisdn, msisdn)
		return result, nil
	} else {
		return TraceRecord{}, ErrSubscriberNotFound{IdType: "MSISDN", Val: msisdn}
	}
}
