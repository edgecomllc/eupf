package core

import (
	"net"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

func HandlePfcpHeartbeatRequest(conn *PfcpConnection, msg message.Message, addr string) (message.Message, error) {
	hbreq := msg.(*message.HeartbeatRequest)
	if association := conn.GetAssociation(addr); association != nil {
		association.RefreshRetries()
		//log.Info().Msgf("============1==========>heartbeat count Refreshed : %v", association.HeartbeatRetries)
	}
	ts, err := hbreq.RecoveryTimeStamp.RecoveryTimeStamp()
	if err != nil {
		log.Info().Msgf("Got Heartbeat Request with invalid TS: %s, from: %s", err, addr)
		return nil, err
	} else {
		log.Debug().Msgf("Got Heartbeat Request with TS: %s, from: %s", ts, addr)
	}

	hbres := message.NewHeartbeatResponse(hbreq.SequenceNumber, ie.NewRecoveryTimeStamp(conn.RecoveryTimestamp))
	log.Debug().Msgf("Sent Heartbeat Response to: %s", addr)
	return hbres, nil
}

func HandlePfcpHeartbeatResponse(conn *PfcpConnection, msg message.Message, addr string) (message.Message, error) {
	hbresp := msg.(*message.HeartbeatResponse)
	ts, err := hbresp.RecoveryTimeStamp.RecoveryTimeStamp()
	//log.Info().Msgf("handle response-------> sequence: %v", msg.Sequence())
	if err != nil {
		log.Info().Msgf("Got Heartbeat Response with invalid TS: %s, from: %s", err, addr)
		return nil, err
	} else {
		log.Debug().Msgf("Got Heartbeat Response with TS: %s, from: %s", ts, addr)
		log.Info().Msgf("----------> HandlePfcpHeartbeatResponse delete from heartbeatCache sequence: %v, addr: %s", msg.Sequence(), addr)
		key := addr + ":" + strconv.Itoa(int(msg.Sequence()))
		delete(conn.heartbeatCache, key)
	}

	if association := conn.GetAssociation(addr); association != nil {
		association.RefreshRetries()
		//log.Info().Msgf("============2==========>heartbeat count Refreshed : %v", association.HeartbeatRetries)
	}
	return nil, err
}

func SendHeartbeatRequest(conn *PfcpConnection, sequenceID uint32, associationAddr string) {
	hbreq := message.NewHeartbeatRequest(sequenceID, ie.NewRecoveryTimeStamp(conn.RecoveryTimestamp), nil)
	log.Debug().Msgf("Sent Heartbeat Request to: %s", associationAddr)
	udpAddr, err := net.ResolveUDPAddr("udp", associationAddr+":8805")
	if err == nil {
		if len(conn.heartbeatCache) >= 5 {
			log.Info().Msgf("=====-----SendHeartbeatRequest heartbeatCache => 5 | delete association: %s", associationAddr)
			conn.DeleteAssociation(associationAddr)
		} else {
			log.Info().Msgf("====++++++SendHeartbeatRequest heartbeatCache len : %d | added association: %s, sequence: %d ", len(conn.heartbeatCache), associationAddr, sequenceID)
			key := associationAddr + ":" + strconv.Itoa(int(sequenceID))
			conn.heartbeatCache[key] = struct{}{}
		}
		if err := conn.SendMessage(hbreq, udpAddr); err != nil {
			log.Info().Msgf("Failed to send Heartbeat Request: %s\n", err.Error())
		}
	} else {
		log.Info().Msgf("Failed to send Heartbeat Request: %s\n", err.Error())
	}
}
