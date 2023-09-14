# Deployment examples
The eUPF can be integrated with different 5G Core implementations in different scenarios.

eUPF pod outbound connection is pure routed at the node. There is no address translation inside pod, so we avoid such lack of throughtput.

>If you need Network Address Translation (NAT) function for subscriber's egress traffic, see the chapter [about NAT](#option-nat-at-the-node) below.

BGP is used to announce the subscriber subnet to the route table of a node.

## [Open5GS + Calico BGP](./open5gs-with-bgp/README.md)

## [Open5GS + Calico BGP with Slices](./open5gs-with-bgp-and-slices/README.md)

## [Free5GC + Calico BGP](./free5gc-with-bgp/README.md)

## [Free5GC UpLink CLassifier (ULCL) architecture](./free5gc-ulcl/README.md)

## Option NAT at the node

If you need NAT (Network Address Translation, or Masqerading) at your node to access Internet, the easiest way is to use standart daemonset [IP Masquerade Agent](https://kubernetes.io/docs/tasks/administer-cluster/ip-masq-agent/):
```powershell
sudo kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/ip-masq-agent/master/ip-masq-agent.yaml
```
   > The below entries show the default set of rules that are applied by the ip-masq-agent:
    ` iptables -t nat -L IP-MASQ-AGENT`
