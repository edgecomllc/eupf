| 3GPP Feature | Description                                                                                                                  | eUPF | 5GOpenUPF | Cisco   | Juniper |
| ------------ | ---------------------------------------------------------------------------------------------------------------------------- | ---- | --------- | ------- | ------- |
| BUCP         | Downlink Data Buffering in CP function is supported by the UP function.                                                      | N    | Y         | Y       | Y       |
| DDND         | The buffering parameter 'Downlink Data Notification Delay' is supported by the UP function.                                  | N    | Y         | Y       | ?       |
| DLBD         | The buffering parameter 'DL Buffering Duration' is supported by the UP function.                                             | N    | Y         | Y       | N       |
| TRST         | Traffic Steering is supported by the UP function.                                                                            | N    | Y         | Y       | Y       |
| FTUP         | F-TEID allocation / release in the UP function is supported by the UP function.                                              | N    | Y         | Y       | ?       |
| PFDM         | The PFD Management procedure is supported by the UP function.                                                                | N    | Y         | Y       | ?       |
| HEEU         | Header Enrichment of Uplink traffic is supported by the UP function.                                                         | N    | Y         | ?       | ?       |
| TREU         | Traffic Redirection Enforcement in the UP function is supported by the UP function.                                          | N    | Y         | Y       | Y       |
| EMPU         | Sending of End Marker packets supported by the UP function.                                                                  | N    | Y         | Y       | Y       |
| PDIU         | Support of PDI optimised signalling in UP function.                                                                          | N    | Y         | Y       | Y       |
| UDBC         | Support of UL/DL Buffering Control.                                                                                          | N    | Y         | Y       | ?       |
| QUOAC        | The UP function supports being provisioned with the Quota Action to apply when reaching quotas.                              | N    | Y         | Y       | ?       |
| TRACE        | The UP function supports Trace.                                                                                              | N    | N         | Y       | Y       |
| FRRT         | The UP function supports Framed Routing.                                                                                     | N    | Y         | Y       | Y       |
| PFDE         | The UP function supports a PFD Contents including a property with multiple values.                                           | N    | Y         | ?       | ?       |
| EPFAR        | The UP function supports the Enhanced PFCP Association Release feature.                                                      | N    | Y         | ?       | N       |
| DPDRA        | The UP function supports Deferred PDR Activation or Deactivation.                                                            | N    | Y         | ?       | N       |
| ADPDP        | The UP function supports the Activation and Deactivation of Pre-defined PDRs.                                                | N    | Y         | Y       | ?       |
| UEIP         | The UPF supports allocating UE IP addresses or prefixes.                                                                     | N    | Y         | via VRF |         |
| SSET         | UPF support of PFCP sessions successively controlled by different SMFs of a same SMF Set.                                    | N    | Y         | ?       |         |
| MNOP         | UPF supports measurement of number of packets which is instructed with the flag 'Measurement of Number of Packets' in a URR. | N    | N         | N       | N       |
