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
	log.Info().Msgf("Got Association Setup Request from: %s", addr)
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
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseMandatoryIEIncorrect)).Inc()
		asres := message.NewAssociationSetupResponse(asreq.SequenceNumber,
			ie.NewCause(ie.CauseMandatoryIEIncorrect),
		)
		return asres, nil
	}

	// Recovery Time Stamp
	if asreq.RecoveryTimeStamp == nil {
		log.Warn().Msgf("Got Association Setup Request without RecoveryTimeStamp from: %s", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseMandatoryIEMissing)).Inc()
		asres := message.NewAssociationSetupResponse(asreq.SequenceNumber,
			ie.NewCause(ie.CauseMandatoryIEMissing),
		)
		return asres, nil
	}
	_, err = asreq.RecoveryTimeStamp.RecoveryTimeStamp()
	if err != nil {
		log.Warn().Msgf("Got Association Setup Request with invalid RecoveryTimeStamp from: %s", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseMandatoryIEIncorrect)).Inc()
		asres := message.NewAssociationSetupResponse(asreq.SequenceNumber,
			ie.NewCause(ie.CauseMandatoryIEIncorrect),
		)
		return asres, nil
	}

	// If the PFCP Association Setup Request contains a Node ID for which a PFCP association was already established
	// proceed with establishing the new PFCP association (regardless of the Recovery AssociationStart received in the request), overwriting the existing association;
	// if the request is accepted:
	// shall store the Node ID of the CP function as the identifier of the PFCP association;

	// Check if the PFCP Association Setup Request contains a Node ID for which a PFCP association was already established
	conn.associationMutex.Lock()
	defer conn.associationMutex.Unlock()
	if _, ok := conn.NodeAssociations[addr]; ok {
		log.Warn().Msgf("Association with NodeID: %s and address: %s already exists", remoteNodeID, addr)
		// retain the PFCP sessions that were established with the existing PFCP association and that are requested to be retained, if the PFCP Session Retention Information IE was received in the request; otherwise, delete the PFCP sessions that were established with the existing PFCP association;
		//log.Warn().Msg("Session retention is not yet implemented")
	} else {
		// Create RemoteNode from AssociationSetupRequest
		remoteNode := NewNodeAssociation(remoteNodeID, addr)
		// Add or replace RemoteNode to NodeAssociationMap
		conn.NodeAssociations[addr] = remoteNode

		log.Info().Msgf("Saving new association: %+v", remoteNode)
		if config.Conf.HeartbeatTimeout != 0 {
			go remoteNode.ScheduleHeartbeat(conn)
		}
	}

	// shall send a PFCP Association Setup Response including:
	asres := message.NewAssociationSetupResponse(asreq.SequenceNumber,
		ie.NewCause(ie.CauseRequestAccepted), // a successful cause
		newIeNodeID(conn.nodeId),             // its Node ID;
		ie.NewRecoveryTimeStamp(conn.RecoveryTimestamp),
		ie.NewUPFunctionFeatures(conn.featuresOctets[:]...),
	)

	// Send AssociationSetupResponse
	PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseRequestAccepted)).Inc()
	return asres, nil
}

func HandlePfcpAssociationUpdateRequest(conn *PfcpConnection, msg message.Message, addr string) (message.Message, error) {
	asreq := msg.(*message.AssociationUpdateRequest)
	log.Info().Msgf("Got Association Update Request from: %s", addr)
	if asreq.NodeID == nil {
		log.Warn().Msgf("Got Association Update Request without NodeID from: %s", addr)
		// Reject with cause

		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseMandatoryIEMissing)).Inc()
		asres := message.NewAssociationUpdateResponse(asreq.SequenceNumber,
			ie.NewCause(ie.CauseMandatoryIEMissing),
		)
		return asres, nil
	}
	printAssociationUpdateRequest(asreq)
	// Get NodeID
	remoteNodeID, err := asreq.NodeID.NodeID()
	if err != nil {
		log.Warn().Msgf("Got Association Update Request with invalid NodeID from: %s", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseMandatoryIEIncorrect)).Inc()
		asres := message.NewAssociationUpdateResponse(asreq.SequenceNumber,
			ie.NewCause(ie.CauseMandatoryIEIncorrect),
		)
		return asres, nil
	}

	conn.associationMutex.Lock()
	defer conn.associationMutex.Unlock()
	if _, ok := conn.NodeAssociations[addr]; !ok {
		log.Warn().Msgf("Association with NodeID: %s and address: %s doesn't exist", remoteNodeID, addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseNoEstablishedPFCPAssociation)).Inc()
		asres := message.NewAssociationUpdateResponse(asreq.SequenceNumber,
			ie.NewCause(ie.CauseNoEstablishedPFCPAssociation),
		)
		return asres, nil
	}

	// shall send a PFCP Association Update Response including:
	asres := message.NewAssociationUpdateResponse(asreq.SequenceNumber,
		newIeNodeID(conn.nodeId),             // its Node ID;
		ie.NewCause(ie.CauseRequestAccepted), // a successful cause
	)

	// Send AssociationUpdateResponse
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

func HandlePfcpAssociationSetupResponse(conn *PfcpConnection, msg message.Message, addr string) (message.Message, error) {
	asres := msg.(*message.AssociationSetupResponse)
	log.Info().Msgf("Got Association Setup Response from: %s", addr)

	// Node ID
	if asres.NodeID == nil {
		log.Warn().Msgf("Got Association Setup Response without NodeID from: %s", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseMandatoryIEMissing)).Inc()
		return nil, nil
	}
	remoteNodeID, err := asres.NodeID.NodeID()
	if err != nil {
		log.Warn().Msgf("Got Association Setup Response with invalid NodeID from: %s", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseMandatoryIEIncorrect)).Inc()
		return nil, err
	}

	// Cause
	if asres.Cause == nil {
		log.Warn().Msgf("Got Association Setup Response without Cause from: %s", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseMandatoryIEMissing)).Inc()
		return nil, nil
	}
	cause, err := asres.Cause.Cause()
	if err != nil {
		log.Warn().Msgf("Got Association Setup Response with invalid Cause from: %s", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseMandatoryIEIncorrect)).Inc()
		return nil, err
	}
	if cause != ie.CauseRequestAccepted {
		log.Warn().Msgf("Got Association Setup Response with rejection in cause from: %s. Cause value: %s", addr, causeToString(cause))
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(cause)).Inc()
		return nil, nil
	}

	// CP Function Features
	if asres.CPFunctionFeatures == nil {
		log.Warn().Msgf("Got Association Setup Response without CPFunctionFeatures from: %s", addr)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseConditionalIEMissing)).Inc()
		return nil, nil
	}
	cpFunctionFeatures, err := asres.CPFunctionFeatures.CPFunctionFeatures()
	if err != nil {
		log.Warn().Msgf("Got Association Setup Response with invalid CPFunctionFeatures from: %s. CPFunctionFeatures: %b", addr, cpFunctionFeatures)
		PfcpMessageRxErrors.WithLabelValues(msg.MessageTypeName(), causeToString(ie.CauseConditionalIEMissing)).Inc()
		return nil, err
	}
	log.Info().Msgf("Got Association Setup Response with CPFunctionFeatures from: %s. CPFunctionFeatures: %b", addr, cpFunctionFeatures)

	// Check if the PFCP Association Setup Request contains a Node ID for which a PFCP association was already established
	conn.associationMutex.Lock()
	defer conn.associationMutex.Unlock()
	if _, ok := conn.NodeAssociations[addr]; ok {
		log.Warn().Msgf("Association with NodeID: %s and address: %s already exists", remoteNodeID, addr)
		// retain the PFCP sessions that were established with the existing PFCP association and that are requested to be retained, if the PFCP Session Retention Information IE was received in the request; otherwise, delete the PFCP sessions that were established with the existing PFCP association;
		//log.Warn().Msg("Session retention is not yet implemented")
	} else {
		// Create RemoteNode from AssociationSetupResponse
		remoteNode := NewNodeAssociation(remoteNodeID, addr)
		// Add or replace RemoteNode to NodeAssociationMap
		conn.NodeAssociations[addr] = remoteNode
		log.Info().Msgf("Saving new association: %+v", remoteNode)

		if config.Conf.HeartbeatTimeout != 0 {
			go remoteNode.ScheduleHeartbeat(conn)
		}
	}

	return nil, nil
}
