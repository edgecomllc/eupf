# eUPF

eUPF is the opensource User Plane Function (UPF) project for using inside or "outside" of any 3GPP 5G core. The goal of the project is to provide high-observability and easily-deployed software for a various cases like multi-access edge computing (MEC) and local traffic breakout. eUPF is built with eBPF to provide high observability and performance. 

 eUPF is tested with the Free5GC and Open5GS 5G cores. 

## What is 5G core and CUPS

5G core uses network virtualized functions (NVF) to provide connectivity and services. 
Control and user plane separation (CUPS) is important architecture enhancement that separates control plane and user plane insde 5G core. 
User plane function (UPF) is the "decapsulating and routing" function that extracts user plane traffic from GPRS tunneling protocol (GTP) and route it to the public data network or local network via the best available path. 

![image](https://user-images.githubusercontent.com/119619173/233130952-e5634aff-b177-4274-a2d7-0e51a5488e5d.png)

## Quick start guide

Read [eUPF intallation guide with Open5GS or Free5GC core](./docs/install.md)

Read [eUPF configuration guide](./docs/Config.md)

Read [eUPF metrics and monitoring guide](./docs/metrics.md)

## eUPF details

eUPF as a part of 5G mobile core network implements data network gateway function. It communicates with SMF via PFCP protocol (N4 interface) and forwards packets between core and data networks(N3 and N6 interfaces correspondingly). These two main UPF parts are implemented in two separate components: control plane and forwarding plane.

The eUPF control plane is an userspace application which receives packet processing rules from SMF and configures forwarding plane for proper forwarding. 

The eUPF forwarding plane is based on eBPF packet processing. When started eUPF adds eBPF XDP hook program in order to process network packets as close to NIC as possible. eBPF program consists of several pipeline steps: determine PDR, apply gating, qos and forwardning rules.

eUPF relies on kernel routing when making routing decision for incomming network packets. When it is not possible to deternime packet route via kernel FIB lookup, eUPF passes such packet to kernel as a fallback path. This approach obviously affects performance but allows maintaining correct kernel routing process (ex., filling arp tables).   

## eUPF architecture

<details><summary>Show me</summary>

### Eagle-eye overview

![UPF-Arch2](https://user-images.githubusercontent.com/20152142/207142700-cc3f17a5-203f-4b43-b712-a518cb627968.png)

### Detailed architecture
![image](https://user-images.githubusercontent.com/20152142/228003420-0a2be83e-095e-4ad4-8635-0eb434951a3e.png)

### Current limitation

- Only one PDR in PFCP session per direction
- Only single FAR supported

### Packet forwarding pipeline

![UPF-Forwarding](https://user-images.githubusercontent.com/20152142/207142725-0af400bb-8ff8-4f36-93bd-3c461c0e7ce4.png)
</details>

## eUPF roadmap

<details><summary>Show me</summary>

### Controlplane Layer

- [x]  PFCP Association Setup/Release and Heartbeats
- [x]  Session Establishment/Modification with support for PFCP entities such as Packet Detection Rules (PDRs), Forwarding Action Rules (FARs), QoS Enforcement Rules (QERs).
- [ ]  UPF-initiated PFCP association
- [ ]  UPF-based UE IP address assignment
- [x]  Integration with Prometheus for exporting PFCP and data plane-level metrics.

### Dataplane Layer

- [x]  IPv4 support
- [x]  N3, N4, N6 interfaces 
- [ ]  N9 interface
- [x]  Single & Multi-port support
- [ ]  Monitoring/Debugging capabilties using
    - tcpdump on individual modules
    - command line shell interface for displaying statistics
- [x]  Static IP routing
- [ ]  I-UPF/A-UPF ULCL/Branching i.e., simultaneous N6/N9 support within PFCP session
- [x]  Basic QoS support with per-session rate limiting
 
 </details>

## Contribution

Please create an issue to report a bug or share an idea.

## License
This project is licensed under the [Apache-2.0 Creative Commons License](https://www.apache.org/licenses/LICENSE-2.0) - see the [LICENSE file](./LICENSE) for details
