# Open5GS + eUPF with Calico BGP

![](./schema.png)

## Requirements

- Kubernetes cluster with Calico and Multus CNI
- [helm](https://helm.sh/docs/intro/install/) installed
- calico backend configured as BIRD

change `calico_backend` parameter to `bird` in configmap with name `calico-config` and then restart all pods with name `calico-node-*`

- configure helm repos

```
helm repo add openverso https://gradiant.github.io/openverso-charts/
helm repo update
```

## Deployment steps

* install eupf

`make upf`

* configure calico BGP settings

`make calico`

* install open5gs

`make open5gs`

* configure SMF

`make smf`

* install gNB

`make gnb`

* install UERANSim

`make ue1`

## Check steps

* exec shell in UE pod

`kubectl -n open5gs exec -ti deployment/ueransim1-ueransim-ues-ues -- /bin/bash`

* run ICMP test

`ping -I uesimtun0 1.1.1.1`

## Undeploy steps

* undeploy all

`make clean`
