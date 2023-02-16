#! /usr/bin/env python3

from scapy.all import Ether, IP, UDP, srp1, sendp, sr, sr1, fuzz
from scapy.contrib.pfcp import *

# TODO: Understand why packet only visible in tcpdump but not reacing software

def test_hearbeat():
    pfcp_heartbeat = Ether() / IP(dst="127.0.0.1") / \
        fuzz(UDP(dport=8805) / PFCP() /
             PFCPHeartbeatRequest(IE_list=[
                 IE_RecoveryTimeStamp()
             ]))
    print(pfcp_heartbeat.show())
    ans = srp1(pfcp_heartbeat, inter=1,  timeout=5, iface='lo')
    print(ans)


test_hearbeat()
