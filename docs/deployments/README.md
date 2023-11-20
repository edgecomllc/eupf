# Deployment examples
The eUPF can be integrated with different 5G Core implementations in different scenarios.

eUPF outbound connections is pure routed at the node. There is no embedded NAT, so external NAT should be used if address translation is needed.

## Docker-compose deployments

## K8s deployments

In K8s BGP is used to announce the subscriber's subnet to the route table of Kubernetes cluster.

- [Open5GS + Calico BGP](./open5gs-with-bgp/README.md)
- [Open5GS + Calico BGP with Slices](./open5gs-with-bgp-and-slices/README.md)
- [Open5GS + Load Balanced eUPF](./open5gs-with-scaling-eupf/README.md)
- [Free5GC + Calico BGP](./free5gc-with-bgp/README.md)
- [Free5GC UpLink CLassifiers (ULCL) architecture](./free5gc-ulcl/README.md)
