#!/usr/bin/env python3

from scapy.all import *
from scapy.contrib.gtp import *
import socket
import time

# https://stackoverflow.com/questions/41166420/sending-a-packet-over-physical-loopback-in-scapy
conf.L3socket=L3RawSocket

ping = Ether()/IP(dst="1.1.1.1", src="2.2.2.2")/UDP(sport=2152,dport=2152)/GTP_U_Header(seq=12345, teid=2)/IPv6(dst="ff02::2", src="fe80::1:0:1f86:9ec4")/ICMPv6ND_RS()
print(ping.show())
ans = srp1(ping, iface='lo')
print(ans.show())