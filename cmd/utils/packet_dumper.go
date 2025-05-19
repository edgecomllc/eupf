package utils

import (
	"encoding/binary"
	"fmt"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/perf"
	"github.com/google/gopacket"
	"github.com/rs/zerolog/log"
)

type Dumper interface {
	DumpRawIn(from netip.AddrPort, to netip.AddrPort, packet []byte)
	DumpRawOut(to netip.AddrPort, from netip.AddrPort, packet []byte)
}

type PacketDumper struct {
	f                *os.File
	w                *pcapgo.NgWriter
	sigInInterface   int
	sigOutInterface  int
	writeC           chan PacketToWrite
	packetsWritten   uint64
	maxFiles         int
	maxSizeBytes     int
	maxPackets       int
	currentFileIndex int
	currentFileSize  int
	currentPacketCnt int
	dumpPath         string
}

func makeNgInterface(name string, linkType layers.LinkType) pcapgo.NgInterface {
	return pcapgo.NgInterface{
		Name:                name,
		OS:                  runtime.GOOS,
		SnapLength:          0, //unlimited
		TimestampResolution: 9,
		LinkType:            linkType}
}

func NewPacketDumper(dumpPath string, maxFiles, maxSizeBytes, maxPackets int) (*PacketDumper, error) {
	dir := filepath.Dir(dumpPath)

	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("can't create directories: %s", err.Error())
	}

	var f *os.File
	if len(dumpPath) == 0 {
		f, err = os.CreateTemp("", "trace-*.pcap")
	} else {
		f, err = os.Create(dumpPath)
	}

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

	if _, err = w.AddInterface(makeNgInterface("in", layers.LinkTypeEthernet)); err != nil {
		f.Close()
		return nil, fmt.Errorf("can't add in ng pcap interface:: %s", err.Error())
	}

	if _, err = w.AddInterface(makeNgInterface("out", layers.LinkTypeEthernet)); err != nil {
		f.Close()
		return nil, fmt.Errorf("can't add out ng pcap interface:: %s", err.Error())
	}

	if _, err = w.AddInterface(makeNgInterface("drop", layers.LinkTypeEthernet)); err != nil {
		f.Close()
		return nil, fmt.Errorf("can't add drop ng pcap interface:: %s", err.Error())
	}

	sigInInterface, err := w.AddInterface(makeNgInterface("sig-in", layers.LinkTypeRaw))
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("can't add sig in ng pcap interface:: %s", err.Error())
	}

	sigOutInterface, err := w.AddInterface(makeNgInterface("sig-out", layers.LinkTypeRaw))
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("can't add sig out ng pcap interface:: %s", err.Error())
	}

	dumper := &PacketDumper{
		f:               f,
		w:               w,
		sigInInterface:  sigInInterface,
		sigOutInterface: sigOutInterface,
		writeC:          make(chan PacketToWrite, 1024),
		packetsWritten:  0,
		maxFiles:        maxFiles,
		maxSizeBytes:    maxSizeBytes,
		maxPackets:      maxPackets,
		dumpPath:        dumpPath,
	}

	if err := dumper.rotate(); err != nil {
		return nil, err
	}
	return dumper, nil
}

func (dumper *PacketDumper) rotate() error {
	if dumper.f != nil {
		dumper.w.Flush()
		dumper.f.Close()
	}

	filename := fmt.Sprintf("%strace-%d-%s.pcap", dumper.dumpPath, dumper.currentFileIndex, time.Now().Format(time.RFC3339))
	dumper.currentFileIndex = (dumper.currentFileIndex + 1) % dumper.maxFiles
	dumper.currentFileSize = 0
	dumper.currentPacketCnt = 0

	return dumper.createDumpFile(filename)
}

func (dumper *PacketDumper) createDumpFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("can't create rotated pcap: %w", err)
	}
	w, err := pcapgo.NewNgWriter(f, layers.LinkTypeEthernet)
	if err != nil {
		f.Close()
		return fmt.Errorf("can't create ng writer: %w", err)
	}
	_, _ = w.AddInterface(makeNgInterface("in", layers.LinkTypeEthernet))
	_, _ = w.AddInterface(makeNgInterface("out", layers.LinkTypeEthernet))
	_, _ = w.AddInterface(makeNgInterface("drop", layers.LinkTypeEthernet))
	sigIn, _ := w.AddInterface(makeNgInterface("sig-in", layers.LinkTypeRaw))
	sigOut, _ := w.AddInterface(makeNgInterface("sig-out", layers.LinkTypeRaw))

	dumper.f = f
	dumper.w = w
	dumper.sigInInterface = sigIn
	dumper.sigOutInterface = sigOut

	return nil
}

func (dumper *PacketDumper) ReadTraceMap(traceMap *ebpf.Map) {

	rd, err := perf.NewReader(traceMap, 4096)
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

		dumper.writeC <- PacketToWrite{gopacket.CaptureInfo{
			Timestamp:      time.Now(),
			Length:         int(packetLength),
			CaptureLength:  int(packetLength),
			InterfaceIndex: int(packetIface),
		}, packet}

		// if err := dumper.w.WritePacket(gopacket.CaptureInfo{
		// 	Timestamp:      time.Now(),
		// 	Length:         int(packetLength),
		// 	CaptureLength:  int(packetLength),
		// 	InterfaceIndex: int(packetIface),
		// }, packet); err != nil {
		// 	log.Error().Msgf(" can't write perf sample to pcap dump: %s", err.Error())
		// }
	}
}

func (dumper *PacketDumper) DumpRawIn(from netip.AddrPort, to netip.AddrPort, packet []byte) {
	dumper.dumpRaw(from.Addr().AsSlice(), from.Port(), to.Addr().AsSlice(), to.Port(), dumper.sigInInterface, packet)
}

func (dumper *PacketDumper) DumpRawOut(to netip.AddrPort, from netip.AddrPort, packet []byte) {
	dumper.dumpRaw(from.Addr().AsSlice(), from.Port(), to.Addr().AsSlice(), to.Port(), dumper.sigOutInterface, packet)
}

type PacketToWrite struct {
	info   gopacket.CaptureInfo
	packet []byte
}

func (dumper *PacketDumper) dumpRaw(srcIP net.IP, srcPort uint16, dstIP net.IP, dstPort uint16, iface int, packet []byte) {

	ip := layers.IPv4{
		Version:  4,
		TTL:      64,
		SrcIP:    srcIP,
		DstIP:    dstIP,
		Protocol: layers.IPProtocolUDP,
	}

	udp := layers.UDP{
		SrcPort: layers.UDPPort(srcPort),
		DstPort: layers.UDPPort(dstPort),
	}
	_ = udp.SetNetworkLayerForChecksum(&ip)

	options := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}

	buffer := gopacket.NewSerializeBuffer()
	err := gopacket.SerializeLayers(buffer, options,
		&ip,
		&udp,
		gopacket.Payload(packet),
	)
	if err != nil {
		log.Error().Err(err).Msgf("can't serialize sig sample")
		return
	}
	outgoingPacket := buffer.Bytes()

	dumper.writeC <- PacketToWrite{gopacket.CaptureInfo{
		Timestamp:      time.Now(),
		Length:         int(len(outgoingPacket)),
		CaptureLength:  int(len(outgoingPacket)),
		InterfaceIndex: int(iface),
	}, outgoingPacket}

	// if err := dumper.w.WritePacket(gopacket.CaptureInfo{
	// 	Timestamp:      time.Now(),
	// 	Length:         int(len(outgoingPacket)),
	// 	CaptureLength:  int(len(outgoingPacket)),
	// 	InterfaceIndex: int(iface),
	// }, outgoingPacket); err != nil {
	// 	log.Error().Err(err).Msgf("can't write sig sample to pcap dump")
	// }
}

func (dumper *PacketDumper) Write() {
	for pkt := range dumper.writeC {
		if dumper.currentPacketCnt >= dumper.maxPackets || dumper.currentFileSize >= dumper.maxSizeBytes {
			if err := dumper.rotate(); err != nil {
				log.Error().Err(err).Msg("failed to rotate dump file")
				break
			}
		}

		if err := dumper.w.WritePacket(pkt.info, pkt.packet); err != nil {
			log.Error().Err(err).Msgf("can't write sample to pcap")
			break
		}

		dumper.packetsWritten++
		dumper.currentPacketCnt++
		dumper.currentFileSize += len(pkt.packet)
	}
}

func (dumper *PacketDumper) GetPacketsWritten() uint64 {

	return dumper.packetsWritten
}

func (dumper *PacketDumper) Close(removeTrace bool) {
	if dumper.w != nil {
		dumper.w.Flush()
	}

	if dumper.f != nil {
		dumper.f.Close()
	}

	close(dumper.writeC)

	if removeTrace {
		os.Remove(dumper.f.Name())
	}
}
