# Load testing

We are testing with 2 tools:

* iperf
* mtr

We have 3 testing scenarios:

* tcp throughput to neighbor pod
* latency to neighbor pod
* latency to google public DNS

## Open5gs
<details><summary>Instructions</summary>
<p>

### iperf

* install iperf server

```bash
helm upgrade --install \
  iperf3 openverso/iperf3 \
  --values docs/examples/open5gs/iperf.yaml \
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

we should use some flags for iperf client (specific for eUPF):
- packet size (`-M`)
- pod address (`-c`)

```bash
$ iperf3 -c 10.233.110.181 -p 5201 -t 30 -R --bind-dev uesimtun0 -M 1350
Connecting to host 10.233.110.181, port 5201
Reverse mode, remote host 10.233.110.181 is sending
...
[ ID] Interval           Transfer     Bitrate         Retr
[  5]   0.00-30.00  sec   490 MBytes   137 Mbits/sec  1181             sender
[  5]   0.00-30.00  sec   490 MBytes   137 Mbits/sec                  receiver
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

* check latency (without upf) to google public dns

```bash
$ mtr --no-dns --report --report-cycles 60 -T -P 443 8.8.8.8
...
HOST: ueransim-ueransim-gnb-ues-5 Loss%   Snt   Last   Avg  Best  Wrst StDev
  1.|-- 188.120.253.172            0.0%    60    0.1   0.2   0.1   0.3   0.0
...
 16.|-- 8.8.8.8                   96.7%    60   16.8  15.7  14.7  16.8   1.5
```

* check latency (with open5gs upf) to iperf3 pod

```bash
$ mtr --no-dns --report --report-cycles 60 -T -P 5201 -I uesimtun0 iperf3
...
HOST: ueransim-ueransim-gnb-ues-5 Loss%   Snt   Last   Avg  Best  Wrst StDev
  1.|-- 10.45.0.1                  0.0%    60    1.0   1.0   0.7   1.7   0.2
  2.|-- 10.233.10.221              0.0%    60    1.0   1.2   0.8   2.5   0.3
```

* check latency (with open5gs upf) to google public dns

```bash
$ mtr --no-dns --report --report-cycles 60 -T -P 443 -I uesimtun0 8.8.8.8
...
HOST: ueransim-ueransim-gnb-ues-5 Loss%   Snt   Last   Avg  Best  Wrst StDev
 1.|-- 10.45.0.1                  0.0%    60    1.3   1.1   0.7   2.4   0.3
...
 17.|-- 8.8.8.8                   96.7%    60   17.2  19.2  17.2  21.2   2.9
```

* check latency (with eUPF) to iperf3 pod

```bash
$ mtr --no-dns --report --report-cycles 60 -T -P 5201 -I uesimtun0 10.233.110.181
HOST: ueransim-ueransim-gnb-ues-5 Loss%   Snt   Last   Avg  Best  Wrst StDev
  3.|-- 10.233.110.181             0.0%    60    0.9   1.0   0.6   1.3   0.1
```

* check latency (with eUPF) to google public dns

```bash
$ mtr --no-dns --report --report-cycles 60 -T -P 443 -I uesimtun0 8.8.8.8
...
HOST: ueransim-ueransim-gnb-ues-5 Loss%   Snt   Last   Avg  Best  Wrst StDev
  1.|-- 10.99.0.254                0.0%    60    1.1   1.0   0.8   1.5   0.1
...
 17.|-- 8.8.8.8                   95.0%    60   14.6  16.9  14.6  21.3   3.8
```

## results

|scenario | raw | open5gs upf | eupf |
|---|---|---|---|
| tcp throughput (to neighbor pod) | 13.1 Gbit/sec | 171 Mbit/sec | 137 Mbit/sec |
| latency (to neighbor pod) | 0.2 | 1.2 | 1.0 |
| latency (to google public DNS) | 15.7 | 19.2 | 16.9 |

</p>

## Free5GC

### iperf

* install iperf server

```bash
helm upgrade --install \
  iperf3 openverso/iperf3 \
  --values docs/examples/free5gc/iperf.yaml \
  --version 0.1.2 \
  --namespace free5gc \
  --wait --timeout 30s --create-namespace
```

* run shell in ueransim ue pod

```
kubectl -n free5gc exec -ti deployment/ueransim-ueransim-gnb-ues -- /bin/bash
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
$ export UESIMTUNO_IP=$(ip -o -4 addr list uesimtun0 | awk '{print $4}' | cut -d/ -f1)
$ iperf3 -c iperf3 -p 5201 -t 30 -R -B ${UESIMTUNO_IP}
?
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

* check latency (without upf) to google public dns

```bash
$ mtr --no-dns --report --report-cycles 60 -T -P 443 8.8.8.8
...
HOST: ueransim-ueransim-gnb-ues-5 Loss%   Snt   Last   Avg  Best  Wrst StDev
  1.|-- 188.120.253.172            0.0%    60    0.1   0.2   0.1   0.3   0.0
...
 16.|-- 8.8.8.8                   96.7%    60   16.8  15.7  14.7  16.8   1.5
```

* check latency (with open5gs upf) to iperf3 pod

```bash
$ mtr --no-dns --report --report-cycles 60 -T -P 5201 -I uesimtun0 iperf3
...
HOST: ueransim-ueransim-gnb-ues-5 Loss%   Snt   Last   Avg  Best  Wrst StDev
  1.|-- 10.45.0.1                  0.0%    60    1.0   1.0   0.7   1.7   0.2
  2.|-- 10.233.10.221              0.0%    60    1.0   1.2   0.8   2.5   0.3
```

* check latency (with open5gs upf) to google public dns

```bash
$ mtr --no-dns --report --report-cycles 60 -T -P 443 -I uesimtun0 8.8.8.8
...
HOST: ueransim-ueransim-gnb-ues-5 Loss%   Snt   Last   Avg  Best  Wrst StDev
 1.|-- 10.45.0.1                  0.0%    60    1.3   1.1   0.7   2.4   0.3
...
 17.|-- 8.8.8.8                   96.7%    60   17.2  19.2  17.2  21.2   2.9
```

* check latency (with eUPF) to iperf3 pod

```bash
$ mtr --no-dns --report --report-cycles 60 -T -P 5201 -I uesimtun0 10.99.0.11
?
```

* check latency (with eUPF) to google public dns

```bash
$ mtr --no-dns --report --report-cycles 60 -T -P 443 -I uesimtun0 8.8.8.8
...
HOST: ueransim-ueransim-gnb-ues-5 Loss%   Snt   Last   Avg  Best  Wrst StDev
  1.|-- 10.99.0.254                0.0%    60    1.1   1.0   0.8   1.5   0.1
...
 17.|-- 8.8.8.8                   95.0%    60   14.6  16.9  14.6  21.3   3.8
```



## results

|scenario | raw | open5gs upf | eupf |
|---|---|---|---|
| tcp throughput (to neighbor pod) | 13.1 Gbit/sec | 171 Mbit/sec | ? |
| latency (to neighbor pod) | 0.2 | 1.2 | ? |
| latency (to google public DNS) | 15.7 | 19.2 | 16.9 |
