# Load testing

we are testing with 2 tools:
* iperf
* mtr

## Open5gs

### iperf

* install iperf server

```bash
helm upgrade --install \
  iperf3 openverso/iperf3 \
  --set-string service.type=ClusterIP \
  --version 0.1.2 \
  --namespace open5gs \
  --wait --timeout 30s --create-namespace
```

* run shell in ueransim ue pod

```
kubectl -n open5gs exec -ti deployment/ueransim-ueransim-gnb-ues -- /bin/bash
```

* install iperf3

```bash
apk add iperf3
```

* check tcp throughput (without upf)

```bash
$ iperf3 -c iperf3 -p 5201 -t 30 -R
Connecting to host iperf3, port 5201
Reverse mode, remote host iperf3 is sending
...
[ ID] Interval           Transfer     Bitrate         Retr
[  5]   0.00-30.00  sec  45.8 GBytes  13.1 Gbits/sec  4369             sender
[  5]   0.00-30.00  sec  45.8 GBytes  13.1 Gbits/sec                  receiver

iperf Done.
```

* check tcp throughput (with open5gs upf)

```bash
$ export UESIMTUNO_IP=$(ip -o -4 addr list uesimtun0 | awk '{print $4}' | cut -d/ -f1)
$ iperf3 -c iperf3 -p 5201 -t 30 -R -B ${UESIMTUNO_IP}
Connecting to host iperf3, port 5201
Reverse mode, remote host iperf3 is sending
...
[ ID] Interval           Transfer     Bitrate         Retr
[  5]   0.00-30.00  sec   612 MBytes   171 Mbits/sec  554             sender
[  5]   0.00-30.00  sec   612 MBytes   171 Mbits/sec                  receiver

iperf Done.
```

* check tcp throughput (with eUPF)

```bash
20:46:43.399431 n6    In  IP 188.120.253.172.2031 > 10.233.10.221.5201:
20:46:43.399455 eth0  Out IP 188.120.253.172.2031 > 10.233.10.221.5201:
```

### mtr

* run shell in ueransim ue pod

```
kubectl -n open5gs exec -ti deployment/ueransim-ueransim-gnb-ues -- /bin/bash
```

* install mtr

```bash
apk add mtr
```

* check latency (without upf) to iperf3 pod

```bash
$ mtr --no-dns --report --report-cycles 60 -T -P 5201 iperf3
...
HOST: ueransim-ueransim-gnb-ues-5 Loss%   Snt   Last   Avg  Best  Wrst StDev
  1.|-- 10.233.10.221              0.0%    60    0.2   0.2   0.1   0.3   0.0
```

* check latency (without upf) to google.com

```bash
$ mtr --no-dns --report --report-cycles 60 -T -P 443 1.1.1.1
...
HOST: ueransim-ueransim-gnb-ues-5 Loss%   Snt   Last   Avg  Best  Wrst StDev
  1.|-- 188.120.253.172            0.0%    60    0.1   0.2   0.1   0.3   0.0
...
  7.|-- 1.1.1.1                    0.0%    60    1.2   1.2   1.1   1.7   0.1
```

==============================================

* check latency (with open5gs upf) to iperf3 pod

```bash
$ mtr --no-dns --report --report-cycles 60 -T -P 5201 -I uesimtun0 iperf3
...
HOST: ueransim-ueransim-gnb-ues-5 Loss%   Snt   Last   Avg  Best  Wrst StDev
  1.|-- 10.45.0.1                  0.0%    60    1.0   1.0   0.7   1.7   0.2
  2.|-- 10.233.10.221              0.0%    60    1.0   1.2   0.8   2.5   0.3
```

* check latency (with open5gs upf) to google.com

```bash
$ mtr --no-dns --report --report-cycles 60 -T -P 443 -I uesimtun0 1.1.1.1
...
HOST: ueransim-ueransim-gnb-ues-5 Loss%   Snt   Last   Avg  Best  Wrst StDev
 1.|-- 10.45.0.1                  0.0%    60    1.3   1.1   0.7   2.4   0.3
...
11.|-- 1.1.1.1                    0.0%    60   21.4  22.3  21.2  25.7   0.8
```

* check latency (with eUPF) to iperf3 pod

```bash
$ mtr --no-dns --report --report-cycles 60 -T -P 5201 -I uesimtun0 iperf3
```

* check latency (with eUPF) to CloudFlare

```bash
$ mtr --no-dns --report --report-cycles 60 -T -P 443 -I uesimtun0 1.1.1.1
...
HOST: ueransim-ueransim-gnb-ues-5 Loss%   Snt   Last   Avg  Best  Wrst StDev
  1.|-- 10.99.0.254                0.0%    60    1.1   1.0   0.8   1.5   0.1
...
  8.|-- 1.1.1.1                    0.0%    60    2.1   2.3   1.8   4.7   0.5
```
