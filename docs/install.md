# How to Install and run eUPF
The easyest way is to use our docker container image. And simple deploy by our helm charts in your kubernetes cluster.

So, to run eUPF you must have:

- kubernetes cluster
- [helm](https://helm.sh/docs/intro/install/) installed
<!-- - deployed 5g core (open5gs or free5gc) -->

## Requirements for a Kubernetes cluster:

cluster should have:

- Calico CNI
- Multus CNI

in our environments, we use one node Kubernetes cluster deployed by [kubespray](https://github.com/kubernetes-sigs/kubespray). You can see configuration examples in this [repo](https://github.com/edgecomllc/ansible)

We have prepared templates to deploy with two opensource environments: open5gs and free5gc, for you to choose. Both with UERANSIM project emulating radio endpoint, so you'll be able to check end-to-end connectivity. 

## How to deploy eUPF with open5gs core:
<details><summary>Instructions</summary>
<p>

### To deploy:

* [install helm](https://helm.sh/docs/intro/install/) if it's not
* add openverso helm repo

   ```
   helm repo add openverso https://gradiant.github.io/openverso-charts/
   helm repo update
   ```

* install eUPF chart

   ```powershell
   helm upgrade --install \
       edgecomllc-eupf .deploy/helm/universal-chart \
       --values docs/examples/open5gs/eupf.yaml \
       -n open5gs \
       --wait --timeout 100s --create-namespace
   ```
   📝Here we use subnet `10.100.111.0/24` for n6 interface as exit to the world, so make sure it's not occupied at your node host.

* install open5gs chart

   ```powershell
   helm upgrade --install \
       open5gs openverso/open5gs \
       --values docs/examples/open5gs/open5gs.yaml \
       -n open5gs \
       --version 2.0.9 \
       --wait --timeout 100s --create-namespace
   ```

* install ueransim chart

   ```powershell
   helm upgrade --install \
       ueransim openverso/ueransim-gnb \
       --values docs/examples/open5gs/ueransim-gnb.yaml \
       -n open5gs \
       --version 0.2.5 \
       --wait --timeout 100s --create-namespace
   ```

### To undeploy everything:

```
helm delete open5gs ueransim edgecomllc-eupf -n open5gs
```
📝 Pod's interconnection. openverso-charts uses default interfaces of your kubernetes cluster. It is Calico CNI interfaces in our environment, type ipvlan. And it uses k8s services names to resolve endpoints.
The only added is ptp type interface `n6` as a door to the outer world for our eUPF. 

For more details refer to openverso-charts [Open5gs and UERANSIM](https://gradiant.github.io/openverso-charts/open5gs-ueransim-gnb.html)

</p>
</details> 

## How to deploy eUPF with free5gc core
<details><summary>Instructions</summary>
<p>

### prepare Kubernetes nodes

You should compile and install gtp5g kernel module on every worker node:

```powershell
apt-get update; apt-get install git build-essential -y; \
cd /tmp; \
git clone --depth 1 --branch v0.7.3 https://github.com/free5gc/gtp5g.git; \
cd gtp5g/; \
make && make install
```

check that the module is loaded:

`lsmod | grep ^gtp5g`

### deploy

0. [install helm](https://helm.sh/docs/intro/install/) if it's not
0. add towards5gs helm repo

	```powershell
	helm repo add towards5gs https://raw.githubusercontent.com/Orange-OpenSource/towards5gs-helm/main/repo/
	helm repo update
	```

0. install eUPF chart

	```powershell
	helm upgrade --install \
		edgecomllc-eupf .deploy/helm/universal-chart \
		--values docs/examples/free5gc/eupf.yaml \
		-n free5gc \
		--wait --timeout 100s --create-namespace
	```
   📝Here we use subnet `10.100.100.0/24` for n6 interface as exit to the world, so make sure it's not occupied at your node host.

0. install free5gc chart

	```powershell
	helm upgrade --install \
		free5gc towards5gs/free5gc \
		--values docs/examples/free5gc/free5gc-single.yaml \
		-n free5gc \
		--version 1.1.6 \
		--wait --timeout 100s --create-namespace
	```

0. create subscriber in free5gc via WebUI

   redirect port from webui pod to localhost

   ```powershell
   kubectl port-forward service/webui-service 5000:5000 -n free5gc
   ```

   open http://127.0.0.1:5000 in your browser (for auth use user "admin" with password "free5gc"), go to menu "subscribers", click "new subscriber", leave all values as is, press "submit"

   close port forward with `Ctrl + C`

0. install ueransim chart

	```powershell
	helm upgrade --install \
		ueransim towards5gs/ueransim \
		--values docs/examples/free5gc/ueransim.yaml \
		-n free5gc \
		--version 2.0.17 \
		--wait --timeout 100s --create-namespace
	```

### To undeploy everything

```
helm delete free5gc ueransim edgecomllc-eupf -n free5gc
```
📝 Pod's interconnection. towards5gs-helm uses separate subnets with ipvlan type interfaces with internal addressing.
The only added is ptp type interface `n6` as a door to the outer world for our eUPF. 


### Architecture is nested from towards5gs-helm project [Setup free5gc](https://github.com/Orange-OpenSource/towards5gs-helm/blob/main/docs/demo/Setup-free5gc-on-multiple-clusters-and-test-with-UERANSIM.md)
![Architecture](pictures/Setup-free5gc-on-multiple-clusters-and-test-with-UERANSIM-Architecture.png)


</p>
</details> 
</p>

## Option NAT at the node
eUPF pod outbound connection is pure routed at the node. There is no address translation inside pod, so we avoid such lack of throughtput.

If you need NAT (Network Address Translation, or Masqerading) at your node to access Internet, the easiest way is to use standart daemonset [IP Masquerade Agent](https://kubernetes.io/docs/tasks/administer-cluster/ip-masq-agent/):
```powershell
sudo kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/ip-masq-agent/master/ip-masq-agent.yaml
```
   > The below entries show the default set of rules that are applied by the ip-masq-agent:
    ` iptables -t nat -L IP-MASQ-AGENT`     

---

## Test scenarios

## case 0

<b>description:</b>

UE can send packet to internet and get response

<b>actions:</b>

1. run shell in pod

   for open5gs:
   ```powershell
   export NS_NAME=open5gs
   export UE_POD_NAME=$(kubectl get pods -l "app.kubernetes.io/name=ueransim-gnb,app.kubernetes.io/component=ues" --output=jsonpath="{.items..metadata.name}" -n ${NS_NAME})
   kubectl exec -n ${NS_NAME} --stdin --tty ${UE_POD_NAME} -- /bin/bash
   ```

   for free5gc:

   ```powershell
   export NS_NAME=free5gc
   export UE_POD_NAME=$(kubectl get pods -l "app=ueransim,component=ue" --output=jsonpath="{.items..metadata.name}" -n ${NS_NAME})
   kubectl exec -n ${NS_NAME} --stdin --tty ${UE_POD_NAME} -- /bin/bash
   ```

1. run command from UE pod's shell. 

   `$ ping -I uesimtun0 google.com`


   <b>expected result:</b>

   ping command successful
