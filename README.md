# eUPF

<div align="center">

[![GitHub Release][release-img]][release]
[![Build][build-img]][build]
[![Test][test-img]][test]
[![Security][security-test-img]][security-test]
[![License: Apache-2.0][license-img]][license]

</div>

eUPF is the opensource User Plane Function (UPF) project for using inside or "outside" of any 3GPP 5G core. The goal of the project is to provide high-observability and easily-deployed software for a various cases like multi-access edge computing (MEC) and local traffic breakout. eUPF is built with eBPF to provide high observability and performance.

 The eUPF has been tested with three different 5G cores: Free5GC, Open5GS and OpenAirInterface. The OpenAirInterface gNB was also used during testing.

## What is 5G core and CUPS

5G core uses network virtualized functions (NVF) to provide connectivity and services.
Control and user plane separation (CUPS) is important architecture enhancement that separates control plane and user plane inside 5G core.
User plane function (UPF) is the "decapsulating and routing" function that extracts user plane traffic from GPRS tunneling protocol (GTP) and route it to the public data network or local network via the best available path.

![image](docs/pictures/eupf.png)

## Quick start guide

Super fast & simple way is to download and run our docker image. It will start standalone eUPF with the default configuration:
```bash
sudo docker run -d --rm --privileged \
  -v /sys/fs/bpf:/sys/fs/bpf \ 
  -v /sys/kernel/debug:/sys/kernel/debug:ro \
  -p 8080 -p 9090 --name your-eupf-def \
   ghcr.io/edgecomllc/eupf:main
```
### Notes
- 📝 *Linux Kernel **5.15.0-25-generic** is the minimum release version it has been tested on. Previous versions are not supported.*
- ℹ The eBPF filesystem must be mounted in the host filesystem allowing the eUPF to persist eBPF resources across restarts so that the datapath can continue to operate while the eUPF is subsequently restarted or upgraded.
Use following command to mount it: `sudo mount bpffs /sys/fs/bpf -t bpf`
- ℹ In order to perform low-level operations like loading ebpf objects some additional privileges are required(NET_ADMIN & SYS_ADMIN)
- ℹ During startup eupf sets rlimits, so corresponding priviledges are required (ulimit)

<details><summary><i>See startup parameters you might want to change</i></summary>
<p>

   - UPF_INTERFACE_NAME=lo    *Network interfaces handling N3 (GTP) & N6 (SGi) traffic.*
   - UPF_N3_ADDRESS=127.0.0.1 *IPv4 address for N3 interface*
   - UPF_XDP_ATTACH_MODE=generic *XDP attach mode. Generic-only at the moment*
   - UPF_API_ADDRESS=:8080    *Local host:port for serving [REST API](api.md) server*
   - UPF_PFCP_ADDRESS=:8805   *Local host:port that PFCP server will listen to*
   - UPF_PFCP_NODE_ID=127.0.0.1  *Local NodeID for PFCP protocol. Format is IPv4 address*
   - UPF_METRICS_ADDRESS=:9090   *Local host:port for serving Prometheus mertrics endpoint*
</p>
</details>
</p>

In a real-world scenario, you would likely need to replace the interface names and IP addresses with values that are applicable to your environment. You can do so with the `-e` option, for example:

```bash
sudo docker run -d --rm -v --privileged \
  -v /sys/fs/bpf:/sys/fs/bpf \ 
  -v /sys/kernel/debug:/sys/kernel/debug:ro \
  -p 8081 -p 9091 --name your-eupf-custom \
  -e UPF_INTERFACE_NAME=[eth0,n6] -e UPF_XDP_ATTACH_MODE=generic \
  -e UPF_API_ADDRESS=:8081 -e UPF_PFCP_ADDRESS=:8806 \
  -e UPF_METRICS_ADDRESS=:9091 -e UPF_PFCP_NODE_ID=10.100.50.241 \
  -e UPF_N3_ADDRESS=10.100.50.233 \
  ghcr.io/edgecomllc/eupf:main
```

## What's next?
Read **[eUPF configuration guide](./docs/Configuration.md)** for more info about how to configure eUPF.

To go further, see the **[eUPF installation guide](./docs/install.md)** to learn how to run eUPF in different environments with different 5G core implementations using docker-compose or Kubernetes cluster.

For statistics you can gather, see the **[eUPF metrics and monitoring guide](./docs/metrics.md)**.

You can find different types of implementation in the **[Implementation expamples](./docs/implementation_examples.md)**.

## Implementation notes

eUPF as a part of 5G mobile core network implements data network gateway function. It communicates with SMF via PFCP protocol (N4 interface) and forwards packets between core and data networks(N3 and N6 interfaces correspondingly). These two main UPF parts are implemented in two separate components: control plane and forwarding plane.

The eUPF control plane is an userspace application which receives packet processing rules from SMF and configures forwarding plane for proper forwarding.

The eUPF forwarding plane is based on eBPF packet processing. When started eUPF adds eBPF XDP hook program in order to process network packets as close to NIC as possible. eBPF program consists of several pipeline steps: determine PDR, apply gating, qos and forwarding rules.

eUPF relies on kernel routing when making routing decision for incoming network packets. When it is not possible to determine packet route via kernel FIB lookup, eUPF passes such packet to kernel as a fallback path. This approach obviously affects performance but allows maintaining correct kernel routing process (ex., filling arp tables).

### Brief functional description

#### FAR support

eUPF supports FAR rules in PDR. Only one FAR rule per PDR is supported.

#### QER support

eUPF supports QER rules in PDR. Currently only one QER rule per PDR is supported.

#### SDF filters support

eUPF is able to apply SDF filters in PDR. Currently only one SDF filter per GTP tunnel is supported.

#### GTP path management

eUPF supports sending GTP Echo requests towards neighbour GTP nodes. Every neighbour GTP node should be explicitly configured. [See](docs/Configuration.md) `gtp_peer` configuration parameter.

### Architecture

<details><summary>Show me</summary>

#### Eagle-eye overview

![UPF-Arch2](https://user-images.githubusercontent.com/20152142/207142700-cc3f17a5-203f-4b43-b712-a518cb627968.png)

#### Detailed architecture
![image](docs/pictures/eupf-arch.png)

</details>

### Roadmap

<details><summary>Show me</summary>

#### Control plane

- [x]  PFCP Association Setup/Release and Heartbeats
- [x]  Session Establishment/Modification with support for PFCP entities such as Packet Detection Rules (PDRs), Forwarding Action Rules (FARs), QoS Enforcement Rules (QERs).
- [ ]  UPF-initiated PFCP association
- [x]  UPF-based UE IP address assignment

#### Data plane

- [x]  IPv4 support
- [x]  N3, N4, N6 interfaces
- [x]  Single & Multi-port support
- [x]  Static IP routing
- [x]  Basic QoS support with per-session rate limiting
- [x]  I-UPF/A-UPF ULCL/Branching (N9 interface)

#### Management plane
- [x]  Free5gc compatibility
- [x]  Open5gs compatibility
- [x]  Integration with Prometheus for exporting PFCP and data plane-level metrics
- [ ]  Monitoring/Debugging capabilities using tcpdump and cli

#### 3GPP specs compatibility
- [x]  `FTUP` F-TEID allocation / release in the UP function is supported by the UP function.
- [x]  `UEIP` Allocating UE IP addresses or prefixes.
- [ ]  `SSET` PFCP sessions successively controlled by different SMFs of a same SMF Set.
- [ ]  `MPAS` Multiple PFCP associations to the SMFs in an SMF set.
- [ ]  `QFQM` Per QoS flow per UE QoS monitoring.
- [ ]  `GPQM` Per GTP-U Path QoS monitoring.
- [ ]  `RTTWP` RTT measurements towards the UE Without PMF.

 </details>

## Running from sources

### Prerequisites

-	Ubuntu 22.04 LTS or higher
-	Git 2.34
-	Golang 1.20.3
-	Clang 14.0.0
-	LLVM 14.0
-	Gcc 11.4.0
-	libbpf-dev 0.5.0
-	Swag 1.8.12
-	Linux Kernel 5.15.0-25

**On Ubuntu 22.04**, you can install these using the following commands:

#### Basic dependencies
```bash
sudo apt install wget git clang llvm gcc-multilib libbpf-dev
```
#### Golang 1.20.3
ℹ Please skip this step if you have golang 1.20.3 already installed.

```bash
sudo rm -rf /usr/local/go
wget https://go.dev/dl/go1.20.3.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.20.3.linux-amd64.tar.gz
export PATH="/usr/local/go/bin:${PATH}"
```

### Manual build

#### Step 1: Install the Swag command line tool for Golang
This is used to automatically generate RESTful API documentation.

```bash
go install github.com/swaggo/swag/cmd/swag@v1.8.12
```

#### Step 2: Clone the eUPF repository and change to the directory

```bash
git clone https://github.com/edgecomllc/eupf.git
cd eupf
```

#### Step 3: Run the code generators

```bash
go generate -v ./cmd/...
```

Sometimes during this step you may see errors like:
```
running "swag": exec: "swag": executable file not found in $PATH
``` 

Make sure that `swag` was successfuly installed(step 1) and path to swag binary is in the PATH environment variable. 

Usually GO Path is supposed to already be on the PATH environment variable. 
Use `export PATH=$(go env GOPATH)/bin:$PATH` otherwise and repeat current step again.


#### Step 4: Build eUPF

```bash
go build -v -o bin/eupf ./cmd/
```
#### Step 5: Run the application

Run binary with privileges allowing to increase [memory-ulimits](https://prototype-kernel.readthedocs.io/en/latest/bpf/troubleshooting.html#memory-ulimits)

```bash
sudo ./bin/eupf
```

This should start application with the default configuration. Please adjust the contents of the configuration file and the command-line arguments as needed for your application and environment.

### Build docker image

Use this command to build eupf's docker image: `docker build -t local/eupf:latest .`

You can also define several build arguments to configure eUPF image: `docker build -t local/eupf:latest --build-arg BPF_ENABLE_LOG=1 --build-arg BPF_ENABLE_ROUTE_CACHE=1 .`

### Hardware requirements

- CPU: any popular CPU is supported, incl. x86, x86_64, x86, ppc64le, armhf, armv7, aarch64, ppc64le, s390x
- CPU_cores: 1 core is enough to run eUPF
- RAM: you need up to 70MB to run eUPF and up to 512MB to run Linux kernel
- HDD: 50MB of free space is required to install eUPF. Different types of storage can be used: HDD, SSD, SD-card, USB-stick
- NIC: Any internal or external networking interface that can be used in Linux

## Contribution

Please create an issue to report a bug or share an idea.

## Translated docs

Please check [this link](./docs/docs-ru_ru/readme.md) to find translated docs.

## License
This project is licensed under the [Apache-2.0 Creative Commons License](https://www.apache.org/licenses/LICENSE-2.0) - see the [LICENSE file](./LICENSE) for details

---

[release]: https://github.com/edgecomllc/eupf/releases
[release-img]: https://img.shields.io/github/release/edgecomllc/eupf.svg?logo=github
[build]: https://github.com/edgecomllc/eupf/actions/workflows/build.yml
[build-img]: https://github.com/edgecomllc/eupf/actions/workflows/build.yml/badge.svg
[test]: https://github.com/edgecomllc/eupf/actions/workflows/test.yml
[test-img]: https://github.com/edgecomllc/eupf/actions/workflows/test.yml/badge.svg
[security-test]: https://github.com/edgecomllc/eupf/actions/workflows/trivy.yml
[security-test-img]: https://github.com/edgecomllc/eupf/actions/workflows/trivy.yml/badge.svg
[license]: https://github.com/edgecomllc/eupf/blob/main/LICENSE
[license-img]: https://img.shields.io/badge/License-Apache%202.0-blue.svg
