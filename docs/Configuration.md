# UPF Config

## Description

Currently UPF have several config parameters shown below.

Parameters can be configured through command line interface, config files (YAML, JSON) or environment variables.

Parameter                                     | Description                                                                                                                                                                                                                                 | yaml                        | env                             | cli arg               | Defaults
--------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------- | ------------------------------- | --------------------- | -----------------------
Interface name `Mandatory`                    | List of network interfaces handling N3 (GTP) & N6 (SGi) traffic. eUPF attaches XDP hook to every interface in this list. Format: `[ifnameA, ifnameB, ...]`.                                                                                 | `interface_name`            | `UPF_INTERFACE_NAME`            | `--iface`             | `lo`
N3 address `Mandatory`                        | IPv4 address for N3 interface                                                                                                                                                                                                               | `n3_address`                | `UPF_N3_ADDRESS`                | `--n3addr`            | `127.0.0.1`
N9 address `Optional`                         | IPv4 address for N9 interface                                                                                                                                                                                                               | `n9_address`                | `UPF_N9_ADDRESS`                | `--n9addr`            | `n3_address`
N3 advertised address `Optional`             | Advertised IPv4 address for N3 interface. Falls back to `n3_address` when unset.                                                                                                                                                            | `n3_adv_address`            | `UPF_N3_ADV_ADDRESS`            | `--n3advaddr`         | `n3_address`
N3 advertised address (5G) `Optional`        | Advertised IPv4 address for N3 interface used for 5G sessions. Falls back to `n3_address` when unset.                                                                                                                                       | `n3_adv_address_5g`         | `UPF_N3_ADV_ADDRESS_5G`         | `--n3advaddr5g`       | `n3_address`
N9 advertised address `Optional`             | Advertised IPv4 address for N9 interface. Falls back to `n9_address` when unset.                                                                                                                                                            | `n9_adv_address`            | `UPF_N9_ADV_ADDRESS`            | `--n9advaddr`         | `n9_address`
XDP mode `Optional`                           | XDP attach mode: ∘ **generic** – kernel-level (evaluation) ∘ **native** – driver-level ∘ **offload** – NIC-level (direct NIC execution). Refer to [How XDP Works](https://www.tigera.io/learn/guides/ebpf/ebpf-xdp/#How-XDP-Works)          | `xdp_attach_mode`           | `UPF_XDP_ATTACH_MODE`           | `--attach`            | `generic`
API address `Optional`                        | Local address for serving [REST API](api.md) server                                                                                                                                                                                         | `api_address`               | `UPF_API_ADDRESS`               | `--aaddr`             | `:8080`
PFCP address `Optional`                       | Local address that PFCP server will listen to                                                                                                                                                                                               | `pfcp_address`              | `UPF_PFCP_ADDRESS`              | `--paddr`             | `:8805`
PFCP NodeID `Optional`                        | Local NodeID for PFCP protocol. Format is IPv4 address.                                                                                                                                                                                     | `pfcp_node_id`              | `UPF_PFCP_NODE_ID`              | `--nodeid`            | `127.0.0.1`
GTP peer `Optional`                           | List of gtp peer's address to send echo requests to. Format is `[hostnameA:portA, hostnameB:portB, ...]`.                                                                                                                                   | `gtp_peer`                  | `UPF_GTP_PEER`                  | `--peer`              | `-`
Echo request iterval `Optional`               | Echo request sending interval. Format is seconds.                                                                                                                                                                                           | `gtp_echo_interval`         | `UPF_GTP_ECHO_INTERVAL`         | `--echo`              | `10`
Metrics address `Optional`                    | Local address for serving Prometheus metrics endpoint.                                                                                                                                                                                      | `metrics_address`           | `UPF_METRICS_ADDRESS`           | `--maddr`             | `:9090`
QER map size `Optional`                       | Size of the QER eBPF map. Overrides value derived from `max_sessions` when set (non-zero).                                                                                                   | `qer_map_size`              | `UPF_QER_MAP_SIZE`              | `--qersize`           | `1024`
FAR map size `Optional`                       | Size of the FAR eBPF map. Overrides value derived from `max_sessions` when set (non-zero).                                                                                                   | `far_map_size`              | `UPF_FAR_MAP_SIZE`              | `--farsize`           | `1024`
PDR map size `Optional`                       | Size of the PDR eBPF map. Overrides value derived from `max_sessions` when set (non-zero).                                                                                                   | `pdr_map_size`              | `UPF_PDR_MAP_SIZE`              | `--pdrsize`           | `1024`
URR map size `Optional`                       | Size of the URR eBPF map. Overrides value derived from `max_sessions` when set (non-zero).                                                                                                   | `urr_map_size`              | `UPF_URR_MAP_SIZE`              | `--urrsize`           | `1024`
Max Sessions `Optional`                       | Maximum number of sessions. Automatically calculates map sizes (PDR = 2×max_sessions, FAR = PDR, QER = max_sessions, URR = 2×max_sessions) when no individual (`qer_map_size`, etc.) is set. | `max_sessions`              | `UPF_MAX_SESSIONS`              | `--maxsessions`       | `65535`
Logging level `Optional`                      | Logs having level <= selected level will be written to stdout                                                                                                                                                                               | `logging_level`             | `UPF_LOGGING_LEVEL`             | `--loglvl`            | `info`
Logging caller `Optional`                    | Add the caller info (file:line) to every log line                                                                                                                                                                                           | `logging_caller`            | `UPF_LOGGING_CALLER`            | `--logcaller`         | `false`
Heartbeat retries `Optional`                 | Number of heartbeat retries before the association is considered lost                                                                                                                                                                       | `heartbeat_retries`         | `UPF_HEARTBEAT_RETRIES`         | `--hbretries`         | `3`
Heartbeat interval `Optional`                | Heartbeat sending interval. Format is seconds.                                                                                                                                                                                              | `heartbeat_interval`        | `UPF_HEARTBEAT_INTERVAL`        | `--hbinterval`        | `5`
Heartbeat timeout `Optional`                 | Heartbeat response timeout. Format is seconds.                                                                                                                                                                                              | `heartbeat_timeout`         | `UPF_HEARTBEAT_TIMEOUT`         | `--hbtimeout`         | `5`
QCI to DSCP mapping `Optional`               | Mapping of QCI values to DSCP marks, e.g. `{"9":46}`.                                                                                                                                                                                       | `qci_dscp_mapping`          | `UPF_QCI_DSCP_MAPPING`          | `--qdmap`             | `{}`
Allowed APNs `Optional`                      | Regex mask of allowed APNs                                                                                                                                                                                                                  | `allowed_apns`              | `UPF_ALLOWED_APNS`              | `--aapns`             | `.*`
Denied APNs `Optional`                       | Regex mask of denied APNs. Empty means no APN is denied.                                                                                                                                                                                    | `denied_apns`               | `UPF_DENIED_APNS`               | `--dapns`             | ``
UEIP Feature `Optional`                       | Support for IP allocation option                                                                                                                                                                                                            | `feature_ueip`              | `UPF_FEATURE_UEIP`              | `--ueip`              | `false`
FTUP Feature `Optional`                       | Support for TEID allocation option                                                                                                                                                                                                          | `feature_ftup`              | `UPF_FEATURE_FTUP`              | `--ftup`              | `false`
UE IP Pool `Optional`                         | Pool of IP addresses, needed to allocate ip when the UEIP option is enabled                                                                                                                                                                 | `ueip_pool`                 | `UPF_UEIP_POOL`                 | `--ueippool`          | `10.60.0.0/24`
TEID Pool `Optional`                          | Pool of TEIDs, needed to allocate TEID when the FTUP option is enabled                                                                                                                                                                      | `teid_pool`                 | `UPF_TEID_POOL`                 | `--teidpool`          | `65535`
PFCP peers `Optional`                         | List of PFCP peers (SMF hostnames or IP addresses) which UPF will try to connect                                                                                                                                                            | `pfcp_remote_node`          | `UPF_PFCP_REMOTE_NODE`          | `--pfcprnode`         | `-`
Association Setup timeout `Optional`          | Timeout between Association Setup Requests initiated by UPF                                                                                                                                                                                 | `association_setup_timeout` | `UPF_ASSOCIATION_SETUP_TIMEOUT` | `--astimeout`         | `5`
PFCP profile `Optional` | PFCP dialect: "default" (standard 3GPP) or "huawei" (Huawei SPGW-C) | `pfcp_profile` | `UPF_PFCP_PROFILE` | `--pfcp_profile` | `default` |
S1-U address `Optional`                       | Address for communication over S1-U interface                                                                                                                                                                                               | `s1u_address`               | `UPF_S1U_ADDRESS`               | `--s1uaddr`           | `127.0.0.1`
S5/S8 address `Optional`                      | Address for communication over S5/S8 interface                                                                                                                                                                                              | `s5s8_address`              | `UPF_S5S8_ADDRESS`              | `--s5s8addr`          | `127.0.0.1`
PA address `Optional`                         | Address for communication over PA interface                                                                                                                                                                                                 | `pa_address`                | `UPF_PA_ADDRESS`                | `--paaddr`            | `127.0.0.1`
Sxa address `Optional`                        | Sxa Address to bind PFCP server to                                                                                                                                                                                                          | `sxa_address`               | `UPF_SXA_ADDRESS`               | `--sxaaddr`           | `127.0.0.2:8805`
List of Sxa peers `Optional`                  | Address of remote Sxa node                                                                                                                                                                                                                  | `sxa_remote_node`           | `UPF_SXA_REMOTE_NODE`           | `--sxanode`           | `-`
Sxa node id `Optional`                        | Sxa PFCP Server Node ID                                                                                                                                                                                                                     | `sxa_node_id`               | `UPF_SXA_NODE_ID`               | `--sxanodeid`         | `127.0.0.2`
Sxb address `Optional`                        | Sxb Address to bind PFCP server to                                                                                                                                                                                                          | `sxb_address`               | `UPF_SXB_ADDRESS`               | `--sxbaddr`           | `127.0.0.3:8805`
List of Sxb peers `Optional`                  | Address of remote Sxb node                                                                                                                                                                                                                  | `sxb_remote_node`           | `UPF_SXB_REMOTE_NODE`           | `--sxbnode`           | `-`
Sxb node id `Optional`                        | Sxb PFCP Server Node ID                                                                                                                                                                                                                     | `sxb_node_id`               | `UPF_SXB_NODE_ID`               | `--sxbnodeid`         | `127.0.0.3`
Trace association `Optional`                  | Trace PFCP Association messages (Establish/Modify/Release)                                                                                                                                                                                  | `trace_association`         | `UPF_TRACE_ASSOCIATION`         | `--traceassoc`        | `true`
Trace heartbeat `Optional`                    | Trace PFCP Heartbeat messages                                                                                                                                                                                                               | `trace_heartbeat`           | `UPF_TRACE_HEARTBEAT`           | `--tracehb`           | `false`
Trace incoming packets `Optional`             | Trace incoming dataplane packets                                                                                                                                                                                                            | `trace_in`                  | `UPF_TRACE_IN`                  | `--tracein`           | `true`
Trace outgoing packets `Optional`             | Trace outgoing dataplane packets                                                                                                                                                                                                            | `trace_out`                 | `UPF_TRACE_OUT`                 | `--traceout`          | `true`
Trace dropped packets `Optional`             | Trace dropped dataplane packets                                                                                                                                                                                                             | `trace_blocked`             | `UPF_TRACE_BLOCKED`             | `--traceblock`        | `false`
Trace files number `Optional`                 | Maximum number of rotated trace dump files                                                                                                                                                                                                  | `trace_files`               | `UPF_TRACE_FILES`               | `--tracefcnt`         | `10`
Trace packets in file `Optional`              | Maximum number of packets in one trace dump file                                                                                                                                                                                            | `trace_max_packets`         | `UPF_TRACE_MAX_PACKETS`         | `--tracefpackets`     | `100000`
Max trace size `Optional`                     | Maximum size (in bytes) of one trace dump file                                                                                                                                                                                              | `trace_max_size`            | `UPF_TRACE_MAX_SIZE`            | `--tracefsize`        | `10485760`
IPv6 router advertisement `Optional`          | Enable or disable IPv6 Router Advertisement support                                                                                                                                                                                         | `ip6_ra_support`            | `UPF_IP6_RA_SUPPORT`            | `--ip6ra`             | `false`
IPv6 subscriber prefix `Optional`             | Subscriber IPv6 prefix                                                                                                                                                                                                                      | `ip6_ra_prefix`             | `UPF_IP6_RA_PREFIX`             | `--ip6rapref`         | `2a03:d000:29a0:509::/64`

## PCC (predefined) rules

Predefined PCC rules are configured in a separate config file (default `./pcc.yaml`, override with `--pcc-config`). The path to the file is set with the `--pcc-config` flag; settings inside the file are read from YAML only. The file defines a list of rules under the `pcc_rules` key:

```yaml
pcc_rules:
  - pcc_name: "block_icmp"
    notify: false
    sdf_filter: "permit out ip from 192.168.1.100/32 to any"
    far:
      action: 2                      # bitmask: 1 = DROP, 2 = FORW, 4 = BUFF, 8 = NOCP, 16 = DUPL
      outer_header_creation: 0       # 0 = none, 1 = GTP-U/UDP/IP IPv4, 2 = GTP-U/UDP/IP IPv6, 4 = UDP/IPv4, 8 = UDP/IPv6
      teid: 0
      remote_ip: 0
      transport_level_marking: 0
    qer:
      qfi: 9                         # QoS Flow Identifier, 1..255
      max_bitrate_ul: 100000         # uplink bitrate, bit/s
      max_bitrate_dl: 100000         # downlink bitrate, bit/s
    urr:
      urrid: 0                       # URR id to report usage for
```

Rule fields:

Field             | Description                                                                                     | yaml                    | Defaults
----------------- | ----------------------------------------------------------------------------------------------- | ----------------------- | --------
Rule name         | Unique name of the predefined rule. Referenced by PDRs via Activate Predefined Rules            | `pcc_name`              | (required)
Notify            | Send a notification when the SDF filter matches                                                | `notify`                | `false`
SDF filter        | Service data flow filter, e.g. `permit out ip from 192.168.1.100/32 to any`                    | `sdf_filter`            | (required)
FAR action        | Forwarding action bitmask: 1 = DROP, 2 = FORW, 4 = BUFF, 8 = NOCP, 16 = DUPL                      | `far.action`            | `0`
Outer header      | Outer header creation for the FAR: 0 = none, 1 = GTP-U/UDP/IP (IPv4), 2 = GTP-U/UDP/IP (IPv6), 4 = UDP/IPv4, 8 = UDP/IPv6 | `far.outer_header_creation` | `0`
FAR TEID          | Tunnel endpoint identifier for the forwarding action                                            | `far.teid`              | `0`
FAR remote IP     | Remote (peer) IPv4 address for the forwarding action, as a 32-bit integer                       | `far.remote_ip`         | `0`
Transport marking | Transport level marking (DSCP/TC mark) applied to forwarded packets                              | `far.transport_level_marking` | `0`
QFI               | QoS Flow Identifier (1..255)                                                                    | `qer.qfi`               | (required)
Max bitrate UL    | Maximum uplink bitrate in bit/s (0 = unlimited)                                                 | `qer.max_bitrate_ul`    | `0`
Max bitrate DL    | Maximum downlink bitrate in bit/s (0 = unlimited)                                               | `qer.max_bitrate_dl`    | `0`
URR id            | URR id whose usage is reported for traffic matching the rule                                    | `urr.urrid`             | `0`

A rule is activated for a PDR when the SMF (or the local API) sends an Activate Predefined Rules IE referencing its `pcc_name`; Deactivate Predefined Rules removes it again.

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
pfcp_profile: default
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
