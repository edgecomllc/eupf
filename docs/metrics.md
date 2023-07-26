# eUPF metrics

### PFCP message metrics
This set of metrics describes how many requests of each type has been processed with outcome specified.
All metrics except for `upf_pfcp_rx_latency` are counters and labeled with `result` indicating if message was successfuly processed or rejected.
**Note:** `upf_pfcp_rx` and `upf_pfcp_rx_errors` have different implementation and counted at different points, we will drop one or another after evaluation, or implement a different counters altogether.
| Metric Name         | Description                                                |
| ------------------- | ---------------------------------------------------------- |
| upf_pfcp_rx         | The total number of received PFCP messages                 |
| upf_pfcp_tx         | The total number of transmitted PFCP messages              |
| upf_pfcp_rx_errors  | The total number of received PFCP messages with cause code |
| upf_pfcp_rx_latency | The total number of PFCP messages processing duration      |

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

| Metric Name        | Description                                |
|--------------------|--------------------------------------------|
| upf_rx_arp         | The total number of received ARP packets   |
| upf_rx_icmp        | The total number of received ICMP packets  |
| upf_rx_icmpv6      | The total number of received ICMPv6 packets|
| upf_rx_ip4         | The total number of received IPv4 packets  |
| upf_rx_ip6         | The total number of received IPv6 packets  |
| upf_rx_tcp         | The total number of received TCP packets   |
| upf_rx_udp         | The total number of received UDP packets   |
| upf_rx_other       | The total number of received other packets |
| upf_rx_gtp_echo    | The total number of received GTP echo packets |
| upf_rx_gtp_pdu     | The total number of received GTP PDU packets |
| upf_rx_gtp_other   | The total number of received GTP other packets |
| upf_rx_gtp_error   | The total number of received GTP error packets |

### PFCP Session metrics

| Metric Name               | Description                                  |
|---------------------------|----------------------------------------------|
| upf_pfcp_sessions         | Number of currently established sessions     |
| upf_pfcp_associations     | Number of currently established associations |