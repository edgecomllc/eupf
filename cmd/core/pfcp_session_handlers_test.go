package core

import (
	"github.com/edgecomllc/eupf/cmd/core/service"
	"github.com/rs/zerolog/log"
	"net"
	"testing"
	"time"

	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

func TestHeartbeat(t *testing.T) {
	// Create pfcp connection struct
	pfcpConn := PfcpConnection{}
	hbReq := message.NewHeartbeatRequest(0,
		ie.NewRecoveryTimeStamp(time.Now()),
		nil,
	)
	response, err := HandlePfcpHeartbeatRequest(&pfcpConn, hbReq, "127.0.0.1")
	if err != nil {
		t.Errorf("Error handling heartbeat request: %s", err)
	}
	if response == nil {
		t.Errorf("No response from heartbeat request")
	}
	ts, err := response.(*message.HeartbeatResponse).RecoveryTimeStamp.RecoveryTimeStamp()
	if err != nil {
		t.Errorf("Error getting timestamp from heartbeat response: %s", err)
	}
	t.Logf("Received response from heartbeat request with timestamp: %s", ts)
}

func TestAssociationSetup(t *testing.T) {
	// Create pfcp connection struct
	pfcpConn := PfcpConnection{
		NodeAssociations: make(map[string]*NodeAssociation),
		nodeId:           "test-node",
	}
	asReq := message.NewAssociationSetupRequest(0,
		ie.NewNodeID("", "", "test"),
	)

	remoteIP := "127.0.0.1"
	response, err := HandlePfcpAssociationSetupRequest(&pfcpConn, asReq, remoteIP)
	if err != nil {
		t.Errorf("Error handling association setup request: %s", err)
	}
	cause, err := response.(*message.AssociationSetupResponse).Cause.Cause()
	if err != nil {
		t.Errorf("Error getting cause from association setup response: %s", err)
	}
	if cause != ie.CauseRequestAccepted {
		t.Errorf("Unexpected cause in association setup response: %d", cause)
	}
	// Check nodeId in response
	nodeId, err := response.(*message.AssociationSetupResponse).NodeID.NodeID()
	if err != nil {
		t.Errorf("Error getting node ID from association setup response: %s", err)
	}
	if nodeId != "test-node" {
		t.Errorf("Unexpected node ID in association setup response: %s", nodeId)
	}
	if _, ok := pfcpConn.NodeAssociations[remoteIP]; !ok {
		t.Errorf("Association not created")
	}
}

func SdfFilterStorePreSetup(t *testing.T) (PfcpConnection, string) {
	mapOps := MapOperationsMock{}

	var pfcpHandlers = PfcpHandlerMap{
		message.MsgTypeHeartbeatRequest:            HandlePfcpHeartbeatRequest,
		message.MsgTypeAssociationSetupRequest:     HandlePfcpAssociationSetupRequest,
		message.MsgTypeSessionEstablishmentRequest: HandlePfcpSessionEstablishmentRequest,
		message.MsgTypeSessionDeletionRequest:      HandlePfcpSessionDeletionRequest,
		message.MsgTypeSessionModificationRequest:  HandlePfcpSessionModificationRequest,
	}

	smfIP := "127.0.0.1"
	pfcpConn := PfcpConnection{
		NodeAssociations: make(map[string]*NodeAssociation),
		nodeId:           "test-node",
		mapOperations:    &mapOps,
		pfcpHandlerMap:   pfcpHandlers,
		n3Address:        net.ParseIP("1.2.3.4"),
	}
	asReq := message.NewAssociationSetupRequest(0,
		ie.NewNodeID("", "", "test"),
	)
	response, err := HandlePfcpAssociationSetupRequest(&pfcpConn, asReq, smfIP)
	if err != nil {
		t.Errorf("Error handling association setup request: %s", err)
	}
	cause, err := response.(*message.AssociationSetupResponse).Cause.Cause()
	if err != nil {
		t.Errorf("Error getting cause from association setup response: %s", err)
	}
	if cause != ie.CauseRequestAccepted {
		t.Errorf("Unexpected cause in association setup response: %d", cause)
	}
	// Check nodeId in response
	nodeId, err := response.(*message.AssociationSetupResponse).NodeID.NodeID()
	if err != nil {
		t.Errorf("Error getting node ID from association setup response: %s", err)
	}
	if nodeId != "test-node" {
		t.Errorf("Unexpected node ID in association setup response: %s", nodeId)
	}
	if _, ok := pfcpConn.NodeAssociations[smfIP]; !ok {
		t.Errorf("Association not created")
	}

	return pfcpConn, smfIP
}

func SendDefaulMappingPdrs(t *testing.T, pfcpConn *PfcpConnection, smfIP string) {
	ip1, _ := net.ResolveIPAddr("ip", "1.1.1.1")
	ip2, _ := net.ResolveIPAddr("ip", "2.2.2.2")

	// Requests for default mapping (without SDF filter)

	// Request with UEIP Address
	seReqPre1 := message.NewSessionEstablishmentRequest(0, 0,
		2, 1, 0,
		ie.NewNodeID("", "", "test"),
		ie.NewFSEID(1, net.ParseIP(smfIP), nil),
		ie.NewCreatePDR(
			ie.NewPDRID(1),
			ie.NewPDI(
				ie.NewSourceInterface(ie.SrcInterfaceCore),
				ie.NewUEIPAddress(2, ip1.IP.String(), "", 0, 0),
			),
		),
	)

	// Request with TEID
	seReqPre2 := message.NewSessionEstablishmentRequest(0, 0,
		3, 1, 0,
		ie.NewNodeID("", "", "test"),
		ie.NewFSEID(2, net.ParseIP(smfIP), nil),
		ie.NewCreatePDR(
			ie.NewPDRID(1),
			ie.NewPDI(
				ie.NewSourceInterface(ie.SrcInterfaceCore),
				ie.NewFTEID(0, 0, ip2.IP, nil, 0),
			),
		),
	)

	var err error
	_, err = HandlePfcpSessionEstablishmentRequest(pfcpConn, seReqPre1, smfIP)
	if err != nil {
		t.Errorf("Error handling session establishment request: %s", err)
	}

	_, err = HandlePfcpSessionEstablishmentRequest(pfcpConn, seReqPre2, smfIP)
	if err != nil {
		t.Errorf("Error handling session establishment request: %s", err)
	}

	// Check that session PDRs are correct
	if pfcpConn.NodeAssociations[smfIP].Sessions[2].PDRs[1].Ipv4.String() != "1.1.1.1" {
		t.Errorf("Session 1, got broken")
	}
	if pfcpConn.NodeAssociations[smfIP].Sessions[3].PDRs[1].Teid != 0 {
		t.Errorf("Session 2, got broken")
	}
}

func TestSdfFilterStoreValid(t *testing.T) {

	pfcpConn, smfIP := SdfFilterStorePreSetup(t)
	SendDefaulMappingPdrs(t, &pfcpConn, smfIP)

	if len(pfcpConn.NodeAssociations[smfIP].Sessions[2].PDRs) != 1 {
		t.Errorf("Session 1, should have already stored 1 PDR")
	}

	if len(pfcpConn.NodeAssociations[smfIP].Sessions[3].PDRs) != 1 {
		t.Errorf("Session 2, should have already stored 1 PDR")
	}

	ip1, _ := net.ResolveIPAddr("ip", "1.1.1.1")
	ip2, _ := net.ResolveIPAddr("ip", "2.2.2.2")

	fd := SdfFilterTestStruct{FlowDescription: "permit out ip from 10.62.0.1 to 8.8.8.8/32", Protocol: 1,
		SrcType: 1, SrcAddress: "10.62.0.1", SrcMask: "ffffffff", SrcPortLower: 0, SrcPortUpper: 65535,
		DstType: 1, DstAddress: "8.8.8.8", DstMask: "ffffffff", DstPortLower: 0, DstPortUpper: 65535}

	// Requests for additional mapping (with SDF filter)

	// Request with UEIP Address
	seReq1 := message.NewSessionModificationRequest(0, 0,
		2, 1, 0,
		ie.NewNodeID("", "", "test"),
		ie.NewFSEID(1, net.ParseIP(smfIP), nil), // Why do we need FSEID?
		ie.NewCreatePDR(
			ie.NewPDRID(2),
			ie.NewPDI(
				ie.NewSourceInterface(ie.SrcInterfaceCore),
				//ie.NewFTEID(0, 0, ip1.IP, nil, 0),
				ie.NewUEIPAddress(2, ip1.IP.String(), "", 0, 0),
				ie.NewSDFFilter(fd.FlowDescription, "", "", "", 0),
			),
		),
	)

	// Request with TEID
	seReq2 := message.NewSessionModificationRequest(0, 0,
		3, 1, 0,
		ie.NewNodeID("", "", "test"),
		ie.NewFSEID(2, net.ParseIP(smfIP), nil),
		ie.NewCreatePDR(
			ie.NewPDRID(2),
			ie.NewPDI(
				ie.NewSourceInterface(ie.SrcInterfaceCore),
				ie.NewFTEID(0, 0, ip2.IP, nil, 0),
				// ie.NewUEIPAddress(2, ip2.IP.String(), "", 0, 0),
				ie.NewSDFFilter(fd.FlowDescription, "", "", "", 0),
			),
		),
	)

	var err error
	_, err = HandlePfcpSessionModificationRequest(&pfcpConn, seReq1, smfIP)
	if err != nil {
		t.Errorf("Error handling session establishment request: %s", err)
	}

	_, err = HandlePfcpSessionModificationRequest(&pfcpConn, seReq2, smfIP)
	if err != nil {
		t.Errorf("Error handling session establishment request: %s", err)
	}

	// Check that session PDRs are correct
	if pfcpConn.NodeAssociations[smfIP].Sessions[2].PDRs[2].Ipv4.String() != "1.1.1.1" {
		t.Errorf("Session 1, got broken")
	}

	if pfcpConn.NodeAssociations[smfIP].Sessions[3].PDRs[2].Teid != 0 {
		t.Errorf("Session 2, got broken")
	}

	// Check that SDF filter is stored inside session
	pdrInfo := pfcpConn.NodeAssociations[smfIP].Sessions[2].PDRs[2].PdrInfo
	err = CheckSdfFilterEquality(pdrInfo.SdfFilter, fd)
	if err != nil {
		t.Error(err.Error())
	}

	pdrInfo = pfcpConn.NodeAssociations[smfIP].Sessions[3].PDRs[2].PdrInfo
	err = CheckSdfFilterEquality(pdrInfo.SdfFilter, fd)
	if err != nil {
		t.Error(err.Error())
	}

	// TODO: Check that FAR and QER are successfully stored in PDR with SDF
}

func TestSdfFilterStoreInvalid(t *testing.T) {

	pfcpConn, smfIP := SdfFilterStorePreSetup(t)
	SendDefaulMappingPdrs(t, &pfcpConn, smfIP)

	if len(pfcpConn.NodeAssociations[smfIP].Sessions[2].PDRs) != 1 {
		t.Errorf("Session 1, should have already stored 1 PDR")
	}

	ip1, _ := net.ResolveIPAddr("ip", "1.1.1.1")

	// Request with bad/unsuported SDF
	seReq1 := message.NewSessionModificationRequest(0, 0,
		2, 1, 0,
		ie.NewNodeID("", "", "test"),
		ie.NewFSEID(1, net.ParseIP(smfIP), nil),
		ie.NewCreatePDR(
			ie.NewPDRID(1),
			ie.NewPDI(
				ie.NewSourceInterface(ie.SrcInterfaceCore),
				ie.NewFTEID(0, 0, ip1.IP, nil, 0),
				ie.NewSDFFilter("deny out ip from 10.62.0.1 to 8.8.8.8/32", "", "", "", 0),
			),
		),
	)

	var err error
	_, err = HandlePfcpSessionModificationRequest(&pfcpConn, seReq1, smfIP)
	if err != nil {
		t.Errorf("No error should appear while handling session establishment request. PDR with bad SDF should be skipped?")
	}

	// Check that session PDR wasn't stored? Now it is, just without SDF.
	if pfcpConn.NodeAssociations[smfIP].Sessions[2].PDRs[2].PdrInfo.SdfFilter != nil {
		t.Errorf("Bad SDF shouldn't be stored")
	}
}

func TestFTUPInAssociationSetupResponse(t *testing.T) {
	pfcpConn, smfIP := SdfFilterStorePreSetup(t)

	// Creating an Association Setup Request
	asReq := message.NewAssociationSetupRequest(1,
		ie.NewNodeID("", "", "test"),
	)

	// Processing Association Setup Request
	response, err := HandlePfcpAssociationSetupRequest(&pfcpConn, asReq, smfIP)
	if err != nil {
		t.Errorf("Error handling Association Setup Request: %s", err)
	}

	//Checking if FTUP is enabled in UP Function Features in response
	asRes, ok := response.(*message.AssociationSetupResponse)
	if !ok {
		t.Error("Unexpected response type")
	}

	ftupEnabled := asRes.UPFunctionFeatures.HasFTUP()
	if !ftupEnabled {
		t.Error("FTUP is not enabled in Association Setup Response")
	}
}

func TestTEIDAllocationInSessionEstablishmentResponse(t *testing.T) {
	pfcpConn, smfIP := SdfFilterStorePreSetup(t)

	ipam, err := service.NewIPAM("10.61.0.0/16")
	if err != nil {
		log.Info().Msgf("[ERROR] Failed to create IPAM. err: %v", err)
	}
	pfcpConn.ipam = ipam

	fteid1 := ie.NewFTEID(0x04, 0, net.ParseIP("127.0.0.1"), nil, 1) // 0x04 - CH true
	createPDR1 := ie.NewCreatePDR(
		ie.NewPDRID(1),
		ie.NewPDI(
			ie.NewSourceInterface(ie.SrcInterfaceCore),
			fteid1,
		),
	)

	fteid2 := ie.NewFTEID(0x04, 0, net.ParseIP("127.0.0.2"), nil, 1)
	createPDR2 := ie.NewCreatePDR(
		ie.NewPDRID(2),
		ie.NewPDI(
			ie.NewSourceInterface(ie.SrcInterfaceCore),
			fteid2,
		),
	)

	// Creating a Session Establishment Request
	seReq := message.NewSessionEstablishmentRequest(0, 0,
		2, 1, 0,
		ie.NewNodeID("", "", "test"),
		ie.NewFSEID(1, net.ParseIP(smfIP), nil),
		createPDR1,
		createPDR2,
	)

	// Processing Session Establishment Request
	response, err := HandlePfcpSessionEstablishmentRequest(&pfcpConn, seReq, smfIP)
	if err != nil {
		t.Errorf("Error handling Session Establishment Request: %s", err)
	}

	// Checking if expected TEIDs are allocated in Session Establishment Response
	seRes, ok := response.(*message.SessionEstablishmentResponse)
	if !ok {
		t.Error("Unexpected response type")
	}

	// Checking TEID for each PDR
	log.Info().Msgf("seRes.CreatedPDR len: %d", len(seRes.CreatedPDR))
	for _, pdr := range seRes.CreatedPDR {

		pdi, err := pdr.PDI()
		if err != nil {
			log.Info().Msgf("[ERROR] PDI IE is missing err: %v", err)
		}

		if teidPdiId := findIEindex(pdi, 21); teidPdiId != -1 { // IE Type F-TEID
			if fteid, err := pdi[teidPdiId].FTEID(); err == nil {
				if fteid.TEID != 1 && fteid.TEID != 2 {
					t.Errorf("Unexpected TEID for PDR ID 2: got %d, expected %d or %d", fteid.TEID, 1, 2)
				}
			} else {
				t.Errorf("err: %v", err)
			}
		} else {
			t.Errorf("teidPdiId = -1")
		}

	}

}
