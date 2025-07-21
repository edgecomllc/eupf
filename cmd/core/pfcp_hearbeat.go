package core

import (
	"net"

	"github.com/rs/zerolog/log"
	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

func HandlePfcpHeartbeatRequest(conn *PfcpConnection, msg message.Message, addr string) (message.Message, bool, error) {
	hbreq := msg.(*message.HeartbeatRequest)
	ts, err := hbreq.RecoveryTimeStamp.RecoveryTimeStamp()
	if err != nil {
		log.Warn().Msgf("Got Heartbeat Request with invalid TS: %s, from: %s", err, addr)
		return nil, true, err
	} else {
		log.Debug().Msgf("Got Heartbeat Request with TS: %s, from: %s", ts, addr)
	}

	_, ok := conn.NodeAssociations[addr]
	if !ok {
		log.Info().Msgf("Rejecting Heartbeat Request from: %s (no association)", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseNoEstablishedPFCPAssociation)).Inc()
		return nil, true, nil
	}

	hbres := message.NewHeartbeatResponse(hbreq.SequenceNumber, ie.NewRecoveryTimeStamp(conn.RecoveryTimestamp))
	log.Debug().Msgf("Sent Heartbeat Response to: %s", addr)
	return hbres, true, nil
}

func HandlePfcpHeartbeatResponse(conn *PfcpConnection, msg message.Message, addr string) (message.Message, bool, error) {
	hbresp := msg.(*message.HeartbeatResponse)
	ts, err := hbresp.RecoveryTimeStamp.RecoveryTimeStamp()
	if err != nil {
		log.Warn().Msgf("Got Heartbeat Response with invalid TS: %s, from: %s", err, addr)
		return nil, true, err
	}

	log.Debug().Msgf("Got Heartbeat Response with TS: %s, from: %s", ts, addr)

	if association := conn.GetAssociation(addr); association != nil {
		association.HandleHeartbeat(msg.Sequence())
	}
	return nil, true, nil
}

func SendHeartbeatRequest(conn *PfcpConnection, sequenceID uint32, associationAddr string) {
	hbreq := message.NewHeartbeatRequest(sequenceID, ie.NewRecoveryTimeStamp(conn.RecoveryTimestamp), nil, ie.NewMetric(25))
	log.Debug().Msgf("Sent Heartbeat Request to: %s", associationAddr)
	udpAddr, err := net.ResolveUDPAddr("udp", associationAddr+":8805")
	if err == nil {
		if err := conn.SendMessageWithTrace(hbreq, udpAddr, true); err != nil {
			log.Warn().Msgf("Failed to send Heartbeat Request: %s\n", err.Error())
		}
	} else {
		log.Warn().Msgf("Failed to send Heartbeat Request: %s\n", err.Error())
	}
}
