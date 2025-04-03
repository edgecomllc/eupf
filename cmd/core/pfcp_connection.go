package core

import (
	"fmt"
	"math"
	"net"
	"net/netip"
	"regexp"
	"sync"
	"time"

	"github.com/edgecomllc/eupf/cmd/config"
	"github.com/edgecomllc/eupf/cmd/core/service"
	"github.com/edgecomllc/eupf/cmd/core/tracing"
	"github.com/edgecomllc/eupf/cmd/ebpf"
	"github.com/edgecomllc/eupf/cmd/utils"

	"github.com/rs/zerolog/log"
	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

type AssociationConnector interface {
	getAddress() string
	sendAssociationSetupRequest(connection *PfcpConnection)
}

var pfcpHandlers = PfcpHandlerMap{
	message.MsgTypeHeartbeatRequest:            HandlePfcpHeartbeatRequest,
	message.MsgTypeHeartbeatResponse:           HandlePfcpHeartbeatResponse,
	message.MsgTypeAssociationSetupRequest:     HandlePfcpAssociationSetupRequest,
	message.MsgTypeAssociationSetupResponse:    HandlePfcpAssociationSetupResponse,
	message.MsgTypeAssociationUpdateRequest:    HandlePfcpAssociationUpdateRequest,
	message.MsgTypeSessionEstablishmentRequest: HandlePfcpSessionEstablishmentRequest,
	message.MsgTypeSessionDeletionRequest:      HandlePfcpSessionDeletionRequest,
	message.MsgTypeSessionModificationRequest:  HandlePfcpSessionModificationRequest,
}

type PfcpConnection struct {
	udpConn                *net.UDPConn
	pfcpHandlerMap         PfcpHandlerMap
	associationMutex       *sync.Mutex
	NodeAssociations       map[string]*NodeAssociation
	nodeId                 string
	nodeAddrV4             netip.AddrPort
	n3Address              net.IP
	n9Address              net.IP
	mapOperations          ebpf.ForwardingPlaneController
	RecoveryTimestamp      time.Time
	featuresOctets         []uint8
	ResourceManager        *service.ResourceManager
	heartbeatFailedC       chan string
	nodes                  []AssociationConnector
	neValidator            *NeValidator
	tracingStorage         tracing.TraceRecordStorage
	dumper                 utils.Dumper
	AssociationSetupTicker *time.Ticker
	PCCRules               map[string]PCCInfo
	PCCRulesNameLinkURRID  map[uint32][]string
}

func NewPfcpConnection(
	addr string,
	nodeId string,
	n3Ip string,
	n9Ip string,
	mapOperations ebpf.ForwardingPlaneController,
	resourceManager *service.ResourceManager,
	dumper utils.Dumper,
) (*PfcpConnection, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Warn().Msgf("Can't resolve UDP address: %s", err.Error())
		return nil, err
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Warn().Msgf("Can't listen UDP address: %s", err.Error())
		return nil, err
	}

	n3Addr := net.ParseIP(n3Ip)
	if n3Addr == nil {
		return nil, fmt.Errorf("failed to parse N3 IP address ID: %s", n3Ip)
	}
	n9Addr := net.ParseIP(n9Ip)
	if n9Addr == nil {
		return nil, fmt.Errorf("failed to parse N9 IP address ID: %s", n9Ip)
	}

	log.Info().Msgf("Starting PFCP connection: %v with Node ID: %v, N3 address: %v, N9 address: %v", udpAddr, nodeId, n3Addr, n9Addr)

	featuresOctets := []uint8{0, 0, 0}
	featuresOctets[1] = setBit(featuresOctets[1], 0)
	if config.Conf.FeatureFTUP {
		featuresOctets[0] = setBit(featuresOctets[0], 4)
	}
	if config.Conf.FeatureUEIP {
		featuresOctets[2] = setBit(featuresOctets[2], 2)
	}

	validator, err := NewNeValidator(config.Conf.AllowedApns, config.Conf.DeniedApns)
	if err != nil {
		return nil, fmt.Errorf("failed to init network instance validator: %s", err.Error())
	}

	pccRule, pccRulesNameLinkURRID := buildPCCRuleMap()

	return &PfcpConnection{
		udpConn:               udpConn,
		pfcpHandlerMap:        pfcpHandlers,
		associationMutex:      &sync.Mutex{},
		NodeAssociations:      map[string]*NodeAssociation{},
		nodeId:                nodeId,
		nodeAddrV4:            udpAddr.AddrPort(),
		n3Address:             n3Addr,
		n9Address:             n9Addr,
		mapOperations:         mapOperations,
		RecoveryTimestamp:     time.Now(),
		featuresOctets:        featuresOctets,
		ResourceManager:       resourceManager,
		heartbeatFailedC:      make(chan string),
		nodes:                 []AssociationConnector{},
		neValidator:           validator,
		tracingStorage:        tracing.NewSimpleTraceRecordStorage(),
		dumper:                dumper,
		PCCRules:              pccRule,
		PCCRulesNameLinkURRID: pccRulesNameLinkURRID,
	}, nil
}

func (connection *PfcpConnection) Update(
	pfcpAddress string,
	pfcpNodeId string,
	pfcpRemoteNodes []string,
) error {
	if connection.udpConn.LocalAddr().String() != pfcpAddress {
		udpAddr, err := net.ResolveUDPAddr("udp", pfcpAddress)
		if err != nil {
			log.Warn().Msgf("Can't resolve UDP address: %s", err.Error())
			return err
		}

		newUdpConn, err := net.ListenUDP("udp", udpAddr)
		if err != nil {
			log.Warn().Msgf("Can't listen UDP address: %s", err.Error())
			return err
		}

		log.Info().Msgf("Starting new PFCP connection: %v", udpAddr)

		oldUDPConn := connection.udpConn
		connection.udpConn = newUdpConn
		connection.nodeId = pfcpNodeId

		err = oldUDPConn.Close()
		if err != nil {
			log.Error().Msgf("Can't close old UDP connection: %s", err.Error())
		}

		log.Info().Msgf("Delete old PFCP connection: %v", udpAddr)
	}

	return nil
}

func (connection *PfcpConnection) GetLocalAddr() string {
	return connection.udpConn.LocalAddr().String()
}

func (connection *PfcpConnection) UpdateN3Address(n3addr net.IP) {
	connection.n3Address = n3addr
}

func (connection *PfcpConnection) UpdateN9Address(n9addr net.IP) {
	connection.n9Address = n9addr
}

func (connection *PfcpConnection) GetN3Addr() string {
	return connection.n3Address.String()
}

func (connection *PfcpConnection) GetN9Addr() string {
	return connection.n9Address.String()
}

func (connection *PfcpConnection) GetMapOperations() ebpf.ForwardingPlaneController {
	return connection.mapOperations
}

func (connection *PfcpConnection) GetResourceManager() *service.ResourceManager {
	return connection.ResourceManager
}

func (connection *PfcpConnection) GetDumper() utils.Dumper {
	return connection.dumper
}

func (connection *PfcpConnection) GetAssociation(assocAddr string) *NodeAssociation {
	if assoc, ok := connection.NodeAssociations[assocAddr]; ok {
		return assoc
	}
	return nil
}

func (connection *PfcpConnection) ValidateNetworkInstance(networkInstance string) bool {
	return connection.neValidator.Validate(networkInstance)
}

func buildPCCRuleMap() (map[string]PCCInfo, map[uint32][]string) {
	pccRules := make(map[string]PCCInfo)
	pccRulesURRLinks := make(map[uint32][]string)

	for i := range config.PCCConf.PccRules {
		sdfFilter, err := ParseSdfFilter(config.PCCConf.PccRules[i].SdfFilter)
		if err != nil {
			log.Error().Msgf("Error parsing SDF filter: %s", err.Error())

			continue
		}

		pccRulesURRLinks[config.PCCConf.PccRules[i].Urr.Urrid] = append(
			pccRulesURRLinks[config.PCCConf.PccRules[i].Urr.Urrid],
			config.PCCConf.PccRules[i].PccName,
		)

		pccRules[config.PCCConf.PccRules[i].PccName] = PCCInfo{
			PCCName:      config.PCCConf.PccRules[i].PccName,
			SDFFilter:    sdfFilter,
			RawSDFFilter: config.PCCConf.PccRules[i].SdfFilter,
			FAR: ebpf.FarInfo{
				Action:                config.PCCConf.PccRules[i].Far.Action,
				OuterHeaderCreation:   config.PCCConf.PccRules[i].Far.OuterHeaderCreation,
				Teid:                  config.PCCConf.PccRules[i].Far.Teid,
				RemoteIP:              config.PCCConf.PccRules[i].Far.RemoteIP,
				LocalIP:               math.MaxUint32, // change on processing
				TransportLevelMarking: config.PCCConf.PccRules[i].Far.TransportLevelMarking,
			},
			QER: ebpf.QerInfo{
				GateStatusDL: 0,
				GateStatusUL: 0,
				Qfi:          config.PCCConf.PccRules[i].Qer.Qfi,
				MaxBitrateDL: uint64(config.PCCConf.PccRules[i].Qer.MaxBitrateDl),
				MaxBitrateUL: uint64(config.PCCConf.PccRules[i].Qer.MaxBitrateUl),
			},
		}
	}

	log.Info().Msgf("loaded pcc rules map %+v", config.PCCConf.PccRules)
	log.Info().Msgf("create pcc rules map %+v", pccRules)

	return pccRules, pccRulesURRLinks
}

func (connection *PfcpConnection) SetRemoteNodes(nodes []AssociationConnector) {
	connection.nodes = nodes
}

func (connection *PfcpConnection) Run() {
	connection.AssociationSetupTicker = time.NewTicker(time.Duration(config.Conf.AssociationSetupTimeout) * time.Second)
	reportTicker := time.NewTicker(time.Duration(60) * time.Second)
	buf := make([]byte, 1500)

	for {
		select {
		case <-connection.AssociationSetupTicker.C:
			log.Debug().Msgf("pfcp address: %s", connection.GetLocalAddr())
			connection.RefreshAssociations()
		case associationAddr := <-connection.heartbeatFailedC:
			connection.DeleteAssociation(associationAddr)
		case <-reportTicker.C:
			connection.SendReports()
		default:
			_ = connection.udpConn.SetReadDeadline(time.Now().Add(time.Second))
			n, addr, err := connection.Receive(buf)
			if err != nil {
				if err.(*net.OpError).Timeout() {
					continue
				}
				log.Warn().Msgf("Error reading from UDP socket: %s", err.Error())
				time.Sleep(1 * time.Second)
				continue
			}
			log.Debug().Msgf("Received %d bytes from %s", n, addr)
			connection.Handle(buf[:n], addr)
		}
	}
}

func (connection *PfcpConnection) Close() {
	err := connection.udpConn.Close()
	if err != nil {
		log.Error().Msgf("Error closing UDP socket: %s", err.Error())
	}
}

func (connection *PfcpConnection) Receive(b []byte) (n int, addr *net.UDPAddr, err error) {
	return connection.udpConn.ReadFromUDP(b)
}

func (connection *PfcpConnection) Handle(b []byte, addr *net.UDPAddr) {
	if err := connection.pfcpHandlerMap.Handle(connection, b, addr); err != nil {
		log.Warn().Msgf("Error handling PFCP message: %s", err.Error())
	}
}

func (connection *PfcpConnection) Send(b []byte, addr *net.UDPAddr) (int, error) {
	return connection.udpConn.WriteTo(b, addr)
}

func (connection *PfcpConnection) SendMessage(msg message.Message, addr *net.UDPAddr) error {
	return connection.SendMessageWithTrace(msg, addr, false)
}

func (connection *PfcpConnection) SendMessageWithTrace(msg message.Message, addr *net.UDPAddr, trace bool) error {
	responseBytes := make([]byte, msg.MarshalLen())
	if err := msg.MarshalTo(responseBytes); err != nil {
		log.Warn().Msg(err.Error())
		return err
	}
	if trace {
		connection.TraceMessage(responseBytes, addr, false)
	}
	if _, err := connection.Send(responseBytes, addr); err != nil {
		log.Warn().Msg(err.Error())
		return err
	}
	PfcpMessageTx.WithLabelValues(msg.MessageTypeName()).Inc()
	return nil
}

func (connection *PfcpConnection) TraceMessage(b []byte, addr *net.UDPAddr, rx bool) {
	if rx {
		connection.dumper.DumpRawIn(addr.AddrPort(), connection.nodeAddrV4, b)
	} else {
		connection.dumper.DumpRawOut(addr.AddrPort(), connection.nodeAddrV4, b)
	}
}

func (connection *PfcpConnection) RefreshAssociations() {
	for _, node := range connection.nodes {
		if connection.GetAssociation(node.getAddress()) == nil {
			node.sendAssociationSetupRequest(connection)
		}
	}
}

// DeleteAssociation deletes an association and all sessions associated with it.
func (connection *PfcpConnection) DeleteAssociation(assocAddr string) {
	assoc := connection.GetAssociation(assocAddr)
	log.Info().Msgf("Pruning expired node association: %s", assocAddr)
	for sessionId, session := range assoc.Sessions {
		log.Info().Msgf("Deleting session: %d", sessionId)
		connection.DeleteSession(session)
	}
	delete(connection.NodeAssociations, assocAddr)
}

// DeleteSession deletes a session and all PDRs, FARs and QERs associated with it.
func (connection *PfcpConnection) DeleteSession(session *Session) {
	for _, far := range session.FARs {
		_ = connection.mapOperations.DeleteFar(far.GlobalId)
	}
	for _, qer := range session.QERs {
		_ = connection.mapOperations.DeleteQer(qer.GlobalId)
	}
	pdrContext := NewPDRCreationContext(session, connection.ResourceManager)
	for _, PDR := range session.PDRs {
		_ = pdrContext.deletePDR(PDR, connection.mapOperations)
	}
}

func (connection *PfcpConnection) GetSessionCount() int {
	count := 0
	for _, assoc := range connection.NodeAssociations {
		count += len(assoc.Sessions)
	}
	return count
}

func (connection *PfcpConnection) GetAssiciationCount() int {
	return len(connection.NodeAssociations)
}

func (connection *PfcpConnection) ReleaseResources(seID uint64) {
	if connection.ResourceManager == nil {
		return
	}

	if connection.ResourceManager.IPAM != nil {
		connection.ResourceManager.IPAM.ReleaseIP(seID)
	}

	if connection.ResourceManager.FTEIDM != nil {
		connection.ResourceManager.FTEIDM.ReleaseTEID(seID)
	}
}

func (connection *PfcpConnection) SendReports() {
	for _, assocaition := range connection.NodeAssociations {
		for _, session := range assocaition.Sessions {
			for urrid, urr := range session.URRs {

				newReport, err := connection.mapOperations.GetUrr(urr.GlobalId)
				if err != nil {
					continue
				}

				uplink := newReport.UplinkVolume - urr.UrrInfo.UplinkVolume
				downlink := newReport.DownlinkVolume - urr.UrrInfo.DownlinkVolume

				ies := make([]*ie.IE, 0)

				if pccRuleNames, ok := connection.PCCRulesNameLinkURRID[urrid]; ok && uplink+downlink > 0 {
					if len(pccRuleNames) > 1 {
						log.Error().Uint32("urrID", urrid).Msgf("number of pcc rules referring to urrid is greater than 1")

						continue
					}

					pccRule, ok := connection.PCCRules[pccRuleNames[0]]
					if !ok {
						log.Error().Uint32("urrID", urrid).Str("pccRule", pccRuleNames[0]).Msgf("pcc rule not found")
					}

					if pccRule.RawSDFFilter != "" {
						ies = append(ies, ie.NewVendorSpecificIE(36017, 2011, []byte(pccRule.RawSDFFilter)))
					} else {
						log.Warn().Uint32("urrID", urrid).Msgf("pcc rule not found")
					}
				} else if !(urr.UrrInfo.VolumeThreshold == 0 || urr.UrrInfo.VolumeThreshold > uplink+downlink) {
					// CHOICE
					// urr-type
					//    enterprise-id: ---- 0x7db(2011)
					//    urr-level-type: ---- bearer(2)
					//    urr-function-type: ---- charging(1)
					//    urr-charging-type: ---- offlinepgw(3)
					ies = append(ies, ie.NewVendorSpecificIE(34000, 2011, []byte{0x02, 0x01, 0x03}))

					// CHOICE
					// bearer-sequence
					//    enterprise-id: ---- 0x7db(2011)
					//    bearer-sequence-value: ---- 0x1(1)
					ies = append(ies, ie.NewVendorSpecificIE(32843, 2011, []byte{1}))

					// CHOICE
					// private-stop-time
					//    enterprise-id: ---- 0x7db(2011)
					//    private-stop-time-value: ---- 0x000001917EEC1EEC
					ies = append(ies, ie.NewVendorSpecificIE(34010, 2011, []byte{0x00, 0x00, 0x01, 0x92, 0x1d, 0xca, 0x24, 0x8c}))

					// CHOICE
					// private-time-of-first-packet
					//    enterprise-id: ---- 0x7db(2011)
					//    time-of-first-packet-value: ---- 0x000001917EEC1BE9
					ies = append(ies, ie.NewVendorSpecificIE(34011, 2011, []byte{0x00, 0x00, 0x01, 0x92, 0x1d, 0xca, 0x20, 0xad}))

					// CHOICE
					// private-time-of-last-packet
					//    enterprise-id: ---- 0x7db(2011)
					//    time-of-last-packet-value: ---- 0x000001917EEC1EEC
					ies = append(ies, ie.NewVendorSpecificIE(34012, 2011, []byte{0x00, 0x00, 0x01, 0x92, 0x1d, 0xca, 0x24, 0x8c}))
				} else {
					continue
				}

				traced := session.IsSessionTraced()
				sequence := assocaition.NewSequenceID()
				urr.ReportSeqNumber += 1
				session.URRSequence += 1 //Huawei !!!

				SendSessionReport(connection, session.RemoteSEID, sequence, assocaition.Addr,
					urrid,
					//urr.ReportSeqNumber,
					session.URRSequence, //Huawei
					uplink,
					downlink,
					traced,
					ies...,
				)

				urr.UrrInfo.UplinkVolume = newReport.UplinkVolume
				urr.UrrInfo.DownlinkVolume = newReport.DownlinkVolume
				session.URRs[urrid] = urr
			}

			for pdrID, pdr := range session.PDRs {
				far, ok := session.FARs[pdr.PdrInfo.FarId]
				if !ok {
					log.Debug().Msgf("far in session not found")

					continue
				}

				if far.FarInfo.Action != 0x08 { // todo maxim: const
					continue
				}

				log.Debug().Uint32("farID", far.GlobalId).Msg("far is FAR_NOCP")

				currentFar, err := connection.mapOperations.GetFar(far.GlobalId)
				if err != nil {
					continue
				}

				if currentFar.Trigger == 1 {
					currentFar.Trigger = 2
					err := connection.mapOperations.UpdateFar(far.GlobalId, currentFar)
					if err != nil {
						log.Warn().Uint32("farID", far.GlobalId).Msgf("can`t update far")

						continue
					}

					traced := session.IsSessionTraced()
					sequence := assocaition.NewSequenceID()
					session.URRSequence += 1

					SendDownlinkNotificationReport(connection, session.RemoteSEID, sequence, assocaition.Addr,
						pdrID,
						traced,
					)
				}
			}
		}

	}
}

func (connection *PfcpConnection) GetConnAddr() string {
	return connection.nodeAddrV4.String()
}

func (connection *PfcpConnection) GetAllTracingRecords() ([]tracing.TraceRecord, error) {
	return connection.tracingStorage.GetTraceRecords()
}

func (connection *PfcpConnection) GetTracingRecordByImsi(imsi string) (tracing.TraceRecord, error) {
	return connection.tracingStorage.GetTraceRecordByImsi(imsi)
}

func (connection *PfcpConnection) GetTracingRecordByMsisdn(msisdn string) (tracing.TraceRecord, error) {
	return connection.tracingStorage.GetTraceRecordByMsisdn(msisdn)
}

func (connection *PfcpConnection) NeedSessionTrace(imsi, msisdn string) bool {
	if imsi != "" {
		if _, err := connection.tracingStorage.GetTraceRecordByImsi(imsi); err == nil {
			return true
		}
	}

	if msisdn != "" {
		if _, err := connection.tracingStorage.GetTraceRecordByMsisdn(msisdn); err == nil {
			return true
		}
	}

	return false
}

func (connection *PfcpConnection) EnableTracingByImsi(imsi string) error {
	storage := connection.tracingStorage

	if err := storage.AddTraceRecordByImsi(imsi); err != nil {
		return fmt.Errorf("error adding tracing record: %w", err)
	}

	return nil
}

func (connection *PfcpConnection) EnableTracingByMsisdn(msisdn string) error {
	storage := connection.tracingStorage

	if err := storage.AddTraceRecordByMsisdn(msisdn); err != nil {
		return fmt.Errorf("error adding tracing record: %w", err)
	}

	return nil
}

func (connection *PfcpConnection) DisableTracing() ([]tracing.TraceRecord, error) {
	return connection.tracingStorage.DeleteTraceRecords()
}

func (connection *PfcpConnection) DisableTracingByImsi(imsi string) (tracing.TraceRecord, error) {
	storage := connection.tracingStorage

	if trace, err := storage.DeleteTraceRecordByImsi(imsi); err == nil {
		return trace, nil
	} else {
		return tracing.TraceRecord{}, fmt.Errorf("error deleting tracing record: %w", err)
	}
}

func (connection *PfcpConnection) DisableTracingByMsisdn(msisdn string) (tracing.TraceRecord, error) {
	storage := connection.tracingStorage

	if trace, err := storage.DeleteTraceRecordByMsisdn(msisdn); err == nil {
		return trace, nil
	} else {
		return tracing.TraceRecord{}, fmt.Errorf("error deleting tracing record: %w", err)
	}
}

func SendSessionReport(conn *PfcpConnection, seid uint64, sequenceID uint32, associationAddr string,
	urrid uint32,
	urSeq uint32,
	uplink uint64,
	downlink uint64,
	traced bool,
	vendorSpecificIEs ...*ie.IE,
) {

	additionalIEs := []*ie.IE{
		ie.NewReportType(0, 0, 1, 0),
		ie.NewUsageReportWithinSessionReportRequest(
			ie.NewURRID(urrid),
			ie.NewURSEQN(urSeq),
			ie.NewUsageReportTrigger(1<<1, 0, 0), //Volume Threshold
			ie.NewEndTime(time.Now()),
			ie.NewVolumeMeasurement(0x6, 0, uplink, downlink, 0, 0, 0),
			ie.NewTimeOfFirstPacket(time.Now()),
			ie.NewTimeOfLastPacket(time.Now()),
		),
	}

	additionalIEs = append(additionalIEs, vendorSpecificIEs...)

	sessionReport := message.NewSessionReportRequest(0, 0, seid, sequenceID, 0, additionalIEs...)
	log.Debug().Msgf("Sent Session Report Request to: %s", associationAddr)
	udpAddr, err := net.ResolveUDPAddr("udp", associationAddr+":8805")
	if err == nil {
		if err := conn.SendMessageWithTrace(sessionReport, udpAddr, traced); err != nil {
			log.Info().Msgf("Failed to send Session Report Request: %s\n", err.Error())
		}
	} else {
		log.Info().Msgf("Failed to send Session Report Request: %s\n", err.Error())
	}
}

func SendDownlinkNotificationReport(
	conn *PfcpConnection,
	seid uint64,
	sequenceID uint32,
	associationAddr string,
	pdrID uint32,
	traced bool,
	vendorSpecificIEs ...*ie.IE,
) {
	if pdrID > math.MaxUint16 {
		log.Warn().Msgf("PDR ID too large (%d)", pdrID)

		return
	}

	additionalIEs := []*ie.IE{
		ie.NewReportType(0, 0, 0, 1),
		ie.NewDownlinkDataReport(
			ie.NewPDRID(uint16(pdrID)),
		),
	}

	additionalIEs = append(additionalIEs, vendorSpecificIEs...)

	sessionReport := message.NewSessionReportRequest(0, 0, seid, sequenceID, 0, additionalIEs...)
	log.Debug().Msgf("Sent DLDR Session Report Request to: %s", associationAddr)
	udpAddr, err := net.ResolveUDPAddr("udp", associationAddr+":8805")
	if err == nil {
		if err := conn.SendMessageWithTrace(sessionReport, udpAddr, traced); err != nil {
			log.Info().Msgf("Failed to send Session Report Request: %s\n", err.Error())
		}
	} else {
		log.Info().Msgf("Failed to send Session Report Request: %s\n", err.Error())
	}
}

type DefaultAssociationConnector struct {
	address string
}

func NewDefaultAssociationConnector(address string) (*DefaultAssociationConnector, error) {
	resolvedAddr, err := net.ResolveIPAddr("ip", address)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve remote node address %s: %w", address, err)
	}

	return &DefaultAssociationConnector{
		address: resolvedAddr.String(),
	}, nil
}

func (connector *DefaultAssociationConnector) getAddress() string {
	return connector.address
}

func (connector *DefaultAssociationConnector) sendAssociationSetupRequest(connection *PfcpConnection) {

	associationAddr := connector.getAddress()
	AssociationSetupRequest := message.NewAssociationSetupRequest(0,
		newIeNodeID(connection.nodeId),
		ie.NewRecoveryTimeStamp(connection.RecoveryTimestamp),
		ie.NewUPFunctionFeatures(connection.featuresOctets[:]...),
	)
	log.Info().Msgf("Sent Association Setup Request to: %s", associationAddr)

	udpAddr, err := net.ResolveUDPAddr("udp", associationAddr+":8805")
	if err != nil {
		log.Error().Msgf("Failed to resolve udp address from PFCP peer address %s. Error: %s\n", associationAddr, err.Error())
		return
	}
	if err := connection.SendMessageWithTrace(AssociationSetupRequest, udpAddr, true); err != nil {
		log.Info().Msgf("Failed to send Association Setup Request: %s\n", err.Error())
	}
}

type SxaAssociationConnector struct {
	address     string
	s1uAddress  string
	s5s8Address string
}

func NewSxaAssociationConnector(address string, s1uAddress string, s5s8Address string) (*SxaAssociationConnector, error) {
	resolvedAddr, err := net.ResolveIPAddr("ip", address)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve remote node address %s: %w", address, err)
	}

	return &SxaAssociationConnector{
		address:     resolvedAddr.String(),
		s1uAddress:  s1uAddress,
		s5s8Address: s5s8Address,
	}, nil
}

func (connector *SxaAssociationConnector) getAddress() string {
	return connector.address
}

func (connector *SxaAssociationConnector) sendAssociationSetupRequest(connection *PfcpConnection) {

	featuresOctets := []uint8{0, 0}
	featuresOctets[0] = setBit(featuresOctets[0], 1)
	featuresOctets[0] = setBit(featuresOctets[0], 2)
	featuresOctets[0] = setBit(featuresOctets[0], 6)
	featuresOctets[0] = setBit(featuresOctets[0], 7)

	s1uIP := net.ParseIP(connector.s1uAddress)
	if s1uIP == nil {
		log.Error().Msgf("failed to parse S1-U IP address ID: %s", connector.s1uAddress)
		return
	}

	s5s8IP := net.ParseIP(connector.s5s8Address)
	if s5s8IP == nil {
		log.Error().Msgf("failed to parse S1/S8 IP address ID: %s", connector.s5s8Address)
		return
	}

	ipSuiteName := "0001" + connection.nodeId
	ipsuitInfo := make([]byte, 0)
	ipsuitInfo = append(ipsuitInfo, 0xA8, 0x00)
	ipsuitInfo = append(ipsuitInfo, (byte)(len(ipSuiteName)))
	ipsuitInfo = append(ipsuitInfo, ipSuiteName...)
	ipsuitInfo = append(ipsuitInfo, s1uIP.To4()...)
	ipsuitInfo = append(ipsuitInfo, s5s8IP.To4()...)
	ipsuitInfo = append(ipsuitInfo, s1uIP.To4()...)

	associationAddr := connector.getAddress()
	AssociationSetupRequest := message.NewAssociationSetupRequest(0,
		newIeNodeIDHuawei(connection.nodeId),
		ie.NewRecoveryTimeStamp(connection.RecoveryTimestamp),
		ie.NewUPFunctionFeatures(featuresOctets[:]...),
		//CHOICE
		// 	ipsuit-info
		// 	enterprise-id: ---- 0x7db(2011)
		// 	s11uIpv4Valid: ---- 0x1(1)
		// 	s11uIpv6Valid: ---- 0x0(0)
		// 	s1uIpv4Valid: ---- 0x1(1)
		// 	s1uIpv6Valid: ---- 0x0(0)
		// 	s5S8Ipv4Valid: ---- 0x1(1)
		// 	s5S8Ipv6Valid: ---- 0x0(0)
		// 	paIpv4Valid: ---- 0x0(0)
		// 	paIpv6Valid: ---- 0x0(0)
		// 	lock-flag: ---- 0x0(0)
		// 	ipsuit-name: ---- 0001dgw1
		// 	s1u-ip-address
		// 		ipv4-address
		// 			uladdr1: ---- 0xa(10)
		// 			uladdr2: ---- 0xa9(169)
		// 			uladdr3: ---- 0x70(112)
		// 			uladdr4: ---- 0x80(128)
		// 	s5s8s-ip-address
		// 		ipv4-address
		// 			uladdr1: ---- 0xa(10)
		// 			uladdr2: ---- 0xa9(169)
		// 			uladdr3: ---- 0x70(112)
		// 			uladdr4: ---- 0x91(145)
		// 	s11u-ip-address
		// 		ipv4-address
		// 			uladdr1: ---- 0xa(10)
		// 			uladdr2: ---- 0xa9(169)
		// 			uladdr3: ---- 0x70(112)
		// 			uladdr4: ---- 0x8a(138)
		ie.NewVendorSpecificIE(32787, 2011, ipsuitInfo),
		//CHOICE
		//	user-plane-element-weight
		//		enterprise-id: ---- 0x7db(2011)
		//		weight-value: ---- 0x1(1)
		ie.NewVendorSpecificIE(32803, 2011, []byte{1}),
		//CHOICE
		//	lock-information
		//		enterprise-id: ---- 0x7db(2011)
		//		lock-information-value: ---- 0x0(0)
		ie.NewVendorSpecificIE(32806, 2011, []byte{0}),
		//CHOICE
		//	apn-support-mode
		//		enterprise-id: ---- 0x7db(2011)
		//		apn-support-mode-value: ---- 0x0(0)
		ie.NewVendorSpecificIE(32857, 2011, []byte{0}),
		//CHOICE
		//	sx-uf-flag
		//		enterprise-id: ---- 0x7db(2011)
		//		spare: ---- 0x0(0)
		//		nb-iot-value: ---- 0x1(1)
		//		dual-connectivity-with-nr-value: ---- 0x1(1)
		ie.NewVendorSpecificIE(32900, 2011, []byte{3}),
		//CHOICE
		//	high-bandwidth
		//		enterprise-id: ---- 0x7db(2011)
		//		high-bandwidth-value: ---- 0x1(1)
		ie.NewVendorSpecificIE(32901, 2011, []byte{1}),
	)
	log.Info().Msgf("Sent Association Setup Request to: %s", associationAddr)

	udpAddr, err := net.ResolveUDPAddr("udp", associationAddr+":8805")
	if err != nil {
		log.Error().Msgf("Failed to resolve udp address from PFCP peer address %s. Error: %s\n", associationAddr, err.Error())
		return
	}
	if err := connection.SendMessageWithTrace(AssociationSetupRequest, udpAddr, true); err != nil {
		log.Info().Msgf("Failed to send Association Setup Request: %s\n", err.Error())
	}
}

type SxbAssociationConnector struct {
	address   string
	paAddress string
}

func NewSxbAssociationConnector(address string, paAddress string) (*SxbAssociationConnector, error) {
	resolvedAddr, err := net.ResolveIPAddr("ip", address)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve remote node address %s: %w", address, err)
	}

	return &SxbAssociationConnector{
		address:   resolvedAddr.String(),
		paAddress: paAddress,
	}, nil
}

func (connector *SxbAssociationConnector) getAddress() string {
	return connector.address
}

func (connector *SxbAssociationConnector) sendAssociationSetupRequest(connection *PfcpConnection) {

	featuresOctets := []uint8{0, 0}
	featuresOctets[0] = setBit(featuresOctets[0], 1)
	featuresOctets[0] = setBit(featuresOctets[0], 2)
	featuresOctets[0] = setBit(featuresOctets[0], 6)
	featuresOctets[0] = setBit(featuresOctets[0], 7)

	paIP := net.ParseIP(connector.paAddress)
	if paIP == nil {
		log.Error().Msgf("failed to parse PA IP address ID: %s", connector.paAddress)
		return
	}

	pSuiteName := "0001" + connection.nodeId
	ipsuitInfo := make([]byte, 0)
	ipsuitInfo = append(ipsuitInfo, 0x02, 0x00)
	ipsuitInfo = append(ipsuitInfo, (byte)(len(ipSuiteName)))
	ipsuitInfo = append(ipsuitInfo, ipSuiteName...)
	ipsuitInfo = append(ipsuitInfo, paIP.To4()...)

	associationAddr := connector.getAddress()
	AssociationSetupRequest := message.NewAssociationSetupRequest(0,
		newIeNodeIDHuawei(connection.nodeId),
		ie.NewRecoveryTimeStamp(connection.RecoveryTimestamp),
		//ie.NewUPFunctionFeatures(connection.featuresOctets[:]...),
		ie.NewUPFunctionFeatures(featuresOctets[:]...),
		//CHOICE
		//	ipsuit-info
		//	enterprise-id: ---- 0x7db(2011)
		//	s11uIpv4Valid: ---- 0x0(0)
		//	s11uIpv6Valid: ---- 0x0(0)
		//	s1uIpv4Valid: ---- 0x0(0)
		//	s1uIpv6Valid: ---- 0x0(0)
		//	s5S8Ipv4Valid: ---- 0x0(0)
		//	s5S8Ipv6Valid: ---- 0x0(0)
		//	paIpv4Valid: ---- 0x1(1)
		//	paIpv6Valid: ---- 0x0(0)
		//	lock-flag: ---- 0x0(0)
		//	ipsuit-name: ---- 0001dgw1
		//	pa-ip-address
		//		ipv4-address
		//			uladdr1: ---- 0xa(10)
		//			uladdr2: ---- 0xa9(169)
		//			uladdr3: ---- 0x70(112)
		//			uladdr4: ---- 0x83(131)
		ie.NewVendorSpecificIE(32787, 2011, ipsuitInfo),
		//CHOICE
		//	user-plane-element-weight
		//		enterprise-id: ---- 0x7db(2011)
		//		weight-value: ---- 0x1(1)
		ie.NewVendorSpecificIE(32803, 2011, []byte{1}),
		//CHOICE
		//	lock-information
		//		enterprise-id: ---- 0x7db(2011)
		//		lock-information-value: ---- 0x0(0)
		ie.NewVendorSpecificIE(32806, 2011, []byte{0}),
		//CHOICE
		//	apn-support-mode
		//		enterprise-id: ---- 0x7db(2011)
		//		apn-support-mode-value: ---- 0x0(0)
		ie.NewVendorSpecificIE(32857, 2011, []byte{0}),
		//CHOICE
		//	sx-uf-flag
		//		enterprise-id: ---- 0x7db(2011)
		//		spare: ---- 0x0(0)
		//		nb-iot-value: ---- 0x1(1)
		//		dual-connectivity-with-nr-value: ---- 0x1(1)
		ie.NewVendorSpecificIE(32900, 2011, []byte{3}),
		//CHOICE
		//	high-bandwidth
		//		enterprise-id: ---- 0x7db(2011)
		//		high-bandwidth-value: ---- 0x1(1)
		ie.NewVendorSpecificIE(32901, 2011, []byte{1}),
	)
	log.Info().Msgf("Sent Association Setup Request to: %s", associationAddr)

	udpAddr, err := net.ResolveUDPAddr("udp", associationAddr+":8805")
	if err != nil {
		log.Error().Msgf("Failed to resolve udp address from PFCP peer address %s. Error: %s\n", associationAddr, err.Error())
		return
	}
	if err := connection.SendMessageWithTrace(AssociationSetupRequest, udpAddr, true); err != nil {
		log.Info().Msgf("Failed to send Association Setup Request: %s\n", err.Error())
	}
}

type NeValidator struct {
	reAllowed *regexp.Regexp
	reDenied  *regexp.Regexp
}

func NewNeValidator(allowed string, denied string) (*NeValidator, error) {

	reAllowed, err := regexp.Compile(allowed)
	if err != nil {
		return nil, fmt.Errorf("network instance validator allowed pattern is invalid: %w", err)
	}

	reDenied, err := regexp.Compile(denied)
	if err != nil {
		return nil, fmt.Errorf("network instance validator denied pattern is invalid: %w", err)
	}

	return &NeValidator{reAllowed: reAllowed, reDenied: reDenied}, nil
}

func (v *NeValidator) IsNeAllowed(networkInstance string) bool {
	return len(v.reAllowed.String()) == 0 || v.reAllowed.MatchString(networkInstance)
}

func (v *NeValidator) IsNeDenied(networkInstance string) bool {
	return len(v.reDenied.String()) > 0 && v.reDenied.MatchString(networkInstance)
}

func (v *NeValidator) Validate(networkInstance string) bool {
	return v.IsNeAllowed(networkInstance) && !v.IsNeDenied(networkInstance)
}
