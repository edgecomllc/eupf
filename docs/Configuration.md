# UPF Config

## Description

Currently UPF have several config parameters shown below. Parameters can be changed through command line interface, config files (YAML, JSON) or environment variables prefixed with `UPF_`.

| Parameter       | Description                                                         | yaml              | env                   | cli arg    | Defaults    |
|-----------------|---------------------------------------------------------------------|-------------------|-----------------------|------------|-------------|
| Interface name  | Interface that eBPF programs will bind to                           | `interface_name`  | `UPF_INTERFACE_NAME`  | `--iface`  | `lo`        |
| XDP mode        | XDP attach mode (generic, native, offload)                          | `xdp_attach_mode` | `UPF_XDP_ATTACH_MODE` | `--attach` | `generic`   |
| API address     | Address for serving REST Api                                        | `api_address`     | `UPF_API_ADDRESS`     | `--aaddr`  | `:8080`     |
| PFCP address    | Address that PFCP server will bind to                               | `pfcp_address`    | `UPF_PFCP_ADDRESS`    | `--paddr`  | `:8805`     |
| PFCP NodeID     | Local NodeID for PFCP protocol                                      | `pfcp_node_id`    | `UPF_PFCP_NODE_ID`    | `--nodeid` | `127.0.0.1` |
| Metrics address | Address for serving Prometheus mertrics                             | `metrics_address` | `UPF_METRICS_ADDRESS` | `--maddr`  | `:9090`     |
| N3 address      | This ip will be used for N3 interface and in ebpf maps as local ip. | `n3_address`      | `UPF_N3_ADDRESS`      | `--n3addr` | `127.0.0.1` |

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
UPF_INTERFACE_NAME=eth0
UPF_XDP_ATTACH_MODE=generic
UPF_API_ADDRESS=:8081
UPF_PFCP_ADDRESS=:8806
UPF_PFCP_NODE_ID=127.0.0.1
UPF_METRICS_ADDRESS=:9091
UPF_N3_ADDRESS=127.0.0.1
```

### CLI

```bash
eupf \
 --iface eth0 \
 --attach generic \
 --aaddr :8081 \
 --paddr :8086 \
 --nodeid 127.0.0.1 \
 --maddr :9090 \
 --n3addr 127.0.0.1
```
