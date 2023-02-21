// Copyright 2019-2022 go-pfcp authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

// Command hb-client sends a HeartbeatRequest and checks response.
//
// Heartbeat exchanging feature is planned be included in the go-pfcp package's
// built-in functions in the future.
package main

import (
	"flag"
	"log"
	"net"
	"time"

	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

func main() {
	var (
		server = flag.String("s", "127.0.0.1:8805", "server's addr/port")
	)
	flag.Parse()

	raddr, err := net.ResolveUDPAddr("udp", *server)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		log.Fatal(err)
	}

	var seq uint32 = 1
	hbreq, err := message.NewHeartbeatRequest(
		seq,
		ie.NewRecoveryTimeStamp(time.Now()),
		ie.NewSourceIPAddress(net.ParseIP("127.0.0.1"), net.ParseIP("2001::1"), 0),
	).Marshal()
	if err != nil {
		log.Fatal(err)
	}

	if _, err := conn.Write(hbreq); err != nil {
		log.Fatal(err)
	}
	log.Printf("sent Heartbeat Request to: %s", raddr)

	if err := conn.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		log.Fatal(err)
	}

	buf := make([]byte, 1500)
	waiting := true
	for waiting {
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Fatal(err)
		}

		msg, err := message.Parse(buf[:n])
		if err != nil {
			log.Printf("ignored undecodable message: %x, error: %s", buf[:n], err)
			continue
		}

		hbres, ok := msg.(*message.HeartbeatResponse)
		if !ok {
			log.Printf("got unexpected message: %s, from: %s", msg.MessageTypeName(), addr)
			continue
		}

		waiting = false
		ts, err := hbres.RecoveryTimeStamp.RecoveryTimeStamp()
		if err != nil {
			log.Printf("got Heartbeat Response with invalid TS: %s, from: %s", err, addr)
			break
		}
		log.Printf("got Heartbeat Response with TS: %s, from: %s", ts, addr)
	}
}
