package utils

import (
	"encoding/binary"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/perf"
	"github.com/google/gopacket"
	"github.com/rs/zerolog/log"
)

type PacketDumper struct {
	praceMap *ebpf.Map
	f        *os.File
	w        *pcapgo.NgWriter
}

func NewPacketDumper(dumpPath string, praceMap *ebpf.Map) (*PacketDumper, error) {

	f, err := os.Create(dumpPath)
	if err != nil {
		return nil, fmt.Errorf("can't create pcap dump: %s", err.Error())
	}

	//w := pcapgo.NewWriterNanos(f)
	//_ = w.WriteFileHeader(65536, layers.LinkTypeEthernet) // new file, must do this.
	w, err := pcapgo.NewNgWriter(f, layers.LinkTypeEthernet)
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("can't create ng pcap writer: %s", err.Error())
	}

	_, err = w.AddInterface(pcapgo.NgInterface{
		Name:                "in",
		OS:                  runtime.GOOS,
		SnapLength:          0, //unlimited
		TimestampResolution: 9,
		LinkType:            layers.LinkTypeEthernet})
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("can't add in ng pcap interface:: %s", err.Error())
	}

	_, err = w.AddInterface(pcapgo.NgInterface{
		Name:                "out",
		OS:                  runtime.GOOS,
		SnapLength:          0, //unlimited
		TimestampResolution: 9,
		LinkType:            layers.LinkTypeEthernet})
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("can't add out ng pcap interface:: %s", err.Error())
	}

	return &PacketDumper{
		praceMap: praceMap,
		f:        f,
		w:        w,
	}, nil
}

func (dumper *PacketDumper) Run() {

	rd, err := perf.NewReader(dumper.praceMap, 4096)
	if err != nil {
		log.Error().Msgf(" can't create perf reader: %s", err.Error())
		return
	}
	defer rd.Close()

	var rec perf.Record
	for {
		if err := rd.ReadInto(&rec); err != nil {
			log.Error().Msgf(" can't read from perf map: %s", err.Error())
			return
		}

		if rec.LostSamples > 0 {
			log.Warn().Msgf(" lost samples from perf map: %d", rec.LostSamples)
		}

		sampleLength := len(rec.RawSample)
		if sampleLength < 9 {
			log.Error().Msgf(" perf sample too small: %d", sampleLength)
		}

		magic := binary.LittleEndian.Uint16(rec.RawSample[:2])
		if magic != 0xdead {
			continue
		}

		packetLength := binary.LittleEndian.Uint16(rec.RawSample[2:4])
		packetIface := binary.LittleEndian.Uint32(rec.RawSample[4:8]) + 1
		packet := rec.RawSample[8 : 8+packetLength]

		//pack := gopacket.NewPacket(packet, layers.LayerTypeEthernet, gopacket.Default)
		//log.Trace().Msgf("Sample lost=%d, remaining=%d, len=%d, packet: %s", rec.LostSamples, rec.Remaining, packetLength, pack.Dump())

		if err := dumper.w.WritePacket(gopacket.CaptureInfo{
			Timestamp:      time.Now(),
			Length:         int(packetLength),
			CaptureLength:  int(packetLength),
			InterfaceIndex: int(packetIface),
		}, packet); err != nil {
			log.Error().Msgf(" can't write perf sample to pcap dump: %s", err.Error())
		}
	}
}

func (dumper *PacketDumper) Close() {
	if dumper.w != nil {
		dumper.w.Flush()
	}

	if dumper.f != nil {
		dumper.f.Close()
	}
}
