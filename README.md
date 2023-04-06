# eupf

## UPF features

<details><summary>3GPP features support</summary>

| Status | Feature    | Description                                                                                                           |
| :----: | :--------- | :-------------------------------------------------------------------------------------------------------------------- |
|   N    | `BUCP`     | Downlink Data Buffering in CP function is supported by the UP function.                                               |
|   N    | `DDND`     | The buffering parameter 'Downlink Data Notification Delay' is supported by the UP function.                           |
|   N    | `DLBD`     | The buffering parameter 'DL Buffering Duration' is supported by the UP function.                                      |
|   N    | `TRST`     | Traffic Steering is supported by the UP function.                                                                     |
|   N    | `FTUP`     | F-TEID allocation / release in the UP function is supported by the UP function.                                       |
|   N    | `PFDM`     | The PFD Management procedure is supported by the UP function.                                                         |
|   N    | `HEEU`     | Header Enrichment of Uplink traffic is supported by the UP function.                                                  |
|   N    | `TREU`     | Traffic Redirection Enforcement in the UP function is supported by the UP function.                                   |
|   N    | `EMPU`     | Sending of End Marker packets supported by the UP function.                                                           |
|   N    | `PDIU`     | Support of PDI optimised signalling in UP function.                                                                   |
|   N    | `UDBC`     | Support of UL/DL Buffering Control.                                                                                   |
|   N    | `QUOAC`    | The UP function supports being provisioned with the Quota Action to apply when reaching quotas.                       |
|   N    | `TRACE`    | The UP function supports Trace.                                                                                       |
|   N    | `FRRT`     | The UP function supports Framed Routing.                                                                              |
|   N    | `PFDE`     | The UP function supports a PFD Contents including a property with multiple values.                                    |
|   N    | `EPFAR`    | The UP function supports the Enhanced PFCP Association Release feature.                                               |
|   N    | `DPDRA`    | The UP function supports Deferred PDR Activation or Deactivation.                                                     |
|   N    | `ADPDP`    | The UP function supports the Activation and Deactivation of Pre-defined PDRs.                                         |
|   N    | `UEIP`     | The UPF supports allocating UE IP addresses or prefixes.                                                              |
|   N    | `SSET`     | UPF support of PFCP sessions successively controlled by different SMFs of a same SMF Set.                             |
|   N    | `MNOP`     | Measurement of number of packets which is instructed with the flag 'Measurement of Number of Packets' in a URR.       |
|   N    | `MTE`      | UPF supports multiple instances of Traffic Endpoint IDs in a PDI.                                                     |
|   N    | `BUNDL`    | PFCP messages bunding is supported by the UP function.                                                                |
|   N    | `GCOM`     | UPF support of 5G VN Group Communication.                                                                             |
|   N    | `MPAS`     | UPF support for multiple PFCP associations to the SMFs in an SMF set.                                                 |
|   N    | `RTTL`     | The UP function supports redundant transmission at transport layer.                                                   |
|   N    | `VTIME`    | UPF support of quota validity time feature.                                                                           |
|   N    | `NORP`     | UP function support of Number of Reports.                                                                             |
|   N    | `IPTV`     | UPF support of IPTV service                                                                                           |
|   N    | `IP6PL`    | UE IPv6 address(es) allocation with IPv6 prefix length other than default /64 (incl. /128 individual IPv6 addresses). |
|   N    | `TSCU`     | Time Sensitive Communication is supported by the UPF.                                                                 |
|   N    | `MPTCP`    | UPF support of MPTCP Proxy functionality.                                                                             |
|   N    | `ATSSS-LL` | UPF support of ATSSS-LLL steering functionality.                                                                      |
|   N    | `QFQM`     | UPF support of per QoS flow per UE QoS monitoring.                                                                    |
|   N    | `GPQM`     | UPF support of per GTP-U Path QoS monitoring.                                                                         |
|   N    | `MT-EDT`   | SGW-U support of reporting the size of DL Data Packets.                                                               |
|   N    | `CIOT`     | UPF support of CIoT feature, e.g. small data packet rate enforcement.                                                 |
|   N    | `ETHAR`    | UPF support of Ethernet PDU Session Anchor Relocation.                                                                |
|   N    | `DDDS`     | Reporting the first buffered/discarded downlink data after buffering / directly dropped downlink data.                |
|   N    | `RDS`      | UP function support of Reliable Data Service                                                                          |
|   N    | `RTTWP`    | UPF support of RTT measurements towards the UE Without PMF.                                                           |
|   N    | `QUASF`    | URR with an Exempted Application ID for Quota Action or an Exempted SDF Filter for Quota Action.                      |
|   N    | `NSPOC`    | UP function supports notifying start of Pause of Charging via user plane.                                             |
|   N    | `L2TP`     | UP function supports the L2TP feature                                                                                 |
|   N    | `UPBER`    | UP function supports the uplink packets buffering during EAS relocation.                                              |
|   N    | `RESPS`    | Restoration of PFCP Sessions associated with one or more PGW-C/SMF FQCSID(s), Group Id(s) or CP IP address(es)        |
|   N    | `IPREP`    | UP function supports IP Address and Port number replacement                                                           |
|   N    | `DNSTS`    | UP function support DNS Traffic Steering based on FQDN in the DNS Query message                                       |
|   N    | `DRQOS`    | UP function supports Direct Reporting of QoS monitoring events to Local NEF or AF                                     |
|   N    | `MBSN4`    | UPF supports sending MBS multicast session data to associated PDU sessions using 5GC individual delivery              |
|   N    | `PSUPRM`   | UP function supports Per Slice UP Resource Management                                                                 |
|   N    | `EPPPI`    | UP function supports Enhanced Provisioning of Paging Policy Indicator feature                                         |
|   N    | `RATP`     | Redirection Address Types with "Port", "IPv4 addr" or "IPv6 addr".                                                    |
|   N    | `UPIDP`    | UP function supports User Plane Inactivity Detection and reporting per PDR feature                                    |
</details>

## Architecture

### Eagle-eye overview
![UPF-Arch2](https://user-images.githubusercontent.com/20152142/207142700-cc3f17a5-203f-4b43-b712-a518cb627968.png)

### Detailed architecture
![image](https://user-images.githubusercontent.com/20152142/228003420-0a2be83e-095e-4ad4-8635-0eb434951a3e.png)


### Current limitation

- Only one PDR in PFCP session per direction
- Only single FAR supported 

### Packet forwarding pipeline

![UPF-Forwarding](https://user-images.githubusercontent.com/20152142/207142725-0af400bb-8ff8-4f36-93bd-3c461c0e7ce4.png)

## Roadmap

### Management Layer

- [ ]  PFCP Association Setup/Release and Heartbeats
- [ ]  Session Establishment/Modification with support for PFCP entities such as Packet Detection Rules (PDRs), Forwarding Action Rules (FARs), QoS Enforcement Rules (QERs).
- [ ]  UPF-initiated PFCP association
- [ ]  UPF-based UE IP address assignment
- [ ]  Integration with Prometheus for exporting PFCP and data plane-level metrics.

### Datapath Layer

- [ ]  IPv4 support
- [ ]  N3, N4, N6, N9 interfacing
- [ ]  Single & Multi-port support
- [ ]  Monitoring/Debugging capabilties using
    - tcpdump on individual modules
    - visualization web interface
    - command line shell interface for displaying statistics
- [ ]  Static IP routing
- [ ]  I-UPF/A-UPF ULCL/Branching i.e., simultaneous N6/N9 support within PFCP session
- [ ]  Basic QoS support, with per-slice and per-session rate limiting

## Backlog

### Management Layer

- [ ]  Application filtering using SDF filters
- [ ]  Generation of End Marker Packets
- [ ]  Downlink Data Notification (DDN) using PFCP Session Report
- [ ]  Application filtering using application PFDs

### Datapath Layer

- [ ]  IPv6 support
- [ ]  Dynamic IP routing
- [ ]  Support for IPv4 datagrams reassembly
- [ ]  Support for IPv4 packets fragmentation
- [ ]  Support for UE IP NAT
- [ ]  Service Data Flow (SDF) configuration via N4/PFCP.
- [ ]  Downlink Data Notification (DDN) - notification only (no buffering)
- [ ]  Per-flow latency and throughput metrics
- [ ]  Network Token Functions

## Metrics

### PFCP message metrics
This set of metrics describes how many requests of each type has been processed with outcome specified.
All metrics except for `upf_message_processing_duration` are counters and labeled with `result` indicating if message was successfuly processed or rejected.
**Note:** `upf_pfcp_msg_rx` and `upf_pfcp_msg_rx_with_cause_code` have different implementation and counted at different points, we will drop one or another after evaluation, or implement a different counters altogether.
| Metric Name                     | Description                                                |
| ------------------------------- | ---------------------------------------------------------- |
| upf_pfcp_msg_rx                 | The total number of received PFCP messages                 |
| upf_pfcp_msg_tx                 | The total number of transmitted PFCP messages              |
| upf_pfcp_msg_rx_with_cause_code | The total number of received PFCP messages with cause code |
| upf_message_processing_duration | The total number of PFCP messages processing duration      |

### XDP Action metrics
This set of metrics are used to count the number of packets with different outcomes, such as the total number of aborted, dropped, passed, transmitted, and redirected packets.

| Metric Name      | Description                             |
| ---------------- | --------------------------------------- |
| upf_xdp_aborted  | The total number of aborted packets     |
| upf_xdp_drop     | The total number of dropped packets     |
| upf_xdp_pass     | The total number of passed packets      |
| upf_xdp_tx       | The total number of transmitted packets |
| upf_xdp_redirect | The total number of redirected packets  |

### Packet metrics
Various packet counters with `packet_type` label.

| Metric Name | Description                              |
| ----------- | ---------------------------------------- |
| upf_rx      | The total number of received ARP packets |
