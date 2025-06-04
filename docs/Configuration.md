# UPF Config

## Description

Currently UPF have several config parameters shown below.

Parameters can be configured through command line interface, config files (YAML, JSON) or environment variables.

Parameter                            | Description                                                                                                                                                                                                                                 | yaml                        | env                             | cli arg         | Defaults
------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------- | ------------------------------- | --------------- | --------------
Interface name `Mandatory`           | List of network interfaces handling N3 (GTP) & N6 (SGi) traffic. eUPF attaches XDP hook to every interface in this list. Format: `[ifnameA, ifnameB, ...]`.                                                                                 | `interface_name`            | `UPF_INTERFACE_NAME`            | `--iface`       | `lo`
N3 address `Mandatory`               | IPv4 address for N3 interface                                                                                                                                                                                                               | `n3_address`                | `UPF_N3_ADDRESS`                | `--n3addr`      | `127.0.0.1`
N9 address `Optional`                | IPv4 address for N9 interface                                                                                                                                                                                                               | `n9_address`                | `UPF_N9_ADDRESS`                | `--n9addr`      | `n3_address`
XDP mode `Optional`                  | XDP attach mode: ∘ **generic** – kernel-level (evaluation) ∘ **native** – driver-level ∘ **offload** – NIC-level (direct NIC execution). Refer to [How XDP Works](https://www.tigera.io/learn/guides/ebpf/ebpf-xdp/#How-XDP-Works)          | `xdp_attach_mode`           | `UPF_XDP_ATTACH_MODE`           | `--attach`      | `generic`
API address `Optional`               | Local address for serving [REST API](api.md) server                                                                                                                                                                                         | `api_address`               | `UPF_API_ADDRESS`               | `--aaddr`       | `:8080`
PFCP address `Optional`              | Local address that PFCP server will listen to                                                                                                                                                                                               | `pfcp_address`              | `UPF_PFCP_ADDRESS`              | `--paddr`       | `:8805`
PFCP NodeID `Optional`               | Local NodeID for PFCP protocol. Format is IPv4 address.                                                                                                                                                                                     | `pfcp_node_id`              | `UPF_PFCP_NODE_ID`              | `--nodeid`      | `127.0.0.1`
GTP peer `Optional`                  | List of gtp peer's address to send echo requests to. Format is `[hostnameA:portA, hostnameB:portB, ...]`.                                                                                                                                   | `gtp_peer`                  | `UPF_GTP_PEER`                  | `--peer`        | `-`
Echo request iterval `Optional`      | Echo request sending interval. Format is seconds.                                                                                                                                                                                           | `echo_interval`             | `UPF_ECHO_INTERVAL`             | `--echo`        | `10`
Metrics address `Optional`           | Local address for serving Prometheus mertrics endpoint.                                                                                                                                                                                     | `metrics_address`           | `UPF_METRICS_ADDRESS`           | `--maddr`       | `:9090`
Resize eBPF maps `Optional`          | Enable custom resizing of eBPF cards. When set to `true`, eBPF map sizes derive from max_sessions, but explicit settings (`qer_map_size`, etc.) override automatic calculations.                                                            | `resize_ebpf_map`           | `UPF_RESIZE_EBPF_MAP`           | `--mapresize`   | `false`
QER map size `Optional`              | Size of the QER eBPF map. Effective only if `resize_ebpf_map` is `true`. Overrides value derived from `max_sessions` when set (non-zero).                                                                                                   | `qer_map_size`              | `UPF_QER_MAP_SIZE`              | `--qersize`     | `0`
FAR map size `Optional`              | Size of the FAR eBPF map. Effective only if `resize_ebpf_map` is `true`. Overrides value derived from `max_sessions` when set (non-zero).                                                                                                   | `far_map_size`              | `UPF_FAR_MAP_SIZE`              | `--farsize`     | `0`
PDR map size `Optional`              | Size of the PDR eBPF map. Effective only if `resize_ebpf_map` is `true`. Overrides value derived from `max_sessions` when set (non-zero).                                                                                                   | `pdr_map_size`              | `UPF_PDR_MAP_SIZE`              | `--pdrsize`     | `0`
URR map size `Optional`              | Size of the URR eBPF map. Effective only if `resize_ebpf_map` is `true`. Overrides value derived from `max_sessions` when set (non-zero).                                                                                                   | `urr_map_size`              | `UPF_URR_MAP_SIZE`              | `--urrsize`     | `0`
Max Sessions `Optional`              | Maximum number of sessions. Effective only if `resize_ebpf_map` is `true`. Automatically calculates map sizes (PDR = 2×max_sessions, FAR = PDR, QER = max_sessions, URR = 2×max_sessions) when no individual (`qer_map_size`, etc.) is set. | `max_sessions`              | `UPF_MAX_SESSIONS`              | `--maxsessions` | `0`
Logging level `Optional`             | Logs having level <= selected level will be written to stdout                                                                                                                                                                               | `logging_level`             | `UPF_LOGGING_LEVEL`             | `--loglvl`      | `info`
UEIP Feature `Optional`              | Support for IP allocation option                                                                                                                                                                                                            | `feature_ueip`              | `UPF_FEATURE_UEIP`              | `--ueip`        | `false`
FTUP Feature `Optional`              | Support for TEID allocation option                                                                                                                                                                                                          | `feature_ftup`              | `UPF_FEATURE_FTUP`              | `--ftup`        | `false`
UE IP Pool `Optional`                | Pool of IP addresses, needed to allocate ip when the UEIP option is enabled                                                                                                                                                                 | `ueip_pool`                 | `UPF_UEIP_POOL`                 | `--ueippool`    | `10.60.0.0/24`
TEID Pool `Optional`                 | Pool of TEIDs, needed to allocate TEID when the FTUP option is enabled                                                                                                                                                                      | `teid_pool`                 | `UPF_TEID_POOL`                 | `--teidpool`    | `65535`
PFCP peers `Optional`                | List of PFCP peers (SMF hostnames or IP addresses) which UPF will try to connect                                                                                                                                                            | `pfcp_node`                 | `UPF_PFCP_NODE`                 | `--pfcpnode`    | `-`
Association Setup timeout `Optional` | Timeout between Association Setup Requests initiated by UPF                                                                                                                                                                                 | `association_setup_timeout` | `UPF_ASSOCIATION_SETUP_TIMEOUT` | `--astimeout`   | `5`

We are using [Viper](https://github.com/spf13/viper) for configuration handling, [Viper](https://github.com/spf13/viper) uses the following precedence order. Each item takes precedence over the item below it:

- CLI argument
- environment variable
- configuration file value
- default value

_NOTE:_ as of [commit](https://github.com/edgecomllc/eupf/commit/ea56431df2f74cb2eabe85052d8762fe95848711) we are currently only support IPv4 NodeID.

## Example configuration

### Default values YAML

```yaml
interface_name: [lo]
xdp_attach_mode: generic
api_address: :8080
pfcp_address: :8805
pfcp_node_id: 127.0.0.1
metrics_address: :9090
n3_address: 127.0.0.1
n9_address: 127.0.0.1
resize_ebpf_maps: true
qer_map_size: 1024
far_map_size: 1024
pdr_map_size: 1024
urr_map_size: 1024
feature_ueip: true
feature_ftup: true
ip_pool: 10.60.0.0/16
teid_pool: 65535
```

### Environment variables

```env
UPF_INTERFACE_NAME="[eth0, n6]"
UPF_XDP_ATTACH_MODE=generic
UPF_API_ADDRESS=:8081
UPF_PFCP_ADDRESS=:8806
UPF_METRICS_ADDRESS=:9091
UPF_PFCP_NODE_ID: 10.100.50.241  # address on n4 interface
UPF_N3_ADDRESS: 10.100.50.233
UPF_N9_ADDRESS: 10.100.50.233
```

### CLI

```bash
eupf \
 --iface n3 \
 --iface n6 \
 --attach generic \
 --aaddr :8081 \
 --paddr :8086 \
 --nodeid 127.0.0.1 \
 --maddr :9090 \
 --n3addr 10.100.50.233
```
