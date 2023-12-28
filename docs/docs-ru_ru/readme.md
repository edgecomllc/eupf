# eUPF

<div align="center">

[![GitHub Release][release-img]][release]
[![Build][build-img]][build]
[![Test][test-img]][test]
[![Security][security-test-img]][security-test]
[![License: Apache-2.0][license-img]][license]

</div>

eUPF — это проект функции userplane (UPF) с открытым исходным кодом для использования внутри или «вне» любого ядра 3GPP 5G. Цель проекта — предоставить "онаблюдаемое" и легко развертываемое программное обеспечение для различных случаев, таких как multi-access edge computing (MEC) и разделение локального трафика (network slicing). eUPF построен на основе eBPF для обеспечения высокой "наблюдаемости" и производительности.

eUPF был протестирован с тремя различными ядрами 5G: Free5GC, Open5GS и OpenAirInterface. Во время тестирования также использовался OpenAirInterface gNB.

## Что такое ядро 5G сети и CPUS

Ядро 5G использует виртуализированные функции сети (NVF) для обеспечения связи и услуг.
Разделение плоскостей управления и пользователя (CUPS) — это важное усовершенствование архитектуры, которое разделяет плоскости управления и плоскости пользователя внутри ядра 5G.
Функция userplane (UPF) — это функция «декапсуляции и маршрутизации», которая извлекает трафик плоскости пользователя из протокола туннелирования GPRS (GTP) и направляет его в общедоступную сеть передачи данных или локальную сеть по наилучшему доступному пути.

![image](../../docs/pictures/eupf.png)

## Quick start guide

Быстрый и простой способ — загрузить и запустить наш докер-образ. Будет запущен автономный eUPF с конфигурацией по умолчанию.:
```bash
docker run -d --rm -v /sys/fs/bpf:/sys/fs/bpf \
  --cap-add SYS_ADMIN --cap-add NET_ADMIN \
  -p 8080 -p 9090 --name your-eupf-def \
  -v /sys/kernel/debug:/sys/kernel/debug:ro ghcr.io/edgecomllc/eupf:main
```
### Замечания
- 📝 *Linux Kernel **5.15.0-25-generic** — это минимальная версия ядра, на которой eUPF был протестирован. Предыдущие версии не поддерживаются.*
- ℹ Для выполнения низкоуровневых операций, таких как загрузка объектов ebpf, требуются некоторые дополнительные привилегии.(NET_ADMIN & SYS_ADMIN)

<details><summary><i>Параметры запуска, которые вы, возможно, захотите изменить.</i></summary>
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

В реальных сценарих вам, скорее всего, придется заменить имена интерфейсов и IP-адреса теми, которые используются в вашей среде. Вы можете сделать это, например, с помощью опции `-e`:

```bash
docker run -d --rm -v /sys/fs/bpf:/sys/fs/bpf \
  --cap-add SYS_ADMIN --cap-add NET_ADMIN \
  -p 8081 -p 9091 --name your-eupf-custom \
  -e UPF_INTERFACE_NAME="[eth0, n6]" -e UPF_XDP_ATTACH_MODE=generic \
  -e UPF_API_ADDRESS=:8081 -e UPF_PFCP_ADDRESS=:8806 \
  -e UPF_METRICS_ADDRESS=:9091 -e UPF_PFCP_NODE_ID=10.100.50.241 \
  -e UPF_N3_ADDRESS=10.100.50.233 \
  -v /sys/kernel/debug:/sys/kernel/debug:ro \
  ghcr.io/edgecomllc/eupf:main
```

## Что дальше??
Read **[eUPF configuration guide](Configuration.md)** for more info about how to configure eUPF.

To go further, see the **[eUPF installation guide](install.md)** to learn how to run eUPF in different environments with different 5G core implementations using docker-compose or Kubernetes cluster.

For statistics you can gather, see the **[eUPF metrics and monitoring guide](metrics.md)**.

You can find different types of implementation in the **[Implementation expamples](implementation_examples.md)**.

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
- [ ]  UPF-based UE IP address assignment

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
- [ ]  `FTUP` F-TEID allocation / release in the UP function is supported by the UP function.
- [ ]  `UEIP` Allocating UE IP addresses or prefixes.
- [ ]  `SSET` PFCP sessions successively controlled by different SMFs of a same SMF Set.
- [ ]  `MPAS` Multiple PFCP associations to the SMFs in an SMF set.
- [ ]  `QFQM` Per QoS flow per UE QoS monitoring.
- [ ]  `GPQM` Per GTP-U Path QoS monitoring.
- [ ]  `RTTWP` RTT measurements towards the UE Without PMF.

 </details>

## Running from sources

### Prerequisites

- Git
- Golang
- Clang
- LLVM
- gcc
- libbpf-dev

**On Ubuntu 22.04**, you can install these using the following command:

```bash
sudo apt install git golang clang llvm gcc-multilib libbpf-dev
```

**On Rocky Linux 9**, use the following command:

```bash
sudo dnf install git golang clang llvm gcc libbpf libbpf-devel libxdp libxdp-devel xdp-tools bpftool kernel-headers
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
