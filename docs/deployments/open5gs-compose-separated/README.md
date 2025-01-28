# open5gs-compose

# install docker, docker-compose

https://docs.docker.com/engine/install/

# configure docker daemon for disable iptables

https://docs.docker.com/network/packet-filtering-firewalls/#prevent-docker-from-manipulating-iptables

# create network for containers

```
docker network create \
    --driver=bridge \
    --subnet=172.20.0.0/24 \
    --gateway=172.20.0.1 \
    -o "com.docker.network.bridge.name"="br-open5gs-main" \
    open5gs-main
sudo ethtool -K br-open5gs-main tx off
```

<details><summary><i> Here we turn off tx offload to avoid TCP checksum error in internal packets for iperf tests.</summary>
<p>

Apparently, checksum offloading is enabled by default and the kernel postpones csum calculation until the last moment, expecting csum to be calculated in the driver when the packet is sent. But we have a virtual environment and the packet eventually goes to the GTP tunnel on UPF. Obviously, this is the reason why csum is not calculated correctly.

However, if we disable offloading, the checksum is calculated immediately on iperf and everything works.
</p>
</details> 

<!---
# configure firewall

`bash fw.sh`
--><br>
# To run multiple iperf servers, use this command

`sudo apt install iperf3`

`for i in $(seq 5201 5208); do (iperf3 -s -p $i &) ; done`

# start Open5GS

1. pull all docker images

`make pull`

2. start infra services (like mongodb)

`make infra`

3. start eUPF

`make eupf`

4. start core

`make core`

5. start gNB (this runs multiple instances of gNB, check parameter `scale` in command)

`make gnb`

6. start UERANSim device works with open5gs UPF

`make ue1`

7. start UERANSim device works with eUPF

`make ue2`

# run iperf tests

1. run iperf test in every UERANSim containers

`make test`

when test ended, you can see reports in directory `.deploy/docker/local-data/iperf`

# stop and remove all containers

`make clean`

`docker network rm open5gs-main`