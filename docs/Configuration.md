# UPF Config

## Description

Currently UPF have several config parameters shown below.

Parameters can be configured through command line interface, config files (YAML, JSON) or environment variables.

Parameter                                     | Description                                                                                                                                                                                                                                 | yaml                        | env                             | cli arg               | Defaults
--------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------- | ------------------------------- | --------------------- | -----------------------
Interface name `Mandatory`                    | List of network interfaces handling N3 (GTP) & N6 (SGi) traffic. eUPF attaches XDP hook to every interface in this list. Format: `[ifnameA, ifnameB, ...]`.                                                                                 | `interface_name`            | `UPF_INTERFACE_NAME`            | `--iface`             | `lo`
N3 address `Mandatory`                        | IPv4 address for N3 interface                                                                                                                                                                                                               | `n3_address`                | `UPF_N3_ADDRESS`                | `--n3addr`            | `127.0.0.1`
N9 address `Optional`                         | IPv4 address for N9 interface                                                                                                                                                                                                               | `n9_address`                | `UPF_N9_ADDRESS`                | `--n9addr`            | `n3_address`
XDP mode `Optional`                           | XDP attach mode: ∘ **generic** – kernel-level (evaluation) ∘ **native** – driver-level ∘ **offload** – NIC-level (direct NIC execution). Refer to [How XDP Works](https://www.tigera.io/learn/guides/ebpf/ebpf-xdp/#How-XDP-Works)          | `xdp_attach_mode`           | `UPF_XDP_ATTACH_MODE`           | `--attach`            | `generic`
API address `Optional`                        | Local address for serving [REST API](api.md) server                                                                                                                                                                                         | `api_address`               | `UPF_API_ADDRESS`               | `--aaddr`             | `:8080`
PFCP address `Optional`                       | Local address that PFCP server will listen to                                                                                                                                                                                               | `pfcp_address`              | `UPF_PFCP_ADDRESS`              | `--paddr`             | `:8805`
PFCP NodeID `Optional`                        | Local NodeID for PFCP protocol. Format is IPv4 address.                                                                                                                                                                                     | `pfcp_node_id`              | `UPF_PFCP_NODE_ID`              | `--nodeid`            | `127.0.0.1`
GTP peer `Optional`                           | List of gtp peer's address to send echo requests to. Format is `[hostnameA:portA, hostnameB:portB, ...]`.                                                                                                                                   | `gtp_peer`                  | `UPF_GTP_PEER`                  | `--peer`              | `-`
Echo request iterval `Optional`               | Echo request sending interval. Format is seconds.                                                                                                                                                                                           | `gtp_echo_interval`         | `UPF_GTP_ECHO_INTERVAL`         | `--echo`              | `10`
Metrics address `Optional`                    | Local address for serving Prometheus mertrics endpoint.                                                                                                                                                                                     | `metrics_address`           | `UPF_METRICS_ADDRESS`           | `--maddr`             | `:9090`
QER map size `Optional`                       | Size of the QER eBPF map. Overrides value derived from `max_sessions` when set (non-zero).                                                                                                   | `qer_map_size`              | `UPF_QER_MAP_SIZE`              | `--qersize`           | `1024`
FAR map size `Optional`                       | Size of the FAR eBPF map. Overrides value derived from `max_sessions` when set (non-zero).                                                                                                   | `far_map_size`              | `UPF_FAR_MAP_SIZE`              | `--farsize`           | `1024`
PDR map size `Optional`                       | Size of the PDR eBPF map. Overrides value derived from `max_sessions` when set (non-zero).                                                                                                   | `pdr_map_size`              | `UPF_PDR_MAP_SIZE`              | `--pdrsize`           | `1024`
URR map size `Optional`                       | Size of the URR eBPF map. Overrides value derived from `max_sessions` when set (non-zero).                                                                                                   | `urr_map_size`              | `UPF_URR_MAP_SIZE`              | `--urrsize`           | `1024`
Max Sessions `Optional`                       | Maximum number of sessions. Automatically calculates map sizes (PDR = 2×max_sessions, FAR = PDR, QER = max_sessions, URR = 3×max_sessions) when no individual (`qer_map_size`, etc.) is set. | `max_sessions`              | `UPF_MAX_SESSIONS`              | `--maxsessions`       | `65535`
Logging level `Optional`                      | Logs having level <= selected level will be written to stdout                                                                                                                                                                               | `logging_level`             | `UPF_LOGGING_LEVEL`             | `--loglvl`            | `info`
UEIP Feature `Optional`                       | Support for IP allocation option                                                                                                                                                                                                            | `feature_ueip`              | `UPF_FEATURE_UEIP`              | `--ueip`              | `false`
FTUP Feature `Optional`                       | Support for TEID allocation option                                                                                                                                                                                                          | `feature_ftup`              | `UPF_FEATURE_FTUP`              | `--ftup`              | `false`
UE IP Pool `Optional`                         | Pool of IP addresses, needed to allocate ip when the UEIP option is enabled                                                                                                                                                                 | `ueip_pool`                 | `UPF_UEIP_POOL`                 | `--ueippool`          | `10.60.0.0/24`
TEID Pool `Optional`                          | Pool of TEIDs, needed to allocate TEID when the FTUP option is enabled                                                                                                                                                                      | `teid_pool`                 | `UPF_TEID_POOL`                 | `--teidpool`          | `65535`
PFCP peers `Optional`                         | List of PFCP peers (SMF hostnames or IP addresses) which UPF will try to connect                                                                                                                                                            | `pfcp_node`                 | `UPF_PFCP_NODE`                 | `--pfcprnode`         | `-`
Association Setup timeout `Optional`          | Timeout between Association Setup Requests initiated by UPF                                                                                                                                                                                 | `association_setup_timeout` | `UPF_ASSOCIATION_SETUP_TIMEOUT` | `--astimeout`         | `5`
Support Huawei proprietary options `Optional` | Enable or disable huawei support                                                                                                                                                                                                            | `huawei_support`            | `UPF_HUAWEI_SUPPORT`            | `--huasupp`           | `true`
S1-U address `Optional`                       | Address for communication over S1-U interface                                                                                                                                                                                               | `s1u_address`               | UPF_S1U_ADDRESS                 | --s1uaddr string      | 127.0.0.1
S5/S8 address `Optional`                      | Address for communication over S5/S8 interface                                                                                                                                                                                              | `s5s8_address`              | UPF_S5S8_ADDRESS                | --s5s8addr string     | 127.0.0.1
PA address `Optional`                         | Address for communication over PA interface                                                                                                                                                                                                 | `pa_address`                | UPF_PA_ADDRESS                  | --paaddr string       | 127.0.0.1
Sxa address `Optional`                        | Sxa Address to bind PFCP server to                                                                                                                                                                                                          | `sxa_address`               | UPF_SXA_ADDRESS                 | --sxaaddr string      | 127.0.0.2:8805
List of Sxa peers `Optional`                  | Address of remote Sxa node                                                                                                                                                                                                                  | `sxa_node`                  | UPF_SXA_NODE                    | --sxanode stringArray | `-`
Sxa node id `Optional`                        | Sxa PFCP Server Node ID                                                                                                                                                                                                                     | `sxa_node_id`               | UPF_SXA_NODE_ID                 | --sxanodeid string    | 127.0.0.2
Sxb address `Optional`                        | Sxb Address to bind PFCP server to                                                                                                                                                                                                          | `sxb_address`               | UPF_SXB_ADDRESS                 | --sxbaddr string      | 127.0.0.3:8805
List of Sxb peers `Optional`                  | Address of remote Sxb node                                                                                                                                                                                                                  | `sxb_node`                  | UPF_SXB_NODE                    | --sxbnode stringArray | `-`
Sxb node id `Optional`                        | Sxb PFCP Server Node ID                                                                                                                                                                                                                     | `sxb_node_id`               | UPF_SXB_NODE_ID                 | --sxbnodeid string    | 127.0.0.3
Trace assocuation `Optional`                  | Trace PFCP Association messages (Establish/Modify/Release)                                                                                                                                                                                  | `trace_association`         | UPF_TRACE_ASSOCIATION           | --traceassoc          | TRUE
Trace heartbeat `Optional`                    | Trace PFCP Heartbeat messages                                                                                                                                                                                                               | `trace_heartbeat`           | UPF_TRACE_HEARTBEAT             | --tracehb             | FALSE
Trace droppped packets `Optional`             | Trace dropped dataplane packets                                                                                                                                                                                                             | `trace_blocked`             | UPF_TRACE_BLOCKED               | --traceblock          | TRUE
Trace files number `Optional`                 | Maximum number of rotated trace dump files                                                                                                                                                                                                  | `trace_files`               | UPF_TRACE_FILES                 | --tracefcnt int       | 10
Trace packets in file `Optional`              | Maximum number of packets in one trace dump file                                                                                                                                                                                            | `trace_max_packets`         | UPF_TRACE_MAX_PACKETS           | --tracefpackets int   | 100000
Max trace size `Optional`                     | Maximum size (in bytes) of one trace dump file                                                                                                                                                                                              | `trace_max_size`            | UPF_TRACE_MAX_SIZE              | --tracefsize int      | 10485760
IPv6 router advertisement `Optional`          | Enable or disable IPv6 Router Advertisement support                                                                                                                                                                                         | `ip6_ra_support`            | UPF_IP6_RA_SUPPORT              | --ip6ra               | FALSE
IPv6 subscriber prefix `Optional`             | Subscriber IPv6 prefix                                                                                                                                                                                                                      | `ip6_ra_prefix`             | UPF_IP6_RA_PREFIX               | --ip6rapref string    | 2a03:d000:29a0:509::/64

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
qer_map_size: 1024
far_map_size: 1024
pdr_map_size: 1024
feature_ueip: true
feature_ftup: true
ip_pool: 10.60.0.0/16
teid_pool: 65535
pfcp_node: 
association_setup_timeout: 5
huawei_support: true
s1u_address: 127.0.0.1
s5s8_address: 127.0.0.1
sxa_address: 127.0.0.2:8805
sxa_node: 
sxa_node_id: 127.0.0.2
sxb_address: 127.0.0.3:8805
sxb_node: 
sxb_node_id: 127.0.0.3
pa_address: 127.0.0.1
trace_association: true
trace_blocked: true
trace_files: 10
trace_max_packets: 100000
trace_max_size: 10485760
trace_heartbeat: false
ip6_ra_support: false
ip6_ra_prefix: 2a03:d000:29a0:509::/64
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
