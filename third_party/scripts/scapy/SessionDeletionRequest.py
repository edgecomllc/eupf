#!/usr/bin/env python3

from scapy.all import *
from scapy.contrib.pfcp import *
import socket
import time

pfcp_session_delete = PFCP(version=1, S=1, seq=2, seid=2, spare_oct=0) / \
     PFCPSessionDeletionRequest(IE_list=[
         IE_FSEID(v4=1, seid=0xffde7230bf97810a, ipv4="172.18.1.1"),
         IE_NodeId(id_type="FQDN", id="BIG-IMPORTANT-CP")
       ])

# https://stackoverflow.com/questions/41166420/sending-a-packet-over-physical-loopback-in-scapy
conf.L3socket=L3RawSocket

target = IP(dst="127.0.0.1")/UDP(sport=33100,dport=8805)

ans = sr1(target/pfcp_session_delete, iface='lo')
print(ans.show())
