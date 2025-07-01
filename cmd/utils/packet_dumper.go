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

	"github.com/gopacket/gopacket/layers"
	"github.com/gopacket/gopacket/pcapgo"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/perf"
	"github.com/gopacket/gopacket"
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
	if dir == "." {
		dir = os.TempDir()
	}

	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("can't create directories: %s", err.Error())
	}

	dumper := &PacketDumper{
		writeC:         make(chan PacketToWrite, 1024),
		packetsWritten: 0,
		maxFiles:       maxFiles,
		maxSizeBytes:   maxSizeBytes,
		maxPackets:     maxPackets,
		dumpPath:       dumpPath,
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

	filename := fmt.Sprintf("%strace.pcap", dumper.dumpPath)
	if _, err := os.Stat(filename); err == nil {
		newFilename := fmt.Sprintf("%strace-%s.pcap", dumper.dumpPath, time.Now().Format(time.RFC3339))
		if err := os.Rename(filename, newFilename); err != nil {
			log.Error().Err(err).Msgf("Can't rename trace file: %s -> %s", filename, newFilename)
		}
	}

	if files, err := filepath.Glob(fmt.Sprintf("%strace-*.pcap", dumper.dumpPath)); err == nil {
		if len(files) > dumper.maxFiles {
			filesToRemove := len(files) - dumper.maxFiles
			for idx, file := range files[0:filesToRemove] {
				log.Info().Msgf("Remove old trace file [%d/%d]: %s", idx+1, filesToRemove, file)
				if err := os.Remove(file); err != nil {
					log.Error().Err(err).Msgf("Can't remove old trace file: %s", file)
				}
			}
		}
	} else {
		return err
	}

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

	if _, err = w.AddInterface(makeNgInterface("in", layers.LinkTypeEthernet)); err != nil {
		f.Close()
		return fmt.Errorf("can't add in ng pcap interface:: %s", err.Error())
	}

	if _, err = w.AddInterface(makeNgInterface("out", layers.LinkTypeEthernet)); err != nil {
		f.Close()
		return fmt.Errorf("can't add out ng pcap interface:: %s", err.Error())
	}

	if _, err = w.AddInterface(makeNgInterface("drop", layers.LinkTypeEthernet)); err != nil {
		f.Close()
		return fmt.Errorf("can't add drop ng pcap interface:: %s", err.Error())
	}

	sigIn, err := w.AddInterface(makeNgInterface("sig-in", layers.LinkTypeRaw))
	if err != nil {
		f.Close()
		return fmt.Errorf("can't add sig in ng pcap interface:: %s", err.Error())
	}

	sigOut, err := w.AddInterface(makeNgInterface("sig-out", layers.LinkTypeRaw))
	if err != nil {
		f.Close()
		return fmt.Errorf("can't add sig out ng pcap interface:: %s", err.Error())
	}

	dumper.f = f
	dumper.w = w
	dumper.sigInInterface = sigIn
	dumper.sigOutInterface = sigOut

	return nil
}

func (dumper *PacketDumper) ReadTraceMap(traceMap *ebpf.Map) {
	rd, err := perf.NewReader(traceMap, 1024*4096)
	//rd, err := ringbuf.NewReader(traceMap)
	if err != nil {
		log.Error().Msgf(" can't create perf reader: %s", err.Error())
		return
	}
	defer rd.Close()

	var rec perf.Record
	//var rec ringbuf.Record
	for {
		if err := rd.ReadInto(&rec); err != nil {
			log.Error().Msgf(" can't read from perf map: %s", err.Error())
			return
		}

		if rec.LostSamples > 0 {
			log.Warn().Msgf(" lost samples from perf map: %d", rec.LostSamples)
			continue
		}

		if len(rec.RawSample) < 9 {
			log.Error().Msgf(" perf sample too small: %d", len(rec.RawSample))
			continue
		}

		magic := binary.LittleEndian.Uint16(rec.RawSample[:2])
		if magic != 0xdead {
			continue
		}

		packetLength := binary.LittleEndian.Uint16(rec.RawSample[2:4])
		packetIface := binary.LittleEndian.Uint32(rec.RawSample[4:8]) + 1
		packet := rec.RawSample[8 : 8+packetLength]

		dumper.writeC <- PacketToWrite{
			gopacket.CaptureInfo{
				Timestamp:      time.Now(),
				Length:         int(packetLength),
				CaptureLength:  int(packetLength),
				InterfaceIndex: int(packetIface),
			},
			packet,
		}
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
