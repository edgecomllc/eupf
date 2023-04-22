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

* run shell in ueransim ue pod, [see instruction here](./install.md#case-0)

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

### mtr

* run shell in ueransim ue pod, [see instruction here](./install.md#case-0)

* install mtr

```bash
apk add mtr
```

* check latency (without mtr)

```
$ mtr --no-dns --report --report-cycles 60 iperf3
...
HOST: ueransim-ueransim-gnb-ues-5 Loss%   Snt   Last   Avg  Best  Wrst StDev
  1.|-- 10.233.57.202              0.0%    60    0.1   0.1   0.1   0.2   0.0
```
