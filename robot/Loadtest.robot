*** Settings ***
Library    ScapyLibrary
Library    RequestsLibrary
Library    Collections
Library    Process

*** Variables ***
${EUPF_API_ADDRESS}    localhost:8080 
@{TCPREPLAY_EXTRA_ARGS}   --limit=70000
${payload}   ${{'a'*1024}} 

*** Test Cases ***
Perform load test
    ${response}=    GET  http://${EUPF_API_ADDRESS}/api/v1/xdp_stats  expected_status=200
    ${TOTAL_START}    Evaluate    ${response.json()}[tx] + ${response.json()}[redirect]
    Log To Console  Total on start: ${TOTAL_START}

    #sudo ./bpftool map update id 53 key hex 00 00 00 00  value hex 00 00 00 00 00 00 00 00 00 00 00 00
    #sudo ./bpftool maupdate id 50 key hex 00 00 00 00  value hex 02 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
    #${RESULT} =    Run Process    ~/projects/bpftool/src/bpftool map update id 53 key hex 00 00 00 00 value hex 00 00 00 00 00 00 00 00 00 00 00 00  shell=True  stderr=STDOUT
    #Log To Console   ${RESULT.stdout} 

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