package core

import (
	"net"
	"time"

	"github.com/edgecomllc/eupf/cmd/config"

	"github.com/rs/zerolog/log"
	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

type PfcpFunc func(conn *PfcpConnection, msg message.Message, addr string) (message.Message, error)

type PfcpHandlerMap map[uint8]PfcpFunc

func (handlerMap PfcpHandlerMap) Handle(conn *PfcpConnection, buf []byte, addr *net.UDPAddr) error {
	log.Debug().Msgf("Handling PFCP message from %s", addr)
	incomingMsg, err := message.Parse(buf)
	if err != nil {
		log.Warn().Msgf("Ignored undecodable message: %x, error: %s", buf, err)
		return err
	}
	PfcpMessageRx.WithLabelValues(incomingMsg.MessageTypeName()).Inc()
	if handler, ok := handlerMap[incomingMsg.MessageType()]; ok {
		startTime := time.Now()
		// TODO: Trim port as a workaround for NAT changing the port. Explore proper solutions.
		stringIpAddr := addr.IP.String()
		outgoingMsg, err := handler(conn, incomingMsg, stringIpAddr)
		if err != nil {
			log.Warn().Msgf("Error handling PFCP message: %s", err.Error())
			return err
		}
		duration := time.Since(startTime)
		UpfMessageRxLatency.WithLabelValues(incomingMsg.MessageTypeName()).Observe(float64(duration.Microseconds()))
		// Now assumption that all handlers will return a message to send is not true.
		if outgoingMsg != nil {
			PfcpMessageTx.WithLabelValues(outgoingMsg.MessageTypeName()).Inc()
			return conn.SendMessage(outgoingMsg, addr)
		}
		return nil
	} else {
		log.Warn().Msgf("Got unexpected message %s: %s, from: %s", incomingMsg.MessageTypeName(), incomingMsg, addr)
	}
	return nil
}

func setBit(n uint8, pos uint) uint8 {
	n |= (1 << pos)
	return n
}

// https://www.etsi.org/deliver/etsi_ts/129200_129299/129244/16.04.00_60/ts_129244v160400p.pdf page 95
func HandlePfcpAssociationSetupRequest(conn *PfcpConnection, msg message.Message, addr string) (message.Message, error) {
	asreq := msg.(*message.AssociationSetupRequest)
	log.Info().Msgf("Got Association Setup Request from: %s. \n", addr)
	if asreq.NodeID == nil {
		log.Warn().Msgf("Got Association Setup Request without NodeID from: %s", addr)
		// Reject with cause

		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseMandatoryIEMissing)).Inc()
		asres := message.NewAssociationSetupResponse(asreq.SequenceNumber,
			ie.NewCause(ie.CauseMandatoryIEMissing),
		)
		return asres, nil
	}
	printAssociationSetupRequest(asreq)
	// Get NodeID
	remoteNodeID, err := asreq.NodeID.NodeID()
	if err != nil {
		log.Warn().Msgf("Got Association Setup Request with invalid NodeID from: %s", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseMandatoryIEMissing)).Inc()
		asres := message.NewAssociationSetupResponse(asreq.SequenceNumber,
			ie.NewCause(ie.CauseMandatoryIEMissing),
		)
		return asres, nil
	}
	// Check if the PFCP Association Setup Request contains a Node ID for which a PFCP association was already established
	if _, ok := conn.NodeAssociations[remoteNodeID]; ok {
		log.Warn().Msgf("Association Setup Request with NodeID: %s from: %s already exists", remoteNodeID, addr)
		// retain the PFCP sessions that were established with the existing PFCP association and that are requested to be retained, if the PFCP Session Retention Information IE was received in the request; otherwise, delete the PFCP sessions that were established with the existing PFCP association;
		log.Warn().Msg("Session retention is not yet implemented")
	}

	// If the PFCP Association Setup Request contains a Node ID for which a PFCP association was already established
	// proceed with establishing the new PFCP association (regardless of the Recovery AssociationStart received in the request), overwriting the existing association;
	// if the request is accepted:
	// shall store the Node ID of the CP function as the identifier of the PFCP association;
	// Create RemoteNode from AssociationSetupRequest
	remoteNode := NewNodeAssociation(remoteNodeID, addr)
	// Add or replace RemoteNode to NodeAssociationMap
	conn.NodeAssociations[addr] = remoteNode
	log.Info().Msgf("Saving new association: %+v", remoteNode)
	featuresOctets := []uint8{0, 0, 0}
	if config.Conf.FeatureFTUP {
		featuresOctets[0] = setBit(featuresOctets[0], 4)
	}
	if config.Conf.FeatureUEIP {
		featuresOctets[2] = setBit(featuresOctets[2], 2)
	}
	upFunctionFeaturesIE := ie.NewUPFunctionFeatures(featuresOctets[:]...)

	// shall send a PFCP Association Setup Response including:
	asres := message.NewAssociationSetupResponse(asreq.SequenceNumber,
		ie.NewCause(ie.CauseRequestAccepted), // a successful cause
		newIeNodeID(conn.nodeId),             // its Node ID;
		ie.NewRecoveryTimeStamp(conn.RecoveryTimestamp),
		upFunctionFeaturesIE,
	)

	// Send AssociationSetupResponse
	PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseRequestAccepted)).Inc()
	return asres, nil
}

func newIeNodeID(nodeID string) *ie.IE {
	ip := net.ParseIP(nodeID)
	if ip != nil {
		if ip.To4() != nil {
			return ie.NewNodeID(nodeID, "", "")
		}
		return ie.NewNodeID("", nodeID, "")
	}
	return ie.NewNodeID("", "", nodeID)
}
