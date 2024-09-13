#!/usr/bin/env python3

from scapy.all import *
from scapy.contrib.gtp import *
import socket
import time

# https://stackoverflow.com/questions/41166420/sending-a-packet-over-physical-loopback-in-scapy
conf.L3socket=L3RawSocket

#ping = IP(dst="127.0.0.1")/UDP(sport=2152,dport=2152)/GTPHeader(seq=12345)/GTPEchoRequest()/IE_Recovery()
#print(ping.show())
#ans = sr1(ping, iface='lo')
#ans = sr1(ping, iface='lo')
#print(ans.show())

ping = Ether()/(IP(dst="127.0.0.1")/UDP(sport=2152,dport=2152)/GTPHeader(seq=12345)/GTPEchoRequest())/Padding(b"\x00\x00\x00\x00\x00\x00")
print(ping.show())
ans = srp1(ping, iface='lo')
print(ans.show())
