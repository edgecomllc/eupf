*** Settings ***
Library    ScapyLibrary
Library    RequestsLibrary
Library    Collections
Library    Process
Library    runKeywordAsync

*** Variables ***
${EUPF_API_ADDRESS}    localhost:8080 
${TCPREPLAY_LIMIT}    ${7000000}
${TCPREPLAY_THREADS}  ${6}
${payload}   ${{'a'*1024}} 

*** Test Cases ***
Perform load test
    ${EUPF_PROCESSED_ON_START}    Get EUPF packet processed    ${EUPF_API_ADDRESS}

    Set eUPF Uplink PDR    teid=0    far_id=${0}    qer_id=${0}
    Set eUPF FAR    far_id=${0}
    Set eUPF QER    qer_id=${0}

    ${PACKET}  Create GTP-U Packet
    ${TCPREPLAY_EXTRA_ARGS}   Create List    --limit=${TCPREPLAY_LIMIT}
    FOR    ${i}    IN RANGE    ${TCPREPLAY_THREADS}
           Exit For Loop If    ${i} == ${TCPREPLAY_THREADS}
           #${handle}=    Run Keyword Async    Sendpfast  packet     pps      mbps     realtime  loop  file_cache  iface  replay_args              parse_results
           ${handle}=     Run Keyword Async    Sendpfast  ${PACKET}  ${None}  ${None}  ${False}  ${0}  ${True}     lo     ${TCPREPLAY_EXTRA_ARGS}  ${True}
    END
    # Give it 15 minutes (900 sec) to finish each Sendpfast process. If not - robot test fails.
    ${return_value}=     Wait Async All     timeout=900

    ${RESULT}=    Create Dictionary    pps=${0}    mbps=${0}    packets=${0}
    FOR    ${i}    IN RANGE    ${TCPREPLAY_THREADS}
           Exit For Loop If    ${i} == ${TCPREPLAY_THREADS}
           ${pps_sum}=    Evaluate    ${RESULT}[pps] + ${return_value}[${i}][pps]
           ${mbps_sum}=    Evaluate    ${RESULT}[mbps] + ${return_value}[${i}][mbps]
           ${packets_sum}=    Evaluate    ${RESULT}[packets] + ${return_value}[${i}][packets]
           Set To Dictionary    ${RESULT}    pps    ${pps_sum}    mbps    ${mbps_sum}    packets    ${packets_sum}
    END
    Log To Console  Resulting pps: ${RESULT}[pps]
    Log To Console  Resulting mbps: ${RESULT}[mbps]
    Log To Console  Resulting packets: ${RESULT}[packets]
    ${TCPREPLAY_SENT}    Set Variable    ${RESULT}[packets]

    ${EUPF_PROCESSED_ON_STOP}    Get EUPF packet processed    ${EUPF_API_ADDRESS}

    Should Be True    ${EUPF_PROCESSED_ON_STOP} > ${EUPF_PROCESSED_ON_START}
    ${EUPF_PROCESSED}    Set Variable    ${${EUPF_PROCESSED_ON_STOP}-${EUPF_PROCESSED_ON_START}}

    Should Be Equal As Integers    ${TCPREPLAY_SENT}    ${EUPF_PROCESSED}    msg=Sent packets count doesn't equals to processed packets 


*** Keywords ***
Get EUPF packet processed
    [Arguments]    ${API_ADDRESS}    
    ${response}    GET  http://${API_ADDRESS}/api/v1/xdp_stats  expected_status=200
    [return]    ${${response.json()}[tx] + ${response.json()}[redirect]}


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

Set eUPF Uplink PDR
    [Arguments]   ${teid}=${0}    ${far_id}=${0}    ${qer_id}=${0}
    ${body}    Create Dictionary    outer_header_removal=${0}    far_id=${far_id}    qer_id=${qer_id}
    ${response}    PUT    url=http://${EUPF_API_ADDRESS}/api/v1/uplink_pdr_map/${teid}    json=${body}
    Log    ${response.json()}

Set eUPF FAR
    [Arguments]   ${far_id}=${0}
    ${body}    Create Dictionary    action=${2}    outer_header_creation=${0}    teid=${0}    remote_ip=${0}    local_ip=${0}    transport_level_marking=${0}
    ${response}    PUT    url=http://${EUPF_API_ADDRESS}/api/v1/far_map/${far_id}    json=${body}
    Log    ${response.json()}

Set eUPF QER
    [Arguments]   ${qer_id}=0
    ${body}    Create Dictionary    gate_status_ul=${0}    gate_status_dl=${0}    qfi=${0}    max_bitrate_ul=${0}    max_bitrate_dl=${0}
    ${response}    PUT    url=http://${EUPF_API_ADDRESS}/api/v1/qer_map/${qer_id}    json=${body}
    Log    ${response.json()}