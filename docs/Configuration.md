# UPF Config

## Desciption

Currently UPF have theese parameters:

- Inteface name: Interface that eBPF programms will bind to.
- Api Address: Address for serving REST Api.
- PFCP Address: Address that PFCP server will bind to.
- PFCP NodeID: Local NodeID for PFCP protocol.
- Metrics Address: Address for serving Prometheus mertrics.

Theese parametrs can be changes through command line interface, config files (YAML, JSON) or enviroment variables prefixed with `UPF_`.

## Example configuration

### Default values YAML

```yaml
interface_name: lo
api_address: :8080
pfcp_address: :8805
pfcp_node_id: localhost
metrics_address: :9090
```

### Enviroment varaibles

```env
UPF_INTERFACE_NAME=eth0
UPF_API_ADDRESS=:8081
UPF_PFCP_ADDRESS=:8806
UPF_PFCP_NODE_ID=example.com
UPF_METRICS_ADDRESS=:9091
```

### CLI

```bash
eupf \
 --iface eth0 \
 --aaddr :8081 \
 --paddr :8086 \
 --nodeid example.com \
 --maddr :9090
```
