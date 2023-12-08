#!/bin/bash 

terminate=0
#_term() { 
#  echo "Caught SIGTERM signal!" 
#  terminate=1
#}
#trap _term SIGTERM


iptables -t nat -A POSTROUTING -s 10.60.0.0/16 -o eth0 -j MASQUERADE
iptables -t nat -A POSTROUTING -s 12.1.1.0/24 -o eth0 -j MASQUERADE
iptables -A FORWARD -j ACCEPT

#echo "1200 n6if" >> /etc/iproute2/rt_tables
#ip rule add to 10.1.0.0/16 table n6if
#ip route add default via upf.free5gc.org dev eth0 table n6if

#ip route add 10.60.0.0/16 via `nslookup upf.free5gc.org | awk '/^Address: / { print $2 }'` dev eth0
ip route add 10.60.0.0/16 via `nslookup upfn6.free5gc.org | awk '/^Address: / { print $2 }'`
ip route add 12.1.1.0/24 via `nslookup upfn6.free5gc.org | awk '/^Address: / { print $2 }'`


while [ $terminate -ne 1 ]
do
    sleep 10;
done


echo "Exit busy-polling script" 