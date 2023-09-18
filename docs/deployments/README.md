# Deployment examples
The eUPF can be integrated with different 5G Core implementations in different scenarios.

eUPF pod outbound connection is pure routed at the node. There is no address translation inside pod, so we avoid such lack of throughtput.

BGP is used to announce the subscriber's subnet to the route table of Kubernetes cluster.

## [Open5GS + Calico BGP](./open5gs-with-bgp/README.md)

## [Open5GS + Calico BGP with Slices](./open5gs-with-bgp-and-slices/README.md)

## [Free5GC + Calico BGP](./free5gc-with-bgp/README.md)

## [Free5GC UpLink CLassifier (ULCL) architecture](./free5gc-ulcl/README.md)
