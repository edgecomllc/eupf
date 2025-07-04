package core

import (
	"net"
	"net/netip"
	"sync"
	"testing"
	"time"

	"github.com/edgecomllc/eupf/cmd/config"
	"github.com/edgecomllc/eupf/cmd/core/service"
	"github.com/edgecomllc/eupf/cmd/core/tracing"
	"github.com/edgecomllc/eupf/cmd/ebpf"
	"github.com/rs/zerolog/log"

	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

func TestHeartbeat(t *testing.T) {
	addr := "127.0.0.1"

	// Create pfcp connection struct with association
	pfcpConn := PfcpConnection{
		NodeAssociations: map[string]*NodeAssociation{
			addr: NewNodeAssociation("test-node", ""),
		},
	}
	hbReq := message.NewHeartbeatRequest(0,
		ie.NewRecoveryTimeStamp(time.Now()),
		nil,
	)
	response, _, err := HandlePfcpHeartbeatRequest(&pfcpConn, hbReq, addr)
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
		associationMutex: &sync.Mutex{},
	}
	asReq := message.NewAssociationSetupRequest(0,
		ie.NewNodeID("", "", "test"),
		ie.NewRecoveryTimeStamp(time.Now()),
	)

	remoteIP := "127.0.0.1"
	response, _, err := HandlePfcpAssociationSetupRequest(&pfcpConn, asReq, remoteIP)
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

func PreparePfcpConnection(t *testing.T) (*PfcpConnection, string) {
	config := config.UpfConfig{}
	return PreparePfcpConnectionWithMock(t, &MapOperationsMock{}, config)
}

func PreparePfcpConnectionWithConfig(t *testing.T, config config.UpfConfig) (*PfcpConnection, string) {
	return PreparePfcpConnectionWithMock(t, &MapOperationsMock{}, config)
}

func PreparePfcpConnectionWithMock(t *testing.T, ebpfMock ebpf.ForwardingPlaneController, config config.UpfConfig) (*PfcpConnection, string) {

	var pfcpHandlers = PfcpHandlerMap{
		message.MsgTypeHeartbeatRequest:            HandlePfcpHeartbeatRequest,
		message.MsgTypeAssociationSetupRequest:     HandlePfcpAssociationSetupRequest,
		message.MsgTypeSessionEstablishmentRequest: HandlePfcpSessionEstablishmentRequest,
		message.MsgTypeSessionDeletionRequest:      HandlePfcpSessionDeletionRequest,
		message.MsgTypeSessionModificationRequest:  HandlePfcpSessionModificationRequest,
	}

	featuresOctets := []uint8{0, 0, 0}
	featuresOctets[0] = setBit(featuresOctets[0], 4)
	featuresOctets[2] = setBit(featuresOctets[2], 2)

	smfIP := "127.0.0.1"

	validator, _ := NewNeValidator(config.AllowedApns, config.DeniedApns)

	pfcpConn := PfcpConnection{
		NodeAssociations: make(map[string]*NodeAssociation),
		nodeId:           "test-node",
		mapOperations:    ebpfMock,
		pfcpHandlerMap:   pfcpHandlers,
		n3Address:        net.ParseIP("1.2.3.4"),
		nodeAddrV4:       netip.MustParseAddrPort("127.0.0.1:8085"),
		associationMutex: &sync.Mutex{},
		featuresOctets:   featuresOctets,
		neValidator:      validator,
	}
	asReq := message.NewAssociationSetupRequest(0,
		ie.NewNodeID("", "", "test"),
		ie.NewRecoveryTimeStamp(time.Now()),
	)
	response, _, err := HandlePfcpAssociationSetupRequest(&pfcpConn, asReq, smfIP)
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

	return &pfcpConn, smfIP
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
	_, _, err = HandlePfcpSessionEstablishmentRequest(pfcpConn, seReqPre1, smfIP)
	if err != nil {
		t.Errorf("Error handling session establishment request: %s", err)
	}

	_, _, err = HandlePfcpSessionEstablishmentRequest(pfcpConn, seReqPre2, smfIP)
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

	pfcpConn, smfIP := PreparePfcpConnection(t)
	SendDefaulMappingPdrs(t, pfcpConn, smfIP)

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
	_, _, err = HandlePfcpSessionModificationRequest(pfcpConn, seReq1, smfIP)
	if err != nil {
		t.Errorf("Error handling session establishment request: %s", err)
	}

	_, _, err = HandlePfcpSessionModificationRequest(pfcpConn, seReq2, smfIP)
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
	err = CheckSdfFilterEquality(&pdrInfo.SdfFilter[0], fd)
	if err != nil {
		t.Error(err.Error())
	}

	pdrInfo = pfcpConn.NodeAssociations[smfIP].Sessions[3].PDRs[2].PdrInfo
	err = CheckSdfFilterEquality(&pdrInfo.SdfFilter[0], fd)
	if err != nil {
		t.Error(err.Error())
	}

	// TODO: Check that FAR and QER are successfully stored in PDR with SDF
}

func TestSdfFilterStoreInvalid(t *testing.T) {

	pfcpConn, smfIP := PreparePfcpConnection(t)
	SendDefaulMappingPdrs(t, pfcpConn, smfIP)

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
	_, _, err = HandlePfcpSessionModificationRequest(pfcpConn, seReq1, smfIP)
	if err != nil {
		t.Errorf("No error should appear while handling session establishment request. PDR with bad SDF should be skipped?")
	}

	// Check that session PDR wasn't stored? Now it is, just without SDF.
	if pfcpConn.NodeAssociations[smfIP].Sessions[2].PDRs[2].PdrInfo.SdfFilter != nil {
		t.Errorf("Bad SDF shouldn't be stored")
	}
}

func TestSeveralSdfFiltersSupport(t *testing.T) {

	pfcpConn, smfIP := PreparePfcpConnection(t)
	SendDefaulMappingPdrs(t, pfcpConn, smfIP)

	ip1, _ := net.ResolveIPAddr("ip", "1.1.1.1")

	fd1 := SdfFilterTestStruct{
		FlowDescription: "permit out 17 from 10.169.247.161/32 13033 to 10.169.145.130/32 1235",
		Protocol:        3,
		SrcType:         1,
		SrcAddress:      "10.169.247.161",
		SrcMask:         "ffffffff",
		SrcPortLower:    13033,
		SrcPortUpper:    13033,
		DstType:         1,
		DstAddress:      "10.169.145.130",
		DstMask:         "ffffffff",
		DstPortLower:    1235,
		DstPortUpper:    1235}

	fd2 := SdfFilterTestStruct{
		FlowDescription: "permit out 17 from 10.169.247.161/32 13034 to 10.169.145.130/32 1236",
		Protocol:        3,
		SrcType:         1,
		SrcAddress:      "10.169.247.161",
		SrcMask:         "ffffffff",
		SrcPortLower:    13034,
		SrcPortUpper:    13034,
		DstType:         1,
		DstAddress:      "10.169.145.130",
		DstMask:         "ffffffff",
		DstPortLower:    1236,
		DstPortUpper:    1236}

	// Request with bad/unsuported SDF
	seReq1 := message.NewSessionModificationRequest(0, 0,
		2, 1, 0,
		ie.NewNodeID("", "", "test"),
		ie.NewFSEID(1, net.ParseIP(smfIP), nil),
		ie.NewCreatePDR(
			ie.NewPDRID(2),
			ie.NewPDI(
				ie.NewSourceInterface(ie.SrcInterfaceCore),
				ie.NewFTEID(0, 0, ip1.IP, nil, 0),
				ie.NewSDFFilter(fd1.FlowDescription, "", "", "", 0),
				ie.NewSDFFilter(fd2.FlowDescription, "", "", "", 0),
			),
		),
	)

	var err error
	_, _, err = HandlePfcpSessionModificationRequest(pfcpConn, seReq1, smfIP)
	if err != nil {
		t.Errorf("No error should appear while handling session establishment request. PDR with bad SDF should be skipped?")
	}

	sdfFilterList := pfcpConn.NodeAssociations[smfIP].Sessions[2].PDRs[2].PdrInfo.SdfFilter
	// Check that session PDR wasn't stored? Now it is, just without SDF.
	if sdfFilterList == nil || len(sdfFilterList) != 2 {
		t.Errorf("SDF Filter have to be stored")
	}

	err = CheckSdfFilterEquality(&sdfFilterList[0], fd1)
	if err != nil {
		t.Error(err.Error())
	}

	err = CheckSdfFilterEquality(&sdfFilterList[1], fd2)
	if err != nil {
		t.Error(err.Error())
	}

}

func TestFTUPInAssociationSetupResponse(t *testing.T) {

	config.Conf = config.UpfConfig{
		FTEIDPool:   65536,
		FeatureFTUP: true,
	}

	pfcpConn, smfIP := PreparePfcpConnection(t)

	// Creating an Association Setup Request
	asReq := message.NewAssociationSetupRequest(1,
		ie.NewNodeID("", "", "test"),
		ie.NewRecoveryTimeStamp(time.Now()),
	)

	// Processing Association Setup Request
	response, _, err := HandlePfcpAssociationSetupRequest(pfcpConn, asReq, smfIP)
	if err != nil {
		t.Errorf("Error handling Association Setup Request: %s", err)
	}

	//Checking if FTUP is enabled in UP Function Features in response
	asRes, ok := response.(*message.AssociationSetupResponse)
	if !ok {
		t.Error("Unexpected response type")
	}

	cause, err := response.(*message.AssociationSetupResponse).Cause.Cause()
	if err != nil {
		t.Errorf("Error getting cause from association setup response: %s", err)
	}
	if cause != ie.CauseRequestAccepted {
		t.Errorf("Unexpected cause in association setup response: %d", cause)
	}

	ftupEnabled := asRes.UPFunctionFeatures.HasFTUP()
	if !ftupEnabled {
		t.Error("FTUP is not enabled in Association Setup Response")
	}
}

func TestTEIDAllocationInSessionEstablishmentResponse(t *testing.T) {
	pfcpConn, smfIP := PreparePfcpConnection(t)

	resourceManager, err := service.NewResourceManager("10.61.0.0/16", 65536)
	if err != nil {
		log.Error().Msgf("failed to create ResourceManager. err: %v", err)
	}
	pfcpConn.ResourceManager = resourceManager

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

	fteid3 := ie.NewFTEID(0, 0, net.ParseIP("127.0.0.2"), nil, 1)
	createPDR3 := ie.NewCreatePDR(
		ie.NewPDRID(2),
		ie.NewPDI(
			ie.NewSourceInterface(ie.SrcInterfaceCore),
			fteid3,
		),
	)

	// Creating a Session Establishment Request
	seReq := message.NewSessionEstablishmentRequest(0, 0,
		2, 1, 0,
		ie.NewNodeID("", "", "test"),
		ie.NewFSEID(1, net.ParseIP(smfIP), nil),
		createPDR1,
		createPDR2,
		createPDR3,
	)

	// Processing Session Establishment Request
	response, _, err := HandlePfcpSessionEstablishmentRequest(pfcpConn, seReq, smfIP)
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
	if len(seRes.CreatedPDR) != 2 {
		t.Errorf("Unexpected count TEIDs: got %d, expected %d", len(seRes.CreatedPDR), 2)
	}

	for _, pdr := range seRes.CreatedPDR {
		fteid, err := pdr.FindByType(ie.FTEID)
		if err != nil {
			log.Fatal().Err(err)
		}

		teid, err := fteid.FTEID()
		if err != nil {
			log.Fatal().Err(err)
		}

		if teid.TEID != 1 && teid.TEID != 2 {
			t.Errorf("Unexpected TEID for PDR ID 2: got %d, expected %d or %d", teid.TEID, 1, 2)
		}

		if !teid.HasIPv4() {
			t.Error("HasIPv4 flag is not enabled in TEID")
		}

		if teid.IPv4Address == nil {
			t.Error("TEID has no ip")
		}
	}
}

func TestIPAllocationInSessionEstablishmentResponse(t *testing.T) {
	if config.Conf.FeatureUEIP {
		pfcpConn, smfIP := PreparePfcpConnection(t)

		resourceManager, err := service.NewResourceManager("10.61.0.0/16", 65536)
		if err != nil {
			log.Error().Msgf("failed to create ResourceManager. err: %v", err)
		}
		pfcpConn.ResourceManager = resourceManager

		ueip1 := ie.NewUEIPAddress(16, "", "", 0, 0)
		createPDR1 := ie.NewCreatePDR(
			ie.NewPDRID(1),
			ie.NewPDI(
				ie.NewSourceInterface(ie.SrcInterfaceCore),
				ueip1,
			),
		)

		// Creating a Session Establishment Request
		seReq := message.NewSessionEstablishmentRequest(0, 0,
			2, 1, 0,
			ie.NewNodeID("", "", "test"),
			ie.NewFSEID(1, net.ParseIP(smfIP), nil),
			createPDR1,
		)

		// Processing Session Establishment Request
		response, _, err := HandlePfcpSessionEstablishmentRequest(pfcpConn, seReq, smfIP)
		if err != nil {
			t.Errorf("Error handling Session Establishment Request: %s", err)
		}

		// Checking if expected IPs are allocated in Session Establishment Response
		seRes, ok := response.(*message.SessionEstablishmentResponse)
		if !ok {
			t.Error("Unexpected response type")
		}

		// Checking UEIP for each PDR
		log.Info().Msgf("seRes.CreatedPDR len: %d", len(seRes.CreatedPDR))
		if len(seRes.CreatedPDR) != 1 {
			t.Errorf("Unexpected count PRD's: got %d, expected %d", len(seRes.CreatedPDR), 1)
		}

		for _, pdr := range seRes.CreatedPDR {

			ueipType, err := pdr.FindByType(ie.UEIPAddress)
			if err != nil {
				t.Errorf("FindByType err: %v", err)
			}

			ueip, err := ueipType.UEIPAddress()
			if err != nil {
				t.Errorf("UEIPAddress err: %v", err)
			}

			if ueip.IPv4Address == nil {
				log.Info().Msg("IPv4Address is nil")
			} else {
				if ueip.IPv4Address.String() == "10.61.0.0" {
					log.Info().Msgf("PASSED. IPv4: %s", ueip.IPv4Address.String())
				} else {
					t.Errorf("Unexpected IPv4, got %s, expected %s", ueip.IPv4Address.String(), "10.61.0.0")
				}
			}

		}
	}
}

func TestUEIPInAssociationSetupResponse(t *testing.T) {

	config.Conf = config.UpfConfig{
		UEIPPool:    "10.61.0.0/16",
		FTEIDPool:   65536,
		FeatureUEIP: true,
		FeatureFTUP: false,
	}

	pfcpConn, smfIP := PreparePfcpConnection(t)

	// Creating an Association Setup Request
	asReq := message.NewAssociationSetupRequest(1,
		ie.NewNodeID("", "", "test"),
		ie.NewRecoveryTimeStamp(time.Now()),
	)

	// Processing Association Setup Request
	response, _, err := HandlePfcpAssociationSetupRequest(pfcpConn, asReq, smfIP)
	if err != nil {
		t.Errorf("Error handling Association Setup Request: %s", err)
	}

	// Checking if UEIP is enabled in UP Function Features in response
	asRes, ok := response.(*message.AssociationSetupResponse)
	if !ok {
		t.Error("Unexpected response type")
	}

	// Verify if UEIP is enabled in UP Function Features in response
	ueipEnabled := asRes.UPFunctionFeatures.HasUEIP()
	if !ueipEnabled {
		t.Error("UEIP is not enabled in Association Setup Response")
	}
}

func TestHandlePfcpSessionEstablishmentRequestWithURR(t *testing.T) {
	pfcpConn, smfIP := PreparePfcpConnection(t)

	estReq := message.NewSessionEstablishmentRequest(0, 0, 2, 1, 0,
		ie.NewNodeID("", "", "test"),
		ie.NewFSEID(1, net.ParseIP(smfIP), nil),
		ie.NewCreatePDR(
			ie.NewPDRID(0xffff),
		),
		ie.NewCreateURR(
			ie.NewURRID(0xf),
		),
	)
	_, _, err := HandlePfcpSessionEstablishmentRequest(pfcpConn, estReq, smfIP)
	if err != nil {
		t.Errorf("Error handling session establishment request: %s", err)
	}

	if _, exists := pfcpConn.NodeAssociations[smfIP].Sessions[2].URRs[0xf]; !exists {
		t.Errorf("URR wasn't stored")
	}
}

func TestHandlePfcpSessionModificationRequestWithURR(t *testing.T) {
	ebpfMock := &MapOperationsMock{}
	pfcpConn, smfIP := PreparePfcpConnectionWithMock(t, ebpfMock, config.UpfConfig{})

	estReq := message.NewSessionEstablishmentRequest(0, 0, 2, 1, 0,
		ie.NewNodeID("", "", "test"),
		ie.NewFSEID(1, net.ParseIP(smfIP), nil),
		ie.NewCreatePDR(
			ie.NewPDRID(0xffff),
		),
	)
	_, _, err := HandlePfcpSessionEstablishmentRequest(pfcpConn, estReq, smfIP)
	if err != nil {
		t.Errorf("Error handling session establishment request: %s", err)
	}

	modReq := message.NewSessionModificationRequest(0, 0, 2, 1, 0,
		ie.NewCreateURR(
			ie.NewURRID(0xf),
		),
	)
	_, _, err = HandlePfcpSessionModificationRequest(pfcpConn, modReq, smfIP)
	if err != nil {
		t.Errorf("Error handling session modification request: %s", err)
	}

	if _, exists := pfcpConn.NodeAssociations[smfIP].Sessions[2].URRs[0xf]; !exists {
		t.Errorf("URR wasn't stored")
	}

	// TODO: add case for UpdateURR
	// modReq = message.NewSessionModificationRequest(0, 0, 2, 1, 0,
	// 	ie.NewUpdateURR(
	// 		ie.NewURRID(0xf),
	// 		ie.NewMeasurementMethod(1, 1, 1),
	// 	),
	// )
	// _, err = HandlePfcpSessionModificationRequest(pfcpConn, modReq, smfIP)
	// if err != nil {
	// 	t.Errorf("Error handling session modification request: %s", err)
	// }
	// if pfcpConn.NodeAssociations[smfIP].Sessions[2].URRs[0xf].UrrInfo.MeasurementMethod != ?  {
	// 	t.Errorf("URR wasn't updated")
	// }

	modReq = message.NewSessionModificationRequest(0, 0, 2, 1, 0,
		ie.NewRemoveURR(
			ie.NewURRID(0xf),
		),
	)

	ebpfMock.urr.UplinkVolume = 1234
	ebpfMock.urr.DownlinkVolume = 5678
	msg, _, err := HandlePfcpSessionModificationRequest(pfcpConn, modReq, smfIP)
	if err != nil {
		t.Errorf("Error handling session modification request: %s", err)
	}

	modResp := msg.(*message.SessionModificationResponse)
	if len(modResp.UsageReport) == 0 {
		t.Errorf("SessionModificationResponse doesn't contain Usage Reports")
	}

	ur := modResp.UsageReport[0]
	urrID, _ := ur.URRID()
	if urrID != 0xf {
		t.Errorf("URRID not equal 0xf")
	}

	vol, _ := ur.VolumeMeasurement()
	if !vol.HasTOVOL() {
		t.Errorf("No TotalVolume")
	}

	if !vol.HasULVOL() {
		t.Errorf("No UplinkVolume")
	}

	if !vol.HasDLVOL() {
		t.Errorf("No DownlinkVolume")
	}

	if vol.UplinkVolume != ebpfMock.urr.UplinkVolume {
		t.Errorf("TotalVolume equals %d", vol.UplinkVolume)
	}

	if vol.DownlinkVolume != ebpfMock.urr.DownlinkVolume {
		t.Errorf("TotalVolume equals %d", vol.DownlinkVolume)
	}

	if vol.TotalVolume != ebpfMock.urr.UplinkVolume+ebpfMock.urr.DownlinkVolume {
		t.Errorf("TotalVolume equals %d", vol.TotalVolume)
	}

	if _, exists := pfcpConn.NodeAssociations[smfIP].Sessions[2].URRs[0xf]; exists {
		t.Errorf("URR wasn't removed")
	}
}

func TestHandlePfcpSessionDeletionRequestWithURR(t *testing.T) {

	ebpfMock := &MapOperationsMock{}
	pfcpConn, smfIP := PreparePfcpConnectionWithMock(t, ebpfMock, config.UpfConfig{})

	estReq := message.NewSessionEstablishmentRequest(0, 0, 2, 1, 0,
		ie.NewNodeID("", "", "test"),
		ie.NewFSEID(1, net.ParseIP(smfIP), nil),
		ie.NewCreatePDR(
			ie.NewPDRID(0xffff),
		),
		ie.NewCreateURR(
			ie.NewURRID(0xf),
		),
	)
	_, _, err := HandlePfcpSessionEstablishmentRequest(pfcpConn, estReq, smfIP)
	if err != nil {
		t.Errorf("Error handling session establishment request: %s", err)
	}

	ebpfMock.urr.UplinkVolume = 100
	ebpfMock.urr.DownlinkVolume = 200
	delReq := message.NewSessionDeletionRequest(0, 0, 2, 1, 0)
	msg, _, err := HandlePfcpSessionDeletionRequest(pfcpConn, delReq, smfIP)
	if err != nil {
		t.Errorf("Error handling session deletion request: %s", err)
	}

	delRes := msg.(*message.SessionDeletionResponse)
	if len(delRes.UsageReport) == 0 {
		t.Errorf("SessionDeletionResponse doesn't contain Usage Reports")
	}

	ur := delRes.UsageReport[0]
	urrID, _ := ur.URRID()
	if urrID != 0xf {
		t.Errorf("URRID not equal 0xf")
	}

	vol, _ := ur.VolumeMeasurement()

	if !vol.HasTOVOL() {
		t.Errorf("No TotalVolume")
	}

	if !vol.HasULVOL() {
		t.Errorf("No UplinkVolume")
	}

	if !vol.HasDLVOL() {
		t.Errorf("No DownlinkVolume")
	}

	if vol.UplinkVolume != ebpfMock.urr.UplinkVolume {
		t.Errorf("TotalVolume equals %d", vol.UplinkVolume)
	}

	if vol.DownlinkVolume != ebpfMock.urr.DownlinkVolume {
		t.Errorf("TotalVolume equals %d", vol.DownlinkVolume)
	}

	if vol.TotalVolume != ebpfMock.urr.UplinkVolume+ebpfMock.urr.DownlinkVolume {
		t.Errorf("TotalVolume equals %d", vol.TotalVolume)
	}
}

func TestHandlePfcpSessionEstablishmentRequestWithNotAllowedAPN(t *testing.T) {
	config := config.UpfConfig{
		AllowedApns: "^(ims|internet1)$",
		DeniedApns:  "",
	}
	pfcpConn, smfIP := PreparePfcpConnectionWithConfig(t, config)

	{
		estReq := message.NewSessionEstablishmentRequest(0, 0, 2, 1, 0,
			ie.NewNodeID("", "", "test"),
			ie.NewFSEID(1, net.ParseIP(smfIP), nil),
			ie.NewCreatePDR(
				ie.NewPDRID(0xffff),
				ie.NewPDI(
					ie.NewSourceInterface(ie.SrcInterfaceCore),
					ie.NewNetworkInstance("internet"),
					ie.NewUEIPAddress(2, "1.2.3.4", "", 0, 0),
				),
			),
		)
		msg, _, err := HandlePfcpSessionEstablishmentRequest(pfcpConn, estReq, smfIP)
		if err != nil {
			t.Errorf("Error handling session establishment request: %s", err)
		}

		response := msg.(*message.SessionEstablishmentResponse)
		cause, err := response.Cause.Cause()
		if err == nil {
			if cause != ie.CauseRuleCreationModificationFailure {
				t.Errorf("Wrong response cause code")
			}
		} else {
			t.Errorf("No response casuse code: %s", err)
		}
	}
	{
		estReqWithAllowed := message.NewSessionEstablishmentRequest(0, 0, 2, 1, 0,
			ie.NewNodeID("", "", "test"),
			ie.NewFSEID(1, net.ParseIP(smfIP), nil),
			ie.NewCreatePDR(
				ie.NewPDRID(0xffff),
				ie.NewPDI(
					ie.NewSourceInterface(ie.SrcInterfaceCore),
					ie.NewNetworkInstance("internet1"),
					ie.NewUEIPAddress(2, "1.2.3.4", "", 0, 0),
				),
			),
		)
		msg, _, err := HandlePfcpSessionEstablishmentRequest(pfcpConn, estReqWithAllowed, smfIP)
		if err != nil {
			t.Errorf("Error handling session establishment request: %s", err)
		}

		response := msg.(*message.SessionEstablishmentResponse)
		cause, err := response.Cause.Cause()
		if err == nil {
			if cause != ie.CauseRequestAccepted {
				t.Errorf("Wrong response cause code")
			}
		} else {
			t.Errorf("No response casuse code: %s", err)
		}
	}
}

func TestHandlePfcpSessionEstablishmentRequestWithTrace(t *testing.T) {

	ebpfMock := &MapOperationsMock{}
	pfcpConn, smfIP := PreparePfcpConnectionWithMock(t, ebpfMock, config.UpfConfig{})
	pfcpConn.tracingStorage = tracing.NewSimpleTraceRecordStorage()

	estReq := message.NewSessionEstablishmentRequest(0, 0, 2, 1, 0,
		ie.NewNodeID("", "", "test"),
		ie.NewFSEID(1, net.ParseIP(smfIP), nil),
		ie.NewCreatePDR(
			ie.NewPDRID(0xffff),
			ie.NewPDI(
				ie.NewSourceInterface(ie.SrcInterfaceCore),
				ie.NewNetworkInstance("internet1"),
				ie.NewUEIPAddress(2, "1.2.3.4", "", 0, 0),
			),
		),
		ie.NewVendorSpecificIE(32769, 2011, []byte{0x52, 0x50, 0x03, 0x00, 0x00, 0x00, 0x20, 0xf3}),
	)
	_, _, err := HandlePfcpSessionEstablishmentRequest(pfcpConn, estReq, smfIP)
	if err != nil {
		t.Errorf("Error handling session establishment request: %s", err.Error())
	}

	if ebpfMock.downlinkPDR.TraceFlag {
		t.Errorf("unexpected trace enabled")
	}

	if err = pfcpConn.EnableTracingByImsi("250530000000023"); err != nil {
		t.Errorf("can't enable trace: %s", err.Error())
	}

	estReq.SetSEID(estReq.SEID() + 1)
	_, _, err = HandlePfcpSessionEstablishmentRequest(pfcpConn, estReq, smfIP)
	if err != nil {
		t.Errorf("Error handling session establishment request: %s", err.Error())
	}

	if !ebpfMock.downlinkPDR.TraceFlag {
		t.Errorf("unexpected trace disabled")
	}
}
