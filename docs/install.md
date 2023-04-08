# Install and run eUPF

## To run eUPF you must have:

- kubernetes cluster
- deployed 5g core (open5gs or free5gc)

## Requirements for a Kubernetes cluster:

cluster should have:

- Calico CNI
- Multus CNI

in our environments, we use one node Kubernetes cluster deployed by [kubespray](https://github.com/kubernetes-sigs/kubespray). You can see configuration examples in this [repo](https://github.com/edgecomllc/ansible)

## How to deploy open5gs core:

### deploy

* [install helm](https://helm.sh/docs/intro/install/)
* add openverso helm repo

```
helm repo add openverso https://gradiant.github.io/openverso-charts/
helm repo update
```

* install eUPF chart

```
helm upgrade --install \
    edgecomllc-eupf .deploy/helm/universal-chart \
    --values docs/examples/open5gs/eupf.yaml \
    -n open5gs \
    --wait --timeout 100s --create-namespace
```

* install open5gs chart

```
helm upgrade --install \
    open5gs openverso/open5gs \
    --values docs/examples/open5gs/open5gs.yaml \
    -n open5gs \
    --version 2.0.9 \
    --wait --timeout 100s --create-namespace
```

* install ueransim chart

```
helm upgrade --install \
    ueransim openverso/ueransim-gnb \
    --values docs/examples/open5gs/ueransim-gnb.yaml \
    -n open5gs \
    --version 0.2.5 \
    --wait --timeout 100s --create-namespace
```

### undeploy everything

```
helm delete open5gs ueransim edgecomllc-eupf -n open5gs
```

## How to deploy free5gc core

### prepare Kubernetes nodes

You should compile and install gtp5g kernel module on every worker node:

```
apt-get update; apt-get install git build-essential -y; \
cd /tmp; \
git clone --depth 1 --branch v0.7.3 https://github.com/free5gc/gtp5g.git; \
cd gtp5g/; \
make && make install
```

check that the module is loaded:

`lsmod | grep ^gtp5g`

### deploy

* [install helm](https://helm.sh/docs/intro/install/)
* add towards5gs helm repo

```
helm repo add towards5gs https://raw.githubusercontent.com/Orange-OpenSource/towards5gs-helm/main/repo/
helm repo update
```

* install eUPF chart

```
helm upgrade --install \
    edgecomllc-eupf .deploy/helm/universal-chart \
    --values docs/examples/free5gc/eupf.yaml \
    -n free5gc \
    --wait --timeout 100s --create-namespace
```

* install free5gc chart

```
helm upgrade --install \
    free5gc towards5gs/free5gc \
    --values docs/examples/free5gc/free5gc-single.yaml \
    -n free5gc \
    --version 1.1.6 \
    --wait --timeout 100s --create-namespace
```

* create susbscriber in freee5gc via WebUI

redirect port from webui pod to localhost

```
kubectl port-forward service/webui-service 5000:5000 -n free5gc
```

open http://127.0.0.1:5000 in your browser (for auth use user "admin" with password "free5gc"), go to menu "subscribers", click "new subscriber", leave all values as is, press "submit"

close port forward with `Ctrl + C`

* install ueransim chart

```
helm upgrade --install \
    ueransim towards5gs/ueransim \
    --values docs/examples/free5gc/ueransim.yaml \
    -n free5gc \
    --version 2.0.17 \
    --wait --timeout 100s --create-namespace
```

### undeploy everything

```
helm delete free5gc ueransim edgecomllc-eupf -n free5gc
```

## Test scenarios

## case 0

<b>description:</b>

UE can send packet to internet and get response

<b>actions:</b>

for open5gs

* get ue pod name and save to variable and run shell in pod

```
export NS_NAME=open5gs
export UE_POD_NAME=$(kubectl get pods -l "app.kubernetes.io/name=ueransim-gnb,app.kubernetes.io/component=ues" --output=jsonpath="{.items..metadata.name}" -n ${NS_NAME})`
kubectl exec -n ${NS_NAME} --stdin --tty ${UE_POD_NAME} -- /bin/bash
```

for free5gc

```
export NS_NAME=free5gc
export UE_POD_NAME=$(kubectl get pods -l "app=ueransim,component=ue" --output=jsonpath="{.items..metadata.name}" -n ${NS_NAME})`
kubectl exec -n ${NS_NAME} --stdin --tty ${UE_POD_NAME} -- /bin/bash
```

* run command

`$ ping -I uesimtun0 google.com`


<b>expected result:</b>

ping command successful
