## eUPF 3GPP compatibility

eUPF implements 5G UPF functions according to 3GPP TS 129 244 version 16.4.0 Release 16.

### N4 interface support

#### PFCP procedures

| Procedure            | Status | 3GPP reference                     |
|:---------------------|:---:|:--------------------------------------|
|Heartbeat             | `Y` | TS 129 244: 6.2.2 Heartbeat Procedure |
|Load Control          | `N` | TS 129 244: 6.2.3 Heartbeat Procedure |
|Overload Control      | `N` | TS 129 244: 6.2.4 Overload Control Procedure |
|PFD Management        | `N` | TS 129 244: 6.2.5 PFCP PFD Management Procedure |
|Association Setup     | `Y` | TS 129 244: 6.2.6 PFCP Association Setup Procedure |
|Association Update    | `N` | TS 129 244: 6.2.7 PFCP Association Update Procedure |
|Association Release   | `N` | TS 129 244: 6.2.8 PFCP Association Release Procedure |
|Node Report           | `N` | TS 129 244: 6.2.9 PFCP Node Report Procedure |
|Session Establishment | `Y` | TS 129 244: 6.3.2 PFCP Session Establishment Procedure |
|Session Modificationt | `Y` | TS 129 244: 6.3.3 PFCP Session Modification Procedure |
|Session Deletion      | `Y` | TS 129 244: 6.3.4 PFCP Session Deletion Procedure |
|Session Report        | `N` | TS 129 244: 6.3.5 PFCP Session Report Procedure |

#### PFCP messages

| Message      | Status | 3GPP reference |
|:-------------|:------------:|:---------------|
| Heartbeat Request              | `Y` | TS 129 244: 7.4.2 Heartbeat Messages |
| Heartbeat Response             | `Y` | TS 129 244: 7.4.2.2 Heartbeat Response |
| PFD Management Request         | `N` | TS 129 244: 7.4.3.1 PFCP PFD Management Request |
| PFD Management Response        | `N` | TS 129 244: 7.4.3.2 PFCP PFD Management Response |
| Association Setup Request      | `Y` | TS 129 244: 7.4.4.1 PFCP Association Setup Request |
| Association Setup Response     | `Y` | TS 129 244: 7.4.4.2 PFCP Association Setup Response|
| Association Update Request     | `N` | TS 129 244: 7.4.4.3 PFCP Association Update Request|
| Association Update Response    | `N` | TS 129 244: 7.4.4.4 PFCP Association Update Response|
| Association Release Request    | `N` | TS 129 244: 7.4.4.5 PFCP Association Release Request|
| Association Release Response   | `N` | TS 129 244: 7.4.4.6 PFCP Association Release Response|
| Version Not Supported Response | `N` | TS 129 244: 7.4.4.7 PFCP Version Not Supported Response|
| Node Report Request            | `N` | TS 129 244: 7.4.5.1 PFCP Node Report Request |
| Node Report Response           | `N` | TS 129 244: 7.4.5.2 PFCP Node Report Response |
| Session Set Deletion Request   | `N` | TS 129 244: 7.4.6.1 PFCP Session Set Deletion Request |
| Session Set Deletion Response  | `N` | TS 129 244: 7.4.6.2 PFCP Session Set Deletion Response  |
| Session Establishment Request  | `Y` | TS 129 244: 7.5.2 PFCP Session Establishment Request|
| Session Establishment Response | `Y` | TS 129 244: 7.5.3 PFCP Session Establishment Response|
| Session Modification Request   | `Y` | TS 129 244: 7.5.4 PFCP Session Modification Request|
| Session Modification Response  | `Y` | TS 129 244: 7.5.5 PFCP Session Modification Response|
| Session Deletion Request       | `Y` | TS 129 244: 7.5.6 PFCP Session Deletion Request|
| Session Deletion Response      | `Y` | TS 129 244: 7.5.7 PFCP Session Deletion Response|
| Session Report Request         | `N` | TS 129 244: 7.5.8 PFCP Session Report Request |
| Session Report Response        | `N` | TS 129 244: 7.5.9 PFCP Session Report Response |

### N3 interface support

eUPF implements N3 interface according to 3GPP TS 29.281 version 16.1.0 Release 16.

#### GTP messages

| Message      | Status | 3GPP reference |
|:-------------|:------------:|:---------------|
| Echo Request                             | `Y` | TS 29.281: 7.2.1 Echo Request |
| Echo Response                            | `Y` | TS 29.281: 7.2.2 Echo Response |
| Supported Extension Headers Notification | `N` | TS 29.281: 7.2.3 Supported Extension Headers Notification |
| Error Indication                         | `N` | TS 29.281: 7.3.1 Error Indication |
| End Marker                               | `N` | TS 29.281: 7.3.2 End Marker |
| G-PDU                                    | `Y` | TS 29.281: 6.1 General |

### 3GPP features support

| **Feature** | **Status** | **Description**|
|-------------|:----------:|-----------------------------------------------------------------------------------------------------------------------|
| `BUCP`      | `N`        | Downlink Data Buffering in CP function is supported by the UP function.                                               |
| `DDND`      | `N`        | The buffering parameter 'Downlink Data Notification Delay' is supported by the UP function.                           |
| `DLBD`      | `N`        | The buffering parameter 'DL Buffering Duration' is supported by the UP function.                                      |
| `TRST`      | `N`        | Traffic Steering is supported by the UP function.                                                                     |
| `FTUP`      | `Y`        | F-TEID allocation / release in the UP function is supported by the UP function.                                       |
| `PFDM`      | `N`        | The PFD Management procedure is supported by the UP function.                                                         |
| `HEEU`      | `N`        | Header Enrichment of Uplink traffic is supported by the UP function.                                                  |
| `TREU`      | `N`        | Traffic Redirection Enforcement in the UP function is supported by the UP function.                                   |
| `EMPU`      | `N`        | Sending of End Marker packets supported by the UP function.                                                           |
| `PDIU`      | `N`        | Support of PDI optimised signalling in UP function.                                                                   |
| `UDBC`      | `N`        | Support of UL/DL Buffering Control.                                                                                   |
| `QUOAC`     | `N`        | The UP function supports being provisioned with the Quota Action to apply when reaching quotas.                       |
| `TRACE`     | `N`        | The UP function supports Trace.                                                                                       |
| `FRRT`      | `N`        | The UP function supports Framed Routing.                                                                              |
| `PFDE`      | `N`        | The UP function supports a PFD Contents including a property with multiple values.                                    |
| `EPFAR`     | `N`        | The UP function supports the Enhanced PFCP Association Release feature.                                               |
| `DPDRA`     | `N`        | The UP function supports Deferred PDR Activation or Deactivation.                                                     |
| `ADPDP`     | `N`        | The UP function supports the Activation and Deactivation of Pre-defined PDRs.                                         |
| `UEIP`      | `Y`        | The UPF supports allocating UE IP addresses or prefixes.                                                              |
| `SSET`      | `N`        | UPF support of PFCP sessions successively controlled by different SMFs of a same SMF Set.                             |
| `MNOP`      | `N`        | Measurement of number of packets which is instructed with the flag 'Measurement of Number of Packets' in a URR.       |
| `MTE`       | `N`        | UPF supports multiple instances of Traffic Endpoint IDs in a PDI.                                                     |
| `BUNDL`     | `N`        | PFCP messages bunding is supported by the UP function.                                                                |
| `GCOM`      | `N`        | UPF support of 5G VN Group Communication.                                                                             |
| `MPAS`      | `N`        | UPF support for multiple PFCP associations to the SMFs in an SMF set.                                                 |
| `RTTL`      | `N`        | The UP function supports redundant transmission at transport layer.                                                   |
| `VTIME`     | `N`        | UPF support of quota validity time feature.                                                                           |
| `NORP`      | `N`        | UP function support of Number of Reports.                                                                             |
| `IPTV`      | `N`        | UPF support of IPTV service                                                                                           |
| `IP6PL`     | `N`        | UE IPv6 address(es) allocation with IPv6 prefix length other than default /64 (incl. /128 individual IPv6 addresses). |
| `TSCU`      | `N`        | Time Sensitive Communication is supported by the UPF.                                                                 |
| `MPTCP`     | `N`        | UPF support of MPTCP Proxy functionality.                                                                             |
| `ATSSS-LL`  | `N`        | UPF support of ATSSS-LLL steering functionality.                                                                      |
| `QFQM`      | `N`        | UPF support of per QoS flow per UE QoS monitoring.                                                                    |
| `GPQM`      | `N`        | UPF support of per GTP-U Path QoS monitoring.                                                                         |
| `MT-EDT`    | `N`        | SGW-U support of reporting the size of DL Data Packets.                                                               |
| `CIOT`      | `N`        | UPF support of CIoT feature, e.g. small data packet rate enforcement.                                                 |
| `ETHAR`     | `N`        | UPF support of Ethernet PDU Session Anchor Relocation.                                                                |
| `DDDS`      | `N`        | Reporting the first buffered/discarded downlink data after buffering / directly dropped downlink data.                |
| `RDS`       | `N`        | UP function support of Reliable Data Service                                                                          |
| `RTTWP`     | `N`        | UPF support of RTT measurements towards the UE Without PMF.                                                           |
| `QUASF`     | `N`        | URR with an Exempted Application ID for Quota Action or an Exempted SDF Filter for Quota Action.                      |
| `NSPOC`     | `N`        | UP function supports notifying start of Pause of Charging via user plane.                                             |
| `L2TP`      | `N`        | UP function supports the L2TP feature                                                                                 |
| `UPBER`     | `N`        | UP function supports the uplink packets buffering during EAS relocation.                                              |
| `RESPS`     | `N`        | Restoration of PFCP Sessions associated with one or more PGW-C/SMF FQCSID(s), Group Id(s) or CP IP address(es)        |
| `IPREP`     | `N`        | UP function supports IP Address and Port number replacement                                                           |
| `DNSTS`     | `N`        | UP function support DNS Traffic Steering based on FQDN in the DNS Query message                                       |
| `DRQOS`     | `N`        | UP function supports Direct Reporting of QoS monitoring events to Local NEF or AF                                     |
| `MBSN4`     | `N`        | UPF supports sending MBS multicast session data to associated PDU sessions using 5GC individual delivery              |
| `PSUPRM`    | `N`        | UP function supports Per Slice UP Resource Management                                                                 |
| `EPPPI`     | `N`        | UP function supports Enhanced Provisioning of Paging Policy Indicator feature                                         |
| `RATP`      | `N`        | Redirection Address Types with "Port", "IPv4 addr" or "IPv6 addr".                                                    |
| `UPIDP`     | `N`        | UP function supports User Plane Inactivity Detection and reporting per PDR feature                                    |
