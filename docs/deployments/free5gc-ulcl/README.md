# README

[[_TOC_]]

## Quick start

### prepare kubernetes nodes - install gtp5g kernel module

compile and install gtp5g kernel module:

```
apt-get update; apt-get install git build-essential -y; \
cd /tmp; \
git clone --depth 1 --branch v0.7.3 https://github.com/free5gc/gtp5g.git; \
cd gtp5g/; \
make && make install
```

check that the module is loaded:

`lsmod | grep ^gtp5g`

### make commands

* [install helm](https://helm.sh/docs/intro/install/)

* add towards5gs helm repo

```
helm repo add towards5gs 'https://raw.githubusercontent.com/Orange-OpenSource/towards5gs-helm/main/repo/'
helm repo update
```

* `make free5gc` for install free5gc
* `make upf` for install free5gc-upf
* `make ueransim` for install ueransim
* `make clean` for delete all components from cluster

default helm values for apps stored in `.deploy/helm/values/dev` directory

after installation, you can run shell into uerasim ue pod:

* `make ueransim_shell`

---
## UpLink CLassifier (ULCL) architecture
Here is configuration in `.deploy/helm/values/dev/ulcl` with three UPFs. 
Traffic routes is:
- upfb--upf1--Internet as default
- upfb--upf2--Internet--1.1.1.1/32 for imsi-208930000000003 specificPath

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