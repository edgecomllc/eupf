package main

import (
	"log"
	"net"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
)

const (
	gtpUPort = "2152"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", ":"+gtpUPort)
	if err != nil {
		log.Fatalf("failed to resolve UDP address: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("failed to listen on UDP port %s: %v", gtpUPort, err)
	}
	defer conn.Close()

	log.Printf("GTP-U Echo Server is listening on UDP port %s...\n", gtpUPort)

	receiveBuffer := make([]byte, 100)

	for {
		n, remoteAddr, err := conn.ReadFromUDP(receiveBuffer)
		if err != nil {
			log.Printf("failed to read from UDP: %v\n", err)
			continue
		}

		packet := gopacket.NewPacket(receiveBuffer[:n], layers.LayerTypeGTPv1U, gopacket.Default)
		if gtpLayer := packet.Layer(layers.LayerTypeGTPv1U); gtpLayer != nil {
			gtp, _ := gtpLayer.(*layers.GTPv1U)

			if gtp.MessageType == 1 { //GTPU_ECHO_REQUEST
				log.Printf("Received Echo Request from %s seq %d\n", remoteAddr, gtp.SequenceNumber)

				response := createEchoResponse(gtp.SequenceNumber)
				if _, err := conn.WriteToUDP(response, remoteAddr); err != nil {
					log.Printf("failed to send Echo Response to %s: %v\n", remoteAddr, err)
				} else {
					log.Printf("sent echo response to %s seq: %d\n", remoteAddr, gtp.SequenceNumber)
				}
			} else {
				log.Printf("received non-Echo Request %d from %s\n", gtp.MessageType, remoteAddr)
			}
		} else {
			log.Printf("received not GTP packet %s\n", remoteAddr)
		}
	}
}

func createEchoResponse(seq uint16) []byte {
	buf := gopacket.NewSerializeBuffer()

	err := gopacket.SerializeLayers(buf, gopacket.SerializeOptions{},
		&layers.GTPv1U{
			Version:            1,
			MessageType:        2,
			MessageLength:      4,
			TEID:               0,
			SequenceNumberFlag: true,
			SequenceNumber:     seq,
		},
	)
	if err != nil {
		log.Fatalf("failed to serialize Echo Response: %v", err)
	}

	return buf.Bytes()
}
