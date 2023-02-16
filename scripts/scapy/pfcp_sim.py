#! /usr/bin/env python3

from scapy.all import Ether, IP, UDP, srp1, sendp, sr, sr1, fuzz
from scapy.contrib.pfcp import *


def test_hearbeat():
    pfcp_heartbeat = IP(dst="localhost") / \
        fuzz(UDP(dport=8805) / PFCP(version=1, S=0, seq=3) /
             PFCPHeartbeatRequest(IE_list=[
                 IE_RecoveryTimeStamp()
             ]))
    print(pfcp_heartbeat.show())
    # ans = sr1(pfcp_heartbeat*10,  inter=1,  timeout=5)
    ans = srp1(pfcp_heartbeat,  timeout=5, iface='lo')
    print(ans)


test_hearbeat()
