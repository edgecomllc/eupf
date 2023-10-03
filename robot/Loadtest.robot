*** Settings ***
Library    ScapyLibrary
Library    RequestsLibrary
Library    Collections

*** Variables ***
@{EXTRA_ARGS}   --limit=70000
${EUPF_API_ADDRESS}    localhost:8456 

*** Test Cases ***
Send And Receive, Return Reply
    ${response}=    GET  http://${EUPF_API_ADDRESS}/api/v1/xdp_stats  expected_status=200
    ${TOTAL_START}    Evaluate    ${response.json()}[tx] + ${response.json()}[redirect]
    Log To Console  Total on start: ${TOTAL_START}

    ${PACKET}  Create GTP-U Packet
    ${RESULT}    Sendpfast  ${PACKET}  iface=lo  file_cache=true  loop=${0}  replay_args=@{EXTRA_ARGS}  parse_results=true
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
    ${ip}         IP      dst=192.168.1.1
    ${udp}        UDP    sport=${2152}     dport=${2152}
    ${gtp}        GTP U Header    
    ${ip-int}     IP      dst=10.60.0.1
    ${icmp}       ICMP
    ${PACKET}    Compose Packet    ${eth}    ${ip}    ${udp}    ${gtp}    ${ip-int}    ${icmp}
    Log Packets    ${PACKET}
    [return]  ${PACKET}