# UPF Config

## Description

Currently UPF have several config parameters shown below.<br>Parameters can be configured through command line interface, config files (YAML, JSON) or environment variables.

| Parameter                      | Description                                                                                                                                                                                                                                                                                                                                     | yaml              | env                   | cli arg     | Defaults        |
|--------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------|-----------------------|-------------|-----------------|
| Interface name<br>`Mandatory`  | List of network interfaces handling N3 (GTP) & N6 (SGi) traffic. eUPF attaches XDP hook to every interface in this list. Format: `[ifnameA, ifnameB, ...]`.                                                                                                                                                                                     | `interface_name`  | `UPF_INTERFACE_NAME`  | `--iface`   | `lo`            |
| N3 address <br>`Mandatory`     | IPv4 address for N3 interface                                                                                                                                                                                                                                                                                                                   | `n3_address`      | `UPF_N3_ADDRESS`      | `--n3addr`  | `127.0.0.1`     |
| XDP mode <br>`Optional`        | XDP attach mode: <br> ∘ **generic** – Kernel-level implementation. For evaluation purpose.  <br> ∘ **native** – Driver-level implenemntaion <br> ∘ **offloaded** – NIC-level implementation. XDP can be loaded and executed directly on the NIC. <br> Refer to [How XDP Works](https://www.tigera.io/learn/guides/ebpf/ebpf-xdp/#How-XDP-Works) | `xdp_attach_mode` | `UPF_XDP_ATTACH_MODE` | `--attach`  | `generic`       |
| API address <br>`Optional`     | Local address for serving [REST API](api.md) server                                                                                                                                                                                                                                                                                             | `api_address`     | `UPF_API_ADDRESS`     | `--aaddr`   | `:8080`         |
| PFCP address <br>`Optional`    | Local address that PFCP server will listen to                                                                                                                                                                                                                                                                                                   | `pfcp_address`    | `UPF_PFCP_ADDRESS`    | `--paddr`   | `:8805`         |
| PFCP NodeID <br>`Optional`     | Local NodeID for PFCP protocol. Format is IPv4 address.                                                                                                                                                                                                                                                                                         | `pfcp_node_id`    | `UPF_PFCP_NODE_ID`    | `--nodeid`  | `127.0.0.1`     |
| Metrics address <br>`Optional` | Local address for serving Prometheus mertrics endpoint.                                                                                                                                                                                                                                                                                         | `metrics_address` | `UPF_METRICS_ADDRESS` | `--maddr`   | `:9090`         |
| QER map size <br>`Optional`    | Size of the eBPF map for QER parameters                                                                                                                                                                                                                                                                                                         | `qer_map_size`    | `UPF_QER_MAP_SIZE`    | `--qersize` | `1024  `        |
| FAR map size <br>`Optional`    | Size of the eBPF map for FAR parameters                                                                                                                                                                                                                                                                                                         | `far_map_size`    | `UPF_FAR_MAP_SIZE`    | `--farsize` | `1024  `        |
| PDR map size <br>`Optional`    | Size of the eBPF map for PDR parameters                                                                                                                                                                                                                                                                                                         | `pdr_map_size`    | `UPF_PDR_MAP_SIZE`    | `--pdrsize` | `1024  `        |
| Logging level <br>`Optional`   | Logs having level <= selected level will be written to stdout                                                                                                                                                                                                                                                                                   | `logging_level`   | `UPF_LOGGING_LEVEL`   | `--loglvl`  | `info`          |
| UEIP Feature <br>`Optional`    | Support for IP allocation option                                                                                                                                                                                                                                                                                                                | `feature_ueip`    | `UPF_FEATURE_UEIP`    | `--ueip`     | `false`         |
| FTUP Feature <br>`Optional`    | Support for TEID allocation option                                                                                                                                                                                                                                                                                                              | `feature_ftup`    | `UPF_FEATURE_FTUP`    | `--ftup`     | `false`         |
| UE IP Pool <br>`Optional`         | Pool of IP addresses, needed to allocate ip when the UEIP option is enabled                                                                                                                                                                                                                                                                     | `ueip_pool`         | `UPF_UEIP_POOL`         | `--ueippool`   | `10.60.0.0/24`  |
| TEID Pool <br>`Optional`       | Pool of TEIDs, needed to allocate TEID when the FTUP option is enabled                                                                                                                                                                                                                                                                          | `teid_pool`       | `UPF_TEID_POOL`       | `--teidpool` | `65535`         |
We are using [Viper](https://github.com/spf13/viper) for configuration handling, [Viper](https://github.com/spf13/viper) uses the following precedence order. Each item takes precedence over the item below it:

- CLI argument
- environment variable
- configuration file value
- default value

*NOTE:* as of [commit](https://github.com/edgecomllc/eupf/commit/ea56431df2f74cb2eabe85052d8762fe95848711) we are currently only support IPv4 NodeID.

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
qer_map_size: 1024
far_map_size: 1024
pdr_map_size: 1024
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