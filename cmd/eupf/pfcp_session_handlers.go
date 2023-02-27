package main

import (
	"fmt"
	"log"
	"net"
	"reflect"

	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

// Это мои тщетные поптыки сделать обобщение для сообщений сессии, я видимо чего-то не знаю о Go.

// type SessionMessageConstraint interface {
// 	message.SessionEstablishmentRequest | message.SessionModificationRequest | message.SessionDeletionRequest | message.SessionReportRequest
// }

// type SessionMessage interface {
// 	message.Message
// 	GetNodeID() (string, error)
// 	GetCPFSEID() (uint64, error)
// }

// type SessionEstablishmentRequest message.SessionEstablishmentRequest

// func (msg *SessionEstablishmentRequest) GetNodeID() (string, error) {
// 	if msg.NodeID == nil {
// 		return "", fmt.Errorf("NodeID is nil")
// 	}
// 	return msg.NodeID.NodeID()
// }

// func (msg *SessionEstablishmentRequest) GetCPFSEID() (*ie.IE, error) {
// 	if msg.CPFSEID == nil {
// 		return nil, fmt.Errorf("CPFSEID is nil")
// 	}
// 	return msg.CPFSEID, nil
// }

// type SessionModificationRequest message.SessionModificationRequest

// func (msg *SessionModificationRequest) GetNodeID() (string, error) {
// 	if msg.NodeID == nil {
// 		return "", fmt.Errorf("NodeID is nil")
// 	}
// 	return msg.NodeID.NodeID()
// }

// Мне не очень нравится рефлексия, но в джаве и C# такие подходы норма.
func session_guard_handler(conn *PfcpConnection, msg message.Message, addr *net.UDPAddr) error {
	relf_msg := reflect.ValueOf(msg)
	refl_ie_nodeID := reflect.Indirect(relf_msg).FieldByName("NodeID")
	refl_ie_cpfseid := reflect.Indirect(relf_msg).FieldByName("CPFSEID")

	if refl_ie_nodeID.IsZero() || refl_ie_cpfseid.IsZero() {
		log.Printf("Rejecting Session Establishment Request from: %s", addr)
		if err := conn.SendMessage(
			message.NewSessionEstablishmentResponse(0,
				0, 0, msg.Sequence(), 0, ie.NewCause(ie.CauseMandatoryIEMissing),
			), addr); err != nil {
			log.Print(err)
			return err
		}
		return fmt.Errorf("missing mandatory IEs")
	}

	ie_nodeID := refl_ie_nodeID.Interface().(*ie.IE)
	ie_cpfseid := refl_ie_cpfseid.Interface().(*ie.IE)

	remote_nodeID, err_node_id := ie_nodeID.NodeID()
	_, err_cpfseid := ie_cpfseid.FSEID()
	// Respond with error if mandatory IEs are missing
	if err_node_id != nil || err_cpfseid != nil {
		log.Printf("Rejecting Session Establishment Request from: %s", addr)
		if err := conn.SendMessage(
			message.NewSessionEstablishmentResponse(0,
				0, 0, msg.Sequence(), 0, ie.NewCause(ie.CauseMandatoryIEMissing),
			), addr); err != nil {
			log.Print(err)
			return err
		}
	}
	if err_node_id != nil {
		return fmt.Errorf("got Session Establishment Request with invalid NodeID from: %s. Err: %s", addr, err_node_id)
	}
	if err_cpfseid != nil {
		return fmt.Errorf("got Session Establishment Request with invalid CPFSEID from: %s. Err: %s", addr, err_cpfseid)
	}

	// Check if the PFCP Session Establishment Request contains a Node ID for which a PFCP association was already established
	if err := conn.checkNodeAssociation(remote_nodeID); err != nil {
		// shall reject any incoming PFCP Session related messages from that CP function, with a cause indicating that no PFCP association exists with the peer entity
		log.Printf("Rejecting Session Establishment Request from: %s", addr)
		// Send SessionEstablishmentResponse with Cause: No PFCP Established Association
		est_resp := message.NewSessionEstablishmentResponse(0,
			0, 0, msg.Sequence(), 0, ie.NewCause(ie.CauseNoEstablishedPFCPAssociation),
		)
		if err := conn.SendMessage(est_resp, addr); err != nil {
			log.Print(err)
			return err
		}
		return err
	}

	switch msg.MessageType() {
	case message.MsgTypeSessionEstablishmentRequest:
		handlePfcpSessionEstablishmentRequest(conn, msg, addr)
	case message.MsgTypeSessionModificationRequest:
		handlePfcpSessionModificationRequest(conn, msg, addr)
	case message.MsgTypeSessionDeletionRequest:
		handlePfcpSessionDeletionRequest(conn, msg, addr)
	case message.MsgTypeSessionReportRequest:
		handlePfcpSessionReportRequest(conn, msg, addr)
	default:
		return fmt.Errorf("got unexpected message %s: %s, from: %s", msg.MessageTypeName(), msg, addr)
	}
	return nil
}

func handlePfcpSessionEstablishmentRequest(conn *PfcpConnection, msg message.Message, addr *net.UDPAddr) error {
	req := msg.(*message.SessionEstablishmentRequest)
	log.Printf("Got Session Establishment Request from: %s. \n %s", addr, req)
	remote_nodeID, err := req.NodeID.NodeID()
	if err != nil {
		log.Printf("Got Session Establishment Request with invalid NodeID from: %s", addr)
		return err
	}
	fseid, err := req.CPFSEID.FSEID()
	if err != nil {
		return err
	}

	// if session already exists, return error
	if _, ok := conn.nodeAssociations[remote_nodeID].Sessions[fseid.SEID]; ok {
		log.Printf("Rejecting Session Establishment Request from: %s", addr)
		est_resp := message.NewSessionEstablishmentResponse(0,
			0, 0, req.SequenceNumber, 0, ie.NewCause(ie.CauseRequestRejected),
		)
		if err := conn.SendMessage(est_resp, addr); err != nil {
			log.Print(err)
			return err
		}
	}
	// We are using same SEID as SMF
	conn.nodeAssociations[remote_nodeID].Sessions[fseid.SEID] = Session{
		SEID: fseid.SEID,
	}

	// #TODO: Actually apply rules to the dataplane
	// #TODO: Handle failed applies and return error

	// #TODO: Implement printing for other IEs
	for _, far := range req.CreateFAR {
		log.Printf("Create FAR: %+v", far)
		CreateFAR, err := far.CreateFAR()
		if err != nil {
			log.Printf("Error: %+v", err)
			continue
		}
		for _, ie := range CreateFAR {
			log.Printf("IE: %+v", ie)
			printIE(ie)
		}
	}

	for _, qer := range req.CreateQER {
		log.Printf("Create QER: %+v", qer)
	}

	for _, urr := range req.CreateURR {
		log.Printf("Create URR: %+v", urr)
	}

	for _, pdr := range req.CreatePDR {
		log.Printf("Create PDR: %+v", pdr)
	}

	if req.CreateBAR != nil {
		log.Printf("Create BAR: %+v", req.CreateBAR)
	}

	// #TODO: support v6
	var v6 net.IP
	// Send SessionEstablishmentResponse
	est_resp := message.NewSessionEstablishmentResponse(
		0, 0,
		fseid.SEID,
		req.SequenceNumber,
		0,
		ie.NewCause(ie.CauseRequestAccepted),
		newIeNodeID(conn.nodeId),
		ie.NewFSEID(fseid.SEID, conn.nodeAddrV4, v6),
	)
	if err := conn.SendMessage(est_resp, addr); err != nil {
		return err
	}
	return nil
}

func handlePfcpSessionModificationRequest(conn *PfcpConnection, msg message.Message, addr *net.UDPAddr) error {

	return fmt.Errorf("not implemented")
}

func handlePfcpSessionDeletionRequest(conn *PfcpConnection, msg message.Message, addr *net.UDPAddr) error {
	return fmt.Errorf("not implemented")
}

func handlePfcpSessionReportRequest(conn *PfcpConnection, msg message.Message, addr *net.UDPAddr) error {
	return fmt.Errorf("not implemented")
}

// Print child IE recursively
func printIE(ie *ie.IE) {
	log.Printf("IE: %+v", ie)
	for _, child := range ie.ChildIEs {
		printIE(child)
	}
}
