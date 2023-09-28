Here is values to deploy Free5GC core + our eUPF at Kubernetes node using helmcharts Orange-OpenSource/towards5gs-helm. 

## UpLink CLassifier (ULCL) architecture

Here is configuration with three UPFs. 
Traffic routes is:
- UE--gNodeB--upfb--upf1--Internet as default
- UE--gNodeB--upfb--upf2--Internet--1.1.1.1/32 for imsi-208930000000003 specificPath

Our eUPF deployed as upfb. upf1 and upf2 are modules from free5gc.

You can see the difference in the first hop of traceroute from UE:
```powershell
bash-5.1# traceroute -i uesimtun0 www.google.com -w 1
traceroute to www.google.com (173.194.222.103), 30 hops max, 46 byte packets
 1  10.233.64.41 (10.233.64.41)  1.518 ms  1.805 ms  1.459 ms
 ......
 bash-5.1# traceroute -i uesimtun0 -w1 1.1.1.1
traceroute to 1.1.1.1 (1.1.1.1), 30 hops max, 46 byte packets
 1  10.233.64.56 (10.233.64.56)  1.512 ms  1.176 ms  0.778 ms
```

## Quick start

### prepare kubernetes nodes - install gtp5g kernel module

compile and install gtp5g kernel module needed for Free5gc UPFs:

```
apt-get update; apt-get install git build-essential -y; \
cd /tmp; \
git clone --depth 1 https://github.com/free5gc/gtp5g.git; \
cd gtp5g/; \
make && make install
```

check that the module is loaded:

`lsmod | grep ^gtp5g`



* [install helm](https://helm.sh/docs/intro/install/)

* add towards5gs helm repo

    ```
    helm repo add towards5gs 'https://raw.githubusercontent.com/Orange-OpenSource/towards5gs-helm/main/repo/'
    helm repo update
    ```

### Use make commands to deploy in NAMESPACE free5gculcl
📝 Other pods deployed by towards5gs in any namespaces should be stopped to avoid conflict of IP addresses of type ipvlan.
1. `make eupf` to install eUPF deploy as upfb
1. `make upf` to install Free5gc UPFs deploy as upf1, upf2
1. `make free5gc` to install free5gc core
1. Open web interface and add new subscriber.

   redirect port from webui pod to localhost

   ```powershell
   kubectl port-forward service/webui-service 5000:5000 -n free5gc
   ```

   open http://127.0.0.1:5000 in your browser (for auth use user "admin" with password "free5gc"), go to menu "subscribers", click "new subscriber", leave all values as is, press "submit"

   close port forward with `Ctrl + C`

1. `make ueransim` to install gNodeB and UE simulators.

after installation, you can run shell into uerasim ue pod:

* `make ueransim_shell`

  `make clean` will delete all components from cluster
