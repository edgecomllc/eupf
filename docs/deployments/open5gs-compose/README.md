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
    open5gs-main
```

# configure firewall

`bash fw.sh`

# for run multiple iperf servers, use this command

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
