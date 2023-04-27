# UPF Config

## Description

Currently UPF have several config parameters shown below. Parameters can be changed through command line interface, config files (YAML, JSON) or environment variables prefixed with `UPF_`.

| Parameter       | Description                                                         | yaml              | env                   | cli arg    | Defaults    |
|-----------------|---------------------------------------------------------------------|-------------------|-----------------------|------------|-------------|
| Interface name * | List of interface names that eBPF programs will bind to. Format: `[ifnameA, ifnameB, ...]`. Here should be a network interfaces handling N3 (GTP) & N6 (WAN) traffic.               | `interface_name`  | `UPF_INTERFACE_NAME`  | `--iface`  | `lo`        |
| N3 address  *    | Local IP address will be used:<br> + as src ip address for outgoing gtp packets over N3 interface and <br> + as local ip in ebpf maps. | `n3_address`      | `UPF_N3_ADDRESS`      | `--n3addr` | `127.0.0.1` |
| XDP mode        | XDP attach mode (generic, native, offload) <br> ∘ Generic – Driver doesn’t have support for XDP, but the kernel fakes it. XDP program works, but its not nearly as efficient as the other modes. <br> ∘ Native – Driver has XDP support and can hand then to XDP without kernel stack interaction <br> ∘ Offloaded – XDP can be loaded and executed directly on the NIC <br> Refer to [How XDP Works](https://www.tigera.io/learn/guides/ebpf/ebpf-xdp/#How-XDP-Works)                      | `xdp_attach_mode` | `UPF_XDP_ATTACH_MODE` | `--attach` | `generic`   |
| API address     | Local Address:port for serving [REST Api](api.md)                                        | `api_address`     | `UPF_API_ADDRESS`     | `--aaddr`  | `:8080`     |
| PFCP address    | Local Address:port that PFCP server will listen to. N4 traffic will be handling according to it.   | `pfcp_address`    | `UPF_PFCP_ADDRESS`    | `--paddr`  | `:8805`     |
| PFCP NodeID     | Local NodeID for PFCP protocol. Format is IPv4 address.                                      | `pfcp_node_id`    | `UPF_PFCP_NODE_ID`    | `--nodeid` | `127.0.0.1` |
| Metrics address | Local Address for serving Prometheus mertrics                             | `metrics_address` | `UPF_METRICS_ADDRESS` | `--maddr`  | `:9090`     |
\* marked parameters are mandatory to be correctly set

We are using [Viper](https://github.com/spf13/viper) for configuration handling, [Viper](https://github.com/spf13/viper) uses the following precedence order. Each item takes precedence over the item below it:

- CLI argument
- environment variable
- configuration file value
- default value

*NOTE:* as of [commit](https://github.com/edgecomllc/eupf/commit/ea56431df2f74cb2eabe85052d8762fe95848711) we are currently only support IPv4 NodeID.

## Example configuration

### Default values YAML

```yaml
interface_name: lo
xdp_attach_mode: generic
api_address: :8080
pfcp_address: :8805
pfcp_node_id: 127.0.0.1
metrics_address: :9090
n3_address: 127.0.0.1
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
 --iface [n3, n6] \
 --attach generic \
 --aaddr :8081 \
 --paddr :8086 \
 --nodeid 127.0.0.1 \
 --maddr :9090 \
 --n3addr 10.100.50.233
```
