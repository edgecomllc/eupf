### PFCP Procedures

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

### PFCP Messages

| Message      | Status | Description |
|:-------------|:------------:|:---------------|
| Heartbeat Request           | `Y` | A message sent by the CP or UP function to check if the peer node is alive. It is sent for each peer with which a PFCP control association is established. |
| Heartbeat Response             | `Y` | A message sent as a response to a Heartbeat Request, indicating that the peer node is alive.                                |
| PFD Management Request         | `N` | Support status |
| PFD Management Response        | `N` | Support status |
| Association Setup Request      | `Y` | A message used to establish a PFCP control association between the CP and UP functions.|
| Association Setup Response     | `Y` | A message used to establish a PFCP control association between the CP and UP functions.|
| Association Update Request     | `N` | A message used to establish a PFCP control association between the CP and UP functions.|
| Association Update Response    | `N` | A message used to establish a PFCP control association between the CP and UP functions.|
| Association Release Request    | `N` | A message used to establish a PFCP control association between the CP and UP functions.|
| Association Release Response   | `N` | A message used to establish a PFCP control association between the CP and UP functions.|
| Version Not Supported Response | `?` | A message used to establish a PFCP control association between the CP and UP functions.|
| Node Report Request            | `N` |  |
| Node Report Response           | `N` |  |
| Session Set Deletion Request   | `N` |  |
| Session Set Deletion Response  | `N` |   |
| Session Establishment Request  | `Y` | A message used to initiate the establishment of a PFCP session for packet forwarding control.|
| Session Establishment Response | `Y` | A message used to initiate the establishment of a PFCP session for packet forwarding control.|
| Session Modification Request   | `Y` | A message used to initiate the modification of a PFCP session.|
| Session Modification Response  | `Y` | A message used to initiate the modification of a PFCP session.|
| Session Deletion Request       | `Y` | A message used to initiate the deletion of a PFCP session.|
| Session Deletion Response      | `Y` | A message used to initiate the deletion of a PFCP session.|
| Session Report Request         | `N` |  |
| Session Report Response        | `N` |  |

### 3GPP features support

| **Feature** | **Status** | **Description**|
|-------------|:----------:|-----------------------------------------------------------------------------------------------------------------------|
| `BUCP`      | `N`        | Downlink Data Buffering in CP function is supported by the UP function.                                               |
| `DDND`      | `N`        | The buffering parameter 'Downlink Data Notification Delay' is supported by the UP function.                           |
| `DLBD`      | `N`        | The buffering parameter 'DL Buffering Duration' is supported by the UP function.                                      |
| `TRST`      | `N`        | Traffic Steering is supported by the UP function.                                                                     |
| `FTUP`      | `N`        | F-TEID allocation / release in the UP function is supported by the UP function.                                       |
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
| `UEIP`      | `N`        | The UPF supports allocating UE IP addresses or prefixes.                                                              |
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
