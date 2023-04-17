# Load testing

## iperf

* install iperf server

```
helm upgrade --install \
  iperf3 openverso/iperf3 \
  --set-string service.type=ClusterIP \
  -n open5gs \
  --version 0.1.2 \
  --wait --timeout 100s --create-namespace
```

* run shell in ueransim ue pod (see instruction above for it)

* run test via default network

```
$ iperf3 -c iperf3 -p 5201 -t 30
...
[ ID] Interval           Transfer     Bitrate         Retr
[  5]   0.00-30.00  sec  40.2 GBytes  11.5 Gbits/sec  9942             sender
[  5]   0.00-30.00  sec  40.2 GBytes  11.5 Gbits/sec                  receiver

iperf Done.
```

* run test via open5gs UPF

```
$ export UESIMTUNO_IP=$(ip -o -4 addr list uesimtun0 | awk '{print $4}' | cut -d/ -f1)
$ iperf3 -c iperf3 -p 5201 -t 30 -B ${UESIMTUNO_IP}
...
[ ID] Interval           Transfer     Bitrate         Retr
[  5]   0.00-30.00  sec   597 MBytes   167 Mbits/sec  730             sender
[  5]   0.00-30.00  sec   597 MBytes   167 Mbits/sec                  receiver
```

* run test via eUPF

TODO

## grafana k6

* create test container with simple http server (nginx)

```
	helm upgrade --install \
		nginx .deploy/helm/universal-chart \
		--values docs/examples/load-testing/nginx.yaml \
		-n free5gc \
		--wait --timeout 100s --create-namespace
```

* copy test scenario `script.js` to ueransim pod

TODO

* install k6

```
wget https://github.com/grafana/k6/releases/download/v0.43.1/k6-v0.43.1-linux-amd64.tar.gz
kubectl cp ./k6-v0.43.1-linux-amd64.tar.gz ueransim-ue-7f76db59c9-r67st:/tmp/k6.tar.gz -n free5gc
```

* run test via k6

`k6 run script.js`

* run shell in ueransim ue pod (see instruction above for it)
