package ebpf

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/perf"
)

type SdfEvent struct {
	srcIP   net.IP
	dstIP   net.IP
	srcPort uint16
	dstPort uint16
	proto   uint8
}

func (e *SdfEvent) UnmarshalIPv4(buffer []byte) {

	e.srcIP = net.IP(buffer[:4])
	e.dstIP = net.IP(buffer[4:8])
	e.srcPort = binary.BigEndian.Uint16(buffer[8:10])
	e.dstPort = binary.BigEndian.Uint16(buffer[10:12])
	e.proto = buffer[12]
}

func (e *SdfEvent) UnmarshalIPv6(buffer []byte) {

	e.srcIP = net.IP(buffer[:16])
	e.dstIP = net.IP(buffer[16:32])
	e.srcPort = binary.BigEndian.Uint16(buffer[32:34])
	e.dstPort = binary.BigEndian.Uint16(buffer[34:36])
	e.proto = buffer[36]
}

func (e *SdfEvent) String() string {

	if len(e.srcIP) != len(e.dstIP) && len(e.srcIP) != 4 && len(e.srcIP) != 16 {
		return ""
	}

	if e.proto == 6 || e.proto == 17 {
		return fmt.Sprintf("permit in %d from %v/%d %d to %v/%d %d", e.proto, e.srcIP, len(e.srcIP)*8, e.srcPort, e.dstIP, len(e.srcIP)*8, e.dstPort)
	} else {
		return fmt.Sprintf("permit in %d from %v/%d to %v/%d", e.proto, e.srcIP, len(e.srcIP)*8, e.dstIP, len(e.srcIP)*8)
	}
}

type SdfFlowNotification struct {
	urrid     uint32
	sdfFilter string
}

func (n *SdfFlowNotification) GetURRID() uint32 {
	return n.urrid
}

func (n *SdfFlowNotification) GetSdfFilter() string {
	return n.sdfFilter
}

type SdfNotifyListener struct {
	reader  *perf.Reader
	notifyC chan SdfFlowNotification
}

func NewSdfNotifyListener(eventMap *ebpf.Map) (*SdfNotifyListener, error) {

	reader, err := perf.NewReader(eventMap, 4096)
	if err != nil {
		return nil, fmt.Errorf("can't create sdf notification reader: %v", err.Error())
	}

	return &SdfNotifyListener{
		reader:  reader,
		notifyC: make(chan SdfFlowNotification, 64),
	}, nil
}
func (l *SdfNotifyListener) Close() {
	if l.reader != nil {
		l.reader.Close()
	}

	close(l.notifyC)
}

func (l *SdfNotifyListener) SetNonblocked() {
	l.reader.SetDeadline(time.Now())
}

func (l *SdfNotifyListener) GetNotificationChannel() <-chan SdfFlowNotification {
	return l.notifyC
}

func (l *SdfNotifyListener) readEvent(rec *perf.Record) (*SdfFlowNotification, error) {
	if err := l.reader.ReadInto(rec); err != nil {
		return nil, fmt.Errorf("can't read from sdf notification map: %v", err.Error())
	}

	if rec.LostSamples > 0 {
		return nil, fmt.Errorf("lost samples from sdf notification map: %d", rec.LostSamples)
	}

	var sdf SdfEvent
	sampleLength := len(rec.RawSample)
	magic := binary.LittleEndian.Uint16(rec.RawSample[:2])
	reference := binary.LittleEndian.Uint32(rec.RawSample[2:6])
	switch magic {
	case 0x1ea1:
		if sampleLength < 15 {
			return nil, fmt.Errorf("perf sample for IPv4 SDF too small: %d", sampleLength)
		}
		sdf.UnmarshalIPv4(rec.RawSample[6:])
	case 0x2ea2:
		if sampleLength < 39 {
			return nil, fmt.Errorf("perf sample for IPv6 SDF too small: %d", sampleLength)
		}
		sdf.UnmarshalIPv6(rec.RawSample[6:])
	default:
		return nil, fmt.Errorf("unknown event: %d", magic)
	}

	return &SdfFlowNotification{reference, sdf.String()}, nil
}

func (l *SdfNotifyListener) ReadEventOnce() (*SdfFlowNotification, error) {
	var rec perf.Record
	return l.readEvent(&rec)
}

func (l *SdfNotifyListener) ReadEventLoop() {

	var rec perf.Record
	for {
		if notification, err := l.readEvent(&rec); err == nil {
			l.notifyC <- *notification //FIXME: send as a pointer?
		}
	}
}
