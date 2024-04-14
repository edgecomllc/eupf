#!/bin/sh

# flush rules
iptables -F
iptables -X
iptables -t raw -F
iptables -t raw -X
iptables -t nat -F
iptables -t nat -X
iptables -t mangle -F
iptables -t mangle -X

# 001 default policies
iptables -P INPUT DROP
iptables -P OUTPUT ACCEPT
iptables -P FORWARD DROP

# 002 allow loopback
iptables -A INPUT -i lo -s 127.0.0.0/8 -d 127.0.0.0/8 -j ACCEPT

# 003 allow ping replies
iptables -A INPUT -p icmp --icmp-type echo-request -j ACCEPT
iptables -A OUTPUT -p icmp --icmp-type echo-reply -j ACCEPT

# INPUT (all)
iptables -A INPUT -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT
iptables -A INPUT -m conntrack --ctstate NEW -p tcp --dport 22 -j ACCEPT
iptables -A INPUT -m conntrack --ctstate NEW -p tcp --dport 3000 -j ACCEPT
# iperf
iptables -A INPUT -m conntrack --ctstate NEW -p tcp --dport 5201 -j ACCEPT
iptables -A INPUT -p udp --dport 5201 -j ACCEPT

# FORWARD (all)
iptables -A FORWARD -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT
iptables -A FORWARD -m conntrack --ctstate NEW -p tcp -s 172.20.0.0/16 -j ACCEPT
iptables -A FORWARD -p udp -s 172.20.0.0/16 -j ACCEPT
iptables -A FORWARD -p icmp -s 172.20.0.0/16 -j ACCEPT
iptables -A FORWARD -p sctp -s 172.20.0.0/16 -j ACCEPT
iptables -A FORWARD -m conntrack --ctstate NEW -p tcp -s 10.0.0.0/8 -j ACCEPT
iptables -A FORWARD -p udp -s 10.0.0.0/8 -j ACCEPT
iptables -A FORWARD -p icmp -s 10.0.0.0/8 -j ACCEPT
iptables -A FORWARD -p sctp -s 10.0.0.0/8 -j ACCEPT

# NAT (snat)
iptables -t nat -A POSTROUTING -s 172.20.0.0/16 -d 0.0.0.0/0 -o eth0 -j MASQUERADE
iptables -t nat -A POSTROUTING -s 10.46.0.0/16 -o eth0 -j MASQUERADE
