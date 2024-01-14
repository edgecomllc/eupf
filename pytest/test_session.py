#!/usr/bin/env python3

from scapy.all import *
from scapy.contrib.pfcp import *
from scapy.layers.inet import IP  # This is to calm down PyCharm's linter
import time

association_request = PFCP(version=1, S=0, seq=1) / \
                      PFCPAssociationSetupRequest(IE_list=[
                          IE_RecoveryTimeStamp(timestamp=3785653512),
                          IE_NodeId(id_type="FQDN", id="BIG-IMPORTANT-CP")
                      ])

session_establish = PFCP(version=1, S=1, seq=2, seid=0, spare_oct=0) / \
                    PFCPSessionEstablishmentRequest(IE_list=[
                        IE_CreateFAR(IE_list=[
                            IE_ApplyAction(FORW=1),
                            IE_FAR_Id(id=1),
                            IE_ForwardingParameters(IE_list=[
                                IE_DestinationInterface(interface="Access"),
                                IE_NetworkInstance(instance="access"),
                                IE_OuterHeaderCreation(GTPUUDPIPV4=1, TEID=0x01000000, ipv4="10.23.118.70"),
                            ])
                        ]),
                        IE_CreateFAR(IE_list=[
                            IE_ApplyAction(DROP=1),
                            IE_FAR_Id(id=2)
                        ]),
                        IE_CreatePDR(IE_list=[
                            IE_FAR_Id(id=1),
                            IE_OuterHeaderRemoval(header="GTP-U/UDP/IPv4"),
                            IE_PDI(IE_list=[
                                IE_FTEID(V4=1, TEID=0x104c9033, ipv4="172.18.1.2"),
                                IE_NetworkInstance(instance="access"),
                                IE_SourceInterface(interface="Access"),
                            ]),
                            IE_PDR_Id(id=1),
                            IE_Precedence(precedence=100)
                        ]),
                        IE_FSEID(v4=1, seid=0xffde7230bf97810a, ipv4="172.18.1.1"),
                        IE_NodeId(id_type="FQDN", id="BIG-IMPORTANT-CP")
                    ])

session_modification = PFCP(version=1, S=1, seq=2, seid=2, spare_oct=0) / \
                       PFCPSessionModificationRequest(IE_list=[
                           IE_UpdateFAR(IE_list=[
                               IE_ApplyAction(FORW=1),
                               IE_FAR_Id(id=2),
                               IE_UpdateForwardingParameters(IE_list=[
                                   IE_DestinationInterface(interface="Access"),
                                   IE_NetworkInstance(instance="access"),
                                   IE_OuterHeaderCreation(GTPUUDPIPV4=1, TEID=0x01000001, ipv4="10.23.118.69"),
                               ])
                           ]),
                           IE_RemoveFAR(IE_list=[
                               IE_ApplyAction(DROP=1),
                               IE_FAR_Id(id=1)
                           ]),
                           IE_UpdatePDR(IE_list=[
                               IE_FAR_Id(id=1),
                               IE_OuterHeaderRemoval(header="GTP-U/UDP/IPv4"),
                               IE_PDI(IE_list=[
                                   IE_FTEID(V4=1, TEID=0x104c9033, ipv4="172.18.1.2"),
                                   IE_NetworkInstance(instance="access"),
                                   IE_SourceInterface(interface="Access"),
                               ]),
                               IE_PDR_Id(id=1),
                               IE_Precedence(precedence=100)
                           ]),
                           IE_FSEID(v4=1, seid=0xffde7230bf97810a, ipv4="172.18.1.1"),
                           IE_NodeId(id_type="FQDN", id="BIG-IMPORTANT-CP")
                       ])

session_delete = PFCP(version=1, S=1, seq=3, seid=2, spare_oct=0) / \
                 PFCPSessionDeletionRequest(IE_list=[
                     IE_FSEID(v4=1, seid=0xffde7230bf97810a, ipv4="172.18.1.1"),
                     IE_NodeId(id_type="FQDN", id="BIG-IMPORTANT-CP")
                 ])

heartbeat_response = PFCP(version=1, S=0, seq=1, seid=2, spare_oct=0) / \
                     PFCPHeartbeatResponse(IE_list=[
                         IE_RecoveryTimeStamp(timestamp=int(time.time()))
                     ])

ue_ip_address = IE_UE_IP_Address(spare=2, SD=0, V4=0)
session_establish_ueip = PFCP(version=1, S=1, seq=2, seid=0, spare_oct=0) / \
                         PFCPSessionEstablishmentRequest(IE_list=[
                             IE_CreatePDR(IE_list=[
                                 IE_FAR_Id(id=1),
                                 IE_OuterHeaderRemoval(header="GTP-U/UDP/IPv4"),
                                 IE_PDI(IE_list=[
                                     ue_ip_address,
                                     # IE_NetworkInstance(instance="access"),
                                     IE_SourceInterface(interface="Access"),
                                 ]),
                                 IE_PDR_Id(id=1),
                                 IE_Precedence(precedence=100)
                             ]),
                             IE_FSEID(v4=1, seid=0xffde7230bf97810a, ipv4="172.18.1.1"),
                             IE_NodeId(id_type="FQDN", id="BIG-IMPORTANT-CP")
                         ])

# https://stackoverflow.com/questions/41166420/sending-a-packet-over-physical-loopback-in-scapy
conf.L3socket = L3RawSocket

target = IP(dst="127.0.0.1") / UDP(sport=33100, dport=8805)


# TODO: Add state checks via eUPF web API

def test_create_association():
    ans = sr1(target / association_request, iface='lo')
    assert ans.haslayer(PFCPAssociationSetupResponse)
    assert ans[PFCPAssociationSetupResponse][IE_Cause].cause == 1


def test_create_session():
    ans = sr1(target / session_establish, iface='lo')
    assert ans.haslayer(PFCPSessionEstablishmentResponse)
    assert ans[PFCPSessionEstablishmentResponse][IE_Cause].cause == 1


def test_modify_session():
    ans = sr1(target / session_modification, iface='lo')
    assert ans.haslayer(PFCPSessionModificationResponse)
    assert ans[PFCPSessionModificationResponse][IE_Cause].cause == 1


def test_delete_session():
    ans = sr1(target / session_delete, iface='lo')
    assert ans.haslayer(PFCPSessionDeletionResponse)
    assert ans[PFCPSessionDeletionResponse][IE_Cause].cause == 1


def test_send_heartbeat():
    # This is imaginary HearBeatResponse, this should not crash eUPF
    send(target / heartbeat_response, iface='lo')


def test_create_session_ueip():
    ans = sr1(target / session_establish_ueip, iface='lo')
    assert ans.haslayer(PFCPSessionEstablishmentResponse)
    assert ans[PFCPSessionEstablishmentResponse][IE_CreatedPDR][IE_UE_IP_Address].ipv4 == "10.60.0.1"