# UPF Config

## Desciption

Currently UPF have several config parameters shown below. Parameters can be changed through command line interface, config files (YAML, JSON) or enviroment variables prefixed with `UPF_`.


| Parameter       | Description                                                         | yaml              | env                   | cli arg    | Defaults    |
| --------------- | ------------------------------------------------------------------- | ----------------- | --------------------- | ---------- | ----------- |
| Inteface name   | Interface that eBPF programms will bind to                          | `interface_name`  | `UPF_INTERFACE_NAME`  | `--iface`  | `lo`        |
| XDP mode        | XDP attach mode (generic, native, offload)                          | `xdp_attach_mode` | `UPF_XDP_ATTACH_MODE` | `--attach` | `generic`   |
| API address     | Address for serving REST Api                                        | `api_address`     | `UPF_API_ADDRESS`     | `--aaddr`  | `:8080`     |
| PFCP address    | Address that PFCP server will bind to                               | `pfcp_address`    | `UPF_PFCP_ADDRESS`    | `--paddr`  | `:8805`     |
| PFCP NodeID     | Local NodeID for PFCP protocol                                      | `pfcp_node_id`    | `UPF_PFCP_NODE_ID`    | `--nodeid` | `localhost` |
| Metrics address | Address for serving Prometheus mertrics                             | `metrics_address` | `UPF_METRICS_ADDRESS` | `--maddr`  | `:9090`     |
| N3IP            | This ip will be used for N3 interface and in ebpf maps as local ip. | `n3_address`      | `UPF_N3_ADDRESS`      | `--n3addr` | `127.0.0.1` |
We are using [Viper](https://github.com/spf13/viper) for configuration handling, [Viper](https://github.com/spf13/viper) uses the following precedence order. Each item takes precedence over the item below it:

- explicit call to Set
- flag
- env
- config
- key/value store
- default

## Example configuration

### Default values YAML

```yaml
interface_name: lo
xdp_attach_mode: generic
api_address: :8080
pfcp_address: :8805
pfcp_node_id: localhost
metrics_address: :9090
```

### Enviroment varaibles

```env
UPF_INTERFACE_NAME=eth0
UPF_XDP_ATTACH_MODE=generic
UPF_API_ADDRESS=:8081
UPF_PFCP_ADDRESS=:8806
UPF_PFCP_NODE_ID=example.com
UPF_METRICS_ADDRESS=:9091
```

### CLI

```bash
eupf \
 --iface eth0 \
 --attach generic \
 --aaddr :8081 \
 --paddr :8086 \
 --nodeid example.com \
 --maddr :9090
```
