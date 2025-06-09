package core

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/cilium/ebpf/link"
	"github.com/edgecomllc/eupf/cmd/config"
	"github.com/edgecomllc/eupf/cmd/ebpf"
	"github.com/rs/zerolog/log"

	"github.com/rs/zerolog"
)

const (
	N4PFCPKeyName  = "n4"
	SxaPFCPKeyName = "Sxa"
	SxbPFCPKeyName = "Sxb"

	GTPPortStr = ":2152"
)

var (
	ErrNilBPFObj         = errors.New("BpfObjects is nil")
	ErrNilGtpPathManager = errors.New("GtpPathManager is nil")
	ErrNilInterface      = errors.New("interface is nil")
	ErrAttachXDP         = errors.New("could not attach XDP program")
	ErrLookupIface       = errors.New("lookup network iface failed")
	ErrCloseOldLink      = errors.New("close old link failed")
	ErrNonePFCPConn      = errors.New("none PFCP connection")
)

func joinErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}

	var messages []string

	for _, err := range errs {
		messages = append(messages, err.Error())
	}

	return errors.New(strings.Join(messages, "; "))
}

func UpdateLogLevel(logLevel string) error {
	if err := SetLoggerLevel(logLevel); err != nil {
		return fmt.Errorf("Logger configuring error: %s. Using '%s' level",
			err.Error(), zerolog.GlobalLevel().String())
	}

	return nil
}

func UpdateLogCaller(enableCaller bool) {
	SetLoggerCaller(enableCaller)
}

func BindDataPlaneInterfaces(
	bpfObjects *ebpf.BpfObjects,
	links *[]link.Link,
	interfaceNames []string,
	xdpAttachMode string,
) error {
	if bpfObjects == nil {
		log.Error().Msg("can't update InterfaceName: BpfObjects is nil")
		return ErrNilBPFObj
	}

	var errList []error
	oldLinks := *links
	newLinks := make([]link.Link, 0)
	existingLinksMap := make(map[link.ID]struct{})

	for _, l := range oldLinks {
		info, err := l.Info()
		if err != nil {
			log.Error().Msgf("Could not get link info: %s", err)
			continue
		}
		existingLinksMap[info.ID] = struct{}{}
	}

	for _, ifaceName := range interfaceNames {
		iface, err := net.InterfaceByName(ifaceName)
		if err != nil {
			log.Error().Msgf("Lookup network iface %q: %s", ifaceName, err)
			errList = append(errList, fmt.Errorf("%w: %q", ErrLookupIface, ifaceName))
			continue
		}

		if _, exists := existingLinksMap[link.ID(iface.Index)]; exists {
			continue
		}

		l, err := link.AttachXDP(link.XDPOptions{
			Program:   bpfObjects.IpEntrypointObjects.UpfIpEntrypointFunc,
			Interface: iface.Index,
			Flags:     StringToXDPAttachMode(xdpAttachMode),
		})
		if err != nil {
			log.Error().Msgf("Could not attach XDP program to %q: %s", ifaceName, err.Error())
			errList = append(errList, fmt.Errorf("%w: %s", ErrAttachXDP, err))
			continue
		}

		newLinks = append(newLinks, l)
		log.Info().Msgf("Attached XDP program to iface %q (index %d)", iface.Name, iface.Index)
	}

	for _, l := range oldLinks {
		info, err := l.Info()
		if err != nil {
			log.Error().Msgf("Could not get link info for closing: %s", err)
			continue
		}

		if _, exists := existingLinksMap[info.ID]; !exists {
			if err := l.Close(); err != nil {
				log.Error().Msgf("Could not close old link %v: %s", info.ID, err)
				errList = append(errList, fmt.Errorf("%w: %s", ErrCloseOldLink, err))
			} else {
				log.Info().Msgf("Closed old link %v", info.ID)
			}
		}
	}

	*links = newLinks
	return joinErrors(errList)
}

func UpdateDataPlaneAddresses(
	pfcpConnection map[string]*PfcpConnection,
	gtpPathManager *GtpPathManager,
	n3Addr net.IP,
	n9Addr net.IP,
) error {
	if gtpPathManager == nil {
		return ErrNilGtpPathManager
	}

	gtpPathManager.UpdateLocalAddress(n3Addr.String() + GTPPortStr)

	for _, conn := range pfcpConnection {
		conn.UpdateN3Address(n3Addr)
		conn.UpdateN9Address(n9Addr)

	}

	return nil
}

func UpdatePFCPConnections(
	pfcpSrv map[string]*PfcpConnection,
	keyName string,
	address string,
	nodeID string,
	remoteNodes []string,
) error {
	conn, ok := pfcpSrv[keyName]
	if ok {
		if err := conn.Update(address, nodeID); err != nil {
			return err
		}
	} else {
		return ErrNonePFCPConn
	}
	if conn == nil {
		return ErrNonePFCPConn
	}

	switch keyName {
	case N4PFCPKeyName:
		remotePFCPNodes := make([]AssociationConnector, 0, len(remoteNodes))
		for _, remoteNode := range remoteNodes {
			connector, err := NewDefaultAssociationConnector(remoteNode)
			if err != nil {
				return fmt.Errorf("failed to create default association connector: %w", err)
			}
			remotePFCPNodes = append(remotePFCPNodes, connector)
		}

		conn.SetRemoteNodes(remotePFCPNodes)
	case SxaPFCPKeyName:
		remotePFCPNodes := make([]AssociationConnector, 0, len(remoteNodes))
		for _, remoteNode := range remoteNodes {
			connector, err := NewSxaAssociationConnector(remoteNode, config.Conf.S1UAddress, config.Conf.S5S8Address)
			if err != nil {
				return fmt.Errorf("failed to create sxa association connector: %w", err)
			}
			remotePFCPNodes = append(remotePFCPNodes, connector)
		}

		conn.SetRemoteNodes(remotePFCPNodes)
	case SxbPFCPKeyName:
		remotePFCPNodes := make([]AssociationConnector, 0, len(remoteNodes))
		for _, remoteNode := range remoteNodes {
			connector, err := NewSxbAssociationConnector(remoteNode, config.Conf.PAAddress)
			if err != nil {
				return fmt.Errorf("failed to create sxb association connector: %w", err)
			}
			remotePFCPNodes = append(remotePFCPNodes, connector)
		}

		conn.SetRemoteNodes(remotePFCPNodes)
	}

	return nil
}

func UpdateAssociationSetupTimeout(
	pfcpSrv map[string]*PfcpConnection,
	associationSetupTimeout uint32,
) {
	for _, connection := range pfcpSrv {
		connection.AssociationSetupTicker = time.NewTicker(time.Duration(associationSetupTimeout) * time.Second)
	}
}

func UpdateHeartbeatTimeout(
	pfcpSrv map[string]*PfcpConnection,
	heartbeatTimeout uint32,
) {
	for _, connection := range pfcpSrv {
		for _, nodeAssociation := range connection.NodeAssociations {
			nodeAssociation.HeartbeatTimeout.Reset(time.Duration(heartbeatTimeout) * time.Second)
		}
	}
}

func UpdateGTPManager(
	gtpPathManager *GtpPathManager,
	gtpPeerAddresses []string,
	GtpEchoInterval uint32,
) {
	gtpPathManager.UpdateGtpPath(gtpPeerAddresses)
	gtpPathManager.UpdateCheckTicker(time.NewTicker(time.Duration(GtpEchoInterval) * time.Second))
}

func StringToXDPAttachMode(Mode string) link.XDPAttachFlags {
	switch Mode {
	case "generic":
		return link.XDPGenericMode
	case "native":
		return link.XDPDriverMode
	case "offload":
		return link.XDPOffloadMode
	default:
		return link.XDPGenericMode
	}
}
