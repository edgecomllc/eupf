*** Settings ***
Library    ScapyLibrary
Library    RequestsLibrary
Library    Collections
Library    Process

*** Variables ***
${EUPF_API_ADDRESS}    localhost:8080 
@{TCPREPLAY_EXTRA_ARGS}   --limit=700000
${payload}   ${{'a'*1024}} 

*** Test Cases ***
Perform load test
    ${response}=    GET  http://${EUPF_API_ADDRESS}/api/v1/xdp_stats  expected_status=200
    ${TOTAL_START}    Evaluate    ${response.json()}[tx] + ${response.json()}[redirect]
    Log To Console  Total on start: ${TOTAL_START}

    Set Uplink PDR    teid=0    far_id=${0}    qer_id=${0}
    Set FAR    far_id=${0}
    Set QER    qer_id=${0}

    ${PACKET}  Create GTP-U Packet
    ${RESULT}    Sendpfast  ${PACKET}  iface=lo  file_cache=true  loop=${0}  replay_args=@{TCPREPLAY_EXTRA_ARGS}  parse_results=true
    Log To Console  Resulting pps: ${RESULT}[pps]
    Log To Console  Resulting mbps: ${RESULT}[mbps]
    Log To Console  Resulting packets: ${RESULT}[packets]

    ${response}=    GET  http://${EUPF_API_ADDRESS}/api/v1/xdp_stats  expected_status=200
    ${TOTAL_STOP}    Evaluate    ${response.json()}[tx] + ${response.json()}[redirect]
    Log To Console  Total on stop: ${TOTAL_STOP}
    Log To Console  Tx by eUPF: ${${TOTAL_STOP}-${TOTAL_START}}


*** Keywords ***
Create GTP-U Packet
    ${eth}        Ether   
    ${ip}         IP      src=127.0.0.1  dst=127.0.0.1
    ${udp}        UDP    sport=${2152}     dport=${2152}
    ${gtp}        GTP U Header    
    ${ip-int}     IP      src=10.60.0.1  dst=8.8.8.8
    ${udp}       UDP     
    ${PACKET}    Compose Packet    ${eth}    ${ip}    ${udp}    ${gtp}    ${ip-int}    ${udp}    ${payload}
    Log Packets    ${PACKET}
    [return]  ${PACKET}

Set Uplink PDR
    [Arguments]   ${teid}=${0}    ${far_id}=${0}    ${qer_id}=${0}
    ${body}    Create Dictionary    outer_header_removal=${2}    far_id=${far_id}    qer_id=${qer_id}
    ${response}    PUT    url=http://${EUPF_API_ADDRESS}/api/v1/uplink_pdr_map/${teid}    json=${body}
    Log    ${response.json()}

Set FAR
    [Arguments]   ${far_id}=${0}
    ${body}    Create Dictionary    action=${2}    outer_header_creation=${0}    teid=${0}    remote_ip=${0}    local_ip=${0}    transport_level_marking=${0}
    ${response}    PUT    url=http://${EUPF_API_ADDRESS}/api/v1/far_map/${far_id}    json=${body}
    Log    ${response.json()}

Set QER
    [Arguments]   ${qer_id}=0
    ${body}    Create Dictionary    gate_status_ul=${0}    gate_status_dl=${0}    qfi=${0}    max_bitrate_ul=${0}    max_bitrate_dl=${0}
    ${response}    PUT    url=http://${EUPF_API_ADDRESS}/api/v1/qer_map/${qer_id}    json=${body}
    Log    ${response.json()}