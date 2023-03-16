#!/usr/bin/env python3

from scapy.all import *
from scapy.contrib.pfcp import *
import socket
import time

pfcpASReq = PFCP(version=1, S=0, seq=1) / \
  PFCPAssociationSetupRequest(IE_list=[
      IE_RecoveryTimeStamp(timestamp=3785653512),
      IE_NodeId(id_type="FQDN", id="BIG-IMPORTANT-CP")
  ])

pfcpSESReq = PFCP(version=1, S=1, seq=2, seid=0, spare_oct=0) / \
     PFCPSessionEstablishmentRequest(IE_list=[
         IE_CreateFAR(IE_list=[
             IE_ApplyAction(FORW=1),
             IE_FAR_Id(id=2),
             IE_ForwardingParameters(IE_list=[
                 IE_DestinationInterface(interface="Access"),
                 IE_NetworkInstance(instance="access"),
             ])
         ]),
         IE_CreateFAR(IE_list=[
             IE_ApplyAction(DROP=1),
             IE_FAR_Id(id=1)
         ]),
         IE_CreatePDR(IE_list=[
             IE_FAR_Id(id=2),
             IE_OuterHeaderRemoval(header="GTP-U/UDP/IPv4"),
             IE_PDI(IE_list=[
                 IE_FTEID(V4=1, TEID=0x104c9033, ipv4="172.18.1.2"),
                 IE_NetworkInstance(instance="cp"),
                 IE_SourceInterface(interface="CP-function"),
             ]),
             IE_PDR_Id(id=2),
             IE_Precedence(precedence=100)
         ]),
         IE_CreatePDR(IE_list=[
             IE_FAR_Id(id=1),
             IE_PDI(IE_list=[
                 IE_SDF_Filter(FD=1, flow_description="permit out ip from any to any"),
                 IE_NetworkInstance(instance="access"),
                 IE_SourceInterface(interface="Access"),
             ]),
             IE_PDR_Id(id=1),
             IE_Precedence(precedence=65000),
             IE_URR_Id(id=1)
         ]),
         IE_CreateURR(IE_list=[
             IE_MeasurementMethod(EVENT=1),
             IE_ReportingTriggers(start_of_traffic=1),
             IE_TimeQuota(quota=60),
             IE_URR_Id(id=1)
         ]),
         IE_FSEID(v4=1, seid=0xffde7230bf97810a, ipv4="172.18.1.1"),
         IE_NodeId(id_type="FQDN", id="BIG-IMPORTANT-CP")
       ])

pfcpSESReqNoSDF = PFCP(version=1, S=1, seq=2, seid=0, spare_oct=0) / \
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


# https://stackoverflow.com/questions/41166420/sending-a-packet-over-physical-loopback-in-scapy
conf.L3socket=L3RawSocket

target = IP(dst="127.0.0.1")/UDP(sport=33100,dport=8805)

# print ("Sending PFCP Session Setup Request")
# ans = sr1(target/pfcpSESReq, iface='lo')
# print(ans.show())

# print ("Sending PFCP Association Establishment Request")
# ans = sr1(target/pfcpASReq, iface='lo')
# print(ans.show())

# print ("Sending PFCP Session Establishment Request")
# ans = sr1(target/pfcpSESReq, iface='lo')
# print(ans.show())

# pfcpASReq_borked = PFCP(version=1, S=0, seq=1) / \
#   PFCPAssociationSetupRequest(IE_list=[
#       IE_RecoveryTimeStamp(timestamp=3785653512),      
#   ])

print ("Sending PFCP Association Setup Request")
ans = sr1(target/pfcpASReq, iface='lo')
print(ans.show())

print ("Sending PFCP Association Setup Request without SDF")
ans = sr1(target/pfcpSESReqNoSDF, iface='lo')
print(ans.show())