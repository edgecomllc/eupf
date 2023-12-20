# Deployment examples
The eUPF can be integrated with different 5G Core implementations in different scenarios.

eUPF outbound connections is pure routed at the node. There is no embedded NAT, so external NAT should be used if address translation is needed.

## Docker-compose deployments

| 5G Core | RAN | Options | Deployment description |
| ------- | --- | ------- | ---------------------- |
| Open5GS | UERANSIM | - | [Open5GS](https://github.com/edgecomllc/open5gs-compose) |
| Open5GS | OpenAirInterface | - | Comming soon... |
| Free5GC | UERANSIM | - | [Free5GC](https://github.com/edgecomllc/free5gc-compose/blob/master/README.md) |
| Free5GC | UERANSIM | ULCL | [Free5GC with UpLink Classifier config throught three eUPFs](https://github.com/edgecomllc/free5gc-compose/tree/ulcl-n9upf-experimetns#ulcl-configuration) |
| OpenAirInterface 5G Core | OpenAirInterface 5G RAN	 | - | [OAI 5G SA mode with L2 nFAPI simulator](./oai-nfapi-sim-compose/README.md) |

## K8s deployments

In K8s BGP is used to announce the subscriber's subnet to the route table of Kubernetes cluster.

| 5G Core | RAN | Options | Deployment description |
| ------- | --- | ------- | ---------------------- |
| Open5GS | UERANSIM | Calico BGP | [Open5GS & Calico BGP](./open5gs-with-bgp/README.md) |
| Open5GS | UERANSIM | Calico BGP with Slices | [Open5GS & Calico BGP with Slices](./open5gs-with-bgp-and-slices/README.md) |
| Open5GS | UERANSIM | Load Balanced eUPF | [Open5GS & Load Balanced eUPF](./open5gs-with-scaling-eupf/README.md) |
| Open5GS | srsRAN | Calico BGP | [Open5GS & srsRAN & Calico BGP](./srsran-gnb/README.md) |
| Free5GC | UERANSIM | Calico BGP | [Free5GC & Calico BGP](./free5gc-with-bgp/README.md) |
| Free5GC | UERANSIM | ULCL | [Free5GC & ULCL](./free5gc-ulcl/README.md) |
| OpenAirInterface 5G Core | OpenAirInterface 5G RAN | - | [OAI](./oai/README.md) |
